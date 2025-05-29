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
import maru.app.Checks.getMinedBlocks
import maru.consensus.ElFork
import maru.core.SealedBeaconBlock
import maru.p2p.Message
import maru.p2p.MessageType
import maru.p2p.NoOpP2PNetwork
import maru.p2p.P2PNetwork
import maru.testutils.InjectableSealedBlocksFakeNetwork
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
import tech.pegasys.teku.infrastructure.async.SafeFuture

class MaruFollowerTest {
  private lateinit var cluster: Cluster
  private lateinit var validatorStack: NetworkParticipantStack
  private lateinit var followerStack: NetworkParticipantStack
  private lateinit var transactionsHelper: BesuTransactionsHelper
  private val log = LogManager.getLogger(this.javaClass)
  private lateinit var injectableSealedBlocksFakeNetwork: InjectableSealedBlocksFakeNetwork

  @BeforeEach
  fun setUp() {
    val elFork = ElFork.Prague
    transactionsHelper = BesuTransactionsHelper()
    cluster =
      Cluster(
        ClusterConfigurationBuilder().build(),
        NetConditions(NetTransactions()),
        ThreadBesuNodeRunner(),
      )

    injectableSealedBlocksFakeNetwork = InjectableSealedBlocksFakeNetwork()
    val validatorP2PNetwork =
      object : P2PNetwork by NoOpP2PNetwork {
        override fun broadcastMessage(message: Message<*>): SafeFuture<Unit> =
          if (message.type == MessageType.BEACON_BLOCK) {
            SafeFuture
              .of(
                SafeFuture
                  .runAsync {
                    // To untie Validator's callstack from follower's concerns
                    injectableSealedBlocksFakeNetwork.injectSealedBlock(
                      message.payload as SealedBeaconBlock,
                    )
                  },
              ).thenApply { }
          } else {
            SafeFuture.completedFuture(Unit)
          }
      }
    validatorStack = NetworkParticipantStack(cluster = cluster, p2pNetwork = validatorP2PNetwork)
    followerStack =
      NetworkParticipantStack(
        cluster = cluster,
        p2pNetwork = injectableSealedBlocksFakeNetwork,
      ) { ethereumJsonRpcBaseUrl, engineRpcUrl, tmpDir, p2pNetwork ->
        MaruFactory.buildTestMaruFollower(
          ethereumJsonRpcUrl = ethereumJsonRpcBaseUrl,
          engineApiRpc = engineRpcUrl,
          elFork = elFork,
          dataDir = tmpDir,
          p2pNetwork = p2pNetwork,
        )
      }
    validatorStack.maruApp.start()
    followerStack.maruApp.start()
  }

  @AfterEach
  fun tearDown() {
    cluster.close()
    followerStack.stop()
    validatorStack.stop()
  }

  @Test
  fun `Maru follower is able to import blocks`() {
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

    await
      .pollDelay(100.milliseconds.toJavaDuration())
      .timeout(1.seconds.toJavaDuration())
      .untilAsserted {
        val blocksProducedByQbftValidator = validatorStack.besuNode.getMinedBlocks(blocksToProduce)
        val blocksImportedByFollower = followerStack.besuNode.getMinedBlocks(blocksToProduce)
        assertThat(blocksImportedByFollower).isEqualTo(blocksProducedByQbftValidator)
      }
  }
}
