package net.consensys.zkevm.ethereum.coordination.aggregation

import io.vertx.core.Vertx
import linea.clients.ProofAggregationProverClientV2
import linea.domain.Aggregation
import linea.domain.AggregationProofIndex
import linea.domain.ProofToFinalize
import linea.timer.TimerSchedule
import net.consensys.zkevm.ethereum.coordination.AbstractProofPoller
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.milliseconds

class AggregationProofPoller(
  private val aggregationProofClient: ProofAggregationProverClientV2,
  private val aggregationProofHandler: AggregationProofHandler,
  log: Logger,
  vertx: Vertx,
  pollingIntervalMs: Long = 100.milliseconds.inWholeMilliseconds,
  name: String = "AggregationProofPoller",
  timerSchedule: TimerSchedule = TimerSchedule.FIXED_DELAY,
) : AbstractProofPoller<AggregationProofIndex, Aggregation, ProofToFinalize>(
  vertx = vertx,
  pollingIntervalMs = pollingIntervalMs,
  log = log,
  name = name,
  timerSchedule = timerSchedule,
) {
  override fun findProofResponse(proofIndex: AggregationProofIndex): SafeFuture<ProofToFinalize?> =
    aggregationProofClient.findProofResponse(proofIndex)

  override fun handleProof(
    proofIndex: AggregationProofIndex,
    unprovenItem: Aggregation,
    proof: ProofToFinalize,
  ): SafeFuture<*> {
    log.info("aggregation proof generated: aggregation={}", unprovenItem.intervalString())
    return aggregationProofHandler.acceptNewAggregation(unprovenItem.copy(aggregationProof = proof))
  }
}
