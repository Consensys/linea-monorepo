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
import kotlin.time.Duration.Companion.seconds
import kotlin.time.ExperimentalTime
import kotlin.time.toJavaDuration
import maru.testutils.MaruFactory
import maru.testutils.besu.BesuFactory
import maru.testutils.besu.BesuTransactionsHelper
import maru.testutils.besu.latestBlock
import maru.testutils.besu.startWithRetry
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
import org.junit.jupiter.api.io.TempDir

class MaruStartFromPragueTest {
  private lateinit var cluster: Cluster
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
    maruNode.close()
  }

  @OptIn(ExperimentalTime::class)
  @Test
  fun `should start QBFT consensus from prague fork allowing empty blocks`() {
    val besuNode = BesuFactory.buildTestBesu()
    cluster.startWithRetry(besuNode)
    maruNode =
      maruFactory.buildTestMaruValidatorWithP2pPeering(
        ethereumJsonRpcUrl = besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = besuNode.engineRpcUrl().get(),
        dataDir = tmpDir.toPath(),
        allowEmptyBlocks = true,
      )
    maruNode.start()

    await
      .atMost(10.seconds.toJavaDuration())
      .ignoreExceptions()
      .until {
        besuNode.latestBlock().number.toLong() > 5
      }
  }

  @OptIn(ExperimentalTime::class)
  @Test
  fun `should start QBFT consensus from prague fork not allowing empty blocks`() {
    val besuNode = BesuFactory.buildTestBesu()
    cluster.startWithRetry(besuNode)
    maruNode =
      maruFactory.buildTestMaruValidatorWithP2pPeering(
        ethereumJsonRpcUrl = besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = besuNode.engineRpcUrl().get(),
        dataDir = tmpDir.toPath(),
        allowEmptyBlocks = false,
      )
    maruNode.start()
    val totalBlocksToProduce = 5
    repeat(totalBlocksToProduce) {
      transactionsHelper.run {
        besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("account"),
          amount = Amount.ether(1),
        )
      }
    }

    await
      .ignoreExceptions()
      .atMost(10.seconds.toJavaDuration())
      .until {
        besuNode.latestBlock().number.toLong() == totalBlocksToProduce.toLong()
      }

    Thread.sleep(3000) // Wait for the node to process the blocks
    // assert no blocks were produced without transactions
    assertThat(besuNode.latestBlock().number.toLong()).isEqualTo(totalBlocksToProduce.toLong())
  }
}
