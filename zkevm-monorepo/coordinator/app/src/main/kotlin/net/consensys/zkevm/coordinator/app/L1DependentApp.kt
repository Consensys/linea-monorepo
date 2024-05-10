package net.consensys.zkevm.coordinator.app

import io.vertx.core.Vertx
import kotlinx.datetime.Clock
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.contract.BoundableFeeCalculator
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.linea.contract.WMAFeesCalculator
import net.consensys.linea.contract.Web3JL2MessageService
import net.consensys.linea.contract.Web3JL2MessageServiceLogsClient
import net.consensys.linea.contract.Web3JLogsClient
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.web3j.SmartContractErrors
import net.consensys.linea.web3j.Web3jBlobExtended
import net.consensys.linea.web3j.okHttpClientBuilder
import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.coordinator.blockcreation.BatchesRepoBasedLastProvenBlockNumberProvider
import net.consensys.zkevm.coordinator.blockcreation.BlockCreationMonitor
import net.consensys.zkevm.coordinator.blockcreation.ExtendedWeb3JImpl
import net.consensys.zkevm.coordinator.blockcreation.GethCliqueSafeBlockProvider
import net.consensys.zkevm.coordinator.blockcreation.TracesFilesManager
import net.consensys.zkevm.coordinator.clients.ShomeiClient
import net.consensys.zkevm.coordinator.clients.TracesGeneratorJsonRpcClientV1
import net.consensys.zkevm.coordinator.clients.Type2StateManagerClient
import net.consensys.zkevm.coordinator.clients.Type2StateManagerJsonRpcClient
import net.consensys.zkevm.coordinator.clients.prover.FileBasedBlobCompressionProverClient
import net.consensys.zkevm.coordinator.clients.prover.FileBasedExecutionProverClient
import net.consensys.zkevm.coordinator.clients.prover.FileBasedProofAggregationClient
import net.consensys.zkevm.domain.BlocksConflation
import net.consensys.zkevm.encoding.ExecutionPayloadV1RLPEncoderByBesuImplementation
import net.consensys.zkevm.ethereum.coordination.HighestAggregationTracker
import net.consensys.zkevm.ethereum.coordination.HighestConflationTracker
import net.consensys.zkevm.ethereum.coordination.HighestProvenBatchTracker
import net.consensys.zkevm.ethereum.coordination.HighestProvenBlobTracker
import net.consensys.zkevm.ethereum.coordination.HighestUnprovenBlobTracker
import net.consensys.zkevm.ethereum.coordination.SimpleCompositeSafeFutureHandler
import net.consensys.zkevm.ethereum.coordination.aggregation.ConsecutiveProvenBlobsProviderWithLastEndBlockNumberTracker
import net.consensys.zkevm.ethereum.coordination.aggregation.ProofAggregationCoordinatorService
import net.consensys.zkevm.ethereum.coordination.blob.BlobCompressionProofCoordinator
import net.consensys.zkevm.ethereum.coordination.blob.BlobCompressionProofUpdate
import net.consensys.zkevm.ethereum.coordination.blob.BlobZkStateProviderImpl
import net.consensys.zkevm.ethereum.coordination.blob.Eip4844SwitchProviderImpl
import net.consensys.zkevm.ethereum.coordination.blob.GoBackedBlobCompressor
import net.consensys.zkevm.ethereum.coordination.blob.GoBackedBlobShnarfCalculator
import net.consensys.zkevm.ethereum.coordination.blob.RollingBlobShnarfCalculator
import net.consensys.zkevm.ethereum.coordination.blockcreation.ForkChoiceUpdaterImpl
import net.consensys.zkevm.ethereum.coordination.conflation.BlockToBatchSubmissionCoordinator
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorByBlockLimit
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorByDataCompressed
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorByExecutionTraces
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorByTimeDeadline
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationService
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationServiceImpl
import net.consensys.zkevm.ethereum.coordination.conflation.DeadlineConflationCalculatorRunner
import net.consensys.zkevm.ethereum.coordination.conflation.GlobalBlobAwareConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.GlobalBlockConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.ProofGeneratingConflationHandlerImpl
import net.consensys.zkevm.ethereum.coordination.conflation.TracesConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.TracesConflationCoordinatorImpl
import net.consensys.zkevm.ethereum.coordination.dynamicgasprice.FeeHistoryFetcherImpl
import net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap.FeeHistoryCachingService
import net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap.GasPriceCapCalculatorImpl
import net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap.GasPriceCapFeeHistoryFetcherImpl
import net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap.GasPriceCapProviderForDataSubmission
import net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap.GasPriceCapProviderForFinalization
import net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap.GasPriceCapProviderImpl
import net.consensys.zkevm.ethereum.coordination.proofcreation.ZkProofCreationCoordinatorImpl
import net.consensys.zkevm.ethereum.finalization.AggregationFinalization
import net.consensys.zkevm.ethereum.finalization.AggregationFinalizationAsCallData
import net.consensys.zkevm.ethereum.finalization.AggregationFinalizationCoordinator
import net.consensys.zkevm.ethereum.finalization.FinalizationMonitor
import net.consensys.zkevm.ethereum.finalization.FinalizationMonitorImpl
import net.consensys.zkevm.ethereum.gaspricing.FeeHistoriesRepositoryWithCache
import net.consensys.zkevm.ethereum.gaspricing.FeesCalculator
import net.consensys.zkevm.ethereum.gaspricing.FeesFetcher
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapCalculator
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapFeeHistoryFetcher
import net.consensys.zkevm.ethereum.gaspricing.L1_BLOCK_TIME_SECONDS
import net.consensys.zkevm.ethereum.submission.BlobSubmissionCoordinatorImpl
import net.consensys.zkevm.ethereum.submission.BlobSubmitterAsCallData
import net.consensys.zkevm.ethereum.submission.BlobSubmitterAsEIP4844
import net.consensys.zkevm.ethereum.submission.Eip4844SwitchAwareBlobSubmitter
import net.consensys.zkevm.persistence.aggregation.AggregationsRepository
import net.consensys.zkevm.persistence.blob.BlobsRepository
import net.consensys.zkevm.persistence.dao.batch.persistence.BatchProofHandlerImpl
import net.consensys.zkevm.persistence.dao.batch.persistence.BatchesRepository
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import org.web3j.protocol.http.HttpService
import org.web3j.utils.Async
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigDecimal
import java.math.BigInteger
import java.util.concurrent.CompletableFuture
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toKotlinDuration

class L1DependentApp(
  private val configs: CoordinatorConfig,
  private val vertx: Vertx,
  private val l2Web3jClient: Web3j,
  private val httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory,
  proverClientV2: FileBasedExecutionProverClient,
  private val batchesRepository: BatchesRepository,
  private val blobsRepository: BlobsRepository,
  private val aggregationsRepository: AggregationsRepository,
  private val l1FeeHistoriesRepository: FeeHistoriesRepositoryWithCache,
  private val smartContractErrors: SmartContractErrors,
  private val metricsFacade: MetricsFacade
) : LongRunningService {
  private val log = LogManager.getLogger(this::class.java)

  init {
    if (configs.messageAnchoringService.disabled) {
      log.warn("Message anchoring service is disabled")
    }
    if (configs.dynamicGasPriceService.disabled) {
      log.warn("Dynamic gas price service is disabled")
    }
  }

  private val l2TransactionManager = createTransactionManager(
    vertx,
    configs.l2Signer,
    l2Web3jClient
  )

  private val l2MessageService = instantiateL2MessageServiceContractClient(
    configs.l2,
    l2TransactionManager,
    l2Web3jClient,
    smartContractErrors
  )

  private val l1Web3jClient = Web3j.build(
    HttpService(
      configs.l1.rpcEndpoint.toString(),
      okHttpClientBuilder(LogManager.getLogger("clients.l1")).build()
    ),
    1000,
    Async.defaultExecutorService()
  )

  private val l1Web3jService = Web3jBlobExtended(HttpService(configs.l1.ethFeeHistoryEndpoint.toString()))

  private val l2ZkGethWeb3jClient: Web3j =
    Web3j.build(
      HttpService(configs.zkGethTraces.ethApi.toString()),
      1000,
      Async.defaultExecutorService()
    )

  private val l2MessageServiceLogsClient = Web3JL2MessageServiceLogsClient(
    logsClient = Web3JLogsClient(vertx, l2Web3jClient),
    l2MessageServiceAddress = configs.l2.messageServiceAddress
  )

  private val l2ExtendedWeb3j = ExtendedWeb3JImpl(l2ZkGethWeb3jClient)

  private val finalizationTransactionManager = createTransactionManager(
    vertx,
    configs.finalizationSigner,
    l1Web3jClient
  )

  private val l1MinPriorityFeeCalculator: FeesCalculator = WMAFeesCalculator(
    WMAFeesCalculator.Config(
      BigDecimal("0.0"),
      BigDecimal.ONE
    )
  )

  private val l1DataSubmissionPriorityFeeCalculator: FeesCalculator = BoundableFeeCalculator(
    BoundableFeeCalculator.Config(
      feeUpperBound = configs.blobSubmission.priorityFeePerGasUpperBound,
      feeLowerBound = configs.blobSubmission.priorityFeePerGasLowerBound,
      feeMargin = BigInteger.ZERO
    ),
    l1MinPriorityFeeCalculator
  )

  private val feesFetcher: FeesFetcher = FeeHistoryFetcherImpl(
    l1Web3jClient,
    l1Web3jService,
    FeeHistoryFetcherImpl.Config(
      configs.l1.feeHistoryBlockCount.toUInt(),
      configs.l1.feeHistoryRewardPercentile
    )
  )

  private val l1FinalizationMonitor = run {
    // To avoid setDefaultBlockParameter clashes
    val zkEvmClientForFinalization: LineaRollupAsyncFriendly = instantiateZkEvmContractClient(
      configs.l1,
      finalizationTransactionManager,
      feesFetcher,
      l1MinPriorityFeeCalculator,
      l1Web3jClient,
      smartContractErrors
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
    )
  }

  private val gasPriceCapProvider = run {
    val feeHistoryPercentileWindowInBlocks =
      configs.l1DynamicGasPriceCapService.gasPriceCapCalculation.baseFeePerGasPercentileWindow
        .toKotlinDuration().inWholeSeconds.div(L1_BLOCK_TIME_SECONDS).toUInt()

    val feeHistoryPercentileWindowLeewayInBlocks =
      configs.l1DynamicGasPriceCapService.gasPriceCapCalculation.baseFeePerGasPercentileWindowLeeway
        .toKotlinDuration().inWholeSeconds.div(L1_BLOCK_TIME_SECONDS).toUInt()

    val l1GasPriceCapCalculator: GasPriceCapCalculator = GasPriceCapCalculatorImpl()

    GasPriceCapProviderImpl(
      config = GasPriceCapProviderImpl.Config(
        enabled = configs.l1DynamicGasPriceCapService.enabled,
        maxFeePerGasCap = configs.l1.maxFeePerGasCap,
        maxFeePerBlobGasCap = configs.l1.maxFeePerBlobGasCap,
        baseFeePerGasPercentile =
        configs.l1DynamicGasPriceCapService.gasPriceCapCalculation.baseFeePerGasPercentile,
        baseFeePerGasPercentileWindowInBlocks = feeHistoryPercentileWindowInBlocks,
        baseFeePerGasPercentileWindowLeewayInBlocks = feeHistoryPercentileWindowLeewayInBlocks,
        timeOfDayMultipliers =
        configs.l1DynamicGasPriceCapService.gasPriceCapCalculation.timeOfDayMultipliers!!,
        adjustmentConstant =
        configs.l1DynamicGasPriceCapService.gasPriceCapCalculation.adjustmentConstant,
        blobAdjustmentConstant =
        configs.l1DynamicGasPriceCapService.gasPriceCapCalculation.blobAdjustmentConstant,
        finalizationTargetMaxDelay =
        configs.l1DynamicGasPriceCapService.gasPriceCapCalculation.finalizationTargetMaxDelay.toKotlinDuration()
      ),
      l1Web3jClient = l1Web3jClient,
      l2ExtendedWeb3JClient = l2ExtendedWeb3j,
      feeHistoriesRepository = l1FeeHistoriesRepository,
      gasPriceCapCalculator = l1GasPriceCapCalculator
    )
  }

  private val gasPriceCapProviderForDataSubmission = GasPriceCapProviderForDataSubmission(
    gasPriceCapProvider = gasPriceCapProvider,
    metricsFacade = metricsFacade
  )

  private val gasPriceCapProviderForFinalization = GasPriceCapProviderForFinalization(
    config = GasPriceCapProviderForFinalization.Config(
      gasPriceCapMultiplier = configs.l1.gasPriceCapMultiplierForFinalization
    ),
    gasPriceCapProvider = gasPriceCapProvider,
    metricsFacade = metricsFacade
  )

  private fun createStateManagerClient(
    stateManagerConfig: StateManagerClientConfig,
    logger: Logger
  ): Type2StateManagerClient {
    return Type2StateManagerJsonRpcClient(
      vertx = vertx,
      rpcClient = httpJsonRpcClientFactory.createWithLoadBalancing(
        endpoints = stateManagerConfig.endpoints.toSet(),
        maxInflightRequestsPerClient = stateManagerConfig.requestLimitPerEndpoint,
        log = logger
      ),
      config = Type2StateManagerJsonRpcClient.Config(
        requestRetry = stateManagerConfig.requestRetryConfig,
        zkStateManagerVersion = stateManagerConfig.version
      ),
      retryConfig = stateManagerConfig.requestRetryConfig,
      log = logger
    )
  }

  private val lastFinalizedBlock = lastFinalizedBlock().get()
  private val lastProcessedBlockNumber = resumeConflationFrom(
    aggregationsRepository,
    lastFinalizedBlock
  ).get()

  private fun createDeadlineConflationCalculatorRunner(): DeadlineConflationCalculatorRunner {
    return DeadlineConflationCalculatorRunner(
      conflationDeadlineCheckInterval = configs.conflation.conflationDeadlineCheckInterval.toKotlinDuration(),
      delegate = ConflationCalculatorByTimeDeadline(
        config = ConflationCalculatorByTimeDeadline.Config(
          conflationDeadline = configs.conflation.conflationDeadline.toKotlinDuration(),
          conflationDeadlineLastBlockConfirmationDelay =
          configs.conflation.conflationDeadlineLastBlockConfirmationDelay.toKotlinDuration()
        ),
        lastBlockNumber = lastProcessedBlockNumber,
        clock = Clock.System,
        latestBlockProvider = GethCliqueSafeBlockProvider(
          l2ExtendedWeb3j,
          GethCliqueSafeBlockProvider.Config(configs.l2.blocksToFinalization.toLong())
        )
      )
    )
  }

  private val deadlineConflationCalculatorRunnerOld = createDeadlineConflationCalculatorRunner()
  private val deadlineConflationCalculatorRunnerNew = createDeadlineConflationCalculatorRunner()

  private fun addBlocksLimitCalculatorIfDefined(calculators: MutableList<ConflationCalculator>) {
    if (configs.conflation.blocksLimit != null) {
      calculators.add(
        ConflationCalculatorByBlockLimit(
          blockLimit = configs.conflation.blocksLimit.toUInt()
        )
      )
    }
  }

  private fun createCalculatorsForBlobsAndConflation(
    logger: Logger,
    compressedBlobCalculator: ConflationCalculatorByDataCompressed
  ): List<ConflationCalculator> {
    val calculators: MutableList<ConflationCalculator> =
      mutableListOf(
        ConflationCalculatorByExecutionTraces(
          tracesCountersLimit = configs.conflation.tracesLimits,
          log = logger
        ),
        compressedBlobCalculator
      )
    addBlocksLimitCalculatorIfDefined(calculators)
    return calculators
  }

  private val conflationCalculator: TracesConflationCalculator = run {
    val logger = LogManager.getLogger(GlobalBlockConflationCalculator::class.java)

    // To fail faster for JNA reasons
    val blobCompressor = GoBackedBlobCompressor.getInstance(configs.blobCompression.blobSizeLimit)

    val compressedBlobCalculator = ConflationCalculatorByDataCompressed(
      blobCompressor = blobCompressor
    )
    val globalCalculator = GlobalBlockConflationCalculator(
      lastBlockNumber = lastProcessedBlockNumber,
      syncCalculators = createCalculatorsForBlobsAndConflation(logger, compressedBlobCalculator),
      deferredTriggerConflationCalculators = listOf(deadlineConflationCalculatorRunnerNew),
      log = logger
    )

    GlobalBlobAwareConflationCalculator(
      conflationCalculator = globalCalculator,
      blobCalculator = compressedBlobCalculator,
      batchesLimit = configs.proofAggregation.aggregationProofsLimit.toUInt() - 1U
    )
  }
  private val conflationService: ConflationService =
    ConflationServiceImpl(calculator = conflationCalculator, metricsFacade = metricsFacade)

  private val zkStateClient: Type2StateManagerClient =
    createStateManagerClient(configs.stateManager, LogManager.getLogger("clients.StateManagerShomeiClient"))

  private val zkEvmClientForDataSubmission: LineaRollupAsyncFriendly = instantiateZkEvmContractClient(
    l1Config = configs.l1,
    transactionManager = createTransactionManager(
      vertx,
      configs.dataSubmissionSigner,
      l1Web3jClient
    ),
    gasFetcher = feesFetcher,
    priorityFeeCalculator = l1DataSubmissionPriorityFeeCalculator,
    client = l1Web3jClient,
    smartContractErrors = smartContractErrors
  )

  private val eip4844SwitchProvider = Eip4844SwitchProviderImpl(configs.eip4844SwitchL2BlockNumber)

  private val blobCompressionProofCoordinator = run {
    val blobCompressionProverClient = FileBasedBlobCompressionProverClient(
      config = FileBasedBlobCompressionProverClient.Config(
        requestFileDirectory = configs.blobCompression.prover.fsRequestsDirectory,
        responseFileDirectory = configs.blobCompression.prover.fsResponsesDirectory,
        inprogressProvingSuffixPattern = configs.blobCompression.prover.fsInprogessProvingSuffixPattern,
        inprogressRequestFileSuffix = configs.blobCompression.prover.fsInprogessRequestWritingSuffix,
        pollingInterval = configs.blobCompression.prover.fsPollingInterval.toKotlinDuration(),
        timeout = configs.blobCompression.prover.fsPollingTimeout.toKotlinDuration(),
        blobCalculatorVersion = configs.conflation.conflationCalculatorVersion,
        conflationCalculatorVersion = configs.conflation.conflationCalculatorVersion
      ),
      vertx = vertx
    )

    val maxProvenBlobCache = run {
      val highestProvenBlobTracker = HighestProvenBlobTracker(lastProcessedBlockNumber)
      metricsFacade.createGauge(
        category = LineaMetricsCategory.BLOB,
        name = "proven.highest.block.number",
        description = "Highest proven blob compression block number",
        measurementSupplier = highestProvenBlobTracker
      )
      highestProvenBlobTracker
    }
    val blobCompressionProofHandler: (BlobCompressionProofUpdate) -> SafeFuture<*> = SimpleCompositeSafeFutureHandler(
      listOf(
        maxProvenBlobCache
      )
    )

    val blobCompresionProofCoordinator = BlobCompressionProofCoordinator(
      vertx = vertx,
      blobsRepository = blobsRepository,
      blobCompressionProverClient = blobCompressionProverClient,
      rollingBlobShnarfCalculator = RollingBlobShnarfCalculator(
        blobShnarfCalculator = GoBackedBlobShnarfCalculator(),
        blobsRepository = blobsRepository,
        genesisShnarf = configs.blobSubmission.genesisShnarf
      ),
      blobZkStateProvider = BlobZkStateProviderImpl(
        zkStateClient = zkStateClient
      ),
      config = BlobCompressionProofCoordinator.Config(
        blobCalculatorVersion = configs.conflation.conflationCalculatorVersion,
        pollingInterval = configs.blobCompression.handlerPollingInterval.toKotlinDuration(),
        conflationCalculatorVersion = configs.conflation.conflationCalculatorVersion
      ),
      eip4844SwitchProvider = eip4844SwitchProvider,
      blobCompressionProofHandler = blobCompressionProofHandler,
      metricsFacade = metricsFacade
    )
    val highestUnprovenBlobTracker = HighestUnprovenBlobTracker(lastProcessedBlockNumber)
    metricsFacade.createGauge(
      category = LineaMetricsCategory.BLOB,
      name = "unproven.highest.block.number",
      description = "Block number of highest unproven blob produced",
      measurementSupplier = highestUnprovenBlobTracker
    )

    val compositeSafeFutureHandler = SimpleCompositeSafeFutureHandler(
      listOf(
        blobCompresionProofCoordinator::handleBlob,
        highestUnprovenBlobTracker
      )
    )
    conflationCalculator.onBlobCreation(compositeSafeFutureHandler)
    blobCompresionProofCoordinator
  }

  private val blobSubmissionCoordinator = run {
    val blobSubmitterAsCallData = BlobSubmitterAsCallData(zkEvmClientForDataSubmission)
    val blobSubmitterAsEIP4844 = BlobSubmitterAsEIP4844(
      contract = zkEvmClientForDataSubmission,
      gasPriceCapProvider = gasPriceCapProviderForDataSubmission
    )
    val eip4844SwitchAwareBlobSubmitter = Eip4844SwitchAwareBlobSubmitter(
      blobSubmitterAsCallData = blobSubmitterAsCallData,
      blobSubmitterAsEIP4844 = blobSubmitterAsEIP4844
    )

    BlobSubmissionCoordinatorImpl(
      config = BlobSubmissionCoordinatorImpl.Config(
        configs.blobSubmission.dbPollingInterval.toKotlinDuration(),
        configs.blobSubmission.proofSubmissionDelay.toKotlinDuration(),
        configs.blobSubmission.maxBlobsToSubmitPerTick.toUInt()
      ),
      blobSubmitter = eip4844SwitchAwareBlobSubmitter,
      blobsRepository = blobsRepository,
      lineaRollup = zkEvmClientForDataSubmission,
      vertx = vertx,
      clock = Clock.System
    )
  }

  private val proofAggregationCoordinatorService: LongRunningService = run {
    val proofAggregationClient = FileBasedProofAggregationClient(
      vertx = vertx,
      config = FileBasedProofAggregationClient.Config(
        requestFileDirectory = configs.proofAggregation.prover.fsRequestsDirectory,
        responseFileDirectory = configs.proofAggregation.prover.fsResponsesDirectory,
        responseFilePollingInterval = configs.proofAggregation.prover.fsPollingInterval.toKotlinDuration(),
        responseFileMonitorTimeout = configs.proofAggregation.prover.fsPollingTimeout.toKotlinDuration(),
        inProgressRequestFileSuffix = configs.proofAggregation.prover.fsInprogessRequestWritingSuffix,
        proverInProgressSuffixPattern = configs.proofAggregation.prover.fsInprogessProvingSuffixPattern
      )
    )
    // it needs it's own client because internally set the blockNumber when making queries.
    // it does not make any transaction
    val messageService = instantiateL2MessageServiceContractClient(
      configs.l2,
      l2TransactionManager,
      l2Web3jClient,
      smartContractErrors
    )
    val l2MessageServiceClient = Web3JL2MessageService(
      vertx = vertx,
      l2MessageServiceLogsClient = l2MessageServiceLogsClient,
      web3jL2MessageService = messageService
    )

    val maxBlobEndBlockNumberTracker = ConsecutiveProvenBlobsProviderWithLastEndBlockNumberTracker(
      aggregationsRepository,
      lastProcessedBlockNumber
    )

    metricsFacade.createGauge(
      category = LineaMetricsCategory.BLOB,
      name = "proven.highest.consecutive.block.number",
      description = "Highest consecutive proven blob compression block number",
      measurementSupplier = maxBlobEndBlockNumberTracker
    )

    val highestAggregationTracker = HighestAggregationTracker(lastProcessedBlockNumber)
    metricsFacade.createGauge(
      category = LineaMetricsCategory.AGGREGATION,
      name = "proven.highest.block.number",
      description = "Highest proven aggregation block number",
      measurementSupplier = highestAggregationTracker
    )

    ProofAggregationCoordinatorService.create(
      vertx = vertx,
      pollerInterval = configs.proofAggregation.pollingInterval.toKotlinDuration(),
      aggregationDeadline = configs.proofAggregation.aggregationDeadline.toKotlinDuration(),
      latestBlockProvider = GethCliqueSafeBlockProvider(
        l2ExtendedWeb3j,
        GethCliqueSafeBlockProvider.Config(configs.l2.blocksToFinalization.toLong())
      ),
      maxProofsPerAggregation = configs.proofAggregation.aggregationProofsLimit.toUInt(),
      startBlockNumberInclusive = lastFinalizedBlock + 1u,
      aggregationCalculatorVersion = configs.proofAggregation.aggregationCalculatorVersion,
      aggregationsRepository = aggregationsRepository,
      consecutiveProvenBlobsProvider = maxBlobEndBlockNumberTracker,
      proofAggregationClient = proofAggregationClient,
      l2web3jClient = l2Web3jClient,
      l2MessageServiceClient = l2MessageServiceClient,
      metricsFacade = metricsFacade,
      provenAggregationEndBlockNumberConsumer = { highestAggregationTracker(it) }
    )
  }

  private val aggregationFinalizationCoordinator = run {
    if (!configs.aggregationFinalization.enabled) {
      DisabledLongRunningService
    } else {
      val zkEvmClientForFinalization: LineaRollupAsyncFriendly = instantiateZkEvmContractClient(
        configs.l1,
        finalizationTransactionManager,
        feesFetcher,
        l1MinPriorityFeeCalculator,
        l1Web3jClient,
        smartContractErrors
      )
      val aggregationFinalization: AggregationFinalization = AggregationFinalizationAsCallData(
        contract = zkEvmClientForFinalization,
        gasPriceCapProvider = gasPriceCapProviderForFinalization
      )

      AggregationFinalizationCoordinator(
        config = AggregationFinalizationCoordinator.Config(
          configs.aggregationFinalization.dbPollingInterval.toKotlinDuration(),
          configs.aggregationFinalization.proofSubmissionDelay.toKotlinDuration(),
          configs.aggregationFinalization.maxAggregationsToFinalizePerTick.toUInt()
        ),
        aggregationFinalization = aggregationFinalization,
        aggregationsRepository = aggregationsRepository,
        lineaRollup = zkEvmClientForFinalization,
        vertx = vertx,
        clock = Clock.System
      )
    }
  }

  private val block2BatchCoordinator = run {
    val tracesFileManager =
      TracesFilesManager(
        vertx,
        TracesFilesManager.Config(
          configs.traces.fileManager.rawTracesDirectory,
          configs.traces.fileManager.nonCanonicalRawTracesDirectory,
          configs.traces.fileManager.pollingInterval.toKotlinDuration(),
          configs.traces.fileManager.tracesFileCreationWaitTimeout.toKotlinDuration(),
          configs.traces.rawExecutionTracesVersion,
          configs.traces.fileManager.tracesFileExtension,
          configs.traces.fileManager.createNonCanonicalDirectory
        )
      )

    val tracesCountersClient = run {
      val log = LogManager.getLogger("clients.TracesCounters")
      TracesGeneratorJsonRpcClientV1(
        vertx = vertx,
        rpcClient = httpJsonRpcClientFactory.createWithLoadBalancing(
          endpoints = configs.traces.counters.endpoints.toSet(),
          maxInflightRequestsPerClient = configs.traces.counters.requestLimitPerEndpoint,
          log = log
        ),
        config = TracesGeneratorJsonRpcClientV1.Config(
          rawExecutionTracesVersion = configs.traces.rawExecutionTracesVersion,
          expectedTracesApiVersion = configs.traces.expectedTracesApiVersion
        ),
        retryConfig = configs.traces.counters.requestRetryConfig,
        log = log
      )
    }
    val tracesConflationClient = run {
      val log = LogManager.getLogger("clients.TracesConflation")

      TracesGeneratorJsonRpcClientV1(
        vertx = vertx,
        rpcClient = httpJsonRpcClientFactory.createWithLoadBalancing(
          endpoints = configs.traces.conflation.endpoints.toSet(),
          maxInflightRequestsPerClient = configs.traces.conflation.requestLimitPerEndpoint,
          log = log
        ),
        config = TracesGeneratorJsonRpcClientV1.Config(
          rawExecutionTracesVersion = configs.traces.rawExecutionTracesVersion,
          expectedTracesApiVersion = configs.traces.expectedTracesApiVersion
        ),
        retryConfig = configs.traces.conflation.requestRetryConfig,
        log = log
      )
    }

    val blobsConflationHandler: (BlocksConflation) -> SafeFuture<*> = run {
      val maxProvenBatchCache = run {
        val highestProvenBatchTracker = HighestProvenBatchTracker(lastProcessedBlockNumber)
        metricsFacade.createGauge(
          category = LineaMetricsCategory.BATCH,
          name = "proven.highest.block.number",
          description = "Highest proven batch execution block number",
          measurementSupplier = highestProvenBatchTracker
        )
        highestProvenBatchTracker
      }

      val batchProofHandler = SimpleCompositeSafeFutureHandler(
        listOf(
          maxProvenBatchCache,
          BatchProofHandlerImpl(batchesRepository)::acceptNewBatch
        )
      )

      val proofGeneratingConflationHandlerImpl = ProofGeneratingConflationHandlerImpl(
        tracesProductionCoordinator = TracesConflationCoordinatorImpl(tracesConflationClient, zkStateClient),
        zkProofProductionCoordinator = ZkProofCreationCoordinatorImpl(
          executionProverClient = proverClientV2,
          config = ZkProofCreationCoordinatorImpl.Config(configs.conflation.conflationCalculatorVersion)
        ),
        batchProofHandler = batchProofHandler,
        vertx = vertx,
        config = ProofGeneratingConflationHandlerImpl.Config(5.seconds)
      )

      val highestConflationTracker = HighestConflationTracker(lastProcessedBlockNumber)
      metricsFacade.createGauge(
        category = LineaMetricsCategory.CONFLATION,
        name = "last.block.number",
        description = "Last conflated block number",
        measurementSupplier = highestConflationTracker
      )
      val conflationsCounter = metricsFacade.createCounter(
        category = LineaMetricsCategory.CONFLATION,
        name = "counter",
        description = "Counter of new conflations"
      )

      SimpleCompositeSafeFutureHandler(
        listOf(
          proofGeneratingConflationHandlerImpl::handleConflatedBatch,
          highestConflationTracker,
          {
            conflationsCounter.increment()
            SafeFuture.COMPLETE
          }
        )
      )
    }

    conflationService.onConflatedBatch(blobsConflationHandler)

    BlockToBatchSubmissionCoordinator(
      conflationService = conflationService,
      tracesFileManager = tracesFileManager,
      tracesCountersClient = tracesCountersClient,
      vertx = vertx,
      payloadEncoder = ExecutionPayloadV1RLPEncoderByBesuImplementation
    )
  }

  private val finalizedBlockNotifier = run {
    val log = LogManager.getLogger("clients.ForkChoiceUpdaterShomeiClient")
    val type2StateProofProviderClients = configs.type2StateProofProvider.endpoints.map {
      ShomeiClient(
        vertx = vertx,
        rpcClient = httpJsonRpcClientFactory.create(it, log = log),
        retryConfig = configs.type2StateProofProvider.requestRetryConfig,
        log = log
      )
    }

    ForkChoiceUpdaterImpl(type2StateProofProviderClients)
  }

  private val lastProvenBlockNumberProvider = run {
    val lastProvenConsecutiveBatchBlockNumberProvider = BatchesRepoBasedLastProvenBlockNumberProvider(
      lastProcessedBlockNumber.toLong(),
      batchesRepository
    )
    metricsFacade.createGauge(
      category = LineaMetricsCategory.BATCH,
      name = "proven.highest.consecutive.block.number",
      description = "Highest proven consecutive execution batch block number",
      measurementSupplier = { lastProvenConsecutiveBatchBlockNumberProvider.getLastKnownProvenBlockNumber() }
    )
    lastProvenConsecutiveBatchBlockNumberProvider
  }

  private val blockCreationMonitor = run {
    log.info("Resuming conflation from block={} inclusive", lastProcessedBlockNumber + 1UL)
    val blockCreationMonitor = BlockCreationMonitor(
      vertx = vertx,
      extendedWeb3j = l2ExtendedWeb3j,
      startingBlockNumberExclusive = lastProcessedBlockNumber.toLong(),
      blockCreationListener = block2BatchCoordinator,
      lastProvenBlockNumberProviderAsync = lastProvenBlockNumberProvider,
      config = BlockCreationMonitor.Config(
        configs.zkGethTraces.newBlockPollingInterval.toKotlinDuration(),
        configs.l2.blocksToFinalization.toLong(),
        configs.conflation.fetchBlocksLimit.toLong()
      )
    )
    blockCreationMonitor
  }

  private fun lastFinalizedBlock(): SafeFuture<ULong> {
    val zkEvmClient: LineaRollupAsyncFriendly = instantiateZkEvmContractClient(
      configs.l1,
      finalizationTransactionManager,
      feesFetcher,
      l1MinPriorityFeeCalculator,
      l1Web3jClient,
      smartContractErrors
    )
    val l1BasedLastFinalizedBlockProvider = L1BasedLastFinalizedBlockProvider(
      vertx,
      zkEvmClient,
      configs.conflation.consistentNumberOfBlocksOnL1ToWait.toUInt()
    )
    return l1BasedLastFinalizedBlockProvider.getLastFinalizedBlock()
  }

  private val messageAnchoringApp: L1toL2MessageAnchoringApp? =
    if (configs.messageAnchoringService.enabled) {
      L1toL2MessageAnchoringApp(
        vertx,
        L1toL2MessageAnchoringApp.Config(
          configs.l1,
          configs.l2,
          configs.finalizationSigner,
          configs.l2Signer,
          configs.messageAnchoringService
        ),
        l1Web3jClient,
        l2Web3jClient,
        smartContractErrors,
        l2MessageService,
        l2TransactionManager
      )
    } else {
      null
    }

  private val gasPriceUpdaterApp: GasPriceUpdaterApp? =
    if (configs.dynamicGasPriceService.enabled) {
      GasPriceUpdaterApp(
        vertx,
        httpJsonRpcClientFactory,
        l1Web3jClient,
        l1Web3jService,
        GasPriceUpdaterApp.Config(configs.dynamicGasPriceService)
      )
    } else {
      null
    }

  private val l1FeeHistoryCachingService: LongRunningService =
    if (configs.l1DynamicGasPriceCapService.enabled) {
      val feeHistoryPercentileWindowInBlocks =
        configs.l1DynamicGasPriceCapService.gasPriceCapCalculation.baseFeePerGasPercentileWindow
          .toKotlinDuration().inWholeSeconds.div(L1_BLOCK_TIME_SECONDS).toUInt()

      val feeHistoryStoragePeriodInBlocks =
        configs.l1DynamicGasPriceCapService.feeHistoryStorage.storagePeriod
          .toKotlinDuration().inWholeSeconds.div(L1_BLOCK_TIME_SECONDS).toUInt()

      val l1FeeHistoryWeb3jBlobExtClient = Web3jBlobExtended(
        HttpService(
          configs.l1DynamicGasPriceCapService.feeHistoryFetcher.endpoint?.toString()
            ?: configs.l1.ethFeeHistoryEndpoint.toString()
        )
      )

      val l1FeeHistoryFetcher: GasPriceCapFeeHistoryFetcher = GasPriceCapFeeHistoryFetcherImpl(
        web3jService = l1FeeHistoryWeb3jBlobExtClient,
        config = GasPriceCapFeeHistoryFetcherImpl.Config(
          maxBlockCount = configs.l1DynamicGasPriceCapService.feeHistoryFetcher.maxBlockCount,
          rewardPercentiles = configs.l1DynamicGasPriceCapService.feeHistoryFetcher.rewardPercentiles
        )
      )

      FeeHistoryCachingService(
        config = FeeHistoryCachingService.Config(
          pollingInterval =
          configs.l1DynamicGasPriceCapService.feeHistoryFetcher.fetchInterval.toKotlinDuration(),
          feeHistoryMaxBlockCount =
          configs.l1DynamicGasPriceCapService.feeHistoryFetcher.maxBlockCount,
          baseFeePerGasPercentile =
          configs.l1DynamicGasPriceCapService.gasPriceCapCalculation.baseFeePerGasPercentile,
          feeHistoryStoragePeriodInBlocks = feeHistoryStoragePeriodInBlocks,
          feeHistoryWindowInBlocks = feeHistoryPercentileWindowInBlocks
        ),
        vertx = vertx,
        web3jClient = l1Web3jClient,
        feeHistoryFetcher = l1FeeHistoryFetcher,
        feeHistoriesRepository = l1FeeHistoriesRepository
      )
    } else {
      DisabledLongRunningService
    }

  private val blockFinalizationHandlerMap = mapOf(
    "finalized records cleanup" to { update: FinalizationMonitor.FinalizationUpdate ->
      val batchesCleanup = batchesRepository.deleteBatchesUpToEndBlockNumber(update.blockNumber.toLong())
      // Subtract 1 from block number because we do not want to delete the last blob as BlobCompressionProofCoordinator
      // needs the shnarf from the previous blob
      val blobsCleanup = blobsRepository.deleteBlobsUpToEndBlockNumber(update.blockNumber - 1u)
      // Subtract 1 from block number because we do not want to delete the last aggregation
      val aggregationsCleanup = aggregationsRepository
        .deleteAggregationsUpToEndBlockNumber(update.blockNumber.toLong() - 1L)

      SafeFuture.allOf(batchesCleanup, blobsCleanup, aggregationsCleanup)
    },
    "type 2 state proof provider finalization updates" to {
      finalizedBlockNotifier.updateFinalizedBlock(
        BlockNumberAndHash(it.blockNumber, it.blockHash)
      )
    },
    "last_proven_block_provider" to { update: FinalizationMonitor.FinalizationUpdate ->
      lastProvenBlockNumberProvider.updateLatestL1FinalizedBlock(update.blockNumber.toLong())
    }
  )

  init {
    blockFinalizationHandlerMap.forEach { (handlerName, handler) ->
      l1FinalizationMonitor.addFinalizationHandler(handlerName, handler)
    }
  }

  private fun cleanupDbDataAfterConflationResumeFromBlock(resumeConflationFrom: ULong): SafeFuture<*> {
    val cleanupBatches = batchesRepository.deleteBatchesAfterBlockNumber(resumeConflationFrom.toLong())
    val cleanupBlobs = blobsRepository.deleteBlobsAfterBlockNumber(resumeConflationFrom)
    val cleanupAggregations = aggregationsRepository.deleteAggregationsAfterBlockNumber(resumeConflationFrom.toLong())

    return SafeFuture.allOf(cleanupBatches, cleanupBlobs, cleanupAggregations)
  }

  override fun start(): CompletableFuture<Unit> {
    return cleanupDbDataAfterConflationResumeFromBlock(lastProcessedBlockNumber)
      .thenCompose { l1FinalizationMonitor.start() }
      .thenCompose { blobSubmissionCoordinator.start() }
      .thenCompose { aggregationFinalizationCoordinator.start() }
      .thenCompose { proofAggregationCoordinatorService.start() }
      .thenCompose { messageAnchoringApp?.start() ?: SafeFuture.completedFuture(Unit) }
      .thenCompose { gasPriceUpdaterApp?.start() ?: SafeFuture.completedFuture(Unit) }
      .thenCompose { l1FeeHistoryCachingService.start() }
      .thenCompose { deadlineConflationCalculatorRunnerOld.start() }
      .thenCompose { deadlineConflationCalculatorRunnerNew.start() }
      .thenCompose { blockCreationMonitor.start() }
      .thenCompose { blobCompressionProofCoordinator.start() }
      .thenPeek {
        log.info("L1App started")
      }
  }

  override fun stop(): CompletableFuture<Unit> {
    return SafeFuture.allOf(
      l1FinalizationMonitor.stop(),
      blobSubmissionCoordinator.stop(),
      aggregationFinalizationCoordinator.stop(),
      proofAggregationCoordinatorService.stop(),
      messageAnchoringApp?.stop() ?: SafeFuture.completedFuture(Unit),
      gasPriceUpdaterApp?.stop() ?: SafeFuture.completedFuture(Unit),
      l1FeeHistoryCachingService.stop(),
      blockCreationMonitor.stop(),
      deadlineConflationCalculatorRunnerOld.stop(),
      deadlineConflationCalculatorRunnerNew.stop(),
      blobCompressionProofCoordinator.stop()
    )
      .thenCompose { SafeFuture.fromRunnable { l1Web3jClient.shutdown() } }
      .thenApply { log.info("L1App Stopped") }
  }

  companion object {
    fun resumeConflationFrom(
      aggregationsRepository: AggregationsRepository,
      lastFinalizedBlock: ULong
    ): SafeFuture<ULong> {
      return aggregationsRepository
        .findConsecutiveProvenBlobs(lastFinalizedBlock.toLong() + 1)
        .thenApply { blobAndBatchCounters ->
          if (blobAndBatchCounters.isNotEmpty()) {
            blobAndBatchCounters.last().blobCounters.endBlockNumber
          } else {
            lastFinalizedBlock
          }
        }
    }
  }
}

private object DisabledLongRunningService : LongRunningService {
  override fun start(): CompletableFuture<Unit> {
    return SafeFuture.completedFuture(Unit)
  }

  override fun stop(): CompletableFuture<Unit> {
    return SafeFuture.completedFuture(Unit)
  }
}
