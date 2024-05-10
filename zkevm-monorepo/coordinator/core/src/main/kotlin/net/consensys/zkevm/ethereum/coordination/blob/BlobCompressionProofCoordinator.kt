package net.consensys.zkevm.ethereum.coordination.blob

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import io.vertx.core.Handler
import io.vertx.core.Vertx
import kotlinx.datetime.Instant
import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.coordinator.clients.BlobCompressionProverClient
import net.consensys.zkevm.domain.Blob
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.BlobStatus
import net.consensys.zkevm.domain.BlockInterval
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ConflationCalculationResult
import net.consensys.zkevm.domain.toBlockIntervalsString
import net.consensys.zkevm.ethereum.coordination.conflation.BlobCreationHandler
import net.consensys.zkevm.persistence.blob.BlobsRepository
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ArrayBlockingQueue
import java.util.concurrent.CompletableFuture
import kotlin.time.Duration

class BlobCompressionProofCoordinator(
  private val vertx: Vertx,
  private val blobsRepository: BlobsRepository,
  private val blobCompressionProverClient: BlobCompressionProverClient,
  private val rollingBlobShnarfCalculator: RollingBlobShnarfCalculator,
  private val blobZkStateProvider: BlobZkStateProvider,
  private val config: Config,
  private val eip4844SwitchProvider: Eip4844SwitchProvider,
  private val blobCompressionProofHandler: BlobCompressionProofHandler,
  metricsFacade: MetricsFacade
) : BlobCreationHandler, LongRunningService {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private val defaultQueueCapacity = 1000 // Should be more than blob submission limit
  private val blobsToHandle = ArrayBlockingQueue<Blob>(defaultQueueCapacity)
  private var timerId: Long? = null
  private lateinit var blobPollingAction: Handler<Long>
  private val blobsCounter = metricsFacade.createCounter(
    LineaMetricsCategory.BLOB,
    "counter",
    "New blobs arriving to blob compression proof coordinator"
  )

  init {
    metricsFacade.createGauge(
      LineaMetricsCategory.BLOB,
      "compression.queue.size",
      "Size of blob compression proving queue",
      { blobsToHandle.size }
    )
  }

  data class Config(
    val blobCalculatorVersion: String,
    val conflationCalculatorVersion: String,
    val pollingInterval: Duration
  )

  /**
   * Handling BlobData events (which have the blob bytes and correspondent batches);
   * Insert the blob metadata in the DB in blobs table, in COMPRESSION_PROVING state.
   * Create the BlobCompressionProofRequest to send to the prover
   * Wait for the prover to generate the proof
   * Update blobs db with status COMPRESSION_PROVEN
   */

  /**
   * Blobs {
   *     start_block_number
   *     end_block_number
   *     conflation_calculator_version
   *     status: COMPRESSION_PROVING, COMPRESSION_PROVEN
   *     blob_hash (same hash that we can query L1 smart contract to identify the blob, check SMC Spec document)
   *     // Meta/Redundant data
   *     start_block_timestamp // useful for blob submission by time limit
   *     end_block_timestamp // useful for blob submission delay
   *     batches_count // useful for proof aggregation
   *     expected_shnarf // the expected new shnarf hash after this blob
   *     blob_compression_proof
   * }
   */

  @Synchronized
  private fun sendBlobToCompressionProver(blob: Blob): SafeFuture<Unit> {
    val eip4844Enabled = eip4844SwitchProvider.isEip4844Enabled(blob)
    log.debug(
      "Going to create the blob compression proof for ${blob.intervalString()} with eip4844Enabled=$eip4844Enabled"
    )
    val blobZkSateAndRollingShnarfFuture = blobZkStateProvider.getBlobZKState(blob.blocksRange)
      .thenCompose { blobZkState ->
        rollingBlobShnarfCalculator.calculateShnarf(
          eip4844Enabled = eip4844Enabled,
          compressedData = blob.compressedData,
          parentStateRootHash = blobZkState.parentStateRootHash,
          finalStateRootHash = blobZkState.finalStateRootHash,
          conflationOrder = BlockIntervals(
            startingBlockNumber = blob.conflations.first().startBlockNumber,
            upperBoundaries = blob.conflations.map { it.endBlockNumber }
          )
        ).thenApply { rollingBlobShnarfResult ->
          Pair(blobZkState, rollingBlobShnarfResult)
        }
      }

    blobZkSateAndRollingShnarfFuture.thenCompose { (blobZkState, rollingBlobShnarfResult) ->
      requestBlobCompressionProof(
        parentStateRootHash = blobZkState.parentStateRootHash,
        finalStateRootHash = blobZkState.finalStateRootHash,
        compressedData = blob.compressedData,
        conflations = blob.conflations,
        parentDataHash = rollingBlobShnarfResult.parentBlobHash,
        prevShnarf = rollingBlobShnarfResult.parentBlobShnarf,
        expectedShnarfResult = rollingBlobShnarfResult.shnarfResult,
        eip4844Enabled = eip4844Enabled,
        commitment = rollingBlobShnarfResult.shnarfResult.commitment,
        kzgProofContract = rollingBlobShnarfResult.shnarfResult.kzgProofContract,
        kzgProofSideCar = rollingBlobShnarfResult.shnarfResult.kzgProofSideCar,
        blobStartBlockTime = blob.startBlockTime,
        blobEndBlockTime = blob.endBlockTime
      ).whenException { exception ->
        log.error(
          "Error in requesting blob compression proof: blob={} errorMessage={} ",
          blob.intervalString(),
          exception.message,
          exception
        )
      }
    }
    // We want to process the next blob without waiting for the compression proof to finish and process the next
    // blob after shnarf calculation of current blob is done
    return blobZkSateAndRollingShnarfFuture.thenApply {}
  }

  private fun requestBlobCompressionProof(
    compressedData: ByteArray,
    conflations: List<ConflationCalculationResult>,
    parentStateRootHash: ByteArray,
    finalStateRootHash: ByteArray,
    parentDataHash: ByteArray,
    prevShnarf: ByteArray,
    expectedShnarfResult: ShnarfResult,
    eip4844Enabled: Boolean,
    commitment: ByteArray,
    kzgProofContract: ByteArray,
    kzgProofSideCar: ByteArray,
    blobStartBlockTime: Instant,
    blobEndBlockTime: Instant
  ): SafeFuture<Unit> {
    return blobCompressionProverClient.requestBlobCompressionProof(
      compressedData = compressedData,
      conflations = conflations,
      parentStateRootHash = parentStateRootHash,
      finalStateRootHash = finalStateRootHash,
      parentDataHash = parentDataHash,
      prevShnarf = prevShnarf,
      expectedShnarfResult = expectedShnarfResult,
      eip4844Enabled = eip4844Enabled,
      commitment = commitment,
      kzgProofContract = kzgProofContract,
      kzgProofSideCar = kzgProofSideCar
    ).thenCompose { result ->
      if (result is Err) {
        SafeFuture.failedFuture(result.error.asException())
      } else {
        val blobCompressionProof = (result as Ok).value
        val blobRecord = BlobRecord(
          startBlockNumber = conflations.first().startBlockNumber,
          endBlockNumber = conflations.last().endBlockNumber,
          conflationCalculatorVersion = config.conflationCalculatorVersion,
          blobHash = expectedShnarfResult.dataHash,
          startBlockTime = blobStartBlockTime,
          endBlockTime = blobEndBlockTime,
          batchesCount = conflations.size.toUInt(),
          status = BlobStatus.COMPRESSION_PROVEN,
          expectedShnarf = expectedShnarfResult.expectedShnarf,
          blobCompressionProof = (result as Ok).value
        )
        SafeFuture.allOf(
          blobsRepository.saveNewBlob(blobRecord),
          blobCompressionProofHandler.acceptNewBlobCompressionProof(
            BlobCompressionProofUpdate(
              blockInterval = BlockInterval.between(
                startBlockNumber = blobRecord.startBlockNumber,
                endBlockNumber = blobRecord.endBlockNumber
              ),
              conflationCalculatorVersion = blobRecord.conflationCalculatorVersion,
              blobCompressionProof = blobCompressionProof
            )
          )
        ).thenApply {}
      }
    }
  }

  @Synchronized
  override fun handleBlob(blob: Blob): SafeFuture<*> {
    blobsCounter.increment()
    log.debug(
      "new blob: blob={} queuedBlobsToProve={} blobBatches={}",
      blob.intervalString(),
      blobsToHandle.size,
      blob.conflations.toBlockIntervalsString()
    )
    blobsToHandle.put(blob)
    log.trace("Blob was added to the handling queue {}", blob)
    return SafeFuture.completedFuture(Unit)
  }

  override fun start(): CompletableFuture<Unit> {
    if (timerId == null) {
      blobPollingAction = Handler<Long> {
        handleBlobsFromTheQueue().whenComplete { _, error ->
          error?.let {
            log.error("Error polling blobs for aggregation: errorMessage={}", error.message, error)
          }
          timerId = vertx.setTimer(config.pollingInterval.inWholeMilliseconds, blobPollingAction)
        }
      }
      timerId = vertx.setTimer(config.pollingInterval.inWholeMilliseconds, blobPollingAction)
    }
    return SafeFuture.completedFuture(Unit)
  }

  private fun handleBlobsFromTheQueue(): SafeFuture<Unit> {
    var blobsHandlingFuture = SafeFuture.completedFuture(Unit)
    if (blobsToHandle.isNotEmpty()) {
      val blobToHandle = blobsToHandle.poll()
      blobsHandlingFuture = blobsHandlingFuture.thenCompose {
        sendBlobToCompressionProver(blobToHandle).whenException { exception ->
          log.error(
            "Error in sending blob to compression prover: blob={} errorMessage={} ",
            blobToHandle.intervalString(),
            exception.message,
            exception
          )
        }
      }
    }
    return blobsHandlingFuture
  }

  override fun stop(): CompletableFuture<Unit> {
    if (timerId != null) {
      vertx.cancelTimer(timerId!!)
      blobPollingAction = Handler<Long> {}
    }
    return SafeFuture.completedFuture(Unit)
  }
}
