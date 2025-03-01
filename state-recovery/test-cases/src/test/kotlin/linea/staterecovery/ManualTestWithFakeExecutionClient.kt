package linea.staterecovery

import build.linea.clients.StateManagerClientV1
import io.vertx.core.Vertx
import linea.domain.BlockNumberAndHash
import linea.domain.BlockParameter
import linea.domain.RetryConfig
import linea.log4j.configureLoggers
import linea.staterecovery.plugin.createAppClients
import linea.staterecovery.test.FakeExecutionLayerClient
import linea.staterecovery.test.FakeStateManagerClientReadFromL1
import net.consensys.linea.async.get
import net.consensys.linea.vertx.VertxFactory
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import java.net.URI
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

class TestRunner(
  private val vertx: Vertx = VertxFactory.createVertx(),
  private val l2RecoveryStartBlockNumber: ULong,
  private val debugForceSyncStopBlockNumber: ULong = ULong.MAX_VALUE
) {
  private val log = LogManager.getLogger("test.case.TestRunner")
  private val infuraAppKey = System.getenv("INFURA_PROJECT_ID")
    .also {
      require(it.isNotEmpty()) { "Please define INFURA_APP_KEY environment variable" }
    }
  private val l1RpcUrl = "https://sepolia.infura.io/v3/$infuraAppKey"
  private val blobScanUrl = "https://api.sepolia.blobscan.com/"
  val appConfig = StateRecoveryApp.Config(
    // l1EarliestSearchBlock = 7236630UL.toBlockParameter(),
    l1EarliestSearchBlock = BlockParameter.Tag.EARLIEST,
    l1LatestSearchBlock = BlockParameter.Tag.LATEST,
    l1PollingInterval = 5.seconds,
    executionClientPollingInterval = 1.seconds,
    smartContractAddress = StateRecoveryApp.Config.lineaSepolia.smartContractAddress,
    l1getLogsChunkSize = 10_000u,
    debugForceSyncStopBlockNumber = debugForceSyncStopBlockNumber
  )
  val appClients = createAppClients(
    vertx = vertx,
    smartContractAddress = appConfig.smartContractAddress,
    l1RpcEndpoint = URI(l1RpcUrl),
    l1SuccessBackoffDelay = 1.milliseconds,
    l1RequestRetryConfig = RetryConfig(backoffDelay = 1.seconds, maxRetries = 1u),
    blobScanEndpoint = URI(blobScanUrl),
    blobScanRequestRetryConfig = RetryConfig(backoffDelay = 10.milliseconds, timeout = 5.seconds),
    stateManagerClientEndpoint = URI("http://it-does-not-matter:5432")
  )
  private val fakeExecutionLayerClient = FakeExecutionLayerClient(
    headBlock = BlockNumberAndHash(number = l2RecoveryStartBlockNumber - 1UL, hash = ByteArray(32) { 0 }),
    initialStateRecoverStartBlockNumber = l2RecoveryStartBlockNumber,
    loggerName = "test.fake.clients.execution-layer"
  )
  var fakeStateManagerClient: StateManagerClientV1 = FakeStateManagerClientReadFromL1(
    headBlockNumber = ULong.MAX_VALUE,
    logsSearcher = appClients.ethLogsSearcher,
    contractAddress = StateRecoveryApp.Config.lineaSepolia.smartContractAddress
  )
  var stateRecoverApp: StateRecoveryApp = StateRecoveryApp(
    vertx = vertx,
    elClient = fakeExecutionLayerClient,
    blobFetcher = appClients.blobScanClient,
    ethLogsSearcher = appClients.ethLogsSearcher,
    stateManagerClient = fakeStateManagerClient,
    transactionDetailsClient = appClients.transactionDetailsClient,
    blockHeaderStaticFields = BlockHeaderStaticFields.localDev,
    lineaContractClient = appClients.lineaContractClient,
    config = appConfig
  )

  init {
    configureLoggers(
      rootLevel = Level.INFO,
      "linea.staterecovery" to Level.TRACE,
      "linea.plugin.staterecovery" to Level.DEBUG,
      "linea.plugin.staterecovery.clients.l1.logs-searcher" to Level.TRACE
    )
  }

  fun run(
    timeout: kotlin.time.Duration = 10.minutes
  ) {
    log.info("Running test case")
    stateRecoverApp.start().get()
    await()
      .atMost(timeout.toJavaDuration())
      .pollInterval(10.seconds.toJavaDuration())
      .untilAsserted {
        val updatedStatus = fakeExecutionLayerClient.lineaGetStateRecoveryStatus().get()
        assertThat(updatedStatus.headBlockNumber).isGreaterThan(debugForceSyncStopBlockNumber)
      }
    log.info("Test case finished")
    vertx.close().get()
  }
}

fun main() {
  TestRunner(
    l2RecoveryStartBlockNumber = 7313000UL,
    debugForceSyncStopBlockNumber = 7313050UL
  ).run(
    timeout = 10.minutes
  )
}
