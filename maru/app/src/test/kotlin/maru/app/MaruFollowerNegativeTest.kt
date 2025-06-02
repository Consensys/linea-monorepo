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

import kotlin.test.Test
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
import maru.core.Seal
import maru.p2p.NoOpP2PNetwork
import maru.p2p.ValidationResult
import maru.testutils.Checks.getMinedBlocks
import maru.testutils.InjectableSealedBlocksFakeNetwork
import maru.testutils.MaruFactory
import maru.testutils.NetworkParticipantStack
import maru.testutils.SpyingP2PNetwork
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
import org.web3j.protocol.core.DefaultBlockParameter

class MaruFollowerNegativeTest {
  private val cluster: Cluster =
    Cluster(
      ClusterConfigurationBuilder().build(),
      NetConditions(NetTransactions()),
      ThreadBesuNodeRunner(),
    )

  private val transactionsHelper: BesuTransactionsHelper = BesuTransactionsHelper()
  private val log = LogManager.getLogger(this.javaClass)
  private val maruFactory = MaruFactory()

  @Test
  fun `Maru follower doesn't import blocks without proper signature`() {
    val spyingP2PNetwork = SpyingP2PNetwork(NoOpP2PNetwork)
    val validatorStack =
      NetworkParticipantStack(cluster = cluster) { ethereumJsonRpcBaseUrl, engineRpcUrl, tmpDir ->
        maruFactory.buildTestMaruValidatorWithoutP2pPeering(
          ethereumJsonRpcUrl = ethereumJsonRpcBaseUrl,
          engineApiRpc = engineRpcUrl,
          dataDir = tmpDir,
          p2pNetwork = spyingP2PNetwork,
        )
      }
    validatorStack.maruApp.start()
    val validatorGenesis =
      validatorStack.besuNode
        .nodeRequests()
        .eth()
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf("earliest"),
          false,
        ).send()
        .block

    val followerP2PNetwork = InjectableSealedBlocksFakeNetwork()
    val followerStack =
      NetworkParticipantStack(
        cluster = cluster,
      ) { ethereumJsonRpcBaseUrl, engineRpcUrl, tmpDir ->
        maruFactory.buildTestMaruFollowerWithoutP2pPeering(
          ethereumJsonRpcUrl = ethereumJsonRpcBaseUrl,
          engineApiRpc = engineRpcUrl,
          dataDir = tmpDir,
          p2pNetwork = followerP2PNetwork,
        )
      }
    followerStack.maruApp.start()
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

    val blocksToProduce = 1
    repeat(blocksToProduce) {
      transactionsHelper.run {
        validatorStack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }

    await
      .pollDelay(100.milliseconds.toJavaDuration())
      .timeout(5.seconds.toJavaDuration())
      .untilAsserted {
        val blocksProducedByQbftValidator = validatorStack.besuNode.getMinedBlocks(blocksToProduce)
        assertThat(blocksProducedByQbftValidator).hasSize(1)
      }

    val originalSealedBlock = spyingP2PNetwork.emittedBlockMessages.first()
    val originalSeal = originalSealedBlock.commitSeals.toList().first()
    val alternatedSignature = originalSeal.signature.reversed().toByteArray()
    val blockWithWrongSeal = originalSealedBlock.copy(commitSeals = setOf(Seal(alternatedSignature)))
    val result = followerP2PNetwork.injectSealedBlock(blockWithWrongSeal).get()

    assertThat(result).isInstanceOf(ValidationResult.Companion.Invalid::class.java)
  }
}
