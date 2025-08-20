/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import java.io.File
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.tests.acceptance.dsl.blockchain.Amount
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode
import org.hyperledger.besu.tests.acceptance.dsl.node.ThreadBesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.io.TempDir
import org.web3j.protocol.core.methods.response.EthBlock
import testutils.Checks.getMinedBlocks
import testutils.Checks.verifyBlockTimeWithAGapOn
import testutils.besu.BesuFactory
import testutils.besu.BesuTransactionsHelper
import testutils.besu.ethGetBlockByNumber
import testutils.besu.startWithRetry
import testutils.maru.MaruFactory
import testutils.maru.awaitTillMaruHasPeers

class MaruConsensusSwitchTest {
  companion object {
    private const val VANILLA_EXTRA_DATA_LENGTH = 32
  }

  private lateinit var cluster: Cluster
  private lateinit var validatorBesuNode: BesuNode
  private lateinit var validatorMaruNode: MaruApp
  private lateinit var followerBesuNode: BesuNode
  private lateinit var followerMaruNode: MaruApp
  private lateinit var transactionsHelper: BesuTransactionsHelper
  private val log = LogManager.getLogger(this.javaClass)

  @TempDir
  private lateinit var validatorMaruTmpDir: File

  @TempDir
  private lateinit var followerMaruTmpDir: File
  private val net = NetConditions(NetTransactions())

  @BeforeEach
  fun setUp() {
    transactionsHelper = BesuTransactionsHelper()
    // We'll set the switchTimestamp in the test method
    cluster =
      Cluster(
        ClusterConfigurationBuilder().build(),
        net,
        ThreadBesuNodeRunner(),
      )
  }

  @AfterEach
  fun tearDown() {
    followerMaruNode.stop()
    validatorMaruNode.stop()
    followerMaruNode.close()
    validatorMaruNode.close()
    cluster.close()
  }

  private fun verifyConsensusSwitch(
    besuNode: BesuNode,
    totalBlocksToProduce: Int,
    switchTimestamp: Long,
  ) {
    val blockProducedByClique = besuNode.ethGetBlockByNumber(1UL)
    assertThat(blockProducedByClique.extraData.length).isGreaterThan(VANILLA_EXTRA_DATA_LENGTH)

    val blockProducedByPrague = besuNode.ethGetBlockByNumber("latest")
    assertThat(blockProducedByPrague.extraData.length).isEqualTo(24)

    val blocks = besuNode.getMinedBlocks(totalBlocksToProduce)
    val switchBlock = blocks.findPragueBlock(switchTimestamp)!!
    blocks.verifyBlockTimeWithAGapOn(switchBlock)
  }

  @Test
  fun `Follower node correctly switches from Clique to POS after peering with Sequencer validator`() {
    val stackStartupMargin = 30
    val expectedBlocksInClique = 5
    val totalBlocksToProduce = expectedBlocksInClique * 4
    var currentTimestamp = System.currentTimeMillis() / 1000
    val shanghaiTimestamp = currentTimestamp + stackStartupMargin + expectedBlocksInClique
    val pragueTimestamp = shanghaiTimestamp + 5L
    log.info(
      "Setting Shanghai switch timestamp to $shanghaiTimestamp, Prague switch timestamp to $pragueTimestamp, " +
        "current timestamp: $currentTimestamp",
    )

    // Initialize Besu with the same switch timestamp
    validatorBesuNode =
      BesuFactory.buildSwitchableBesu(
        switchTimestamp = shanghaiTimestamp,
        pragueTimestamp = pragueTimestamp,
        expectedBlocksInClique = expectedBlocksInClique,
        validator = true,
      )
    followerBesuNode =
      BesuFactory.buildSwitchableBesu(
        switchTimestamp = shanghaiTimestamp,
        pragueTimestamp = pragueTimestamp,
        expectedBlocksInClique = expectedBlocksInClique,
        validator = false,
      )
    cluster.startWithRetry(validatorBesuNode, followerBesuNode)

    // Create a new Maru node with consensus switch configuration
    val validatorEthereumJsonRpcBaseUrl = validatorBesuNode.jsonRpcBaseUrl().get()
    val validatorEngineRpcUrl = validatorBesuNode.engineRpcUrl().get()

    val maruFactory = MaruFactory(shanghaiTimestamp = shanghaiTimestamp, pragueTimestamp = pragueTimestamp)
    validatorMaruNode =
      maruFactory.buildSwitchableTestMaruValidatorWithP2pPeering(
        ethereumJsonRpcUrl = validatorEthereumJsonRpcBaseUrl,
        engineApiRpc = validatorEngineRpcUrl,
        dataDir = validatorMaruTmpDir.toPath(),
      )
    validatorMaruNode.start()

    val followerEthereumJsonRpcBaseUrl = followerBesuNode.jsonRpcBaseUrl().get()
    val followerEngineRpcUrl = followerBesuNode.engineRpcUrl().get()

    followerMaruNode =
      maruFactory.buildTestMaruFollowerWithConsensusSwitch(
        ethereumJsonRpcUrl = followerEthereumJsonRpcBaseUrl,
        engineApiRpc = followerEngineRpcUrl,
        dataDir = followerMaruTmpDir.toPath(),
        validatorPortForStaticPeering = validatorMaruNode.p2pPort(),
      )
    followerMaruNode.start()

    followerMaruNode.awaitTillMaruHasPeers(1u)
    validatorMaruNode.awaitTillMaruHasPeers(1u)
    log.info("Sending transactions")
    repeat(totalBlocksToProduce) {
      transactionsHelper.run {
        validatorBesuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("pre-switch account"),
          amount = Amount.ether(100),
        )
      }
    }

    currentTimestamp = System.currentTimeMillis() / 1000
    log.info("Current timestamp: $currentTimestamp, shanghai switch timestamp: $shanghaiTimestamp")
    assertThat(currentTimestamp).isGreaterThan(shanghaiTimestamp)
    log.info("Current timestamp: $currentTimestamp, prague switch timestamp: $pragueTimestamp")
    assertThat(currentTimestamp).isGreaterThan(pragueTimestamp)

    verifyConsensusSwitch(validatorBesuNode, totalBlocksToProduce, shanghaiTimestamp)
    verifyConsensusSwitch(followerBesuNode, totalBlocksToProduce, shanghaiTimestamp)
  }

  private fun List<EthBlock.Block>.findPragueBlock(expectedSwitchTimestamp: Long): Int? =
    this
      .indexOfFirst {
        it.timestamp.toLong() >= expectedSwitchTimestamp
      }.takeIf { it != -1 }
}
