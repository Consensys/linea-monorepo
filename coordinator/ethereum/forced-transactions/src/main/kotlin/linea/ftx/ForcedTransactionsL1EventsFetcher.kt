package linea.ftx

import linea.LongRunningService
import linea.contract.events.ForcedTransactionAddedEvent
import linea.domain.BlockParameter
import linea.domain.BlockParameter.Companion.toBlockParameter
import linea.domain.EthLog
import linea.ethapi.EthLogsClient
import linea.ethapi.EthLogsFilterOptions
import linea.ethapi.extensions.EthLogsFilterSubscriptionFactory
import linea.ethapi.extensions.EthLogsFilterSubscriptionManager
import linea.kotlin.toHexStringUInt256
import net.consensys.linea.async.toSafeFuture
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.Queue
import java.util.concurrent.CompletableFuture

class ForcedTransactionsL1EventsFetcher(
  private val address: String,
  private val ethLogsClient: EthLogsClient,
  private val resumePointProvider: ForcedTransactionsResumePointProvider,
  private val ethLogsFilterSubscriptionFactory: EthLogsFilterSubscriptionFactory,
  private val l1EarliestBlock: BlockParameter = BlockParameter.Tag.EARLIEST,
  private val l1HighestBlock: BlockParameter = BlockParameter.Tag.FINALIZED,
  private val ftxQueue: Queue<ForcedTransactionAddedEvent>,
  private val log: Logger = LogManager.getLogger(ForcedTransactionsL1EventsFetcher::class.java),
) : LongRunningService {
  private lateinit var eventsSubscription: EthLogsFilterSubscriptionManager

  fun getStartBlockNumberToListenForLogs(): SafeFuture<BlockParameter> {
    return resumePointProvider
      .lastFinalizedForcedTransaction()
      .thenCompose { finalizedForcedTransactionNumber ->
        val l1BlockToStartListeningFromFuture = if (finalizedForcedTransactionNumber == 0UL) {
          // no finalized forced transactions yet, start from the beginning
          SafeFuture.completedFuture(l1EarliestBlock)
        } else {
          // get l1 block number of the latest finalized forced transaction,
          // and start listening from there
          ethLogsClient.getLogs(
            fromBlock = l1EarliestBlock,
            toBlock = l1HighestBlock,
            address = address,
            topics = listOf(
              ForcedTransactionAddedEvent.topic,
              finalizedForcedTransactionNumber.toHexStringUInt256(),
            ),
          )
            .thenApply { logs ->
              logs.firstOrNull()
                ?.also {
                  val event = ForcedTransactionAddedEvent.fromEthLog(it)
                  log.debug(
                    "latest finalized forced transaction: l1Block={} event={}",
                    event.log.blockNumber,
                    event.event,
                  )
                }
                // only one FTX per L1 block, so the next one can be on the next L1 block at best
                ?.blockNumber?.inc()?.toBlockParameter()
                ?: l1EarliestBlock
            }
        }

        l1BlockToStartListeningFromFuture
          .thenPeek { l1BlockToStartListeningFrom ->
            log.info(
              "start listening to ForcedTransactionAdded: from l1BlockNumber={} finalizedFtx={}",
              l1BlockToStartListeningFrom,
              finalizedForcedTransactionNumber,
            )
          }
      }
  }

  @Synchronized
  fun startPoller(): CompletableFuture<Unit> {
    if (::eventsSubscription.isInitialized) {
      // already started
      return CompletableFuture.completedFuture(Unit)
    }

    return getStartBlockNumberToListenForLogs()
      .thenCompose { l1BlockSinceLastFinalizedFtx ->
        val filterOptions = EthLogsFilterOptions(
          fromBlock = l1BlockSinceLastFinalizedFtx,
          toBlock = l1HighestBlock,
          address = address,
          topics = listOf(ForcedTransactionAddedEvent.topic),
        )
        this.log.info("start listening to ForcedTransactionAdded filter={}", filterOptions)
        val tmpEventsSubscription = ethLogsFilterSubscriptionFactory
          .create(
            filterOptions = filterOptions,
            logsConsumer = ::onNewForcedTransaction,
          )
        tmpEventsSubscription
          .start()
          .toSafeFuture()
          .thenPeek {
            // only update local reference if subscription started successfully
            this.eventsSubscription = tmpEventsSubscription
          }
      }
  }

  override fun start(): CompletableFuture<Unit> = startPoller()

  private fun onNewForcedTransaction(ethLog: EthLog) {
    val event = ForcedTransactionAddedEvent.fromEthLog(ethLog)
    log.info("event ForcedTransactionAdded: l1Block={} event={}", event.log.blockNumber, event.event)
    // if queue is full, will throw and events fetch will retry on next tick
    ftxQueue.add(event.event)
  }

  override fun stop(): CompletableFuture<Unit> {
    return if (this::eventsSubscription.isInitialized) {
      this.eventsSubscription.stop()
    } else {
      CompletableFuture.completedFuture(Unit)
    }
  }
}
