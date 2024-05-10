package net.consensys.zkevm.ethereum.submission

import io.vertx.core.Vertx
import kotlinx.datetime.Clock
import net.consensys.encodeHex
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.zkevm.PeriodicPollingService
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.toBlockIntervalsString
import net.consensys.zkevm.ethereum.settlement.BlobSubmitter
import net.consensys.zkevm.persistence.blob.BlobsRepository
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import kotlin.time.Duration

class BlobSubmissionCoordinatorImpl(
  private val config: Config,
  private val blobSubmitter: BlobSubmitter,
  private val blobsRepository: BlobsRepository,
  private val lineaRollup: LineaRollupAsyncFriendly,
  private val vertx: Vertx,
  private val clock: Clock,
  private val log: Logger = LogManager.getLogger(BlobSubmissionCoordinatorImpl::class.java)
) : PeriodicPollingService(
  vertx = vertx,
  pollingIntervalMs = config.pollingInterval.inWholeMilliseconds,
  log = log
) {
  class Config(
    val pollingInterval: Duration,
    val proofSubmissionDelay: Duration,
    val maxBlobsToSubmitPerTick: UInt
  )

  override fun action(): SafeFuture<Unit> {
    return lineaRollup.updateNonceAndReferenceBlockToLastL1Block()
      .thenCompose {
        SafeFuture.of(lineaRollup.currentL2BlockNumber().sendAsync())
          .thenCompose(::getBlobsToSubmit)
          .thenCompose(::submitBlobsAfterEthCall)
      }
  }

  private fun getBlobsToSubmit(lastFinalizedBlockNumber: BigInteger): SafeFuture<List<BlobRecord>> {
    return blobsRepository.getConsecutiveBlobsFromBlockNumber(
      lastFinalizedBlockNumber.inc().toLong(),
      clock.now().minus(config.proofSubmissionDelay)
    ).thenCompose { blobRecords ->
      filterOutAlreadySubmittedBlobRecords(blobRecords)
        .thenApply { filterOutRecordsThatDoNotFollowSameSubmissionMethod(it) }
        .thenApply { blobRecordsToSubmit ->
          val blobRecordsToSubmitCappedToLimit = blobRecordsToSubmit.take(config.maxBlobsToSubmitPerTick.toInt())
          if (blobRecordsToSubmit.isNotEmpty()) {
            log.info(
              "blobs to submit: lastFinalizedBlockNumber={} totalBlobs={} maxBlobsToSubmitPerTick={} " +
                "newBlobsToSubmit={} alreadySubmittedBlobs={}",
              lastFinalizedBlockNumber,
              blobRecords.size,
              config.maxBlobsToSubmitPerTick,
              blobRecordsToSubmitCappedToLimit.toBlockIntervalsString(),
              (blobRecords - blobRecordsToSubmit.toSet()).toBlockIntervalsString()
            )
          }
          blobRecordsToSubmitCappedToLimit
        }
    }
  }

  private fun filterOutAlreadySubmittedBlobRecords(
    blobRecords: List<BlobRecord>
  ): SafeFuture<List<BlobRecord>> {
    val ethCallFutures = blobRecords.map { blobRecord ->
      SafeFuture.of(lineaRollup.dataShnarfHashes(blobRecord.blobHash).sendAsync()).thenApply { shnarfFromContract ->
        blobRecord to shnarfFromContract
      }
    }.toTypedArray()
    return SafeFuture.collectAll(*ethCallFutures).thenApply { blobsAndShnarfsFromContract ->
      blobsAndShnarfsFromContract.filter { it.second.contentEquals(zeroHash) }.map { it.first }
    }
  }

  private fun submitBlobsAfterEthCall(
    blobRecords: List<BlobRecord>
  ): SafeFuture<Unit> {
    return if (blobRecords.isEmpty()) {
      SafeFuture.completedFuture(Unit)
    } else {
      blobSubmitter.submitBlobCall(blobRecords.first())
        .whenException { th -> logSubmissionError(log, blobRecords.first().intervalString(), th, isEthCall = true) }
        .thenCompose {
          blobRecords.fold(SafeFuture.completedFuture("")) { chain, blob ->
            chain.thenCompose {
              val nonce = lineaRollup.currentNonce()
              log.debug(
                "Submitting blob: blob={} nonce={}",
                blob.intervalString(),
                nonce
              )
              blobSubmitter.submitBlob(blob)
                .thenPeek { transactionHash ->
                  log.info(
                    "blob submitted: blob={} dataHash={} transactionHash={}, nonce={}",
                    blob.intervalString(),
                    blob.blobCompressionProof!!.dataHash.encodeHex(),
                    transactionHash,
                    nonce
                  )
                }
                .whenException { th -> logSubmissionError(log, blob.intervalString(), th, isEthCall = false) }
            }
          }
        }
    }.thenApply { Unit }
  }

  override fun handleError(error: Throwable) {
    log.error("Error from blob submission coordinator: errorMessage={}", error.message, error)
  }

  companion object {
    internal val zeroHash = ByteArray(32)
  }
}
