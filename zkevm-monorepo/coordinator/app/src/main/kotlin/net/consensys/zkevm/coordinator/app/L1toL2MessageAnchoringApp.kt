package net.consensys.zkevm.coordinator.app

import io.vertx.core.Vertx
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.EIP1559GasProvider
import net.consensys.linea.contract.L2MessageService
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.linea.contract.RollupSmartContractClientWeb3JImpl
import net.consensys.linea.contract.Web3JLogsClient
import net.consensys.linea.web3j.SmartContractErrors
import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.ethereum.coordination.messageanchoring.L1EventQuerier
import net.consensys.zkevm.ethereum.coordination.messageanchoring.L1EventQuerierImpl
import net.consensys.zkevm.ethereum.coordination.messageanchoring.L2MessageAnchorer
import net.consensys.zkevm.ethereum.coordination.messageanchoring.L2MessageAnchorerImpl
import net.consensys.zkevm.ethereum.coordination.messageanchoring.L2Querier
import net.consensys.zkevm.ethereum.coordination.messageanchoring.L2QuerierImpl
import net.consensys.zkevm.ethereum.coordination.messageanchoring.MessageAnchoringService
import org.apache.logging.log4j.LogManager
import org.web3j.protocol.Web3j
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.toKotlinDuration

class L1toL2MessageAnchoringApp(
  vertx: Vertx,
  configs: Config,
  l1Web3jClient: Web3j,
  l2Web3jClient: Web3j,
  smartContractErrors: SmartContractErrors,
  l2MessageService: L2MessageService,
  l2TransactionManager: AsyncFriendlyTransactionManager
) : LongRunningService {
  private val log = LogManager.getLogger(this::class.java)

  private var l1GasProvider = EIP1559GasProvider(
    l1Web3jClient,
    EIP1559GasProvider.Config(
      configs.l1.gasLimit,
      configs.l1.maxFeePerGasCap,
      configs.l1.feeHistoryBlockCount.toUInt(),
      configs.l1.feeHistoryRewardPercentile
    )
  )

  data class Config(
    val l1: L1Config,
    val l2: L2Config,
    val l1Signer: SignerConfig,
    val l2Signer: SignerConfig,
    val messageAnchoringService: MessageAnchoringServiceConfig
  )

  private val messageAnchoringService: MessageAnchoringService = run {
    val l1TransactionManager = createTransactionManager(
      vertx,
      configs.l1Signer,
      l1Web3jClient
    )
    val lineaRollup = instantiateLineaRollupContractClient(
      configs.l1,
      l1TransactionManager,
      l1Web3jClient,
      smartContractErrors
    )

    val l1EventQuerier: L1EventQuerier = L1EventQuerierImpl(
      vertx,
      L1EventQuerierImpl.Config(
        configs.l1.sendMessageEventPollingInterval.toKotlinDuration(),
        configs.l1.maxEventScrapingTime.toKotlinDuration(),
        configs.l1.earliestBlock,
        configs.l1.maxMessagesToCollect,
        configs.l1.zkEvmContractAddress,
        configs.l1.finalizedBlockTag,
        configs.l1.blockRangeLoopLimit
      ),
      l1Web3jClient
    )

    val l2Querier: L2Querier = L2QuerierImpl(
      l2Web3jClient,
      l2MessageService,
      L2QuerierImpl.Config(
        configs.l2.blocksToFinalization,
        configs.l2.lastHashSearchWindow,
        configs.l2.lastHashSearchMaxBlocksBack,
        l2MessageService.contractAddress
      ),
      vertx
    )

    val l2MessageAnchorer: L2MessageAnchorer = L2MessageAnchorerImpl(
      vertx,
      l2Web3jClient,
      l2MessageService,
      L2MessageAnchorerImpl.Config(
        configs.l2.anchoringReceiptPollingInterval.toKotlinDuration(),
        configs.l2.maxReceiptRetries,
        configs.l2.blocksToFinalization.toLong()
      )
    )

    val anchoringConfig = MessageAnchoringService.Config(
      configs.messageAnchoringService.pollingInterval.toKotlinDuration(),
      configs.messageAnchoringService.maxMessagesToAnchor
    )

    MessageAnchoringService(
      anchoringConfig,
      vertx,
      l1EventQuerier,
      l2MessageAnchorer,
      l2Querier,
      RollupSmartContractClientWeb3JImpl(
        Web3JLogsClient(vertx, l1Web3jClient),
        lineaRollup
      ),
      l2MessageService,
      l2TransactionManager
    )
  }

  override fun start(): SafeFuture<Unit> {
    return messageAnchoringService.start().thenPeek {
      log.info("L1toL2MessageAnchoringApp started")
    }
  }

  override fun stop(): SafeFuture<Unit> {
    return messageAnchoringService.stop().thenPeek {
      log.info("L1toL2MessageAnchoringApp stopped")
    }
  }

  private fun instantiateLineaRollupContractClient(
    l1Config: L1Config,
    transactionManager: AsyncFriendlyTransactionManager,
    l1Client: Web3j,
    smartContractErrors: SmartContractErrors
  ): LineaRollupAsyncFriendly {
    return LineaRollupAsyncFriendly.load(
      l1Config.zkEvmContractAddress,
      l1Client,
      transactionManager,
      l1GasProvider,
      smartContractErrors
    )
  }
}
