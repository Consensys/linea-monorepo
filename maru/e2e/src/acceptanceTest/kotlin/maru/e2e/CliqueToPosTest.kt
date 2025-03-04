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
package maru.e2e

import com.fasterxml.jackson.databind.ObjectMapper
import com.palantir.docker.compose.DockerComposeRule
import com.palantir.docker.compose.configuration.ProjectName
import com.palantir.docker.compose.connection.waiting.HealthChecks
import java.io.File
import java.math.BigInteger
import java.nio.file.Path
import java.util.Optional
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
import maru.e2e.Mappers.executionPayloadV3FromBlock
import maru.e2e.TestEnvironment.createWeb3jClient
import maru.e2e.TestEnvironment.waitForInclusion
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.kotlin.await
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.DefaultBlockParameterName
import org.web3j.protocol.core.methods.response.EthBlock
import tech.pegasys.teku.ethereum.executionclient.auth.JwtConfig
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV3
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

class CliqueToPosTest {
  companion object {
    private val qbftCluster =
      DockerComposeRule
        .Builder()
        .file(Path.of("./../docker/compose.yaml").toString())
        .projectName(ProjectName.random())
        .waitingForService("sequencer", HealthChecks.toHaveAllPortsOpen())
        .build()
    private val maru = MaruFactory.buildTestMaru()

    @BeforeAll
    @JvmStatic
    fun beforeAll() {
      qbftCluster.before()
      maru.start()
    }

    @AfterAll
    @JvmStatic
    fun afterAll() {
      qbftCluster.after()
      maru.stop()
    }
  }

  private val log: Logger = LogManager.getLogger(CliqueToPosTest::class.java)
  private val besuFollowerExecutionEngineClient = createExecutionClient("http://localhost:9550")
  private val nethermindFollowerExecutionEngineClient =
    createExecutionClient(
      "http://localhost:10550",
      TestEnvironment.jwtConfig,
    )
  private val erigonFollowerExecutionEngineClient =
    createExecutionClient(
      "http://localhost:11551",
      TestEnvironment.jwtConfig,
    )
  private val geth1ExecutionEngineClient = createExecutionClient("http://localhost:8561", TestEnvironment.jwtConfig)
  private val geth2ExecutionEngineClient = createExecutionClient("http://localhost:8571", TestEnvironment.jwtConfig)
  private val gethSnapServerExecutionEngineClient =
    createExecutionClient("http://localhost:8581", TestEnvironment.jwtConfig)
  private val gethExecutionEngineClients =
    mapOf(
      "follower-geth" to geth1ExecutionEngineClient,
      "follower-geth-2" to geth2ExecutionEngineClient,
      "follower-geth-snap-server" to gethSnapServerExecutionEngineClient,
    )
  private val followerExecutionEngineClients =
    mapOf(
      "follower-besu" to besuFollowerExecutionEngineClient,
      "follower-erigon" to erigonFollowerExecutionEngineClient,
      "follower-nethermind" to nethermindFollowerExecutionEngineClient,
    ) + gethExecutionEngineClients

  @Test
  fun networkCanBeSwitched() {
    sealPreMergeBlocks()
    everyoneArePeered()
    val newBlockTimestamp = UInt64.valueOf(parseCancunTimestamp())

    val preMergeBlock =
      TestEnvironment.sequencerL2Client
        .ethGetBlockByNumber(DefaultBlockParameterName.LATEST, false)
        .send()
    val lastPreMergeBlockHash = preMergeBlock.block.hash

    await
      .timeout(1.minutes.toJavaDuration())
      .pollInterval(5.seconds.toJavaDuration())
      .untilAsserted {
        val unixTimestamp = System.currentTimeMillis() / 1000
        log.info(
          "Waiting for Cancun switch {} seconds until the switch ",
          { newBlockTimestamp.longValue() - unixTimestamp },
        )
        assertThat(unixTimestamp).isGreaterThan(newBlockTimestamp.longValue())
      }

    fcuFollowersToBlockHash(lastPreMergeBlockHash)
    waitForAllBlockHeightsToMatch()

    log.info("Marked last pre merge block as finalized")

    // Next block's content
    val firstPosBlockTransaction = TestEnvironment.sendArbitraryTransaction()

    log.info("Sequencer has switched to PoS")

    firstPosBlockTransaction.waitForInclusion()

    val postMergeBlock =
      TestEnvironment.sequencerL2Client
        .ethGetBlockByNumber(DefaultBlockParameter.valueOf(BigInteger.valueOf(6)), true)
        .send()

    val blockHash = postMergeBlock.block.hash
    val getNewPayloadFromLastBlockNumber = executionPayloadV3FromBlock(postMergeBlock.block)
    sendNewPayloadToFollowers(getNewPayloadFromLastBlockNumber)
    fcuFollowersToBlockHash(blockHash)
    val postPreMergeBlockHash = postMergeBlock.block.hash

    await.untilAsserted {
      fcuFollowersToBlockHash(postPreMergeBlockHash)
      waitForAllBlockHeightsToMatch()
    }
  }

  // This is more of a debug method rather than an independent test. Thus disabling it
  @Disabled
  fun runFCU() {
    val headBlockNumber = 6L
    val currentBLock = getBlockByNumber(headBlockNumber, true)
    val blockHash = currentBLock.hash

    val getNewPayloadFromLastBlockNumber = executionPayloadV3FromBlock(currentBLock)
    sendNewPayloadToFollowers(getNewPayloadFromLastBlockNumber)
    fcuFollowersToBlockHash(blockHash)
  }

  // This is more of a debug method rather than an independent test. Useful to test if a node can sync from scratch
  @Disabled
  fun fullSync() {
    val target = geth1ExecutionEngineClient

    val lastPreMergeBlockNumber = 5L

    val headBlockNumber = 6L
    for (blockNumber in lastPreMergeBlockNumber + 1..headBlockNumber) {
      val block = getBlockByNumber(blockNumber, true)

      val newPayloadV3 = executionPayloadV3FromBlock(block)
      target.newPayloadV3(newPayloadV3, emptyList(), Bytes32.ZERO).get()
    }

    val currentBLock =
      TestEnvironment.sequencerL2Client
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf(BigInteger.valueOf(headBlockNumber)),
          false,
        ).send()
    val blockHash = currentBLock.block.hash

    val lastBlockHashBytes = Bytes32.fromHexString(blockHash)
    target.forkChoiceUpdatedV3(
      ForkChoiceStateV1(lastBlockHashBytes, lastBlockHashBytes, lastBlockHashBytes),
      Optional.empty(),
    )
  }

  private fun parseCancunTimestamp(): Long {
    val objectMapper = ObjectMapper()
    val genesisTree = objectMapper.readTree(File("../docker/initialization/genesis-besu.json"))
    return genesisTree.at("/config/shanghaiTime").asLong()
  }

  private fun sealPreMergeBlocks() {
    val sequencerBlock = TestEnvironment.sequencerL2Client.ethBlockNumber().send()
    if (sequencerBlock.blockNumber >= BigInteger.valueOf(5)) {
      return
    }
    repeat(5) { TestEnvironment.sendArbitraryTransaction().waitForInclusion() }
  }

  private fun createExecutionClient(
    eeEndpoint: String,
    jwtConfig: Optional<JwtConfig> = Optional.empty(),
  ): Web3JExecutionEngineClient = Web3JExecutionEngineClient(createWeb3jClient(eeEndpoint, jwtConfig))

  @Disabled
  fun listBlockHeights() {
    val blockHeights =
      TestEnvironment.followerClients.entries.map { entry ->
        entry.key to SafeFuture.of(entry.value.ethBlockNumber().sendAsync())
      }

    blockHeights.forEach { log.info("${it.first} block height is ${it.second.get().blockNumber}") }
  }

  private fun waitForAllBlockHeightsToMatch() {
    val sequencerBlockHeight = TestEnvironment.sequencerL2Client.ethBlockNumber().send()

    await.untilAsserted {
      val blockHeights =
        TestEnvironment.followerClients.entries
          .map { entry ->
            entry.key to
              SafeFuture.of(
                entry.value.ethBlockNumber().sendAsync(),
              )
          }.map { it.first to it.second.get() }

      blockHeights.forEach {
        assertThat(it.second.blockNumber)
          .withFailMessage {
            "Block height doesn't match for ${it.first}. Found ${it.second.blockNumber} " +
              "while expecting ${sequencerBlockHeight.blockNumber}."
          }.isEqualTo(sequencerBlockHeight.blockNumber)
      }
    }
  }

  private fun everyoneArePeered() {
    log.info("Call add peer on all nodes and wait for peering to happen.")
    await.pollInterval(1.seconds.toJavaDuration()).timeout(1.minutes.toJavaDuration()).untilAsserted {
      TestEnvironment.followerClients.forEach {
        try {
          it.value
            .adminAddPeer(
              "enode://14408801a444dafc44afbccce2eb755f902aed3b5743fed787b3c790e021fef28b8c827ed896aa4e8fb46e2" +
                "2bd67c39f994a73768b4b382f8597b0d44370e15d@11.11.11.101:30303",
            ).send()
        } catch (e: Exception) {
          if (it.key.contains("nethermind")) {
            log.debug("Nethermind returns response to admin_addPeer that is incompatible with Web3J")
          } else {
            throw e
          }
        }
        val peersResult =
          it.value
            .adminPeers()
            .send()
            .result
        val peers = peersResult.size
        log.info("Peers from node ${it.key}: $peers")
        assertThat(peers).withFailMessage("${it.key} isn't peered! Peers: $peersResult").isGreaterThan(0)
      }
    }
  }

  private fun getBlockByNumber(
    blockNumber: Long,
    retreiveTransactions: Boolean = false,
  ): EthBlock.Block =
    TestEnvironment.sequencerL2Client
      .ethGetBlockByNumber(
        DefaultBlockParameter.valueOf(BigInteger.valueOf(blockNumber)),
        retreiveTransactions,
      ).send()
      .block

  private fun fcuFollowersToBlockHash(blockHash: String) {
    val lastBlockHashBytes = Bytes32.fromHexString(blockHash)
    // Cutting off the merge first
    val mergeForkChoiceState = ForkChoiceStateV1(lastBlockHashBytes, lastBlockHashBytes, lastBlockHashBytes)

    followerExecutionEngineClients.entries
      .map {
        it.key to
          it.value.forkChoiceUpdatedV3(
            mergeForkChoiceState,
            Optional.empty(),
          )
      }.forEach {
        log.info("FCU for block hash $blockHash, node: ${it.first} response: ${it.second.get()}")
      }
  }

  private fun sendNewPayloadToFollowers(newPayloadV3: ExecutionPayloadV3) {
    followerExecutionEngineClients.entries
      .map {
        it.key to
          it.value.newPayloadV3(
            newPayloadV3,
            emptyList(),
            Bytes32.ZERO,
          )
      }.forEach { log.info("New payload for node: ${it.first} response: ${it.second.get()}") }
  }
}
