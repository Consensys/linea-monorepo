package net.consensys.zkevm.ethereum.finalization

import build.linea.contract.LineaRollupV5
import io.vertx.core.Vertx
import net.consensys.linea.async.AsyncRetryer
import net.consensys.linea.async.toSafeFuture
import net.consensys.toULong
import net.consensys.zkevm.PeriodicPollingService
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.DefaultBlockParameterName
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.CompletableFuture
import java.util.concurrent.atomic.AtomicReference
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

data class FinalizationUpdatePollerConfig(
  val pollingInterval: Duration = 12.seconds,
  val blockTag: String
) {
  init {
    require(pollingInterval >= 0.seconds) {
      "pollingInterval must be greater than 0"
    }

    require(DefaultBlockParameterName.fromString(blockTag) != null) {
      "Invalid blockTag='$blockTag', " +
        "valid values: ${DefaultBlockParameterName.values().joinToString(", ")}"
    }
  }
}

class FinalizationUpdatePoller(
  private val vertx: Vertx,
  private val config: FinalizationUpdatePollerConfig,
  private val lineaRollup: LineaRollupV5,
  private val finalizationHandler: (ULong) -> CompletableFuture<*>,
  private val log: Logger = LogManager.getLogger(FinalizationUpdatePoller::class.java)
) : PeriodicPollingService(
  vertx = vertx,
  pollingIntervalMs = config.pollingInterval.inWholeMilliseconds,
  log = log
) {
  private val lastFinalizationRef: AtomicReference<ULong> = AtomicReference(null)

  init {
    lineaRollup.setDefaultBlockParameter(DefaultBlockParameter.valueOf(config.blockTag))
  }

  override fun action(): SafeFuture<*> {
    return AsyncRetryer.retry(
      vertx,
      backoffDelay = config.pollingInterval
    ) {
      SafeFuture.of(lineaRollup.currentL2BlockNumber().sendAsync())
    }
      .thenCompose { lineaFinalizedBlockNumber ->
        val prevFinalizedBlockNumber = lastFinalizationRef.get()
        lastFinalizationRef.set(lineaFinalizedBlockNumber.toULong())
        if (prevFinalizedBlockNumber != lineaFinalizedBlockNumber.toULong()) {
          finalizationHandler(lineaFinalizedBlockNumber.toULong()).thenApply { Unit }
        } else {
          CompletableFuture.completedFuture(Unit)
        }
      }
      .toSafeFuture()
  }

  override fun handleError(error: Throwable) {
    if (error.cause is UnsupportedOperationException) {
      log.error(
        "\"setFinalizedBlock\" and \"setSafeBlock\" methods are not supported in the hosting Besu client, " +
          "the poller will stop now, please check the Besu client's settings"
      )
      super.stop()
    } else {
      log.warn("Error when polling/handling Linea finalized block number", error)
    }
  }

  fun finalizedBlockNumber(): ULong {
    return lastFinalizationRef.get() ?: throw IllegalStateException("No finalization update available")
  }
}
