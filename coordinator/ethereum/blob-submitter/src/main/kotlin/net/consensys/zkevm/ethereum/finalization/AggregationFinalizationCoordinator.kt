package net.consensys.zkevm.ethereum.finalization

import io.vertx.core.Vertx
import kotlinx.datetime.Clock
import net.consensys.encodeHex
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.toULong
import net.consensys.zkevm.PeriodicPollingService
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.toBlockIntervalsString
import net.consensys.zkevm.ethereum.submission.getErrorAppropriateLevel
import net.consensys.zkevm.ethereum.submission.logSubmissionError
import net.consensys.zkevm.persistence.aggregation.AggregationsRepository
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration

class AggregationFinalizationCoordinator(
  private val config: Config,
  private val aggregationFinalization: AggregationFinalization,
  private val aggregationsRepository: AggregationsRepository,
  private val lineaRollup: LineaRollupAsyncFriendly,
  private val vertx: Vertx,
  private val clock: Clock,
  private val log: Logger = LogManager.getLogger(AggregationFinalizationCoordinator::class.java)
) : PeriodicPollingService(
  vertx = vertx,
  pollingIntervalMs = config.pollingInterval.inWholeMilliseconds,
  log = log
) {
  class Config(
    val pollingInterval: Duration,
    val proofSubmissionDelay: Duration,
    val maxAggregationsToFinalizePerIteration: UInt
  ) {
    init {
      require(maxAggregationsToFinalizePerIteration == 1u) {
        "only maxAggregationsToFinalizePerIteration=1 is supported for now"
      }
    }
  }

  override fun action(): SafeFuture<Unit> {
    return lineaRollup.resetNonce()
      .thenComposeCombined(SafeFuture.of(lineaRollup.currentL2BlockNumber().sendAsync())) { _, lastFinalizedBlock ->
        log.debug(
          "fetching aggregation proofs for finalization: lastFinalizedBlock={}",
          lastFinalizedBlock.toLong()
        )
        aggregationsRepository.getProofsToFinalize(
          lastFinalizedBlock.toLong() + 1,
          Aggregation.Status.Proven,
          clock.now().minus(config.proofSubmissionDelay),
          config.maxAggregationsToFinalizePerIteration.toInt()
        )
          .thenCompose { aggregationsProofs ->
            if (aggregationsProofs.isNullOrEmpty()) {
              log.debug("Skipping finalization as there are no new aggregation proofs to finalize")
              SafeFuture.completedFuture(Unit)
            } else {
              filterAggregationsWithAllDataSubmitted(aggregationsProofs).thenCompose { checkedAggregations ->
                if (checkedAggregations.isNullOrEmpty()) {
                  log.debug("Skipping finalization as there are no aggregations ready to finalize")
                  SafeFuture.completedFuture(Unit)
                } else {
                  log.info(
                    "aggregations ready to submit: lastFinalizedBlock={} submitting aggregation={} " +
                      "all aggregations={} ",
                    lastFinalizedBlock.toULong(),
                    checkedAggregations.first().intervalString(),
                    checkedAggregations.toBlockIntervalsString()
                  )
                  finalizeAggregationAfterEthCall(checkedAggregations.first())
                }
              }
            }
          }
      }
  }

  private fun logAggregationSubmissionError(
    intervalString: String,
    error: Throwable,
    isEthCall: Boolean = false
  ) {
    val logLevel = if (!error.message.isNullOrEmpty() &&
      error.message!!.contains("Contract Call has been reverted by the EVM with the reason", ignoreCase = true) &&
      error.message!!.contains("FinalBlockStateEqualsZeroHash", ignoreCase = true)
    ) {
      // this means current aggregation has blobs that were submitted but not finalized due to high gas price
      Level.DEBUG
    } else {
      getErrorAppropriateLevel(error, isEthCall)
    }

    logSubmissionError(
      log,
      "{} for aggregation finalization failed: aggregation={} errorMessage={}",
      intervalString,
      error,
      isEthCall,
      logLevel
    )
  }

  private fun finalizeAggregationAfterEthCall(
    aggregationProof: ProofToFinalize
  ): SafeFuture<Unit> {
    return aggregationFinalization
      .finalizeAggregationEthCall(aggregationProof)
      .whenException { th -> logAggregationSubmissionError(aggregationProof.intervalString(), th, isEthCall = true) }
      .thenCompose {
        aggregationFinalization.finalizeAggregation(aggregationProof)
          .whenException { th -> logAggregationSubmissionError(aggregationProof.intervalString(), th) }
          .whenSuccess {
            log.info(
              "finalization transaction accepted: aggregation={}",
              aggregationProof.intervalString()
            )
          }
          .thenApply { }
      }
  }

  private fun filterAggregationsWithAllDataSubmitted(
    aggregationProofs: List<ProofToFinalize>
  ): SafeFuture<List<ProofToFinalize>> {
    data class CheckedAggregationSubmissionData(
      val aggregationProof: ProofToFinalize,
      val verified: Boolean
    )

    val checkedProofsFutures = aggregationProofs.map { proof ->
      checkForUnsubmittedData(proof.dataHashes)
        .thenApply { unsubmittedData ->
          if (unsubmittedData.isEmpty()) {
            CheckedAggregationSubmissionData(proof, true)
          } else {
            log.info(
              "aggregation not ready to submit aggregation={} unsubmitted data: dataHashes={}",
              proof.intervalString(),
              unsubmittedData.joinToString(separator = ", ", prefix = "[", postfix = "]") { byteArray ->
                byteArray.encodeHex()
              }
            )
            CheckedAggregationSubmissionData(proof, false)
          }
        }
    }.toTypedArray()
    val checkedProofs = SafeFuture.collectAll(*checkedProofsFutures)
    return checkedProofs.thenApply { proofs ->
      proofs.takeWhile { it.verified }.map { it.aggregationProof }
    }
  }

  private fun checkForUnsubmittedData(
    dataHashes: List<ByteArray>
  ): SafeFuture<List<ByteArray>> {
    val ethCallFutures = dataHashes.map { hash ->
      SafeFuture.of(lineaRollup.dataShnarfHashes(hash).sendAsync()).thenApply { shnarfFromContract ->
        hash to shnarfFromContract
      }
    }.toTypedArray()
    return SafeFuture.collectAll(*ethCallFutures).thenApply { hashToShnarf ->
      hashToShnarf.filter { it.second.contentEquals(Bytes32.ZERO.toArray()) }.map { it.first }
    }
  }

  override fun handleError(error: Throwable) {
    log.error("Error from aggregation finalization coordinator: errorMessage={}", error.message, error)
  }
}
