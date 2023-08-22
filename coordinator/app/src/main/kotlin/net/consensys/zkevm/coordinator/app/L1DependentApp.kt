package net.consensys.zkevm.coordinator.app

import io.vertx.core.Vertx
import kotlinx.datetime.Clock
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeeHistoryFetcherImpl
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeesCalculator
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeesFetcher
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.WMAFeesCalculator
import net.consensys.zkevm.ethereum.finalization.FinalizationMonitor
import net.consensys.zkevm.ethereum.finalization.FinalizationMonitorImpl
import net.consensys.zkevm.ethereum.settlement.BatchSubmissionCoordinatorService
import net.consensys.zkevm.ethereum.settlement.BatchSubmitter
import net.consensys.zkevm.ethereum.settlement.ZkEvmBatchSubmissionCoordinator
import net.consensys.zkevm.ethereum.settlement.ZkEvmBatchSubmitter
import net.consensys.zkevm.ethereum.settlement.persistence.BatchesRepository
import org.apache.logging.log4j.LogManager
import org.web3j.protocol.Web3j
import org.web3j.protocol.http.HttpService
import org.web3j.utils.Async
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigDecimal
import java.util.concurrent.CompletableFuture
import kotlin.time.toKotlinDuration

class L1DependentApp(
  private val configs: CoordinatorConfig,
  private val vertx: Vertx,
  private val l2Web3jClient: Web3j,
  httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory,
  batchesRepository: BatchesRepository,
  blockFinalizationHandler: (FinalizationMonitor.FinalizationUpdate) -> SafeFuture<*>
) : LongRunningService {
  private val log = LogManager.getLogger(this::class.java)

  init {
    if (configs.messageAnchoringService.disabled) {
      log.warn("Message anchoring service is disabled")
    }
    if (configs.dynamicGasPriceService.disabled) {
      log.warn("Dynamic gas price service is disabled")
    }
    if (configs.batchSubmission.disabled) {
      log.warn("Batch submission is disabled")
    }
  }

  private val l1Web3jClient =
    Web3j.build(
      HttpService(configs.l1.rpcEndpoint.toString()),
      1000,
      Async.defaultExecutorService()
    )

  private val l1TransactionManager = createTransactionManager(vertx, configs.l1Signer, l1Web3jClient)

  private val l1MinMinerTipCalculator: FeesCalculator = WMAFeesCalculator(
    WMAFeesCalculator.Config(
      BigDecimal("0.0"),
      BigDecimal.ONE
    )
  )

  private val feesFetcher: FeesFetcher = FeeHistoryFetcherImpl(
    l1Web3jClient,
    FeeHistoryFetcherImpl.Config(
      configs.l1.feeHistoryBlockCount.toUInt(),
      configs.l1.feeHistoryRewardPercentile
    )
  )

  private val finalizationMonitor = run {
    // To avoid setDefaultBlockParameter clashes
    val zkEvmClientForFinalization: ZkEvmV2AsyncFriendly = instantiateZkEvmContractClient(
      configs.l1,
      l1TransactionManager,
      feesFetcher,
      l1MinMinerTipCalculator,
      l1Web3jClient
    )

    FinalizationMonitorImpl(
      config =
      FinalizationMonitorImpl.Config(
        configs.l1.finalizationPollingInterval.toKotlinDuration(),
        configs.l1.blocksToFinalization
      ),
      contract = zkEvmClientForFinalization,
      l1Client = l1Web3jClient,
      l2Client = l2Web3jClient,
      vertx = vertx
    ).apply {
      addFinalizationHandler("status update", blockFinalizationHandler)
    }
  }

  val batchSubmissionCoordinator = run {
    val zkEvmClientForSubmission: ZkEvmV2AsyncFriendly = instantiateZkEvmContractClient(
      configs.l1,
      l1TransactionManager,
      feesFetcher,
      l1MinMinerTipCalculator,
      l1Web3jClient
    )
    val batchSubmitter: BatchSubmitter = ZkEvmBatchSubmitter(zkEvmClientForSubmission)

    if (configs.batchSubmission.enabled) {
      ZkEvmBatchSubmissionCoordinator(
        ZkEvmBatchSubmissionCoordinator.Config(
          configs.l1.newBatchPollingInterval.toKotlinDuration(),
          configs.batchSubmission.proofSubmissionDelay.toKotlinDuration()
        ),
        batchSubmitter,
        batchesRepository,
        zkEvmClientForSubmission,
        vertx,
        Clock.System
      )
    } else {
      // instantiate a dummy batch submitter when is disabled
      DisabledBatchSubmissionCoordinator()
    }
  }

  fun lastFinalizedBlock(): SafeFuture<ULong> {
    val zkEvmClient: ZkEvmV2AsyncFriendly = instantiateZkEvmContractClient(
      configs.l1,
      l1TransactionManager,
      feesFetcher,
      l1MinMinerTipCalculator,
      l1Web3jClient
    )
    val l1BasedLastFinalizedBlockProvider = L1BasedLastFinalizedBlockProvider(
      vertx,
      zkEvmClient,
      configs.conflation.consistentNumberOfBlocksOnL1ToWait.toUInt()
    )

    return l1BasedLastFinalizedBlockProvider.getLastFinalizedBlock()
  }

  private val messageAnchoringApp: L1toL2MessageAnchoringApp? = run {
    if (configs.messageAnchoringService.enabled) {
      L1toL2MessageAnchoringApp(
        vertx,
        L1toL2MessageAnchoringApp.Config(
          configs.l1,
          configs.l2,
          configs.l2Signer,
          configs.messageAnchoringService
        ),
        l1Web3jClient,
        l2Web3jClient
      )
    } else {
      null
    }
  }

  private val gasPriceUpdaterApp: GasPriceUpdaterApp? = if (configs.dynamicGasPriceService.enabled) {
    GasPriceUpdaterApp(
      vertx,
      httpJsonRpcClientFactory,
      l1Web3jClient,
      GasPriceUpdaterApp.Config(configs.dynamicGasPriceService)
    )
  } else {
    null
  }

  override fun start(): CompletableFuture<Unit> {
    return finalizationMonitor.start()
      .thenCompose { batchSubmissionCoordinator.start() }
      .thenCompose { messageAnchoringApp?.start() ?: SafeFuture.completedFuture(Unit) }
      .thenCompose { gasPriceUpdaterApp?.start() ?: SafeFuture.completedFuture(Unit) }
      .thenPeek {
        log.info("L1App started")
      }
  }

  override fun stop(): CompletableFuture<Unit> {
    return SafeFuture.allOf(
      finalizationMonitor.stop(),
      batchSubmissionCoordinator.stop(),
      messageAnchoringApp?.stop() ?: SafeFuture.completedFuture(Unit),
      gasPriceUpdaterApp?.stop() ?: SafeFuture.completedFuture(Unit)
    )
      .thenCompose { SafeFuture.fromRunnable { l1Web3jClient.shutdown() } }
      .thenApply {
        log.info("L1App Stopped")
        Unit
      }
  }
}

private class DisabledBatchSubmissionCoordinator : BatchSubmissionCoordinatorService {
  private val log = LogManager.getLogger(this::class.java)

  override fun acceptNewBatch(batch: Batch): SafeFuture<Unit> {
    log.info("Batch submission is disabled. Ignoring batch: {}", batch.intervalString())
    return SafeFuture.completedFuture(Unit)
  }

  override fun start(): CompletableFuture<Unit> {
    return SafeFuture.completedFuture(Unit)
  }

  override fun stop(): CompletableFuture<Unit> {
    return SafeFuture.completedFuture(Unit)
  }
}
