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
import java.nio.file.Files
import java.nio.file.Path
import java.nio.file.StandardCopyOption
import java.util.Optional
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
import maru.app.MaruApp
import maru.e2e.Mappers.executionPayloadV3FromBlock
import maru.e2e.TestEnvironment.waitForInclusion
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.kotlin.await
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Order
import org.junit.jupiter.api.Test
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.MethodSource
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.methods.response.EthBlock
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV3
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.infrastructure.async.SafeFuture

class CliqueToPosTest {
  companion object {
    private val qbftCluster =
      DockerComposeRule
        .Builder()
        .file(Path.of("./../docker/compose.yaml").toString())
        .projectName(ProjectName.random())
        .waitingForService("sequencer", HealthChecks.toHaveAllPortsOpen())
        .build()
    private lateinit var maru: MaruApp
    private var pragueSwitchTimestamp: Long = 0
    private val genesisDir = File("../docker/initialization")

    private fun parsePragueSwitchTimestamp(): Long {
      val objectMapper = ObjectMapper()
      val genesisTree = objectMapper.readTree(File(genesisDir, "genesis-besu.json"))
      val switchTime = genesisTree.at("/config/pragueTime").asLong()
      return if (switchTime == 0L) System.currentTimeMillis() / 1000 else switchTime
    }

    private fun deleteGenesisFiles() {
      val besuGenesis = File(genesisDir, "genesis-besu.json")
      val gethGenesis = File(genesisDir, "genesis-geth.json")
      val nethermindGenesis = File(genesisDir, "genesis-nethermind.json")
      besuGenesis.delete()
      gethGenesis.delete()
      nethermindGenesis.delete()
    }

    @BeforeAll
    @JvmStatic
    fun beforeAll() {
      deleteGenesisFiles()
      qbftCluster.before()

      pragueSwitchTimestamp = parsePragueSwitchTimestamp()
      maru = MaruFactory.buildTestMaru(pragueSwitchTimestamp)
    }

    @AfterAll
    @JvmStatic
    fun afterAll() {
      File("docker_logs").mkdirs()
      TestEnvironment.allClients.forEach {
        val containerShortName = it.key
        val logsInputStream =
          qbftCluster.dockerExecutable().execute("logs", containerShortNameToFullId(containerShortName)).inputStream
        Files.copy(
          logsInputStream,
          Path.of("docker_logs/$containerShortName.log"),
          StandardCopyOption.REPLACE_EXISTING,
        )
      }

      qbftCluster.after()
    }

    private fun containerShortNameToFullId(containerShortName: String) =
      "${qbftCluster.projectName().asString()}-$containerShortName-1"

    private val log: Logger = LogManager.getLogger(CliqueToPosTest::class.java)

    @JvmStatic
    fun followerNodes(): List<Arguments> =
      TestEnvironment.followerExecutionClientsPostMerge.map {
        Arguments.of(it.key, it.value)
      }
  }

  @Order(1)
  @Test
  fun networkCanBeSwitched() {
    maru.start()
    sendCliqueTransactions()
    everyoneArePeered()
    waitTillTimestamp("Prague", pragueSwitchTimestamp)

    repeat(4) {
      TestEnvironment.sendArbitraryTransaction().waitForInclusion()
    }
    log.info("Sequencer has switched to PoS")

    val latestBlock = getBlockByNumber(7)!!
    assertThat(latestBlock.timestamp.toLong()).isGreaterThan(parsePragueSwitchTimestamp())
    (6L..9L).forEach {
      setAllFollowersHeadToBlockNumberPrague(it)
    }
    assertNodeBlockHeight(9, TestEnvironment.sequencerL2Client)

    waitForAllBlockHeightsToMatch()
    maru.stop()
  }

  @Order(2)
  @ParameterizedTest
  @MethodSource("followerNodes")
  fun syncFromScratch(
    nodeName: String,
    nodeEngineApiClient: Web3JExecutionEngineClient,
  ) {
    val nodeEthereumClient = TestEnvironment.followerClientsPostMerge[nodeName]!!
    restartNodeFromScratch(nodeName, nodeEthereumClient)
    log.info("Container $nodeName restarted")

    val awaitCondition =
      if (nodeName.contains("nethermind")) {
        await
          .pollInterval(10.seconds.toJavaDuration())
          .timeout(60.seconds.toJavaDuration())
      } else {
        await
          .pollInterval(1.seconds.toJavaDuration())
          .timeout(30.seconds.toJavaDuration())
      }

    awaitCondition
      .ignoreExceptions()
      .alias(nodeName)
      .untilAsserted {
        if (nodeName.contains("erigon") || nodeName.contains("nethermind")) {
          // For some reason Erigon needs a restart after PoS transition
          restartNodeKeepingState(nodeName, nodeEthereumClient)
        }
        syncTarget(nodeEngineApiClient, 9)
        if (nodeName.contains("follower-geth")) {
          // For some reason it doesn't set latest block correctly, but the block is available
          val blockByNumber =
            getBlockByNumber(
              blockNumber = 9,
              retreiveTransactions = false,
              ethClient = nodeEthereumClient,
            )
          assertThat(blockByNumber).isNotNull
        } else {
          assertNodeBlockHeight(9, nodeEthereumClient)
        }
      }
  }

  private fun waitTillTimestamp(
    label: String,
    timestamp: Long,
  ) {
    await.timeout(1.minutes.toJavaDuration()).pollInterval(5.seconds.toJavaDuration()).untilAsserted {
      val unixTimestamp = System.currentTimeMillis() / 1000
      log.info(
        "Waiting {} seconds for the {} switch",
        timestamp - unixTimestamp,
        label,
      )
      assertThat(unixTimestamp).isGreaterThan(timestamp)
    }
  }

  private fun restartNodeFromScratch(
    nodeName: String,
    nodeEthereumClient: Web3j,
  ) {
    qbftCluster.docker().rm(containerShortNameToFullId(nodeName))
    qbftCluster.dockerCompose().up()
    awaitExpectedBlockNumberAfterStartup(nodeName, nodeEthereumClient)
  }

  private fun restartNodeKeepingState(
    nodeName: String,
    nodeEthereumClient: Web3j,
  ) {
    log.debug("Restarting $nodeName keeping state")
    val container = qbftCluster.containers().container(nodeName)
    container.stop()
    container.start()
    awaitExpectedBlockNumberAfterStartup(nodeName, nodeEthereumClient)
  }

  private fun awaitExpectedBlockNumberAfterStartup(
    nodeName: String,
    ethereumClient: Web3j,
  ) {
    val expectedBlockNumber =
      when {
        nodeName.contains("erigon") -> 5L
        nodeName.contains("nethermind") -> 5L
        else -> 0L
      }
    await
      .pollInterval(1.seconds.toJavaDuration())
      .timeout(30.seconds.toJavaDuration())
      .ignoreExceptions()
      .alias(nodeName)
      .untilAsserted {
        log.debug("Waiting till the node $nodeName is up and synced")
        assertThat(
          ethereumClient
            .ethBlockNumber()
            .send()
            .blockNumber
            .toLong(),
        ).isGreaterThanOrEqualTo(expectedBlockNumber)
          .withFailMessage("Node is unexpectedly synced after restart! Was its state flushed?")
      }
  }

  private fun sendNewPayloadByBlockNumber(
    blockNumber: Long,
    target: Web3JExecutionEngineClient,
  ): ExecutionPayloadV3 {
    val targetBlock = getBlockByNumber(blockNumber = blockNumber, retreiveTransactions = true)!!
    val blockPayload = executionPayloadV3FromBlock(targetBlock)
    val newPayloadResult = target.newPayloadV4(blockPayload, listOf(), Bytes32.ZERO, emptyList()).get()
    log.debug("New payload result: {}", newPayloadResult)
    return blockPayload
  }

  private fun syncTarget(
    target: Web3JExecutionEngineClient,
    headBlockNumber: Long,
  ) {
    val headPayload = sendNewPayloadByBlockNumber(headBlockNumber, target)

    val fcuResult =
      target
        .forkChoiceUpdatedV1(
          ForkChoiceStateV1(headPayload.blockHash, headPayload.blockHash, headPayload.blockHash),
          Optional.empty(),
        ).get()

    log.debug("Fork choice updated result: {}", fcuResult)
  }

  private fun setAllFollowersHeadToBlockNumberPrague(blockNumber: Long): String {
    val postMergeBlock = getBlockByNumber(blockNumber = blockNumber, retreiveTransactions = true)!!
    val getNewPayloadFromPostMergeBlockNumber = executionPayloadV3FromBlock(postMergeBlock)
    sendNewPayloadToFollowersPrague(getNewPayloadFromPostMergeBlockNumber)
    fcuFollowersToBlockHashPrague(postMergeBlock.hash)
    return postMergeBlock.hash
  }

  private fun sendCliqueTransactions() {
    val sequencerBlock = TestEnvironment.sequencerL2Client.ethBlockNumber().send()
    if (sequencerBlock.blockNumber >= BigInteger.valueOf(5)) {
      return
    }
    repeat(5) { TestEnvironment.sendArbitraryTransaction().waitForInclusion() }
  }

  private fun assertNodeBlockHeight(
    expectedBlockNumber: Long,
    web3j: Web3j,
  ) {
    val targetNodeBlockHeight = web3j.ethBlockNumber().send().blockNumber
    assertThat(targetNodeBlockHeight).isEqualTo(expectedBlockNumber)
  }

  private fun waitForAllBlockHeightsToMatch() {
    val sequencerBlockHeight =
      TestEnvironment.sequencerL2Client
        .ethBlockNumber()
        .send()
        .blockNumber
        .toLong()

    await.untilAsserted {
      val blockHeights =
        TestEnvironment.clientsSyncablePreMergeAndPostMerge.entries
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
              "while expecting $sequencerBlockHeight."
          }.isEqualTo(sequencerBlockHeight)
      }
    }
  }

  private fun everyoneArePeered() {
    log.info("Call add peer on all nodes and wait for peering to happen.")
    await.timeout(1.minutes.toJavaDuration()).untilAsserted {
      TestEnvironment.preMergeFollowerClients.forEach {
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
    ethClient: Web3j = TestEnvironment.sequencerL2Client,
  ): EthBlock.Block? =
    ethClient
      .ethGetBlockByNumber(
        DefaultBlockParameter.valueOf(BigInteger.valueOf(blockNumber)),
        retreiveTransactions,
      ).send()
      .block

  private fun fcuFollowersToBlockHashPrague(blockHash: String) {
    val lastBlockHashBytes = Bytes32.fromHexString(blockHash)
    // Cutting off the merge first
    val mergeForkChoiceState = ForkChoiceStateV1(lastBlockHashBytes, lastBlockHashBytes, lastBlockHashBytes)

    TestEnvironment.followerExecutionClientsPostMerge.entries
      .map {
        it.key to
          it.value.forkChoiceUpdatedV1(
            mergeForkChoiceState,
            Optional.empty(),
          )
      }.forEach {
        log.info("FCU for block hash $blockHash, node: ${it.first} response: ${it.second.get()}")
      }
  }

  private fun sendNewPayloadToFollowersPrague(newPayloadV3: ExecutionPayloadV3) {
    TestEnvironment.followerExecutionClientsPostMerge.entries
      .map {
        it.key to
          it.value.newPayloadV4(
            newPayloadV3,
            emptyList(),
            Bytes32.ZERO,
            emptyList(),
          )
      }.forEach { log.info("New payload for node: ${it.first} response: ${it.second.get()}") }
  }
}
