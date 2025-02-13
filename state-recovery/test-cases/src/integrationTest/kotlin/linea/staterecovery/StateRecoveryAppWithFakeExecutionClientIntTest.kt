package linea.staterecovery

import build.linea.contract.l1.LineaContractVersion
import build.linea.contract.l1.LineaRollupSmartContractClientReadOnly
import build.linea.contract.l1.Web3JLineaRollupSmartContractClientReadOnly
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import linea.domain.RetryConfig
import linea.log4j.configureLoggers
import linea.staterecovery.clients.VertxTransactionDetailsClient
import linea.staterecovery.clients.blobscan.BlobScanClient
import linea.staterecovery.test.FakeExecutionLayerClient
import linea.staterecovery.test.FakeStateManagerClient
import linea.staterecovery.test.FakeStateManagerClientBasedOnBlobsRecords
import linea.web3j.Web3JLogsSearcher
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.BlockParameter
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.metrics.micrometer.MicrometerMetricsFacade
import net.consensys.linea.testing.submission.AggregationAndBlobs
import net.consensys.linea.testing.submission.loadBlobsAndAggregationsSortedAndGrouped
import net.consensys.linea.testing.submission.submitBlobsAndAggregationsAndWaitExecution
import net.consensys.zkevm.coordinator.clients.smartcontract.LineaRollupSmartContractClient
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.MakeFileDelegatedContractsManager.connectToLineaRollupContract
import net.consensys.zkevm.ethereum.MakeFileDelegatedContractsManager.lineaRollupContractErrors
import net.consensys.zkevm.ethereum.Web3jClientManager
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import java.net.URI
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

@ExtendWith(VertxExtension::class)
class StateRecoveryAppWithFakeExecutionClientIntTest {
  private val log = LogManager.getLogger("test.case.StateRecoverAppWithFakeExecutionClientIntTest")
  private lateinit var stateRecoverApp: StateRecoveryApp
  private lateinit var aggregationsAndBlobs: List<AggregationAndBlobs>
  private lateinit var executionLayerClient: FakeExecutionLayerClient
  private lateinit var fakeStateManagerClient: FakeStateManagerClient
  private lateinit var transactionDetailsClient: TransactionDetailsClient
  private lateinit var lineaContractClient: LineaRollupSmartContractClientReadOnly

  private lateinit var contractClientForBlobSubmissions: LineaRollupSmartContractClient
  private lateinit var contractClientForAggregationSubmissions: LineaRollupSmartContractClient
  private lateinit var blobScanClient: BlobScanClient
  private lateinit var logsSearcher: Web3JLogsSearcher
  private lateinit var vertx: Vertx

  private val testDataDir = run {
    "testdata/coordinator/prover/v3"
  }

  private val l1RpcUrl = "http://localhost:8445"
  private val blobScanUrl = "http://localhost:4001"

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    this.vertx = vertx
    val jsonRpcFactory = VertxHttpJsonRpcClientFactory(
      vertx = vertx,
      metricsFacade = MicrometerMetricsFacade(SimpleMeterRegistry())
    )
    aggregationsAndBlobs = loadBlobsAndAggregationsSortedAndGrouped(
      blobsResponsesDir = "$testDataDir/compression/responses",
      aggregationsResponsesDir = "$testDataDir/aggregation/responses"
    )
    executionLayerClient = FakeExecutionLayerClient(
      headBlock = BlockNumberAndHash(number = 0uL, hash = ByteArray(32) { 0 }),
      initialStateRecoverStartBlockNumber = null,
      loggerName = "test.fake.clients.l1.fake-execution-layer"
    )
    fakeStateManagerClient =
      FakeStateManagerClientBasedOnBlobsRecords(blobRecords = aggregationsAndBlobs.flatMap { it.blobs })
    transactionDetailsClient = VertxTransactionDetailsClient.create(
      jsonRpcClientFactory = jsonRpcFactory,
      endpoint = URI(l1RpcUrl),
      retryConfig = RequestRetryConfig(
        backoffDelay = 10.milliseconds,
        timeout = 2.seconds
      ),
      logger = LogManager.getLogger("test.clients.l1.transaction-details")
    )

    val rollupDeploymentResult = ContractsManager.get()
      .deployLineaRollup(numberOfOperators = 2, contractVersion = LineaContractVersion.V6).get()

    lineaContractClient = Web3JLineaRollupSmartContractClientReadOnly(
      web3j = Web3jClientManager.buildL1Client(
        log = LogManager.getLogger("test.clients.l1.linea-contract"),
        requestResponseLogLevel = Level.INFO,
        failuresLogLevel = Level.WARN
      ),
      contractAddress = rollupDeploymentResult.contractAddress
    )
    this.logsSearcher = Web3JLogsSearcher(
      vertx = vertx,
      web3jClient = Web3jClientManager.buildL1Client(
        log = LogManager.getLogger("test.clients.l1.events-fetcher"),
        requestResponseLogLevel = Level.TRACE,
        failuresLogLevel = Level.WARN
      ),
      Web3JLogsSearcher.Config(
        backoffDelay = 1.milliseconds,
        requestRetryConfig = RetryConfig.noRetries
      ),
      log = LogManager.getLogger("test.clients.l1.events-fetcher")
    )

    contractClientForBlobSubmissions = rollupDeploymentResult.rollupOperatorClient
    contractClientForAggregationSubmissions = connectToLineaRollupContract(
      rollupDeploymentResult.contractAddress,
      rollupDeploymentResult.rollupOperators[1].txManager,
      smartContractErrors = lineaRollupContractErrors
    )
    this.blobScanClient = BlobScanClient.create(
      vertx = vertx,
      endpoint = URI(blobScanUrl),
      requestRetryConfig = RequestRetryConfig(
        backoffDelay = 10.milliseconds,
        timeout = 2.seconds
      ),
      responseLogMaxSize = 1000u,
      logger = LogManager.getLogger("test.clients.l1.blobscan")
    )

    instantiateStateRecoveryApp()

    configureLoggers(
      rootLevel = Level.INFO,
      log.name to Level.INFO,
      "linea.testing.submission" to Level.INFO,
      "net.consensys.linea.contract.Web3JContractAsyncHelper" to Level.WARN, // silence noisy gasPrice Caps logs
      "test.clients.l1.executionlayer" to Level.DEBUG,
      "test.clients.l1.web3j-default" to Level.INFO,
      "test.clients.l1.state-manager" to Level.INFO,
      "test.clients.l1.transaction-details" to Level.INFO,
      "test.clients.l1.linea-contract" to Level.INFO,
      "test.clients.l1.events-fetcher" to Level.INFO,
      "test.clients.l1.blobscan" to Level.INFO,
      "net.consensys.linea.contract.l1" to Level.INFO,
      "test.fake.clients.l1.fake-execution-layer" to Level.INFO
    )
  }

  fun instantiateStateRecoveryApp(
    debugForceSyncStopBlockNumber: ULong? = null
  ) {
    stateRecoverApp = StateRecoveryApp(
      vertx = vertx,
      elClient = executionLayerClient,
      blobFetcher = blobScanClient,
      ethLogsSearcher = logsSearcher,
      stateManagerClient = fakeStateManagerClient,
      transactionDetailsClient = transactionDetailsClient,
      blockHeaderStaticFields = BlockHeaderStaticFields.localDev,
      lineaContractClient = lineaContractClient,
      config = StateRecoveryApp.Config(
        l1LatestSearchBlock = BlockParameter.Tag.LATEST,
        l1PollingInterval = 10.milliseconds,
        executionClientPollingInterval = 1.seconds,
        smartContractAddress = lineaContractClient.getAddress(),
        debugForceSyncStopBlockNumber = debugForceSyncStopBlockNumber
      )
    )
  }

  private fun submitDataToL1ContactAndWaitExecution(
    aggregationsAndBlobs: List<AggregationAndBlobs> = this.aggregationsAndBlobs,
    blobChunksSize: Int = 6,
    waitTimeout: Duration = 4.minutes
  ) {
    submitBlobsAndAggregationsAndWaitExecution(
      contractClientForBlobSubmission = contractClientForBlobSubmissions,
      contractClientForAggregationSubmission = contractClientForAggregationSubmissions,
      aggregationsAndBlobs = aggregationsAndBlobs,
      blobChunksMaxSize = blobChunksSize,
      waitTimeout = waitTimeout,
      l1Web3jClient = Web3jClientManager.l1Client,
      log = log
    )
  }

  /*
  1. when state recovery disabled: enables with lastFinalizedBlock + 1
  1.1 no finalizations and start from genesis
  1.2 headBlockNumber > lastFinalizedBlock -> enable with headBlockNumber + 1
  1.3 headBlockNumber < lastFinalizedBlock -> enable with lastFinalizedBlock + 1

  2. when state recovery enabled:
  2.1 recoveryStartBlockNumber > headBlockNumber: pull for head block number until is reached and start recovery there
  2.2 recoveryStartBlockNumber <= headBlockNumber: resume recovery from headBlockNumber
  */
  @Test
  fun `when state recovery disabled and is starting from genesis`() {
    stateRecoverApp.start().get()
    submitDataToL1ContactAndWaitExecution()

    val lastAggregation = aggregationsAndBlobs.findLast { it.aggregation != null }!!.aggregation
    await()
      .atMost(1.minutes.toJavaDuration())
      .untilAsserted {
        assertThat(stateRecoverApp.lastSuccessfullyRecoveredFinalization?.event?.endBlockNumber)
          .isEqualTo(lastAggregation!!.endBlockNumber)
      }

    assertThat(executionLayerClient.lineaGetStateRecoveryStatus().get())
      .isEqualTo(
        StateRecoveryStatus(
          headBlockNumber = lastAggregation!!.endBlockNumber,
          stateRecoverStartBlockNumber = 1UL
        )
      )
  }

  @Test
  fun `when recovery is disabled and headBlock is before lastFinalizedBlock resumes from lastFinalizedBlock+1`() {
    val finalizationToResumeFrom = aggregationsAndBlobs.get(1).aggregation!!
    // assert that the finalization event to resume from has at least 1 middle block
    assertThat(finalizationToResumeFrom.endBlockNumber)
      .isGreaterThan(finalizationToResumeFrom.startBlockNumber + 1UL)
      .withFailMessage("finalizationEventToResumeFrom must at least 3 blocks for this test")

    val finalizationsBeforeCutOff = aggregationsAndBlobs
      .filter { it.aggregation != null }
      .filter { it.aggregation!!.endBlockNumber < finalizationToResumeFrom.startBlockNumber }

    val finalizationsAfterCutOff = aggregationsAndBlobs
      .filter { it.aggregation != null }
      .filter { it.aggregation!!.startBlockNumber >= finalizationToResumeFrom.startBlockNumber }

    log.debug(
      "finalizations={} finalizationToStartRecoveryFrom={}",
      aggregationsAndBlobs.map { it.aggregation?.intervalString() },
      finalizationToResumeFrom.intervalString()
    )

    submitDataToL1ContactAndWaitExecution(
      aggregationsAndBlobs = finalizationsBeforeCutOff
    )

    executionLayerClient.headBlock = BlockNumberAndHash(
      number = 1UL,
      hash = ByteArray(32) { 0 }
    )

    val lastFinalizedBlockNumber = finalizationsBeforeCutOff.last().aggregation!!.endBlockNumber
    val expectedStateRecoverStartBlockNumber = lastFinalizedBlockNumber + 1UL
    stateRecoverApp.start().get()

    await()
      .atMost(4.minutes.toJavaDuration())
      .pollInterval(1.seconds.toJavaDuration())
      .untilAsserted {
        assertThat(executionLayerClient.stateRecoverStatus).isEqualTo(
          StateRecoveryStatus(
            headBlockNumber = 1UL,
            stateRecoverStartBlockNumber = expectedStateRecoverStartBlockNumber
          )
        )
        log.info("stateRecoverStatus={}", executionLayerClient.stateRecoverStatus)
      }

    // simulate that execution client has synced up to the last finalized block through P2P network
    executionLayerClient.headBlock = BlockNumberAndHash(
      number = lastFinalizedBlockNumber,
      hash = ByteArray(32) { 0 }
    )

    // continue finalizing the rest of the aggregations
    submitDataToL1ContactAndWaitExecution(
      aggregationsAndBlobs = finalizationsAfterCutOff
    )

    val lastAggregation = aggregationsAndBlobs.findLast { it.aggregation != null }!!.aggregation
    await()
      .atMost(1.minutes.toJavaDuration())
      .pollInterval(1.seconds.toJavaDuration())
      .untilAsserted {
        assertThat(executionLayerClient.headBlock.number).isEqualTo(lastAggregation!!.endBlockNumber)
      }

    // assert it imports correct blocks
    val importedBlocks = executionLayerClient.importedBlockNumbersInRecoveryMode
    assertThat(importedBlocks.first()).isEqualTo(expectedStateRecoverStartBlockNumber)
    assertThat(importedBlocks.last()).isEqualTo(lastAggregation!!.endBlockNumber)
  }

  @Test
  fun `when node starts with headblock greater lastFinalizedBlock`() {
    val finalizationToResumeFrom = aggregationsAndBlobs.get(1).aggregation!!
    // assert that the finalization event to resume from has at least 1 middle block
    assertThat(finalizationToResumeFrom.endBlockNumber)
      .isGreaterThan(finalizationToResumeFrom.startBlockNumber + 1UL)
      .withFailMessage("finalizationEventToResumeFrom must at least 3 blocks for this test")

    val finalizationsBeforeCutOff = aggregationsAndBlobs
      .filter { it.aggregation != null }
      .filter { it.aggregation!!.endBlockNumber < finalizationToResumeFrom.startBlockNumber }

    val finalizationsAfterCutOff = aggregationsAndBlobs
      .filter { it.aggregation != null }
      .filter { it.aggregation!!.startBlockNumber >= finalizationToResumeFrom.startBlockNumber }

    log.debug(
      "finalizations={} finalizationToStartRecoveryFrom={}",
      aggregationsAndBlobs.map { it.aggregation?.intervalString() },
      finalizationToResumeFrom.intervalString()
    )

    submitDataToL1ContactAndWaitExecution(
      aggregationsAndBlobs = finalizationsBeforeCutOff
    )

    // set execution layer head block after latest finalization
    val headBlockNumberAtStart = finalizationsBeforeCutOff.last().aggregation!!.endBlockNumber + 1UL
    executionLayerClient.headBlock = BlockNumberAndHash(
      number = headBlockNumberAtStart,
      hash = ByteArray(32) { 0 }
    )

    stateRecoverApp.start().get()
    await()
      .atMost(2.minutes.toJavaDuration())
      .pollInterval(1.seconds.toJavaDuration())
      .untilAsserted {
        assertThat(executionLayerClient.stateRecoverStatus).isEqualTo(
          StateRecoveryStatus(
            headBlockNumber = headBlockNumberAtStart,
            stateRecoverStartBlockNumber = headBlockNumberAtStart + 1UL
          )
        )
        log.debug("stateRecoverStatus={}", executionLayerClient.stateRecoverStatus)
      }

    // continue finalizing the rest of the aggregations
    submitDataToL1ContactAndWaitExecution(
      aggregationsAndBlobs = finalizationsAfterCutOff
    )

    val lastAggregation = aggregationsAndBlobs.findLast { it.aggregation != null }!!.aggregation
    await()
      .atMost(2.minutes.toJavaDuration())
      .pollInterval(1.seconds.toJavaDuration())
      .untilAsserted {
        assertThat(executionLayerClient.stateRecoverStatus)
          .isEqualTo(
            StateRecoveryStatus(
              headBlockNumber = lastAggregation!!.endBlockNumber,
              stateRecoverStartBlockNumber = headBlockNumberAtStart + 1UL
            )
          )
      }
    // assert it does not try to import blocks behind the head block
    assertThat(executionLayerClient.importedBlockNumbersInRecoveryMode.minOrNull())
      .isEqualTo(headBlockNumberAtStart + 1UL)
  }

  @Test
  fun `should stop recovery as soon as stateroot mismatches`() {
    fakeStateManagerClient.setBlockStateRootHash(
      aggregationsAndBlobs[1].aggregation!!.endBlockNumber,
      ByteArray(32) { 1 }
    )
    log.debug(
      "aggregations={} forcedMismatchAggregation={}",
      aggregationsAndBlobs.map { it.aggregation?.intervalString() },
      aggregationsAndBlobs[1].aggregation!!.intervalString()
    )

    stateRecoverApp.start().get()
    submitDataToL1ContactAndWaitExecution()

    await()
      .atMost(1.minutes.toJavaDuration())
      .untilAsserted {
        assertThat(stateRecoverApp.stateRootMismatchFound).isTrue()
      }

    assertThat(executionLayerClient.headBlock.number)
      .isEqualTo(aggregationsAndBlobs[1].aggregation!!.endBlockNumber)
  }

  @Test
  fun `should stop synch at forceSyncStopBlockNumber`() {
    val debugForceSyncStopBlockNumber = aggregationsAndBlobs[2].aggregation!!.startBlockNumber
    instantiateStateRecoveryApp(debugForceSyncStopBlockNumber = debugForceSyncStopBlockNumber)
    log.debug("forceSyncStopBlockNumber={}", fakeStateManagerClient)

    stateRecoverApp.start().get()
    submitDataToL1ContactAndWaitExecution(waitTimeout = 1.minutes)

    await()
      .atMost(1.minutes.toJavaDuration())
      .untilAsserted {
        println(executionLayerClient.headBlock.number)
        assertThat(executionLayerClient.headBlock.number).isGreaterThanOrEqualTo(debugForceSyncStopBlockNumber)
      }

    assertThat(executionLayerClient.headBlock.number).isEqualTo(debugForceSyncStopBlockNumber)
  }
}
