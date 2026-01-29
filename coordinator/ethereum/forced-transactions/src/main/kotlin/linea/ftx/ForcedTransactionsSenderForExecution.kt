package linea.ftx

import io.vertx.core.Vertx
import linea.LongRunningService
import linea.contract.events.ForcedTransactionAddedEvent
import linea.forcedtx.ForcedTransactionRequest
import linea.forcedtx.ForcedTransactionsClient
import linea.timer.TimerSchedule
import linea.timer.VertxPeriodicPollingService
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration

/**
 * Takes ForcedTransactionEventAdded from the queue and sends them to the sequencer.
 */
class ForcedTransactionsSenderForExecution(
  vertx: Vertx,
  val alreadyProcessed: ProcessedTransactionsFilter,
  val ftxClient: ForcedTransactionsClient,
  val ftxQueue: MutableMap<ULong, ForcedTransactionAddedEvent>,
  val pollingInterval: Duration,
  val log: Logger = LogManager.getLogger(ForcedTransactionsSenderForExecution::class.java),
  val txLimitToSendPerTick: Int = 10,
) :
  VertxPeriodicPollingService(
    vertx = vertx,
    pollingIntervalMs = pollingInterval.inWholeMilliseconds,
    log = log,
    timerSchedule = TimerSchedule.FIXED_DELAY,
    name = "ForcedTransactionsRelayerForExecution",
  ),
  LongRunningService {

  override fun action(): SafeFuture<*> {
    val allFtx = ftxQueue.values.toList()
    log.trace("all ftxs in queue={}", allFtx.map { it.forcedTransactionNumber })
    if (allFtx.isEmpty()) {
      return SafeFuture.completedFuture(Unit)
    }

    return alreadyProcessed
      .filterOutAlreadyProcessed(allFtx)
      .thenCompose { unprocessedTxs ->
        log.debug("unprocessed ftxs={}", unprocessedTxs.map { it.forcedTransactionNumber })
        this.sendTransactions(unprocessedTransactions = unprocessedTxs.take(txLimitToSendPerTick))
      }
  }

  fun sendTransactions(unprocessedTransactions: List<ForcedTransactionAddedEvent>): SafeFuture<*> {
    log.info("sending forced transactions for execution: ftxs={}", unprocessedTransactions)
    val requests = unprocessedTransactions.map {
      ForcedTransactionRequest(
        ftxNumber = it.forcedTransactionNumber,
        deadlineBlockNumber = it.blockNumberDeadline,
        ftxRlp = it.rlpEncodedSignedTransaction,
      )
    }
    return ftxClient
      .lineaSendForcedRawTransaction(requests)
      .thenPeek { responses ->
        log.debug("successfully sent forced transactions for execution: responses={}", responses)
      }
      .whenException { exception ->
        log.warn(
          "Failed sending forced transactions to sequencer, will retry in {}. errorMessage={} ftx={}",
          pollingInterval,
          exception.message,
          unprocessedTransactions,
          exception,
        )
      }
  }
}
