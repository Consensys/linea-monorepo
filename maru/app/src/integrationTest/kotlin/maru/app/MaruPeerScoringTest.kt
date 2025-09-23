/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import java.lang.Thread.sleep
import java.net.ServerSocket
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.launch
import linea.domain.BlockParameter
import linea.ethapi.EthApiClient
import linea.web3j.ethapi.createEthApiClient
import maru.p2p.messages.BlockRetrievalStrategy
import maru.p2p.messages.DefaultBlockRetrievalStrategy
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.kotlin.await
import org.hyperledger.besu.tests.acceptance.dsl.blockchain.Amount
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions
import org.hyperledger.besu.tests.acceptance.dsl.node.ThreadBesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Test
import testutils.FourEmptyResponsesStrategy
import testutils.MisbehavingP2PNetwork
import testutils.PeeringNodeNetworkStack
import testutils.TimeOutResponsesStrategy
import testutils.besu.BesuFactory
import testutils.besu.BesuTransactionsHelper
import testutils.besu.ethGetBlockByNumber
import testutils.maru.MaruFactory
import testutils.maru.awaitTillMaruHasPeers

class MaruPeerScoringTest {
  private lateinit var cluster: Cluster
  private lateinit var validatorStack: PeeringNodeNetworkStack
  private lateinit var followerStack: PeeringNodeNetworkStack
  private lateinit var transactionsHelper: BesuTransactionsHelper
  private val log = LogManager.getLogger(this.javaClass)
  private val maruFactory = MaruFactory()
  private lateinit var fakeLineaContract: FakeLineaRollupSmartContractClient
  private lateinit var validatorEthApiClient: EthApiClient
  private lateinit var followerEthApiClient: EthApiClient

  @AfterEach
  fun tearDown() {
    followerStack.maruApp.stop()
    validatorStack.maruApp.stop()
    followerStack.maruApp.close()
    validatorStack.maruApp.close()
    cluster.close()
  }

  @Test
  fun `node gets in sync with default block retrieval strategy`() {
    val maruNodeSetup =
      setUpNodes(blockRetrievalStrategy = DefaultBlockRetrievalStrategy())

    try {
      await
        .atMost(20.seconds.toJavaDuration())
        .pollInterval(200.milliseconds.toJavaDuration())
        .ignoreExceptions()
        .untilAsserted {
          assertThat(
            followerEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.LATEST).get().number,
          ).isGreaterThanOrEqualTo(15UL)
        }
    } finally {
      maruNodeSetup.job.cancel()
    }
  }

  @Test
  fun `node disconnects validator when receiving too many empty responses`() {
    val maruNodeSetup =
      setUpNodes(
        blockRetrievalStrategy = FourEmptyResponsesStrategy(),
        banPeriod = 10.seconds,
        followerCooldownPeriod = 10.minutes,
      )

    try {
      // In setUpNodes we have made sure that the validator and the follower have 1 peer
      // Now wait until it is disconnected because of empty responses
      await
        .atMost(2.seconds.toJavaDuration())
        .pollInterval(250.milliseconds.toJavaDuration())
        .ignoreExceptions()
        .untilAsserted {
          assertThat(
            maruNodeSetup.followerMaruApp.p2pNetwork.peerCount == 0,
          )
        }
      // reconnects after ban period and finishes syncing
      await
        .atMost(20.seconds.toJavaDuration())
        .pollInterval(200.milliseconds.toJavaDuration())
        .ignoreExceptions()
        .untilAsserted {
          assertThat(
            followerEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.LATEST).get().number,
          ).isGreaterThanOrEqualTo(18UL)
        }
    } finally {
      maruNodeSetup.job.cancel()
    }
  }

  @Test
  fun `node disconnects validator when requests time out`() {
    val timeout = 3.seconds
    val delay = timeout + 2.seconds
    val maruNodeSetup =
      setUpNodes(
        blockRangeRequestTimeout = timeout,
        blockRetrievalStrategy = TimeOutResponsesStrategy(delay = delay),
        validatorCooldownPeriod = 20.seconds,
      )
    try {
      sleep((delay - 1.seconds).inWholeMilliseconds)
      assertThat(maruNodeSetup.followerMaruApp.p2pNetwork.peerCount == 1)

      await.untilAsserted {
        assertThat(maruNodeSetup.followerMaruApp.p2pNetwork.peerCount == 0)
      }
    } finally {
      maruNodeSetup.job.cancel()
    }
  }

  data class MaruNodeSetup(
    val validatorMaruApp: MaruApp,
    val followerMaruApp: MaruApp,
    val job: Job,
  )

  fun setUpNodes(
    blockRetrievalStrategy: BlockRetrievalStrategy,
    banPeriod: Duration = 10.seconds,
    validatorCooldownPeriod: Duration = 10.seconds,
    followerCooldownPeriod: Duration = 10.seconds,
    blockRangeRequestTimeout: Duration = 5.seconds,
  ): MaruNodeSetup {
    fakeLineaContract = FakeLineaRollupSmartContractClient()
    transactionsHelper = BesuTransactionsHelper()
    cluster =
      Cluster(
        ClusterConfigurationBuilder().build(),
        NetConditions(NetTransactions()),
        ThreadBesuNodeRunner(),
      )

    validatorStack = PeeringNodeNetworkStack()
    followerStack =
      PeeringNodeNetworkStack(
        besuBuilder = { BesuFactory.buildTestBesu(validator = false) },
      )

    PeeringNodeNetworkStack.startBesuNodes(cluster, validatorStack, followerStack)

    val tcpPort = findFreePort()
    val udpPort = findFreePort()
    val validatorMaruApp =
      maruFactory.buildTestMaruValidatorWithDiscovery(
        ethereumJsonRpcUrl = validatorStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = validatorStack.besuNode.engineRpcUrl().get(),
        dataDir = validatorStack.tmpDir,
        overridingLineaContractClient = fakeLineaContract,
        p2pPort = tcpPort,
        discoveryPort = udpPort,
        cooldownPeriod = validatorCooldownPeriod,
        p2pNetworkFactory = {
          privateKeyBytes,
          p2pConfig,
          chainId,
          serDe,
          metricsFacade,
          metricsSystem,
          smf,
          chain,
          forkIdHashProvider,
          forkIdHasher,
          isBlockImportEnabledProvider,
          p2pState,
          ->
          MisbehavingP2PNetwork(
            privateKeyBytes = privateKeyBytes,
            p2pConfig = p2pConfig,
            chainId = chainId,
            serDe = serDe,
            metricsFacade = metricsFacade,
            metricsSystem = metricsSystem,
            smf = smf,
            chain = chain,
            forkIdHashProvider = forkIdHashProvider,
            forkIdHasher = forkIdHasher,
            isBlockImportEnabledProvider = isBlockImportEnabledProvider,
            p2pState = p2pState,
            blockRetrievalStrategy = blockRetrievalStrategy,
          ).p2pNetwork
        },
      )

    validatorStack.setMaruApp(validatorMaruApp)
    validatorStack.maruApp.start()

    val bootnodeEnr =
      validatorStack.maruApp.p2pNetwork.localNodeRecord
        ?.asEnr()

    validatorEthApiClient =
      createEthApiClient(
        rpcUrl = validatorStack.besuNode.jsonRpcBaseUrl().get(),
        log = LogManager.getLogger("clients.l2.test.validator"),
        requestRetryConfig = null,
        vertx = null,
      )

    await
      .atMost(20.seconds.toJavaDuration())
      .pollInterval(200.milliseconds.toJavaDuration())
      .ignoreExceptions()
      .untilAsserted {
        assertThat(
          validatorEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.LATEST).get().number,
        ).isGreaterThanOrEqualTo(0UL)
      }

    val validatorGenesis = validatorStack.besuNode.ethGetBlockByNumber("earliest", false)
    val followerGenesis = followerStack.besuNode.ethGetBlockByNumber("earliest", false)
    assertThat(validatorGenesis).isEqualTo(followerGenesis)

    val tcpPortFollower = findFreePort()
    val udpPortFollower = findFreePort()
    val followerMaruApp =
      maruFactory.buildTestMaruFollowerWithDiscovery(
        ethereumJsonRpcUrl = followerStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = followerStack.besuNode.engineRpcUrl().get(),
        dataDir = followerStack.tmpDir,
        overridingLineaContractClient = fakeLineaContract,
        bootnode = bootnodeEnr,
        p2pPort = tcpPortFollower,
        discoveryPort = udpPortFollower,
        banPeriod = banPeriod,
        cooldownPeriod = followerCooldownPeriod,
        blockRangeRequestTimeout = blockRangeRequestTimeout,
      )
    followerStack.setMaruApp(followerMaruApp)

    val job =
      CoroutineScope(Dispatchers.Default).launch {
        while (true) {
          transactionsHelper.run {
            validatorStack.besuNode.sendTransactionAndAssertExecution(
              logger = log,
              recipient = createAccount("another account"),
              amount = Amount.ether(1),
            )
          }
        }
      }

    try {
      await
        .atMost(20.seconds.toJavaDuration())
        .pollInterval(200.milliseconds.toJavaDuration())
        .ignoreExceptions()
        .untilAsserted {
          assertThat(
            validatorEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.LATEST).get().number,
          ).isGreaterThanOrEqualTo(10UL)
        }

      followerStack.maruApp.start()

      validatorStack.maruApp.awaitTillMaruHasPeers(1u)
      followerStack.maruApp.awaitTillMaruHasPeers(1u)

      followerEthApiClient =
        createEthApiClient(
          rpcUrl = followerStack.besuNode.jsonRpcBaseUrl().get(),
          log = LogManager.getLogger("clients.l2.test.follower"),
          requestRetryConfig = null,
          vertx = null,
        )
      // wait for Besu to be fully started and synced,
      // to avoid CI flakiness due to low resources sometimes
      await
        .atMost(20.seconds.toJavaDuration())
        .pollInterval(200.milliseconds.toJavaDuration())
        .ignoreExceptions()
        .untilAsserted {
          assertThat(
            followerEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.LATEST).get().number,
          ).isGreaterThanOrEqualTo(0UL)
        }
    } catch (e: Exception) {
      job.cancel()
    }
    return MaruNodeSetup(validatorMaruApp = validatorMaruApp, followerMaruApp = followerMaruApp, job = job)
  }

  private fun findFreePort(): UInt =
    runCatching {
      ServerSocket(0).use { socket ->
        socket.reuseAddress = true
        socket.localPort.toUInt()
      }
    }.getOrElse {
      throw IllegalStateException("Could not find a free port", it)
    }
}
