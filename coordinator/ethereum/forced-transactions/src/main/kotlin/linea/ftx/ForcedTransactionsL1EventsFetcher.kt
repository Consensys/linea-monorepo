package linea.ftx

import linea.LongRunningService
import linea.contract.events.ForcedTransactionAddedEvent
import linea.contract.l1.LineaRollupSmartContractClientReadOnlyFinalizedStateProvider
import linea.domain.BlockParameter
import linea.domain.BlockParameter.Companion.toBlockParameter
import linea.domain.EthLog
import linea.ethapi.EthLogsClient
import linea.ethapi.EthLogsFilterOptions
import linea.ethapi.EthLogsFilterSubscriptionFactoryPollingBased
import linea.ethapi.extensions.EthLogsFilterSubscriptionManager
import linea.kotlin.toHexStringUInt256
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.CompletableFuture

class ForcedTransactionsL1EventsFetcher(
  private val address: String,
  private val finalizedStateProvider: LineaRollupSmartContractClientReadOnlyFinalizedStateProvider,
  private val ethLogsClient: EthLogsClient,
  private val ethLogsFilterSubscriptionFactory: EthLogsFilterSubscriptionFactoryPollingBased,
  private val l1EarliestBlock: BlockParameter = BlockParameter.Tag.EARLIEST,
  private val l1HighestBlock: BlockParameter = BlockParameter.Tag.FINALIZED,
  private val ftxQueue: MutableMap<ULong, ForcedTransactionAddedEvent>,
  private val log: Logger = LogManager.getLogger(ForcedTransactionsL1EventsFetcher::class.java),
) : LongRunningService {
  private lateinit var eventsSubscription: EthLogsFilterSubscriptionManager

  private fun lastFinalizedForcedTransaction(): SafeFuture<ULong> {
    return finalizedStateProvider
      .getLatestFinalizedState(blockParameter = l1HighestBlock)
      .thenApply { finalizedState -> finalizedState.forcedTransactionNumber }
      .exceptionally { th ->
        if (th is UnsupportedOperationException) {
          // contract is still before V8 or no finalization event yet, let's default to ftx0
          log.info(
            "failed to get finalized forcedTransactionNumber, " +
              "will default forcedTransactionNumber=0, errorMessage={}",
            th.message,
          )
          0UL
        } else {
          throw th
        }
      }
  }

  fun getStartBlockNumberToListenForLogs(): SafeFuture<BlockParameter> {
    return lastFinalizedForcedTransaction()
      .thenCompose { finalizedForcedTransactionNumber ->
        this.log.info(
          "start listening to ForcedTransactionAdded from finalized forcedTransactionNumber={}",
          finalizedForcedTransactionNumber,
        )
        if (finalizedForcedTransactionNumber == 0UL) {
          SafeFuture.completedFuture(l1EarliestBlock)
        } else {
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
                    "lastest finalized forced transaction: l1Block={} event={}",
                    event.log.blockNumber,
                    event.event,
                  )
                }
                // only one FTX per L1 block, so next one can be on next L1 block at best
                ?.blockNumber?.inc()?.toBlockParameter()
                ?: l1EarliestBlock
            }
        }
      }
  }

  override fun start(): CompletableFuture<Unit> {
    if (::eventsSubscription.isInitialized) {
      // already stated
      return CompletableFuture.completedFuture(Unit)
    }

    return getStartBlockNumberToListenForLogs()
      .thenCompose { l1EarliestBlock ->
        val filterOptions = EthLogsFilterOptions(
          fromBlock = l1EarliestBlock,
          toBlock = l1HighestBlock,
          address = address,
          topics = listOf(ForcedTransactionAddedEvent.topic),
        )
        this.log.info("start listening to ForcedTransactionAdded filter={}", filterOptions)
        this.eventsSubscription = ethLogsFilterSubscriptionFactory.create(filterOptions)
        this.eventsSubscription.setConsumer(::onNewForcedTransaction)
        this.eventsSubscription.start()
      }
  }

  fun onNewForcedTransaction(ethLog: EthLog) {
    val event = ForcedTransactionAddedEvent.fromEthLog(ethLog)
    log.debug("ForcedTransactionAdded l1Block={} event={}", event.log.blockNumber, event.event)
    ftxQueue[event.event.forcedTransactionNumber] = event.event
  }

  override fun stop(): CompletableFuture<Unit> {
    return if (this::eventsSubscription.isInitialized) {
      this.eventsSubscription.stop()
    } else {
      CompletableFuture.completedFuture(Unit)
    }
  }
}
