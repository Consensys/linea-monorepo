package net.consensys.zkevm.coordinator.app.conflation

import build.linea.clients.StateManagerClientV1
import build.linea.clients.StateManagerV1JsonRpcClient
import io.vertx.core.Vertx
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import linea.blob.ShnarfCalculatorVersion
import linea.contract.l2.Web3JL2MessageServiceSmartContractClient
import linea.coordinator.config.toJsonRpcRetry
import linea.coordinator.config.v2.CoordinatorConfig
import linea.coordinator.config.v2.isDisabled
import linea.domain.BlockParameter
import linea.domain.RetryConfig
import linea.encoding.BlockRLPEncoder
import linea.web3j.ExtendedWeb3JImpl
import linea.web3j.createWeb3jHttpClient
import linea.web3j.ethapi.createEthApiClient
import net.consensys.linea.contract.l1.GenesisStateProvider
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.traces.TracesCountersV2
import net.consensys.zkevm.LongRunningService
import net.consensys.zkevm.coordinator.app.conflation.ConflationAppHelper.cleanupDbDataAfterBlockNumbers
import net.consensys.zkevm.coordinator.app.conflation.ConflationAppHelper.resumeAggregationFrom
import net.consensys.zkevm.coordinator.app.conflation.ConflationAppHelper.resumeConflationFrom
import net.consensys.zkevm.coordinator.app.conflation.TracesClientFactory.createTracesClients
import net.consensys.zkevm.coordinator.blockcreation.BatchesRepoBasedLastProvenBlockNumberProvider
import net.consensys.zkevm.coordinator.blockcreation.BlockCreationMonitor
import net.consensys.zkevm.coordinator.blockcreation.GethCliqueSafeBlockProvider
import net.consensys.zkevm.coordinator.clients.ExecutionProverClientV2
import net.consensys.zkevm.coordinator.clients.prover.ProverClientFactory
import net.consensys.zkevm.domain.Batch
import net.consensys.zkevm.domain.BlocksConflation
import net.consensys.zkevm.domain.ProofIndex
import net.consensys.zkevm.ethereum.coordination.HighestConflationTracker
import net.consensys.zkevm.ethereum.coordination.HighestProvenBatchTracker
import net.consensys.zkevm.ethereum.coordination.HighestProvenBlobTracker
import net.consensys.zkevm.ethereum.coordination.HighestULongTracker
import net.consensys.zkevm.ethereum.coordination.HighestUnprovenBlobTracker
import net.consensys.zkevm.ethereum.coordination.SimpleCompositeSafeFutureHandler
import net.consensys.zkevm.ethereum.coordination.aggregation.ConsecutiveProvenBlobsProviderWithLastEndBlockNumberTracker
import net.consensys.zkevm.ethereum.coordination.aggregation.ProofAggregationCoordinatorService
import net.consensys.zkevm.ethereum.coordination.blob.BlobCompressionProofCoordinator
import net.consensys.zkevm.ethereum.coordination.blob.BlobCompressionProofUpdate
import net.consensys.zkevm.ethereum.coordination.blob.BlobZkStateProviderImpl
import net.consensys.zkevm.ethereum.coordination.blob.GoBackedBlobCompressor
import net.consensys.zkevm.ethereum.coordination.blob.GoBackedBlobShnarfCalculator
import net.consensys.zkevm.ethereum.coordination.blob.RollingBlobShnarfCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.BlockToBatchSubmissionCoordinator
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorByBlockLimit
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorByDataCompressed
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorByExecutionTraces
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorByTargetBlockNumbers
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorByTimeDeadline
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationService
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationServiceImpl
import net.consensys.zkevm.ethereum.coordination.conflation.DeadlineConflationCalculatorRunner
import net.consensys.zkevm.ethereum.coordination.conflation.GlobalBlobAwareConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.GlobalBlockConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.ProofGeneratingConflationHandlerImpl
import net.consensys.zkevm.ethereum.coordination.conflation.TimestampHardForkConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.TracesConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.TracesConflationCoordinatorImpl
import net.consensys.zkevm.ethereum.coordination.proofcreation.ZkProofCreationCoordinatorImpl
import net.consensys.zkevm.persistence.AggregationsRepository
import net.consensys.zkevm.persistence.BatchesRepository
import net.consensys.zkevm.persistence.BlobsRepository
import net.consensys.zkevm.persistence.dao.batch.persistence.BatchProofHandlerImpl
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.CompletableFuture
import kotlin.time.Duration.Companion.seconds

class ConflationApp(
  val vertx: Vertx,
  val batchesRepository: BatchesRepository,
  val blobsRepository: BlobsRepository,
  val aggregationsRepository: AggregationsRepository,
  val lastFinalizedBlock: ULong,
  val configs: CoordinatorConfig,
  val metricsFacade: MetricsFacade,
  val httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory,
) : LongRunningService {
  private val log = LogManager.getLogger("conflation.app")

  private val lastProcessedBlockNumber = resumeConflationFrom(
    aggregationsRepository,
    lastFinalizedBlock,
  ).get()
  private val lastConsecutiveAggregatedBlockNumber = resumeAggregationFrom(
    aggregationsRepository,
    lastFinalizedBlock,
  ).get()

  val l2Web3jClient: Web3j = createWeb3jHttpClient(
    rpcUrl = configs.conflation.l2Endpoint.toString(),
    log = LogManager.getLogger("clients.l2.eth.conflation"),
  )

  private val extendedWeb3J = ExtendedWeb3JImpl(l2Web3jClient)
  private val lastProcessedBlock = extendedWeb3J.ethGetBlock(
    BlockParameter.fromNumber(lastProcessedBlockNumber),
  ).get()

  init {
    require(lastProcessedBlock != null) {
      "lastProcessedBlock=$lastProcessedBlock is null! Unable to instantiate conflation calculators!"
    }
  }

  private val lastProcessedTimestamp = Instant.fromEpochSeconds(lastProcessedBlock!!.timestamp.toLong())

  private val deadlineConflationCalculatorRunner = createDeadlineConflationCalculatorRunner(l2Web3jClient)

  private val conflationCalculator: TracesConflationCalculator = run {
    val logger = LogManager.getLogger(GlobalBlockConflationCalculator::class.java)

    // To fail faster for JNA reasons
    val blobCompressor = GoBackedBlobCompressor.Companion.getInstance(
      compressorVersion = configs.conflation.blobCompression.blobCompressorVersion,
      dataLimit = configs.conflation.blobCompression.blobSizeLimit,
      metricsFacade = metricsFacade,
    )

    val compressedBlobCalculator = ConflationCalculatorByDataCompressed(
      blobCompressor = blobCompressor,
    )
    val globalCalculator = GlobalBlockConflationCalculator(
      lastBlockNumber = lastProcessedBlockNumber,
      syncCalculators = createCalculatorsForBlobsAndConflation(logger, compressedBlobCalculator),
      deferredTriggerConflationCalculators = listOfNotNull(deadlineConflationCalculatorRunner),
      emptyTracesCounters = TracesCountersV2.Companion.EMPTY_TRACES_COUNT,
      log = logger,
    )

    val batchesLimit = configs.conflation.blobCompression.batchesLimit
      ?: (configs.conflation.proofAggregation.proofsLimit - 1U)
    GlobalBlobAwareConflationCalculator(
      conflationCalculator = globalCalculator,
      blobCalculator = compressedBlobCalculator,
      metricsFacade = metricsFacade,
      batchesLimit = batchesLimit,
    )
  }
  private val conflationService: ConflationService =
    ConflationServiceImpl(calculator = conflationCalculator, metricsFacade = metricsFacade)

  private val zkStateClient: StateManagerClientV1 = StateManagerV1JsonRpcClient.Companion.create(
    rpcClientFactory = httpJsonRpcClientFactory,
    endpoints = configs.stateManager.endpoints.map { it.toURI() },
    maxInflightRequestsPerClient = configs.stateManager.requestLimitPerEndpoint,
    requestRetry = configs.stateManager.requestRetries.toJsonRpcRetry(),
    requestTimeout = configs.stateManager.requestTimeout?.inWholeMilliseconds,
    zkStateManagerVersion = configs.stateManager.version,
    logger = LogManager.getLogger("clients.StateManagerShomeiClient"),
  )

  private val proverClientFactory = ProverClientFactory(
    vertx = vertx,
    config = configs.proversConfig,
    metricsFacade = metricsFacade,
  )

  private val blobCompressionProofCoordinator = run {
    val maxProvenBlobCache = run {
      val highestProvenBlobTracker = HighestProvenBlobTracker(lastProcessedBlockNumber)
      metricsFacade.createGauge(
        category = LineaMetricsCategory.BLOB,
        name = "proven.highest.block.number",
        description = "Highest proven blob compression block number",
        measurementSupplier = highestProvenBlobTracker,
      )
      highestProvenBlobTracker
    }
    val blobCompressionProofHandler: (BlobCompressionProofUpdate) -> SafeFuture<*> = SimpleCompositeSafeFutureHandler(
      listOf(
        maxProvenBlobCache,
      ),
    )
    val genesisStateProvider = GenesisStateProvider(
      stateRootHash = configs.protocol.genesis.genesisStateRootHash,
      shnarf = configs.protocol.genesis.genesisShnarf,
    )

    val blobCompressionProofCoordinator = BlobCompressionProofCoordinator(
      vertx = vertx,
      blobsRepository = blobsRepository,
      blobCompressionProverClient = proverClientFactory.blobCompressionProverClient(),
      rollingBlobShnarfCalculator = RollingBlobShnarfCalculator(
        blobShnarfCalculator = GoBackedBlobShnarfCalculator(
          version = ShnarfCalculatorVersion.V1_2,
          metricsFacade = metricsFacade,
        ),
        blobsRepository = blobsRepository,
        genesisShnarf = genesisStateProvider.shnarf,
      ),
      blobZkStateProvider = BlobZkStateProviderImpl(
        zkStateClient = zkStateClient,
      ),
      config = BlobCompressionProofCoordinator.Config(
        pollingInterval = configs.conflation.blobCompression.handlerPollingInterval,
      ),
      blobCompressionProofHandler = blobCompressionProofHandler,
      metricsFacade = metricsFacade,
    )
    val highestUnprovenBlobTracker = HighestUnprovenBlobTracker(lastProcessedBlockNumber)
    metricsFacade.createGauge(
      category = LineaMetricsCategory.BLOB,
      name = "unproven.highest.block.number",
      description = "Block number of highest unproven blob produced",
      measurementSupplier = highestUnprovenBlobTracker,
    )

    val compositeSafeFutureHandler = SimpleCompositeSafeFutureHandler(
      listOf(
        blobCompressionProofCoordinator::handleBlob,
        highestUnprovenBlobTracker,
      ),
    )
    conflationCalculator.onBlobCreation(compositeSafeFutureHandler)
    blobCompressionProofCoordinator
  }

  private val proofAggregationCoordinatorService: LongRunningService = run {
    val maxBlobEndBlockNumberTracker = ConsecutiveProvenBlobsProviderWithLastEndBlockNumberTracker(
      aggregationsRepository,
      lastProcessedBlockNumber,
    )

    metricsFacade.createGauge(
      category = LineaMetricsCategory.BLOB,
      name = "proven.highest.consecutive.block.number",
      description = "Highest consecutive proven blob compression block number",
      measurementSupplier = maxBlobEndBlockNumberTracker,
    )

    val highestAggregationTracker = HighestULongTracker(lastConsecutiveAggregatedBlockNumber)
    metricsFacade.createGauge(
      category = LineaMetricsCategory.AGGREGATION,
      name = "proven.highest.block.number",
      description = "Highest proven aggregation block number",
      measurementSupplier = highestAggregationTracker,
    )

    ProofAggregationCoordinatorService.Companion
      .create(
        vertx = vertx,
        aggregationCoordinatorPollingInterval = configs.conflation.proofAggregation.coordinatorPollingInterval,
        deadlineCheckInterval = configs.conflation.proofAggregation.deadlineCheckInterval,
        aggregationDeadline = configs.conflation.proofAggregation.deadline,
        latestBlockProvider = GethCliqueSafeBlockProvider(
          web3j = l2Web3jClient,
          config = GethCliqueSafeBlockProvider.Config(0),
        ),
        maxProofsPerAggregation = configs.conflation.proofAggregation.proofsLimit,
        startBlockNumberInclusive = lastConsecutiveAggregatedBlockNumber + 1u,
        aggregationsRepository = aggregationsRepository,
        consecutiveProvenBlobsProvider = maxBlobEndBlockNumberTracker,
        proofAggregationClient = proverClientFactory.proofAggregationProverClient(),
        l2EthApiClient = createEthApiClient(
          l2Web3jClient,
          requestRetryConfig = RetryConfig(
            backoffDelay = 1.seconds,
            failuresWarningThreshold = 3u,
          ),
          vertx = vertx,
        ),
        l2MessageService = Web3JL2MessageServiceSmartContractClient.createReadOnly(
          web3jClient = l2Web3jClient,
          contractAddress = configs.protocol.l2.contractAddress,
          smartContractErrors = configs.smartContractErrors,
          smartContractDeploymentBlockNumber = configs.protocol.l2.contractDeploymentBlockNumber?.getNumber(),
        ),
        aggregationDeadlineDelay = configs.conflation.conflationDeadlineLastBlockConfirmationDelay,
        targetEndBlockNumbers = configs.conflation.proofAggregation.targetEndBlocks ?: emptyList(),
        metricsFacade = metricsFacade,
        provenAggregationEndBlockNumberConsumer = { aggEndBlockNumber -> highestAggregationTracker(aggEndBlockNumber) },
        aggregationSizeMultipleOf = configs.conflation.proofAggregation.aggregationSizeMultipleOf,
        hardForkTimestamps = configs.conflation.proofAggregation.timestampBasedHardForks,
        initialTimestamp = lastProcessedTimestamp,
      )
  }

  private val block2BatchCoordinator = run {
    val (tracesCountersClient, tracesConflationClient) = createTracesClients(
      vertx = vertx,
      rpcClientFactory = httpJsonRpcClientFactory,
      configs = configs.traces,
    )

    val blobsConflationHandler: (BlocksConflation) -> SafeFuture<*> = run {
      val maxProvenBatchCache = run {
        val highestProvenBatchTracker = HighestProvenBatchTracker(lastProcessedBlockNumber)
        metricsFacade.createGauge(
          category = LineaMetricsCategory.BATCH,
          name = "proven.highest.block.number",
          description = "Highest proven batch execution block number",
          measurementSupplier = highestProvenBatchTracker,
        )
        highestProvenBatchTracker
      }

      val batchProofHandler = SimpleCompositeSafeFutureHandler(
        listOf(
          maxProvenBatchCache,
          BatchProofHandlerImpl(batchesRepository)::acceptNewBatch,
        ),
      )
      val executionProverClient: ExecutionProverClientV2 = proverClientFactory.executionProverClient(
        // we cannot use configs.traces.expectedTracesApiVersion because it breaks prover expected version pattern
        tracesVersion = "2.1.0",
        stateManagerVersion = configs.stateManager.version,
      )

      val proofGeneratingConflationHandlerImpl = ProofGeneratingConflationHandlerImpl(
        tracesProductionCoordinator = TracesConflationCoordinatorImpl(tracesConflationClient, zkStateClient),
        zkProofProductionCoordinator = ZkProofCreationCoordinatorImpl(
          executionProverClient = executionProverClient,
          l2EthApiClient = createEthApiClient(
            rpcUrl = configs.conflation.l2Endpoint.toString(),
            log = LogManager.getLogger("clients.l2.eth.conflation"),
            requestRetryConfig = configs.conflation.l2RequestRetries,
            vertx = vertx,
          ),
          messageServiceAddress = configs.protocol.l2.contractAddress,
        ),
        batchProofHandler = batchProofHandler,
        vertx = vertx,
        batchAlreadyProvenSupplier = { batch: Batch ->
          executionProverClient.isProofAlreadyDone(
            proofRequestId = ProofIndex(
              batch.startBlockNumber,
              batch.endBlockNumber,
            ),
          )
        },
        config = ProofGeneratingConflationHandlerImpl.Config(5.seconds),
      )

      val highestConflationTracker = HighestConflationTracker(lastProcessedBlockNumber)
      metricsFacade.createGauge(
        category = LineaMetricsCategory.CONFLATION,
        name = "last.block.number",
        description = "Last conflated block number",
        measurementSupplier = highestConflationTracker,
      )
      val conflationsCounter = metricsFacade.createCounter(
        category = LineaMetricsCategory.CONFLATION,
        name = "counter",
        description = "Counter of new conflations",
      )

      SimpleCompositeSafeFutureHandler(
        listOf(
          proofGeneratingConflationHandlerImpl::handleConflatedBatch,
          highestConflationTracker,
          {
            conflationsCounter.increment()
            SafeFuture.COMPLETE
          },
        ),
      )
    }

    conflationService.onConflatedBatch(blobsConflationHandler)

    BlockToBatchSubmissionCoordinator(
      conflationService = conflationService,
      tracesCountersClient = tracesCountersClient,
      vertx = vertx,
      encoder = BlockRLPEncoder,
    )
  }

  private val lastProvenBlockNumberProvider = run {
    val lastProvenConsecutiveBatchBlockNumberProvider = BatchesRepoBasedLastProvenBlockNumberProvider(
      lastProcessedBlockNumber.toLong(),
      batchesRepository,
    )
    metricsFacade.createGauge(
      category = LineaMetricsCategory.BATCH,
      name = "proven.highest.consecutive.block.number",
      description = "Highest proven consecutive execution batch block number",
      measurementSupplier = { lastProvenConsecutiveBatchBlockNumberProvider.getLastKnownProvenBlockNumber() },
    )
    lastProvenConsecutiveBatchBlockNumberProvider
  }

  private val blockCreationMonitor = run {
    log.info("Resuming conflation from block={} inclusive", lastProcessedBlockNumber + 1UL)
    val blockCreationMonitor = BlockCreationMonitor(
      vertx = vertx,
      web3j = extendedWeb3J,
      startingBlockNumberExclusive = lastProcessedBlockNumber.toLong(),
      blockCreationListener = block2BatchCoordinator,
      lastProvenBlockNumberProviderAsync = lastProvenBlockNumberProvider,
      config = BlockCreationMonitor.Config(
        pollingInterval = configs.conflation.blocksPollingInterval,
        blocksToFinalization = 0L,
        blocksFetchLimit = configs.conflation.l2FetchBlocksLimit.toLong(),
        // We need to add 1 to forceStopConflationAtBlockInclusive because conflation calculator requires
        // block_number = forceStopConflationAtBlockInclusive + 1 to trigger conflation at
        // forceStopConflationAtBlockInclusive
        lastL2BlockNumberToProcessInclusive = configs.conflation.forceStopConflationAtBlockInclusive?.inc(),
      ),
    )
    blockCreationMonitor
  }

  override fun start(): CompletableFuture<Unit> {
    return cleanupDbDataAfterBlockNumbers(
      lastProcessedBlockNumber = lastProcessedBlockNumber,
      lastConsecutiveAggregatedBlockNumber = lastConsecutiveAggregatedBlockNumber,
      batchesRepository = batchesRepository,
      blobsRepository = blobsRepository,
      aggregationsRepository = aggregationsRepository,
    )
      .thenCompose { proofAggregationCoordinatorService.start() }
      .thenCompose { deadlineConflationCalculatorRunner?.start() ?: SafeFuture.completedFuture(Unit) }
      .thenCompose { blockCreationMonitor.start() }
      .thenCompose { blobCompressionProofCoordinator.start() }
      .thenPeek {
        log.info("Conflation started")
      }
  }

  override fun stop(): CompletableFuture<Unit> {
    return SafeFuture.allOf(
      proofAggregationCoordinatorService.stop(),
      blockCreationMonitor.stop(),
      deadlineConflationCalculatorRunner?.stop() ?: SafeFuture.completedFuture(Unit),
      blobCompressionProofCoordinator.stop(),
    )
      .thenApply { log.info("Conflation Stopped") }
  }

  fun updateLatestL1FinalizedBlock(blockNumber: Long): SafeFuture<Unit> {
    return lastProvenBlockNumberProvider.updateLatestL1FinalizedBlock(blockNumber)
  }

  private fun createDeadlineConflationCalculatorRunner(
    l2Web3jClient: Web3j,
  ): DeadlineConflationCalculatorRunner? {
    if (configs.conflation.isDisabled() || configs.conflation.conflationDeadline == null) {
      log.info("Conflation deadline calculator is disabled")
      return null
    }

    return DeadlineConflationCalculatorRunner(
      conflationDeadlineCheckInterval = configs.conflation.conflationDeadlineCheckInterval,
      delegate = ConflationCalculatorByTimeDeadline(
        config = ConflationCalculatorByTimeDeadline.Config(
          conflationDeadline = configs.conflation.conflationDeadline,
          conflationDeadlineLastBlockConfirmationDelay =
          configs.conflation.conflationDeadlineLastBlockConfirmationDelay,
        ),
        lastBlockNumber = lastProcessedBlockNumber,
        clock = Clock.System,
        latestBlockProvider = GethCliqueSafeBlockProvider(
          l2Web3jClient,
          GethCliqueSafeBlockProvider.Config(blocksToFinalization = 0),
        ),
      ),
    )
  }

  private fun addBlocksLimitCalculatorIfDefined(calculators: MutableList<ConflationCalculator>) {
    if (configs.conflation.blocksLimit != null) {
      calculators.add(
        ConflationCalculatorByBlockLimit(
          blockLimit = configs.conflation.blocksLimit,
        ),
      )
    }
  }

  private fun addTargetEndBlockConflationCalculatorIfDefined(calculators: MutableList<ConflationCalculator>) {
    if (configs.conflation.proofAggregation.targetEndBlocks?.isNotEmpty() ?: false) {
      calculators.add(
        ConflationCalculatorByTargetBlockNumbers(
          targetEndBlockNumbers = configs.conflation.proofAggregation.targetEndBlocks.toSet(),
        ),
      )
    }
  }

  private fun addTimestampHardForkCalculatorIfDefined(
    calculators: MutableList<ConflationCalculator>,
  ) {
    if (configs.conflation.proofAggregation.timestampBasedHardForks.isNotEmpty()) {
      calculators.add(
        TimestampHardForkConflationCalculator(
          hardForkTimestamps = configs.conflation.proofAggregation.timestampBasedHardForks,
          initialTimestamp = lastProcessedTimestamp,
        ),
      )
      log.info(
        "Added timestamp-based hard fork calculator with {} timestamps, initialized at {}, timestamps={}",
        configs.conflation.proofAggregation.timestampBasedHardForks.size,
        lastProcessedTimestamp,
        configs.conflation.proofAggregation.timestampBasedHardForks,
      )
    }
  }

  private fun createCalculatorsForBlobsAndConflation(
    logger: Logger,
    compressedBlobCalculator: ConflationCalculatorByDataCompressed,
  ): List<ConflationCalculator> {
    val calculators: MutableList<ConflationCalculator> =
      mutableListOf(
        ConflationCalculatorByExecutionTraces(
          tracesCountersLimit = configs.conflation.tracesLimitsV2,
          emptyTracesCounters = TracesCountersV2.EMPTY_TRACES_COUNT,
          metricsFacade = metricsFacade,
          log = logger,
        ),
        compressedBlobCalculator,
      )
    addBlocksLimitCalculatorIfDefined(calculators)
    addTargetEndBlockConflationCalculatorIfDefined(calculators)
    addTimestampHardForkCalculatorIfDefined(calculators)
    return calculators
  }
}
