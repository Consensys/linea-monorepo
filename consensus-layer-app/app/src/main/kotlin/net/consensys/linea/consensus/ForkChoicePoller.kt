package net.consensys.linea.consensus

import io.vertx.core.Vertx
import net.consensys.linea.forkchoicestate.ForkChoiceState
import net.consensys.linea.forkchoicestate.ForkChoiceStateInfoV0
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.ethereum.executionclient.ExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.time.Duration
import java.util.Optional

fun ForkChoiceState.asTekuForkChoiceStateV1(): ForkChoiceStateV1 =
  ForkChoiceStateV1(this.headBlockHash, this.safeBlockHash, this.finalizedBlockHash)

class ForkChoicePoller(
  private val vertx: Vertx,
  private val pollingInterval: Duration,
  private val forkChoiceStateClient: ForkChoiceStateClient,
  private val executionClient: ExecutionEngineClient,
  private val allowRevert: Boolean = false
) {
  private val log: Logger = LogManager.getLogger()
  private var timerId: Long? = null
  private var lastForkChoiceState: ForkChoiceStateInfoV0 =
    forkChoiceStateClient.getForkChoiceState().get()

  private fun updateInternalForkChoiceState(newForkChoice: ForkChoiceStateInfoV0): Boolean {
    if (newForkChoice == lastForkChoiceState) return false

    val isRevert = newForkChoice.headBlockNumber <= lastForkChoiceState.headBlockNumber
    if (isRevert) {
      log.warn("Revert detected: currentState={} newState={}", lastForkChoiceState, newForkChoice)
      if (!allowRevert) {
        return false
      }
    }
    log.info("ForkChoiceState updated {} --> {}", lastForkChoiceState, newForkChoice)
    lastForkChoiceState = newForkChoice
    return true
  }

  fun updateExecutionLayer(fcu: ForkChoiceStateInfoV0): SafeFuture<Unit> {
    return executionClient
      .forkChoiceUpdatedV2(fcu.forkChoiceState.asTekuForkChoiceStateV1(), Optional.empty())
      .whenException { th -> log.error("Execution client update failed: {}", th) }
      .thenAccept { result -> log.info("Execution client update result: {}", result) }
      .thenApply { null }
  }

  fun updateFlow() {
    forkChoiceStateClient.getForkChoiceState().thenApply {
        forkChoiceStateInfo: ForkChoiceStateInfoV0 ->
      if (updateInternalForkChoiceState(forkChoiceStateInfo)) {
        this.updateExecutionLayer(forkChoiceStateInfo)
      }
    }
  }

  fun startPoller() {
    timerId = vertx.setPeriodic(pollingInterval.toMillis()) { _ -> updateFlow() }
  }

  fun stopPoller() {
    timerId?.let { vertx.cancelTimer(it) }
  }
}
