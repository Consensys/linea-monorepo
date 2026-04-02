package linea.ftx.conflation

import io.vertx.core.Vertx
import linea.contract.l1.LineaRollupSmartContractClientReadOnlyFinalizedStateProvider
import linea.domain.BlockParameter
import linea.persistence.ftx.ForcedTransactionsDao
import linea.timer.TimerSchedule
import linea.timer.VertxPeriodicPollingService
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicReference
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

/**
 * Cleans up forced transaction database records periodically based on latest finalized ftx number.
 *
 * Responsibilities:
 * 1. Monitor finalized FTX number (polling)
 * 2. Remove FTX records with FTX number <= finalized FTX number
 */
class ForcedTransactionsDbCleanUpService(
  private val ftxDao: ForcedTransactionsDao,
  private val finalizedStateProvider: LineaRollupSmartContractClientReadOnlyFinalizedStateProvider,
  vertx: Vertx,
  pollingInterval: Duration = 5.seconds,
  log: Logger = LogManager.getLogger(ForcedTransactionsDbCleanUpService::class.java),
) : VertxPeriodicPollingService(
  vertx = vertx,
  pollingIntervalMs = pollingInterval.inWholeMilliseconds,
  log = log,
  name = "ForcedTransactionsDbCleanUpService",
  timerSchedule = TimerSchedule.FIXED_DELAY,
) {
  private val lastFinalizedFtxNumber = AtomicReference<ULong>(null)
  override fun action(): SafeFuture<*> = tick()

  @Synchronized
  private fun tick(): SafeFuture<Unit> {
    return finalizedStateProvider
      .getLatestFinalizedState(blockParameter = BlockParameter.Tag.FINALIZED)
      .thenCompose { currentFinalizationUpdate ->
        if (lastFinalizedFtxNumber.get() != currentFinalizationUpdate.forcedTransactionNumber) {
          lastFinalizedFtxNumber.set(currentFinalizationUpdate.forcedTransactionNumber)
          ftxDao.deleteFtxUpToInclusive(currentFinalizationUpdate.forcedTransactionNumber)
        }
        SafeFuture.completedFuture(Unit)
      }
  }
}
