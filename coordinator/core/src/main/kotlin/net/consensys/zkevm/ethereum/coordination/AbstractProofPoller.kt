package net.consensys.zkevm.ethereum.coordination

import io.vertx.core.Vertx
import linea.timer.TimerSchedule
import linea.timer.VertxPeriodicPollingService
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ConcurrentLinkedDeque

abstract class AbstractProofPoller<ProofIndex, UnprovenItem, Proof>(
  vertx: Vertx,
  pollingIntervalMs: Long,
  log: Logger,
  name: String,
  timerSchedule: TimerSchedule,
) : VertxPeriodicPollingService(
  vertx = vertx,
  pollingIntervalMs = pollingIntervalMs,
  log = log,
  name = name,
  timerSchedule = timerSchedule,
) {
  protected val proofRequestsInProgress = ConcurrentLinkedDeque<Pair<ProofIndex, UnprovenItem>>()

  @Synchronized
  fun addProofRequestsInProgressForPolling(proofIndex: ProofIndex, unprovenItem: UnprovenItem) {
    proofRequestsInProgress.add(Pair(proofIndex, unprovenItem))
  }

  protected abstract fun findProofResponse(proofIndex: ProofIndex): SafeFuture<Proof?>

  protected abstract fun handleProof(proofIndex: ProofIndex, unprovenItem: UnprovenItem, proof: Proof): SafeFuture<*>

  override fun action(): SafeFuture<*> {
    if (proofRequestsInProgress.isEmpty()) return SafeFuture.completedFuture(Unit)
    val entry = proofRequestsInProgress.peekFirst()
    val (proofIndex, unprovenItem) = entry
    return findProofResponse(proofIndex).thenCompose { proof ->
      if (proof != null) {
        handleProof(proofIndex, unprovenItem, proof).thenApply {
          proofRequestsInProgress.remove(entry)
        }
      } else {
        SafeFuture.completedFuture(Unit)
      }
    }
  }
}
