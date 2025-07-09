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
import maru.testutils.Checks.getMinedBlocks
import maru.testutils.Checks.verifyBlockTimeWithAGapOn
import maru.testutils.MaruFactory
import maru.testutils.besu.BesuFactory
import maru.testutils.besu.BesuTransactionsHelper
import maru.testutils.besu.ethGetBlockByNumber
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

class MaruConsensusSwitchTest {
  companion object {
    private const val VANILLA_EXTRA_DATA_LENGTH = 32
  }

  private lateinit var cluster: Cluster
  private lateinit var besuNode: BesuNode
  private lateinit var maruNode: MaruApp
  private lateinit var transactionsHelper: BesuTransactionsHelper
  private val log = LogManager.getLogger(this.javaClass)
  private val maruFactory = MaruFactory()

  @TempDir
  private lateinit var tmpDir: File
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
    cluster.close()
    maruNode.stop()
  }

  @Test
  fun `Maru is capable of switching from Delegated to QBFT consensus without block pauses`() {
    val stackStartupMargin = 15
    val expectedBlocksInClique = 5
    val totalBlocksToProduce = expectedBlocksInClique * 2
    var currentTimestamp = System.currentTimeMillis() / 1000
    val switchTimestamp = currentTimestamp + stackStartupMargin + expectedBlocksInClique
    log.info("Setting Prague switch timestamp to $switchTimestamp, current timestamp: $currentTimestamp")

    // Initialize Besu with the same switch timestamp
    besuNode =
      BesuFactory.buildSwitchableBesu(
        switchTimestamp = switchTimestamp,
        expectedBlocksInClique = expectedBlocksInClique,
      )
    cluster.start(besuNode)

    // Create a new Maru node with consensus switch configuration
    val ethereumJsonRpcBaseUrl = besuNode.jsonRpcBaseUrl().get()
    val engineRpcUrl = besuNode.engineRpcUrl().get()

    maruNode =
      maruFactory.buildTestMaruValidatorWithConsensusSwitch(
        ethereumJsonRpcUrl = ethereumJsonRpcBaseUrl,
        engineApiRpc = engineRpcUrl,
        dataDir = tmpDir.toPath(),
        switchTimestamp = switchTimestamp,
      )
    maruNode.start()

    log.info("Sending transactions")
    repeat(totalBlocksToProduce) {
      transactionsHelper.run {
        besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("pre-switch account"),
          amount = Amount.ether(100),
        )
      }
    }

    currentTimestamp = System.currentTimeMillis() / 1000
    log.info("Current timestamp: $currentTimestamp, switch timestamp: $switchTimestamp")
    assertThat(currentTimestamp).isGreaterThan(switchTimestamp)

    val blockProducedByClique = besuNode.ethGetBlockByNumber(1UL)

    assertThat(blockProducedByClique.extraData.length).isGreaterThan(VANILLA_EXTRA_DATA_LENGTH)

    val blockProducedByPrague = besuNode.ethGetBlockByNumber("latest")

    assertThat(blockProducedByPrague.extraData.length).isEqualTo(24)

    val blocks = besuNode.getMinedBlocks(totalBlocksToProduce)

    val switchBlock = blocks.findPragueBlock(switchTimestamp)!!

    blocks.verifyBlockTimeWithAGapOn(switchBlock)
  }

  private fun List<EthBlock.Block>.findPragueBlock(expectedSwitchTimestamp: Long): Int? =
    this
      .indexOfFirst {
        it.timestamp.toLong() >= expectedSwitchTimestamp
      }.takeIf { it != -1 }
}
