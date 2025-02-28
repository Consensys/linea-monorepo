/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.app

import java.math.BigInteger
import maru.testutils.MaruFactory
import maru.testutils.TransactionsHelper
import maru.testutils.besu.BesuFactory
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.tests.acceptance.dsl.account.Account
import org.hyperledger.besu.tests.acceptance.dsl.blockchain.Amount
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.ThreadBesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.web3j.protocol.core.DefaultBlockParameter

class MaruDummyConsensusTest {
  private val besuNodeRunner: BesuNodeRunner = ThreadBesuNodeRunner()
  private var cluster =
    Cluster(
      ClusterConfigurationBuilder().build(),
      NetConditions(NetTransactions()),
      besuNodeRunner,
    )
  private lateinit var besuNode: BesuNode
  private lateinit var maruNode: MaruApp
  private lateinit var transactionsHelper: TransactionsHelper
  private val log = LogManager.getLogger(this.javaClass)

  @BeforeEach
  fun setUp() {
    transactionsHelper = TransactionsHelper()
    besuNode = BesuFactory.buildTestBesu()
    cluster = Cluster(NetConditions(NetTransactions()))

    cluster.start(besuNode)
    val ethereumJsonRpcBaseUrl = besuNode.jsonRpcBaseUrl().get()
    val engineRpcUrl = besuNode.engineRpcUrl().get()
    maruNode = MaruFactory.buildTestMaru(ethereumJsonRpcUrl = ethereumJsonRpcBaseUrl, engineApiRpc = engineRpcUrl)
    maruNode.start()
  }

  @AfterEach
  fun tearDown() {
    cluster.close()
    maruNode.stop()
  }

  private fun sendTransactionAndAssertExecution(
    recipient: Account,
    amount: Amount,
  ) {
    val transfer = transactionsHelper.createTransfer(recipient, amount)
    val txHash = besuNode.execute(transfer)
    assertThat(txHash).isNotNull()
    log.info("Sending transaction {}, transaction data ", txHash)
    transactionsHelper.ethConditions.expectSuccessfulTransactionReceipt(txHash.toString()).verify(besuNode)
    log.info("Transaction {} was mined", txHash)
  }

  @Test
  fun `dummyConsensus is able to produce blocks with the expected block time`() {
    val blocksToProduce = 10
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }

    verifyBlockHeaders(blocksToProduce)
  }

  @Test
  fun `dummyConsensus works if Besu stops mid flight`() {
    val blocksToProduce = 5
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }
    cluster.stop()
    cluster.start(besuNode)

    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }
  }

  @Test
  fun `dummyConsensus works if Maru stops mid flight`() {
    val blocksToProduce = 5
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }
    maruNode.stop()
    maruNode.start()
    repeat(blocksToProduce) {
      sendTransactionAndAssertExecution(transactionsHelper.createAccount("another account"), Amount.ether(100))
    }
  }

  private fun verifyBlockHeaders(blocksProduced: Int) {
    val blocks =
      (1..blocksProduced)
        .map {
          besuNode
            .nodeRequests()
            .eth()
            .ethGetBlockByNumber(
              DefaultBlockParameter.valueOf(BigInteger.valueOf(it.toLong())),
              false,
            ).sendAsync()
        }.map { it.get().block }

    val blockTimeSeconds = 1L
    val timestamps = blocks.map { it.timestamp.toLong() }
    (1.until(blocks.size)).forEach {
      assertThat(timestamps[it - 1]).isLessThan(timestamps[it])
      assertThat(timestamps[it] - timestamps[it - 1]).isEqualTo(blockTimeSeconds)
    }
  }
}
