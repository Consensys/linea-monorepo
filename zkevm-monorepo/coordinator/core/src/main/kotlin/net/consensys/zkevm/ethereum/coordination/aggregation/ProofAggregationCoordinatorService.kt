package net.consensys.zkevm.ethereum.coordination.aggregation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import io.vertx.core.Vertx
import kotlinx.datetime.Clock
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.PeriodicPollingService
import net.consensys.zkevm.coordinator.clients.L2MessageServiceClient
import net.consensys.zkevm.coordinator.clients.ProofAggregationClient
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.BlobAndBatchCounters
import net.consensys.zkevm.domain.BlobsToAggregate
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ProofsToAggregate
import net.consensys.zkevm.domain.VersionedExecutionProofs
import net.consensys.zkevm.domain.toBlockIntervalsString
import net.consensys.zkevm.ethereum.coordination.blockcreation.SafeBlockProvider
import net.consensys.zkevm.persistence.aggregation.AggregationsRepository
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ConcurrentLinkedQueue
import java.util.function.Consumer
import kotlin.time.Duration

class ProofAggregationCoordinatorService(
  private val vertx: Vertx,
  private val config: Config,
  private var nextBlockNumberToPoll: Long,
  private val aggregationCalculator: AggregationCalculator,
  private val aggregationsRepository: AggregationsRepository,
  private val consecutiveProvenBlobsProvider: ConsecutiveProvenBlobsProvider,
  private val proofAggregationClient: ProofAggregationClient,
  private val aggregationL2StateProvider: AggregationL2StateProvider,
  private val log: Logger = LogManager.getLogger(ProofAggregationCoordinatorService::class.java),
  private val provenAggregationEndBlockNumberConsumer: Consumer<ULong> = Consumer<ULong> { }
) : AggregationHandler, PeriodicPollingService(
  vertx = vertx,
  pollingIntervalMs = config.pollingInterval.inWholeMilliseconds,
  log = log
) {
  data class Config(
    val pollingInterval: Duration,
    val aggregationCalculatorVersion: String,
    val proofsLimit: UInt
  )

  private val pendingBlobs = ConcurrentLinkedQueue<BlobAndBatchCounters>()

  init {
    aggregationCalculator.onAggregation(this)
  }

  /**
   * This component is responsible for periodically monitoring the blobs table to check how many blobs were PROVEN
   *  and aggregate all the corresponding batch proofs into a single final proof.
   *  The aggregation proof can be triggered by 2 conditions:
   *    Number of batches - if number of all batches combined for all subsequent blobs above a given threshold;
   *    Finalization deadline: the time elapsed between first block’s timestamp in the first submitted blob and
   *    current time is greater than a finalization deadline configured.
   *  This is delegated to AggregationCalculator
   * Note: do we need to take into account the 18 days blob eviction on L1?
   *    It should not be a problem in practice, assuming conflation of 20 blocks (over estimation),
   *    500 batches proving limit gives (20*500*12)/3600 ~ 33 hours max.
   *
   * High level steps:
   *   Poll blobs table for blobs, respecting
   *      Blob status is COMPRESSION_PROVEN and Join with batches table, PROVEN status
   *      Blob submission start block number order
   *   Send blob to aggregation calculator. Which will trigger aggregation based on:
   *      Sum(blob.numberOfBatches) < prover_batches_limit
   *      blob.firstBlockTimestamp < current_time - submission_deadline;
   *        (we don't wait to fill prover capacity of the submission_deadline is reached)
   *   When aggregation calculator fires an aggregation (onAggregation), do:
   *    Build aggregation proof request and send it to the prover;
   *    Insert a record in aggregation proof response table PROVING status
   *    Wait for the prover response and update record with PROVEN status;
   *
   *    Aggregations Table {
   *     start_block_number,
   *     end_block_number,
   *     status: PROVING, PROVEN
   *     // Meta/Redundant data
   *     start_block_timestamp // useful for submitter to implement submission by deadline
   *     batch_count
   *    }
   */

  @Synchronized
  override fun action(): SafeFuture<*> {
    log.debug("Polling blobs for aggregation calculator from block={}", nextBlockNumberToPoll)
    return consecutiveProvenBlobsProvider.findConsecutiveProvenBlobs(nextBlockNumberToPoll).thenCompose { blobs ->
      val newBlobs = blobs.map {
        pendingBlobs.offer(it)
        aggregationCalculator.newBlob(it.blobCounters)
      }
      if (blobs.isNotEmpty()) {
        nextBlockNumberToPoll = blobs.last().blobCounters.endBlockNumber.toLong() + 1
        val numberOfBatches = blobs.sumOf { it.blobCounters.numberOfBatches }
        log.info(
          "new blobs sent to aggregation calculator: " +
            "nextBlockToPoll={} numberOfBlobs={} numberOfBatches={} total={} blobs={}",
          nextBlockNumberToPoll,
          blobs.size,
          numberOfBatches,
          blobs.size + numberOfBatches.toInt(),
          blobs.map(BlobAndBatchCounters::blobCounters).toBlockIntervalsString()
        )
      } else {
        log.debug("Found no new blobs for aggregation. nextBlockToPoll={}", nextBlockNumberToPoll)
      }
      SafeFuture.allOf(newBlobs.stream())
    }
  }

  @Synchronized
  override fun onAggregation(blobsToAggregate: BlobsToAggregate): SafeFuture<Unit> {
    log.debug("new aggregation={}", blobsToAggregate.intervalString())
    val predicateToFilterCompressionBlobs: (BlobAndBatchCounters) -> Boolean = {
      it.blobCounters.startBlockNumber >= blobsToAggregate.startBlockNumber &&
        it.blobCounters.endBlockNumber <= blobsToAggregate.endBlockNumber
    }
    val compressionBlobs = mutableListOf<BlobAndBatchCounters>()
    while (pendingBlobs.isNotEmpty() && predicateToFilterCompressionBlobs(pendingBlobs.peek())) {
      compressionBlobs.add(pendingBlobs.poll())
    }
    assert(compressionBlobs.first().blobCounters.startBlockNumber == blobsToAggregate.startBlockNumber)
    assert(compressionBlobs.last().blobCounters.endBlockNumber == blobsToAggregate.endBlockNumber)

    val batchCount = compressionBlobs.sumOf { blobCounters ->
      blobCounters.blobCounters.numberOfBatches
    }
    val compressionProofs = BlockIntervals(
      startingBlockNumber = blobsToAggregate.startBlockNumber,
      upperBoundaries = compressionBlobs.map { it.blobCounters.endBlockNumber }
    )
    val aggregation = Aggregation(
      startBlockNumber = blobsToAggregate.startBlockNumber,
      endBlockNumber = blobsToAggregate.endBlockNumber,
      status = Aggregation.Status.Proving,
      aggregationCalculatorVersion = config.aggregationCalculatorVersion,
      batchCount = batchCount.toULong(),
      aggregationProof = null
    )

    val startingBlockNumber = compressionBlobs.first().versionedExecutionProofs.executionProofs.startingBlockNumber
    val upperBoundaries = compressionBlobs.flatMap {
      it.versionedExecutionProofs.executionProofs.upperBoundaries
    }
    val blockIntervals = BlockIntervals(startingBlockNumber, upperBoundaries)
    val executionProofVersions = compressionBlobs.flatMap {
      it.versionedExecutionProofs.executionVersion
    }
    val versionedExecutionProofs = VersionedExecutionProofs(blockIntervals, executionProofVersions)

    return aggregationL2StateProvider
      .getAggregationL2State(blockNumber = aggregation.startBlockNumber.toLong() - 1)
      .whenException {
        log.error(
          "failed to get parent aggregation l2 message rolling hash: aggregation={} errorMessage={}",
          aggregation.intervalString(),
          it.message,
          it
        )
      }
      .thenApply { rollingInfo ->
        ProofsToAggregate(
          compressionProofs = compressionProofs,
          executionProofs = versionedExecutionProofs.executionProofs,
          executionVersion = versionedExecutionProofs.executionVersion,
          parentAggregationLastBlockTimestamp = rollingInfo.parentAggregationLastBlockTimestamp,
          parentAggregationLastL1RollingHashMessageNumber = rollingInfo.parentAggregationLastL1RollingHashMessageNumber,
          parentAggregationLastL1RollingHash = rollingInfo.parentAggregationLastL1RollingHash
        )
      }
      .thenCompose { proofsToAggregate ->
        proofAggregationClient.getAggregatedProof(proofsToAggregate)
      }
      .whenException {
        log.error(
          "Error getting aggregation proof: aggregation={} errorMessage={}",
          aggregation.intervalString(),
          it.message,
          it
        )
      }
      .thenCompose {
        when (it) {
          is Ok -> {
            val updatedAggregation = Aggregation(
              startBlockNumber = blobsToAggregate.startBlockNumber,
              endBlockNumber = blobsToAggregate.endBlockNumber,
              status = Aggregation.Status.Proven,
              aggregationCalculatorVersion = config.aggregationCalculatorVersion,
              batchCount = batchCount.toULong(),
              aggregationProof = it.value
            )
            aggregationsRepository.saveNewAggregation(aggregation = updatedAggregation)
              .thenPeek {
                provenAggregationEndBlockNumberConsumer.accept(updatedAggregation.endBlockNumber)
              }
              .whenException {
                log.error(
                  "Error saving proven aggregation to DB: aggregation={} errorMessage={}",
                  aggregation.intervalString(),
                  it.message,
                  it
                )
              }
          }

          is Err -> {
            log.error(it.error)
            SafeFuture.failedFuture(it.error.asException())
          }
        }
      }
  }

  companion object {
    fun create(
      vertx: Vertx,
      pollerInterval: Duration,
      aggregationDeadline: Duration,
      latestBlockProvider: SafeBlockProvider,
      maxProofsPerAggregation: UInt,
      startBlockNumberInclusive: ULong,
      aggregationCalculatorVersion: String,
      aggregationsRepository: AggregationsRepository,
      consecutiveProvenBlobsProvider: ConsecutiveProvenBlobsProvider,
      proofAggregationClient: ProofAggregationClient,
      l2web3jClient: Web3j,
      l2MessageServiceClient: L2MessageServiceClient,
      metricsFacade: MetricsFacade,
      provenAggregationEndBlockNumberConsumer: Consumer<ULong>
    ): LongRunningService {
      val aggregationCalculatorByDeadline =
        AggregationTriggerCalculatorByDeadline(
          config = AggregationTriggerCalculatorByDeadline.Config(
            aggregationDeadline = aggregationDeadline
          ),
          clock = Clock.System,
          latestBlockProvider = latestBlockProvider
        )
      val aggregationCalculatorByBatch = AggregationTriggerCalculatorByProofLimit(
        maxProofsPerAggregation = maxProofsPerAggregation
      )
      val globalAggregationCalculator = GlobalAggregationCalculator(
        lastBlockNumber = startBlockNumberInclusive - 1UL,
        syncAggregationTrigger = listOf(aggregationCalculatorByBatch),
        deferredAggregationTrigger = listOf(aggregationCalculatorByDeadline),
        metricsFacade = metricsFacade
      )

      val deadlineCheckRunner = AggregationTriggerCalculatorByDeadlineRunner(
        vertx = vertx,
        config = AggregationTriggerCalculatorByDeadlineRunner.Config(
          deadlineCheckInterval = pollerInterval
        ),
        aggregationTriggerByDeadline = aggregationCalculatorByDeadline
      )

      val proofAggregationService = ProofAggregationCoordinatorService(
        vertx,
        config = Config(
          pollingInterval = pollerInterval,
          aggregationCalculatorVersion = aggregationCalculatorVersion,
          proofsLimit = maxProofsPerAggregation
        ),
        nextBlockNumberToPoll = startBlockNumberInclusive.toLong(),
        aggregationCalculator = globalAggregationCalculator,
        aggregationsRepository = aggregationsRepository,
        consecutiveProvenBlobsProvider = consecutiveProvenBlobsProvider,
        proofAggregationClient = proofAggregationClient,
        aggregationL2StateProvider = AggregationL2StateProviderImpl(
          vertx = vertx,
          l2web3jClient = l2web3jClient,
          l2MessageServiceClient = l2MessageServiceClient
        ),
        provenAggregationEndBlockNumberConsumer = provenAggregationEndBlockNumberConsumer
      )

      return LongRunningService.compose(deadlineCheckRunner, proofAggregationService)
    }
  }

  override fun handleError(error: Throwable) {
    log.error("Error polling blobs for aggregation: errorMessage={}", error.message, error)
  }
}
