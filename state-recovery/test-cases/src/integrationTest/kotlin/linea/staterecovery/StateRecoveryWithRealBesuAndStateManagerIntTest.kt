package linea.staterecovery

import build.linea.clients.StateManagerClientV1
import build.linea.clients.StateManagerV1JsonRpcClient
import build.linea.contract.l1.LineaContractVersion
import build.linea.domain.BlockInterval
import build.linea.domain.EthLogEvent
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import linea.domain.RetryConfig
import linea.log4j.configureLoggers
import linea.testing.CommandResult
import linea.testing.Runner
import linea.web3j.Web3JLogsSearcher
import linea.web3j.createWeb3jHttpClient
import linea.web3j.waitForTxReceipt
import net.consensys.gwei
import net.consensys.linea.BlockParameter
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.metrics.micrometer.MicrometerMetricsFacade
import net.consensys.linea.testing.submission.AggregationAndBlobs
import net.consensys.linea.testing.submission.loadBlobsAndAggregationsSortedAndGrouped
import net.consensys.linea.testing.submission.submitBlobsAndAggregationsAndWaitExecution
import net.consensys.toBigInteger
import net.consensys.toULong
import net.consensys.zkevm.coordinator.clients.smartcontract.LineaRollupSmartContractClient
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.L2AccountManager
import net.consensys.zkevm.ethereum.LineaRollupDeploymentResult
import net.consensys.zkevm.ethereum.MakeFileDelegatedContractsManager.connectToLineaRollupContract
import net.consensys.zkevm.ethereum.MakeFileDelegatedContractsManager.lineaRollupContractErrors
import net.consensys.zkevm.ethereum.Web3jClientManager
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Order
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.web3j.protocol.Web3j
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.net.URI
import java.util.concurrent.atomic.AtomicBoolean
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

// tests rely on same Besu/Shomei docker containers, so only one test can run at a time
// @Execution(ExecutionMode.SAME_THREAD)
@ExtendWith(VertxExtension::class)
class StateRecoveryWithRealBesuAndStateManagerIntTest {
  private val log = LogManager.getLogger("test.case.StateRecoverAppWithLocalStackIntTest")
  private lateinit var stateManagerClient: StateManagerClientV1
  private val testDataDir = "testdata/coordinator/prover/v3"
  private val aggregationsAndBlobs: List<AggregationAndBlobs> = loadBlobsAndAggregationsSortedAndGrouped(
    blobsResponsesDir = "$testDataDir/compression/responses",
    aggregationsResponsesDir = "$testDataDir/aggregation/responses"
  )
  private lateinit var rollupDeploymentResult: LineaRollupDeploymentResult
  private lateinit var contractClientForBlobSubmission: LineaRollupSmartContractClient
  private lateinit var contractClientForAggregationSubmission: LineaRollupSmartContractClient
  private val executionLayerUrl = "http://localhost:9145"
  private val stateManagerUrl = "http://localhost:8890"

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    val jsonRpcFactory = VertxHttpJsonRpcClientFactory(
      vertx = vertx,
      metricsFacade = MicrometerMetricsFacade(SimpleMeterRegistry())
    )

    stateManagerClient = StateManagerV1JsonRpcClient.create(
      rpcClientFactory = jsonRpcFactory,
      endpoints = listOf(URI(stateManagerUrl)),
      maxInflightRequestsPerClient = 1U,
      requestRetry = RequestRetryConfig(
        backoffDelay = 10.milliseconds,
        timeout = 2.seconds
      ),
      zkStateManagerVersion = "2.3.0",
      logger = LogManager.getLogger("test.clients.l1.state-manager")
    )

    configureLoggers(
      rootLevel = Level.INFO,
      log.name to Level.DEBUG,
      "net.consensys.linea.contract.Web3JContractAsyncHelper" to Level.WARN,
      "test.clients.l1.executionlayer" to Level.INFO,
      "test.clients.l1.web3j-default" to Level.INFO,
      "test.clients.l1.state-manager" to Level.DEBUG,
      "test.clients.l1.transaction-details" to Level.INFO,
      "test.clients.l1.linea-contract" to Level.INFO,
      "test.clients.l1.events-fetcher" to Level.INFO,
      "test.clients.l1.blobscan" to Level.INFO,
      "net.consensys.linea.contract.l1" to Level.INFO
    )
  }

  fun freshStartOfStack() {
  }

  @Test
  @Order(1)
  fun `should recover status from genesis`() {
    Runner.executeCommandFailOnNonZeroExitCode(
      "make fresh-start-all-staterecovery"
    ).get()
    log.debug("stack restarted")

    this.rollupDeploymentResult = ContractsManager.get()
      .deployLineaRollup(numberOfOperators = 2, contractVersion = LineaContractVersion.V6)
      .get()
    log.info("LineaRollup address={}", rollupDeploymentResult.contractAddress)
    contractClientForBlobSubmission = rollupDeploymentResult.rollupOperatorClient
    contractClientForAggregationSubmission = connectToLineaRollupContract(
      rollupDeploymentResult.contractAddress,
      // index 0 is the first operator in rollupOperatorClient
      rollupDeploymentResult.rollupOperators[1].txManager,
      smartContractErrors = lineaRollupContractErrors
    )
    log.info("starting stack for recovery of state pushed to L1")
    val staterecoveryNodesStartFuture = execCommandAndAssertSuccess(
      "make staterecovery-replay-from-block " +
        "L1_ROLLUP_CONTRACT_ADDRESS=${rollupDeploymentResult.contractAddress} " +
        "STATERECOVERY_OVERRIDE_START_BLOCK_NUMBER=1"
    )
    val lastAggregationAndBlobs = aggregationsAndBlobs.findLast { it.aggregation != null }!!
    val lastAggregation = lastAggregationAndBlobs.aggregation!!

    val blobsSubmissionFuture = SafeFuture.supplyAsync {
      submitBlobsAndAggregationsAndWaitExecution(
        contractClientForBlobSubmission = contractClientForBlobSubmission,
        contractClientForAggregationSubmission = contractClientForAggregationSubmission,
        aggregationsAndBlobs = aggregationsAndBlobs,
        blobChunksSize = 6,
        l1Web3jClient = Web3jClientManager.l1Client,
        waitTimeout = 4.minutes
      )
      log.info("finalization={} executed on l1", lastAggregation.intervalString())
    }
    SafeFuture.allOf(staterecoveryNodesStartFuture, blobsSubmissionFuture).get()

    val web3jElClient = createWeb3jHttpClient(executionLayerUrl)

    // wait for Besu to be up and running
    waitExecutionLayerToBeUpAndRunning(web3jElClient)

    assertBesuAndShomeiStateRootMatches(web3jElClient, stateManagerClient, lastAggregationAndBlobs)
  }

  private fun execCommandAndAssertSuccess(
    command: String
  ): SafeFuture<CommandResult> {
    return Runner
      .executeCommandFailOnNonZeroExitCode(command, log = log)
      .thenPeek { execResult ->
        log.debug("STDOUT: {}", execResult.stdOutStr)
        log.debug("STDERR: {}", execResult.stdErrStr)
        assertThat(execResult.isSuccess).isTrue()
      }
  }

  @Test
  @Order(2)
  fun `should recover from middle of chain and be resilient to node restarts`(
    vertx: Vertx
  ) {
    // Part A:
    // we shall have multiple finalizations on L1
    // restart FRESH (empty state) Besu & Shomei with recovery block somewhere in the middle of those finalizations
    // will partially sync through P2P network
    // then will trigger recovery mode and sync the remaining blocks
    // assert Besu and Shomei are in sync
    // Part B:
    // send some txs to L2 to trigger coordinator to finalize on L1
    // wait for at least 1 more finalization on L1
    // assert Besu and Shomei are in sync
    // Part C:
    // restart zkbesu-node only
    // send some txs to L2 to trigger coordinator to finalize on L1
    // wait for at least 1 more finalization on L1
    // assert Besu and Shomei are in sync

    Runner.executeCommandFailOnNonZeroExitCode(
      "make fresh-start-all-staterecovery"
    ).get()
    log.debug("stack restarted")

    val localStackL1ContractAddress = "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9"
    val logsSearcher = Web3JLogsSearcher(
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
    val web3jElClient = createWeb3jHttpClient(executionLayerUrl)
    log.info("starting: besuBlock={}", web3jElClient.ethBlockNumber().send().blockNumber)
    // generate some activity on L2 for coordinator to finalize on L1
    val keepSendingTxToL2 = AtomicBoolean(true)
    sendTxToL2(keepSendingTxToL2::get)

    // A
    // await for 3 finalizations to happen on L1
    val lastFinalizationA = run {
      var finalizationLogs: List<EthLogEvent<DataFinalizedV3>> = emptyList()
      await()
        .atMost(2.minutes.toJavaDuration())
        .untilAsserted {
          finalizationLogs = getFinalizationsOnL1(logsSearcher, localStackL1ContractAddress)
          assertThat(finalizationLogs.size).isGreaterThan(2)
        }
      finalizationLogs.last()
    }

    log.info("lastFinalizationA={}", lastFinalizationA.event.intervalString())

    // await for coordinator to finalize on L1 at least once
    val stateRecoveryStartBlockNumber = lastFinalizationA.event.startBlockNumber - 2UL

    log.info("starting stack for recovery of state pushed to L1")
    execCommandAndAssertSuccess(
      "make staterecovery-replay-from-block " +
        "L1_ROLLUP_CONTRACT_ADDRESS=$localStackL1ContractAddress " +
        "STATERECOVERY_OVERRIDE_START_BLOCK_NUMBER=$stateRecoveryStartBlockNumber"
    ).get()

    // wait for Besu to be up and running
    waitExecutionLayerToBeUpAndRunning(web3jElClient)
    // assert besu and shomei could sync through P2P network
    assertBesuAndShomeiStateRootMatches(
      web3jElClient,
      stateManagerClient,
      lastFinalizationA.event.endBlockNumber,
      lastFinalizationA.event.finalStateRootHash
    )

    // B
    // await for coordinator to finalize on L1 again
    await()
      .atMost(2.minutes.toJavaDuration())
      .untilAsserted {
        getLastFinalizationOnL1(logsSearcher, localStackL1ContractAddress)
          .also { finalization ->
            assertThat(finalization.event.startBlockNumber)
              .isGreaterThan(lastFinalizationA.event.endBlockNumber)
          }
      }
    val lastFinalizationB = getLastFinalizationOnL1(logsSearcher, localStackL1ContractAddress)
    log.info("lastFinalizationB={}", lastFinalizationB.event.intervalString())
    assertBesuAndShomeiStateRootMatches(
      web3jElClient,
      stateManagerClient,
      lastFinalizationB.event.endBlockNumber,
      lastFinalizationB.event.finalStateRootHash
    )

    // C.
    // restart besu node with non-graceful shutdown
    log.info("Restarting zkbesu-shomei node")
    execCommandAndAssertSuccess("docker restart -s 9 zkbesu-shomei-sr").get()

    waitExecutionLayerToBeUpAndRunning(web3jElClient)
    val lastRecoveredBlock = web3jElClient.ethBlockNumber().send().blockNumber.toULong()
    // await coordinator to finalize on L1, beyond what's already in sync
    await()
      .atMost(2.minutes.toJavaDuration())
      .untilAsserted {
        getLastFinalizationOnL1(logsSearcher, localStackL1ContractAddress)
          .also { lastFinalizationC ->
            assertThat(lastFinalizationC.event.startBlockNumber)
              .isGreaterThan(lastFinalizationB.event.endBlockNumber)
            assertThat(lastFinalizationC.event.endBlockNumber).isGreaterThan(lastRecoveredBlock)
          }
      }
    keepSendingTxToL2.set(false)

    val lastFinalizationC = getLastFinalizationOnL1(logsSearcher, localStackL1ContractAddress)
    log.info("lastFinalizationC={}", lastFinalizationC.event.intervalString())

    assertBesuAndShomeiStateRootMatches(
      web3jElClient,
      stateManagerClient,
      lastFinalizationC.event.endBlockNumber,
      lastFinalizationC.event.finalStateRootHash
    )
  }

  private fun waitExecutionLayerToBeUpAndRunning(
    web3jElClient: Web3j,
    expectedHeadBlockNumber: ULong = 0UL
  ) {
    await()
      .pollInterval(1.seconds.toJavaDuration())
      .atMost(5.minutes.toJavaDuration())
      .untilAsserted {
        kotlin.runCatching {
          assertThat(web3jElClient.ethBlockNumber().send().blockNumber.toULong())
            .isGreaterThanOrEqualTo(expectedHeadBlockNumber)
        }.getOrElse {
          log.info("waiting for Besu to start, trying to connect to $executionLayerUrl")
          throw AssertionError("could not connect to $executionLayerUrl", it)
        }
      }
  }

  private fun assertBesuAndShomeiStateRootMatches(
    web3jElClient: Web3j,
    stateManagerClient: StateManagerClientV1,
    targetAggregationAndBlobs: AggregationAndBlobs
  ) {
    val targetAggregation = targetAggregationAndBlobs.aggregation!!
    val expectedZkEndStateRootHash = targetAggregationAndBlobs.blobs.last().blobCompressionProof!!.finalStateRootHash

    assertBesuAndShomeiStateRootMatches(
      web3jElClient,
      stateManagerClient,
      targetAggregation.endBlockNumber,
      expectedZkEndStateRootHash
    )
  }

  private fun assertBesuAndShomeiStateRootMatches(
    web3jElClient: Web3j,
    stateManagerClient: StateManagerClientV1,
    expectedBlockNumber: ULong,
    expectedZkEndStateRootHash: ByteArray
  ) {
    await()
      .pollInterval(1.seconds.toJavaDuration())
      .atMost(5.minutes.toJavaDuration())
      .untilAsserted {
        assertThat(web3jElClient.ethBlockNumber().send().blockNumber.toULong())
          .isGreaterThanOrEqualTo(expectedBlockNumber)
        val blockInterval = BlockInterval(expectedBlockNumber, expectedBlockNumber)
        assertThat(stateManagerClient.rollupGetStateMerkleProof(blockInterval).get().zkEndStateRootHash)
          .isEqualTo(expectedZkEndStateRootHash)
      }
  }

  private fun getLastFinalizationOnL1(
    logsSearcher: Web3JLogsSearcher,
    contractAddress: String
  ): EthLogEvent<DataFinalizedV3> {
    return getFinalizationsOnL1(logsSearcher, contractAddress)
      .lastOrNull()
      ?: error("no finalization found")
  }

  private fun getFinalizationsOnL1(
    logsSearcher: Web3JLogsSearcher,
    contractAddress: String
  ): List<EthLogEvent<DataFinalizedV3>> {
    return logsSearcher.getLogs(
      fromBlock = BlockParameter.Tag.EARLIEST,
      toBlock = BlockParameter.Tag.LATEST,
      address = contractAddress,
      topics = listOf(DataFinalizedV3.topic)
    ).get().map(DataFinalizedV3::fromEthLog)
  }

  private fun sendTxToL2(
    keepSendingPredicate: () -> Boolean
  ) {
    val account = L2AccountManager.generateAccount()
    val txManager = L2AccountManager.getTransactionManager(account)
    Thread {
      while (keepSendingPredicate()) {
        val txHash = txManager.sendTransaction(
          /*gasPrice*/ 150UL.gwei.toBigInteger(),
          /*gasLimit*/ 25_000UL.toBigInteger(),
          /*to*/ account.address,
          /*data*/ "",
          /*value*/ 1UL.toBigInteger()
        ).transactionHash
        log.trace("sent tx to L2, txHash={}", txHash)
        Web3jClientManager.l2Client.waitForTxReceipt(
          txHash = txHash,
          timeout = 5.seconds,
          pollingInterval = 500.milliseconds
        )
      }
    }.start()
  }
}
