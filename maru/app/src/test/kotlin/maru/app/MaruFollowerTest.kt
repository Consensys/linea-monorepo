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
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
import maru.testutils.Checks.getMinedBlocks
import maru.testutils.MaruFactory
import maru.testutils.NetworkParticipantStack
import maru.testutils.besu.BesuTransactionsHelper
import maru.testutils.besu.ethGetBlockByNumber
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

class MaruFollowerTest {
  private lateinit var cluster: Cluster
  private lateinit var validatorStack: NetworkParticipantStack
  private lateinit var followerStack: NetworkParticipantStack
  private lateinit var transactionsHelper: BesuTransactionsHelper
  private val log = LogManager.getLogger(this.javaClass)
  private val maruFactory = MaruFactory()

  @BeforeEach
  fun setUp() {
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
        )
      }
    followerStack.maruApp.start()

    val validatorGenesis = validatorStack.besuNode.ethGetBlockByNumber("earliest", false)
    val followerGenesis = followerStack.besuNode.ethGetBlockByNumber("earliest", false)
    assertThat(validatorGenesis).isEqualTo(followerGenesis)
  }

  @AfterEach
  fun tearDown() {
    cluster.close()
    followerStack.stop()
    validatorStack.stop()
  }

  @Test
  fun `Maru follower is able to import blocks`() {
    val blocksToProduce = 5
    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }

    checkValidatorAndFollowerBlocks(blocksToProduce)
  }

  @Test
  fun `Maru follower is able to import blocks after going down`() {
    val blocksToProduce = 5
    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }

    // This is here mainly to wait until block propagation is complete
    checkValidatorAndFollowerBlocks(blocksToProduce)

    followerStack.maruApp.stop()
    followerStack.maruApp.close()
    followerStack.maruApp =
      maruFactory.buildTestMaruFollowerWithP2pPeering(
        ethereumJsonRpcUrl = followerStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = followerStack.besuNode.engineRpcUrl().get(),
        dataDir = followerStack.tmpDir,
        validatorPortForStaticPeering = validatorStack.p2pPort,
      )
    followerStack.maruApp.start()

    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }

    checkValidatorAndFollowerBlocks(blocksToProduce * 2)
  }

  @Test
  fun `Maru follower is able to import blocks after Validator stack goes down`() {
    val blocksToProduce = 5
    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }

    val validatorP2pPort = validatorStack.p2pPort
    // This is here mainly to wait until block propagation is complete
    checkValidatorAndFollowerBlocks(blocksToProduce)

    validatorStack.maruApp.stop()
    validatorStack.maruApp.close()
    validatorStack.maruApp =
      maruFactory.buildTestMaruValidatorWithP2pPeering(
        ethereumJsonRpcUrl = validatorStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = validatorStack.besuNode.engineRpcUrl().get(),
        dataDir = validatorStack.tmpDir,
        p2pPort = validatorP2pPort,
      )
    validatorStack.maruApp.start()
    // TODO: This is to guarantee reconnection. It should actually go away once syncing is implemented
    Thread.sleep(MaruFactory.defaultReconnectDelay.inWholeMilliseconds)

    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }

    checkValidatorAndFollowerBlocks(blocksToProduce * 2)
  }

  @Test
  fun `Maru follower is able to import blocks after its validator el node goes down`() {
    val blocksToProduce = 5
    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }

    // This is here mainly to wait until block propagation is complete
    checkValidatorAndFollowerBlocks(blocksToProduce)

    cluster.stop()
    Thread.sleep(3000)
    cluster.start(followerStack.besuNode)

    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }

    checkValidatorAndFollowerBlocks(blocksToProduce * 2)
  }

  private fun checkValidatorAndFollowerBlocks(blocksToProduce: Int) {
    await
      .pollDelay(100.milliseconds.toJavaDuration())
      .timeout(10.seconds.toJavaDuration())
      .untilAsserted {
        val blocksProducedByQbftValidator = validatorStack.besuNode.getMinedBlocks(blocksToProduce)
        val blocksImportedByFollower = followerStack.besuNode.getMinedBlocks(blocksToProduce)
        assertThat(blocksImportedByFollower).isEqualTo(blocksProducedByQbftValidator)
      }
  }
}
