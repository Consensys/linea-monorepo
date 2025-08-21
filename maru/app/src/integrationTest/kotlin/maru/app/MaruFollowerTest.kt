/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
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
import org.junit.jupiter.api.Test
import testutils.Checks.getMinedBlocks
import testutils.PeeringNodeNetworkStack
import testutils.besu.BesuFactory
import testutils.besu.BesuTransactionsHelper
import testutils.besu.ethGetBlockByNumber
import testutils.besu.startWithRetry
import testutils.maru.MaruFactory
import testutils.maru.awaitTillMaruHasPeers

class MaruFollowerTest {
  private lateinit var cluster: Cluster
  private lateinit var validatorStack: PeeringNodeNetworkStack
  private lateinit var followerStack: PeeringNodeNetworkStack
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

    validatorStack = PeeringNodeNetworkStack()

    followerStack =
      PeeringNodeNetworkStack(
        besuBuilder = { BesuFactory.buildTestBesu(validator = false) },
      )

    // Start both Besu nodes together for proper peering
    PeeringNodeNetworkStack.startBesuNodes(cluster, validatorStack, followerStack)

    // Create and start validator Maru app first
    val validatorMaruApp =
      maruFactory.buildTestMaruValidatorWithP2pPeering(
        ethereumJsonRpcUrl = validatorStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = validatorStack.besuNode.engineRpcUrl().get(),
        dataDir = validatorStack.tmpDir,
      )
    validatorStack.setMaruApp(validatorMaruApp)
    validatorStack.maruApp.start()

    // Get the validator's p2p port after it's started
    val validatorP2pPort = validatorStack.p2pPort

    // Create follower Maru app with the validator's p2p port for static peering
    val followerMaruApp =
      maruFactory.buildTestMaruFollowerWithP2pPeering(
        ethereumJsonRpcUrl = followerStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = followerStack.besuNode.engineRpcUrl().get(),
        dataDir = followerStack.tmpDir,
        validatorPortForStaticPeering = validatorP2pPort,
        syncPeerChainGranularity = 1u,
      )
    followerStack.setMaruApp(followerMaruApp)
    followerStack.maruApp.start()

    log.info("Nodes are peered")
    followerStack.maruApp.awaitTillMaruHasPeers(1u)
    validatorStack.maruApp.awaitTillMaruHasPeers(1u)
    val validatorGenesis = validatorStack.besuNode.ethGetBlockByNumber("earliest", false)
    val followerGenesis = followerStack.besuNode.ethGetBlockByNumber("earliest", false)
    assertThat(validatorGenesis).isEqualTo(followerGenesis)
  }

  @AfterEach
  fun tearDown() {
    followerStack.maruApp.stop()
    validatorStack.maruApp.stop()
    followerStack.maruApp.close()
    validatorStack.maruApp.close()
    cluster.close()
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
    followerStack.setMaruApp(
      maruFactory.buildTestMaruFollowerWithP2pPeering(
        ethereumJsonRpcUrl = followerStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = followerStack.besuNode.engineRpcUrl().get(),
        dataDir = followerStack.tmpDir,
        validatorPortForStaticPeering = validatorStack.p2pPort,
      ),
    )
    followerStack.maruApp.start()

    followerStack.maruApp.awaitTillMaruHasPeers(1u)
    validatorStack.maruApp.awaitTillMaruHasPeers(1u)

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
    validatorStack.setMaruApp(
      maruFactory.buildTestMaruValidatorWithP2pPeering(
        ethereumJsonRpcUrl = validatorStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = validatorStack.besuNode.engineRpcUrl().get(),
        dataDir = validatorStack.tmpDir,
        p2pPort = validatorP2pPort,
      ),
    )
    validatorStack.maruApp.start()

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
    cluster.startWithRetry(followerStack.besuNode)

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
  fun `Maru follower is able to complete initial syncing`() {
    followerStack.maruApp.stop()
    followerStack.maruApp.close()
    val blocksToProduce = 20
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
    checkNetworkStackBlocksProduced(validatorStack, blocksToProduce)

    followerStack.setMaruApp(
      maruFactory.buildTestMaruFollowerWithP2pPeering(
        ethereumJsonRpcUrl = followerStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = followerStack.besuNode.engineRpcUrl().get(),
        dataDir = followerStack.tmpDir,
        validatorPortForStaticPeering = validatorStack.p2pPort,
      ),
    )
    followerStack.maruApp.start()

    checkValidatorAndFollowerBlocks(blocksToProduce)
  }

  @Test
  fun `Maru follower is able to complete syncing after restarted`() {
    val blocksToProduce = 20
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

    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }
    checkNetworkStackBlocksProduced(validatorStack, 2 * blocksToProduce)

    followerStack.setMaruApp(
      maruFactory.buildTestMaruFollowerWithP2pPeering(
        ethereumJsonRpcUrl = followerStack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = followerStack.besuNode.engineRpcUrl().get(),
        dataDir = followerStack.tmpDir,
        validatorPortForStaticPeering = validatorStack.p2pPort,
      ),
    )
    followerStack.maruApp.start()

    checkValidatorAndFollowerBlocks(2 * blocksToProduce)
  }

  @Test
  fun `Maru follower is able to complete syncing after disconnect peers`() {
    val blocksToProduce = 20
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

    val followerP2PNetwork = followerStack.maruApp.p2pNetwork()
    val peers = followerP2PNetwork.getPeers()
    peers.forEach {
      followerP2PNetwork.dropPeer(it)
    }

    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }
    checkNetworkStackBlocksProduced(validatorStack, 2 * blocksToProduce)
    checkNetworkStackBlocksProduced(followerStack, blocksToProduce)
    peers.forEach {
      followerP2PNetwork.addPeer("${it.address}/p2p/${it.nodeId}")
    }
    checkNetworkStackBlocksProduced(followerStack, 2 * blocksToProduce)
  }

  private fun checkValidatorAndFollowerBlocks(blocksToProduce: Int) {
    await
      .pollDelay(100.milliseconds.toJavaDuration())
      .timeout(30.seconds.toJavaDuration())
      .untilAsserted {
        val blocksProducedByQbftValidator = validatorStack.besuNode.getMinedBlocks(blocksToProduce)
        val blocksImportedByFollower = followerStack.besuNode.getMinedBlocks(blocksToProduce)
        assertThat(blocksImportedByFollower).isEqualTo(blocksProducedByQbftValidator)
      }
  }

  private fun checkNetworkStackBlocksProduced(
    stack: PeeringNodeNetworkStack,
    blocksProduced: Int,
  ) {
    await
      .pollDelay(100.milliseconds.toJavaDuration())
      .timeout(30.seconds.toJavaDuration())
      .untilAsserted {
        val blocksOnStack = stack.besuNode.getMinedBlocks(blocksProduced)
        assertThat(blocksOnStack.size).isEqualTo(blocksProduced)
      }
  }
}
