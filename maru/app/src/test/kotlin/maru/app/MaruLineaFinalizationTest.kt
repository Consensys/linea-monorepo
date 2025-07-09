/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import kotlin.test.Test
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
import linea.domain.BlockParameter
import linea.ethapi.EthApiClient
import linea.web3j.ethapi.createEthApiClient
import maru.testutils.MaruFactory
import maru.testutils.NetworkParticipantStack
import maru.testutils.besu.BesuTransactionsHelper
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
import org.junit.jupiter.api.BeforeEach
import org.web3j.protocol.core.DefaultBlockParameter

class MaruLineaFinalizationTest {
  private lateinit var cluster: Cluster
  private lateinit var validatorStack: NetworkParticipantStack
  private lateinit var followerStack: NetworkParticipantStack
  private lateinit var transactionsHelper: BesuTransactionsHelper
  private val log = LogManager.getLogger(this.javaClass)
  private val maruFactory = MaruFactory()
  private lateinit var fakeLineaContract: FakeLineaRollupSmartContractClient
  private lateinit var validatorEthApiClient: EthApiClient
  private lateinit var followerEthApiClient: EthApiClient

  @BeforeEach
  fun setUp() {
    fakeLineaContract = FakeLineaRollupSmartContractClient()
    transactionsHelper = BesuTransactionsHelper()
    cluster =
      Cluster(
        ClusterConfigurationBuilder().build(),
        NetConditions(NetTransactions()),
        ThreadBesuNodeRunner(),
      )

    validatorStack =
      NetworkParticipantStack(cluster = cluster) { ethereumJsonRpcBaseUrl, engineRpcUrl, tmpDir ->
        maruFactory.buildTestMaruValidatorWithP2pPeering(
          ethereumJsonRpcUrl = ethereumJsonRpcBaseUrl,
          engineApiRpc = engineRpcUrl,
          dataDir = tmpDir,
          overridingLineaContractClient = fakeLineaContract,
        )
      }
    validatorStack.maruApp.start()
    followerStack =
      NetworkParticipantStack(
        cluster = cluster,
      ) { ethereumJsonRpcBaseUrl, engineRpcUrl, tmpDir ->
        maruFactory.buildTestMaruFollowerWithP2pPeering(
          ethereumJsonRpcUrl = ethereumJsonRpcBaseUrl,
          engineApiRpc = engineRpcUrl,
          dataDir = tmpDir,
          validatorPortForStaticPeering = validatorStack.p2pPort,
          overridingLineaContractClient = fakeLineaContract,
        )
      }
    followerStack.maruApp.start()

    val validatorGenesis =
      validatorStack.besuNode
        .nodeRequests()
        .eth()
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf("earliest"),
          false,
        ).send()
        .block
    val followerGenesis =
      followerStack.besuNode
        .nodeRequests()
        .eth()
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf("earliest"),
          false,
        ).send()
        .block
    assertThat(validatorGenesis).isEqualTo(followerGenesis)

    validatorEthApiClient =
      createEthApiClient(
        rpcUrl = validatorStack.besuNode.jsonRpcBaseUrl().get(),
        log = LogManager.getLogger("clients.l2.test.validator"),
        requestRetryConfig = null,
        vertx = null,
      )
    followerEthApiClient =
      createEthApiClient(
        rpcUrl = followerStack.besuNode.jsonRpcBaseUrl().get(),
        log = LogManager.getLogger("clients.l2.test.follower"),
        requestRetryConfig = null,
        vertx = null,
      )
  }

  @AfterEach
  fun tearDown() {
    cluster.close()
    followerStack.stop()
    validatorStack.stop()
  }

  @Test
  fun `should finalize current block right away when syncing and behind finalized on L1`() {
    fakeLineaContract.setFinalizedBlock(4UL)

    repeat(3) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(1),
        )
      }
    }

    await
      .atMost(5.seconds.toJavaDuration())
      .until {
        followerEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.LATEST).get().number == 3UL
      }

    // FIXME: it should be 3, but it seems that only calls FCU.finalized with parent block?
    assertThat(validatorEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.FINALIZED).get().number)
      .isEqualTo(2UL)
    assertThat(followerEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.FINALIZED).get().number)
      .isEqualTo(2UL)

    // Propagating the Head of the chain further than the Finalization height
    repeat(4) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(1),
        )
      }
    }

    await
      .atMost(5.seconds.toJavaDuration())
      .untilAsserted {
        assertThat(followerEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.LATEST).get().number)
          .isGreaterThan(4UL)
      }

    assertThat(validatorEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.FINALIZED).get().number)
      .isEqualTo(4UL)
    assertThat(followerEthApiClient.getBlockByNumberWithoutTransactionsData(BlockParameter.Tag.FINALIZED).get().number)
      .isEqualTo(4UL)
  }
}
