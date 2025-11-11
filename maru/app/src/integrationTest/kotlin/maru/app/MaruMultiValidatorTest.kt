/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import io.libp2p.etc.types.fromHex
import java.math.BigInteger
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
import maru.consensus.qbft.ProposerSelectorImpl
import maru.core.SealedBeaconBlock
import maru.crypto.SecpCrypto
import maru.database.BeaconChain
import maru.extensions.encodeHex
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.kotlin.await
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.hyperledger.besu.tests.acceptance.dsl.blockchain.Amount
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions
import org.hyperledger.besu.tests.acceptance.dsl.node.ThreadBesuNodeRunner
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.web3j.protocol.core.methods.response.EthBlock
import testutils.Checks.getMinedBlocks
import testutils.PeeringNodeNetworkStack
import testutils.besu.BesuFactory
import testutils.besu.BesuTransactionsHelper
import testutils.besu.ethGetBlockByNumber
import testutils.maru.MaruFactory
import testutils.maru.awaitTillMaruHasPeers

class MaruMultiValidatorTest {
  private val key1 = "0802122012c0b113e2b0c37388e2b484112e13f05c92c4471e3ee1dfaa368fa5045325b2".fromHex()
  private val key2 = "0802122100f3d2fffa99dc8906823866d96316492ebf7a8478713a89a58b7385af85b088a1".fromHex()

  private lateinit var cluster: Cluster
  private lateinit var validator1Stack: PeeringNodeNetworkStack
  private lateinit var validator2Stack: PeeringNodeNetworkStack
  private lateinit var transactionsHelper: BesuTransactionsHelper
  private val log = LogManager.getLogger(this.javaClass)
  private val maruFactory1 = MaruFactory(validatorPrivateKey = key1)
  private val maruFactory2 = MaruFactory(validatorPrivateKey = key2)

  @BeforeEach
  fun setUp() {
    transactionsHelper = BesuTransactionsHelper()
    cluster =
      Cluster(
        ClusterConfigurationBuilder().build(),
        NetConditions(NetTransactions()),
        ThreadBesuNodeRunner(),
      )

    val besuBuilder = { BesuFactory.buildTestBesu(validator = false) }
    validator1Stack = PeeringNodeNetworkStack(besuBuilder)
    validator2Stack = PeeringNodeNetworkStack(besuBuilder)

    // Start all Besu nodes together for proper peering
    PeeringNodeNetworkStack.startBesuNodes(cluster, validator1Stack, validator2Stack)
  }

  @AfterEach
  fun tearDown() {
    validator2Stack.maruApp.stop()
    validator1Stack.maruApp.stop()
    validator2Stack.maruApp.close()
    validator1Stack.maruApp.close()
    cluster.close()
  }

  @Test
  fun `maru with multiple validators is able to produce blocks`() {
    val validator1Address = SecpCrypto.privateKeyToValidator(SecpCrypto.privateKeyBytesWithoutPrefix(key1))
    val validator2Address = SecpCrypto.privateKeyToValidator(SecpCrypto.privateKeyBytesWithoutPrefix(key2))
    log.info("Validator 1 (key1) address: ${validator1Address.address.encodeHex()}")
    log.info("Validator 2 (key2) address: ${validator2Address.address.encodeHex()}")
    val initialValidators = setOf(validator1Address, validator2Address)

    // Create and start validator 1 Maru app first
    val validator1MaruApp =
      maruFactory1.buildTestMaruValidatorWithP2pPeering(
        ethereumJsonRpcUrl = validator1Stack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = validator1Stack.besuNode.engineRpcUrl().get(),
        dataDir = validator1Stack.tmpDir,
        syncingConfig = MaruFactory.defaultSyncingConfig,
        allowEmptyBlocks = true,
        initialValidators = initialValidators,
      )
    validator1Stack.setMaruApp(validator1MaruApp)
    validator1Stack.maruApp.start()
    // Get validator 1 p2p port and node ID after it's started
    val validator1P2pPort = validator1Stack.p2pPort
    val validator1NodeId = validator1Stack.maruApp.p2pNetwork.nodeId

    // Create validator 2 Maru app with the validator 1 p2p port and node ID for static peering
    val validator2MaruApp =
      maruFactory2.buildTestMaruValidatorWithP2pPeering(
        ethereumJsonRpcUrl = validator2Stack.besuNode.jsonRpcBaseUrl().get(),
        engineApiRpc = validator2Stack.besuNode.engineRpcUrl().get(),
        dataDir = validator2Stack.tmpDir,
        validatorPortForStaticPeering = validator1P2pPort,
        validatorNodeIdForStaticPeering = validator1NodeId,
        syncingConfig = MaruFactory.defaultSyncingConfig,
        allowEmptyBlocks = true,
        initialValidators = initialValidators,
      )
    validator2Stack.setMaruApp(validator2MaruApp)
    validator2Stack.maruApp.start()

    validator2Stack.maruApp.awaitTillMaruHasPeers(1u)
    validator1Stack.maruApp.awaitTillMaruHasPeers(1u)
    log.info("Nodes are peered")

    val validator1besuGenesis = validator1Stack.besuNode.ethGetBlockByNumber("earliest", false)
    val validator2besuGenesis = validator2Stack.besuNode.ethGetBlockByNumber("earliest", false)
    assertThat(validator1besuGenesis).isEqualTo(validator2besuGenesis)

    val validator1MaruGenesis = validator1Stack.maruApp.beaconChain.getBeaconState(0u)
    val validator2MaruGenesis = validator2Stack.maruApp.beaconChain.getBeaconState(0u)
    assertThat(validator1MaruGenesis).isEqualTo(validator2MaruGenesis)

    val blocksToProduce = 5
    repeat(blocksToProduce) {
      transactionsHelper.run {
        validator1Stack.besuNode.sendTransactionAndAssertExecution(
          logger = log,
          recipient = createAccount("another account"),
          amount = Amount.ether(100),
        )
      }
    }

    val beaconChain = validator1Stack.maruApp.beaconChain
    waitForBlocksToBeProduced(blocksToProduce, beaconChain)

    // verify that all EL blocks are the same
    checkAllValidatorBlocksAreTheSame(
      getValidator1Blocks = { validator1Stack.besuNode.getMinedBlocks(blocksToProduce) },
      getValidator2Blocks = { validator2Stack.besuNode.getMinedBlocks(blocksToProduce) },
      blocksToMetadata = ::elBlocksToMetadata,
    )

    // verify that all CL blocks are the same
    checkAllValidatorBlocksAreTheSame(
      getValidator1Blocks = { beaconChain.getSealedBeaconBlocks(1uL, blocksToProduce.toULong()) },
      getValidator2Blocks = { beaconChain.getSealedBeaconBlocks(1uL, blocksToProduce.toULong()) },
      blocksToMetadata = ::clBlocksToMetadata,
    )

    checkBlockProposersMatchExpectedProposers(
      beaconChain = beaconChain,
      blocksToProduce = blocksToProduce,
    )
  }

  private fun waitForBlocksToBeProduced(
    blocksToProduce: Int,
    beaconChain: BeaconChain,
  ) {
    await
      .pollDelay(1.seconds.toJavaDuration())
      .timeout(30.seconds.toJavaDuration())
      .untilAsserted {
        assertThat(validator1Stack.besuNode.getMinedBlocks(blocksToProduce))
          .hasSize(blocksToProduce)
        assertThat(beaconChain.getSealedBeaconBlocks(1uL, blocksToProduce.toULong()))
          .hasSize(blocksToProduce)
      }
  }

  private fun <T, M> checkAllValidatorBlocksAreTheSame(
    getValidator1Blocks: () -> List<T>,
    getValidator2Blocks: () -> List<T>,
    blocksToMetadata: (List<T>) -> List<M>,
  ) {
    val blocksProducedByQbftValidator1 = blocksToMetadata(getValidator1Blocks())
    val blocksProducedByQbftValidator2 = blocksToMetadata(getValidator2Blocks())
    assertThat(blocksProducedByQbftValidator2)
      .isEqualTo(blocksProducedByQbftValidator1)
  }

  private fun checkBlockProposersMatchExpectedProposers(
    beaconChain: BeaconChain,
    blocksToProduce: Int,
  ) {
    val blocks = beaconChain.getSealedBeaconBlocks(1uL, blocksToProduce.toULong())
    val proposerSelector = ProposerSelectorImpl

    blocks.forEachIndexed { index, block ->
      val beaconBlockHeader = block.beaconBlock.beaconBlockHeader
      val roundIdentifier = ConsensusRoundIdentifier(beaconBlockHeader.number.toLong(), beaconBlockHeader.round.toInt())
      val parentBeaconState = beaconChain.getBeaconState(beaconBlockHeader.number - 1uL)
      val expectedProposer = proposerSelector.getProposerForBlock(parentBeaconState!!, roundIdentifier).get()

      assertThat(beaconBlockHeader.proposer)
        .withFailMessage {
          "Block ${beaconBlockHeader.number} should be proposed by ${expectedProposer.address.encodeHex()} " +
            "but was proposed by ${beaconBlockHeader.proposer.address.encodeHex()}"
        }.isEqualTo(expectedProposer)
    }
  }

  private fun elBlocksToMetadata(blocks: List<EthBlock.Block>): List<Pair<BigInteger, String>> =
    blocks.map {
      it.number to it.hash
    }

  private fun clBlocksToMetadata(blocks: List<SealedBeaconBlock>): List<Pair<ULong, String>> =
    blocks.map {
      it.beaconBlock.beaconBlockHeader.number to
        it.beaconBlock.beaconBlockHeader.hash
          .encodeHex()
    }
}
