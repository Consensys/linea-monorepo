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
import linea.kotlin.decodeHex
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
import testutils.Checks.verifyBlockTime
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
    expectedBlocksInClique: Int,
    totalBlocksToProduce: Int,
  ) {
    val blockProducedByClique = besuNode.ethGetBlockByNumber(1UL)
    assertThat(blockProducedByClique.extraData.decodeHex().size).isGreaterThan(VANILLA_EXTRA_DATA_LENGTH)

    val blockProducedAfterSwitch = besuNode.ethGetBlockByNumber("latest")
    assertThat(blockProducedAfterSwitch.extraData.decodeHex().size).isLessThanOrEqualTo(VANILLA_EXTRA_DATA_LENGTH)

    val blocks = besuNode.getMinedBlocks(totalBlocksToProduce)
    val parisSwitchBlock = blocks.findSwitchBlock()!!
    val cliqueBlocks = blocks.subList(0, parisSwitchBlock)
    cliqueBlocks.verifyBlockTime()
    assertThat(cliqueBlocks).hasSize(expectedBlocksInClique)
    // Check that there are Prague blocks
    blocks.subList(parisSwitchBlock, blocks.size).verifyBlockTime()
  }

  @Test
  fun `follower node correctly switches from Clique to POS after peering with Sequencer validator`() {
    val stackStartupMargin = 40UL
    val expectedBlocksInClique = 5
    var currentTimestamp = (System.currentTimeMillis() / 1000).toULong()
    val shanghaiTimestamp = currentTimestamp + stackStartupMargin + expectedBlocksInClique.toULong()
    val cancunTimestamp = shanghaiTimestamp + 10u
    val pragueTimestamp = cancunTimestamp + 10u
    val totalBlocksToProduce = (pragueTimestamp - currentTimestamp).toInt()
    val ttd = expectedBlocksInClique.toULong() * 2UL
    log.info(
      "Setting Prague switch timestamp to $pragueTimestamp, shanghai switch to $shanghaiTimestamp, Cancun switch to " +
        "$cancunTimestamp, current timestamp: $currentTimestamp",
    )

    // Initialize Besu with the same switch timestamp
    validatorBesuNode =
      BesuFactory.buildSwitchableBesu(
        shanghaiTimestamp = shanghaiTimestamp,
        cancunTimestamp = cancunTimestamp,
        pragueTimestamp = pragueTimestamp,
        ttd = ttd,
        validator = true,
      )
    followerBesuNode =
      BesuFactory.buildSwitchableBesu(
        shanghaiTimestamp = shanghaiTimestamp,
        cancunTimestamp = cancunTimestamp,
        pragueTimestamp = pragueTimestamp,
        ttd = ttd,
        validator = false,
      )
    cluster.startWithRetry(validatorBesuNode, followerBesuNode)

    // Create a new Maru node with consensus switch configuration
    val validatorEthereumJsonRpcBaseUrl = validatorBesuNode.jsonRpcBaseUrl().get()
    val validatorEngineRpcUrl = validatorBesuNode.engineRpcUrl().get()

    val maruFactory =
      MaruFactory(
        pragueTimestamp = pragueTimestamp,
        cancunTimestamp = cancunTimestamp,
        shanghaiTimestamp = shanghaiTimestamp,
        ttd = ttd,
      )
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

    currentTimestamp = (System.currentTimeMillis() / 1000).toULong()
    log.info("Current timestamp: $currentTimestamp, prague switch timestamp: $pragueTimestamp")
    assertThat(currentTimestamp).isGreaterThan(pragueTimestamp)

    verifyConsensusSwitch(
      besuNode = validatorBesuNode,
      expectedBlocksInClique = expectedBlocksInClique,
      totalBlocksToProduce = totalBlocksToProduce,
    )
    verifyConsensusSwitch(
      besuNode = followerBesuNode,
      expectedBlocksInClique = expectedBlocksInClique,
      totalBlocksToProduce = totalBlocksToProduce,
    )
  }

  private fun List<EthBlock.Block>.findSwitchBlock(): Int? =
    this
      .indexOfFirst {
        it.difficulty.toInt() == 0
      }.takeIf { it != -1 }
}
