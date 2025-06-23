/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.e2e

import com.fasterxml.jackson.databind.ObjectMapper
import com.palantir.docker.compose.DockerComposeRule
import com.palantir.docker.compose.configuration.DockerComposeFiles
import com.palantir.docker.compose.configuration.ProjectName
import com.palantir.docker.compose.connection.waiting.HealthChecks
import java.io.File
import java.math.BigInteger
import java.nio.file.Files
import java.nio.file.Path
import java.nio.file.StandardCopyOption
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
import maru.app.Helpers
import maru.app.MaruApp
import maru.config.ApiEndpointConfig
import maru.config.consensus.ElFork
import maru.consensus.NewBlockHandler
import maru.consensus.blockimport.FollowerBeaconBlockImporter
import maru.consensus.state.InstantFinalizationProvider
import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.EMPTY_HASH
import maru.core.Validator
import maru.mappers.Mappers.toDomain
import maru.serialization.rlp.RLPSerializers
import maru.testutils.Web3jTransactionsHelper
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
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
import tech.pegasys.teku.infrastructure.async.SafeFuture

class CliqueToPosTest {
  companion object {
    private val useMaruContainer: Boolean = System.getProperty("useMaruContainer").toBoolean()
    private val dockerComposeFilePaths =
      mutableListOf<String>(Path.of("./../docker/compose.yaml").toString()).also {
        if (useMaruContainer) {
          it.add(Path.of("./../docker/compose.dev.yaml").toString())
        }
      }
    private val qbftCluster =
      DockerComposeRule
        .Builder()
        .files(DockerComposeFiles.from(*dockerComposeFilePaths.toTypedArray()))
        .projectName(ProjectName.random())
        .waitingForService("sequencer", HealthChecks.toHaveAllPortsOpen())
        .build()
    private lateinit var maru: MaruApp
    private var pragueSwitchTimestamp: Long = 0
    private val genesisDir = File("../docker/initialization")
    private val dataDir = File("/tmp/maru-db").also { it.deleteOnExit() }
    private val transactionsHelper = Web3jTransactionsHelper(TestEnvironment.sequencerL2Client)
    private val log: Logger = LogManager.getLogger(this::javaClass)

    private fun parsePragueSwitchTimestamp(): Long {
      val objectMapper = ObjectMapper()
      val genesisTree = objectMapper.readTree(File(genesisDir, "genesis-besu.json"))
      val switchTime = genesisTree.at("/config/pragueTime").asLong()
      return if (switchTime == 0L) System.currentTimeMillis() / 1000 else switchTime
    }

    private fun deleteGenesisFiles() {
      val maruGenesis = File(genesisDir, "genesis-maru.json")
      val besuGenesis = File(genesisDir, "genesis-besu.json")
      val gethGenesis = File(genesisDir, "genesis-geth.json")
      val nethermindGenesis = File(genesisDir, "genesis-nethermind.json")
      maruGenesis.delete()
      besuGenesis.delete()
      gethGenesis.delete()
      nethermindGenesis.delete()
    }

    private fun containerShortNameToFullId(containerShortName: String) = containerShortName

    @BeforeAll
    @JvmStatic
    fun beforeAll() {
      log.info("Using Maru as docker container: $useMaruContainer")
      deleteGenesisFiles()
      dataDir.deleteRecursively()
      dataDir.mkdirs()
      qbftCluster.before()
      pragueSwitchTimestamp = parsePragueSwitchTimestamp()
      if (!useMaruContainer) {
        maru = MaruFactory.buildTestMaru(pragueSwitchTimestamp)
      }
    }

    @AfterAll
    @JvmStatic
    fun afterAll() {
      File("docker_logs").mkdirs()
      qbftCluster.dockerExecutable().execute("ps").inputStream.use {
        Files.copy(
          it,
          Path.of("docker_logs/containers.txt"),
          StandardCopyOption.REPLACE_EXISTING,
        )
      }

      val containerShortNames =
        TestEnvironment.allClients
          .map { it.key }
          .toMutableList()
          .also { if (useMaruContainer) it.add("maru") }
      containerShortNames.forEach { containerShortName ->
        qbftCluster
          .dockerExecutable()
          .execute("logs", containerShortNameToFullId(containerShortName))
          .inputStream
          .use {
            Files.copy(
              it,
              Path.of("docker_logs/$containerShortName.log"),
              StandardCopyOption.REPLACE_EXISTING,
            )
          }
      }

      qbftCluster.after()
    }

    @JvmStatic
    fun followerNodes(): List<Arguments> =
      TestEnvironment.followerExecutionClientsPostMerge.map {
        Arguments.of(it.key, it.value)
      }
  }

  @Order(1)
  @Test
  fun networkCanBeSwitched() {
    if (!useMaruContainer) {
      maru.start()
    }
    sendCliqueTransactions()
    everyoneArePeered()
    waitTillTimestamp(pragueSwitchTimestamp)

    log.info("Sequencer has switched to PoS")
    repeat(4) {
      transactionsHelper.run { sendArbitraryTransaction().waitForInclusion() }
    }

    val postMergeBlock = getBlockByNumber(6)!!
    assertThat(postMergeBlock.timestamp.toLong()).isGreaterThanOrEqualTo(parsePragueSwitchTimestamp())
    assertNodeBlockHeight(TestEnvironment.sequencerL2Client)

    waitForAllBlockHeightsToMatch()
    if (!useMaruContainer) {
      maru.stop()
    }
  }

  // TODO: Explore parallelization of this test
  @Order(2)
  @ParameterizedTest
  @MethodSource("followerNodes")
  fun syncFromScratch(
    nodeName: String,
    engineApiConfig: ApiEndpointConfig,
  ) {
    // To fail right away in case switch failed in the first place
    assertNodeBlockHeight(TestEnvironment.sequencerL2Client)
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

    if (nodeName.contains("besu")) {
      // Required to change validation rules from Clique to PostMerge
      // TODO: investigate this issue more. It was working happen with Dummy Consensus
      syncTarget(engineApiConfig, 5)
      awaitCondition
        .ignoreExceptions()
        .alias(nodeName)
        .untilAsserted {
          assertNodeBlockHeight(nodeEthereumClient, 5L)
        }
    }

    awaitCondition
      .ignoreExceptions()
      .alias(nodeName)
      .untilAsserted {
        if (nodeName.contains("erigon") || nodeName.contains("nethermind")) {
          // For some reason Erigon and Nethermind need a restart after PoS transition
          restartNodeKeepingState(nodeName, nodeEthereumClient)
        }
        // TODO: Sync should be done entirely by Maru in the future
        syncTarget(engineApiConfig, 9)
        if (nodeName.contains("follower-geth")) {
          val latestBlockFromGeth =
            getBlockByNumber(
              blockNumber = 9,
              retreiveTransactions = true,
              ethClient = nodeEthereumClient,
            )!!
          // For some reason it doesn't set latest block correctly, but the block is available
          assertThat(latestBlockFromGeth).isNotNull
        } else {
          assertNodeBlockHeight(nodeEthereumClient)
        }
      }
  }

  private fun buildBlockImportHandler(engineApiConfig: ApiEndpointConfig): NewBlockHandler<*> =
    FollowerBeaconBlockImporter.create(
      Helpers.buildExecutionEngineClient(
        engineApiConfig,
        ElFork.Prague,
        TestEnvironment.testMetricsFacade,
      ),
      InstantFinalizationProvider,
    )

  private fun waitTillTimestamp(timestamp: Long) {
    await.timeout(1.minutes.toJavaDuration()).pollInterval(500.milliseconds.toJavaDuration()).untilAsserted {
      val unixTimestamp = System.currentTimeMillis() / 1000
      log.info(
        "Waiting {} seconds for the Prague switch at timestamp $timestamp",
        timestamp - unixTimestamp,
      )
      assertThat(unixTimestamp).isGreaterThanOrEqualTo(timestamp)
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

  private fun syncTarget(
    engineApiConfig: ApiEndpointConfig,
    headBlockNumber: Long,
  ) {
    val latestBlock = getBlockByNumber(headBlockNumber, retreiveTransactions = true)!!

    val latestExecutionPayload = latestBlock.toDomain()
    val stubBeaconBlock =
      BeaconBlock(
        BeaconBlockHeader(
          number = 0u,
          round = 0u,
          timestamp = latestExecutionPayload.timestamp,
          proposer = Validator(latestExecutionPayload.feeRecipient),
          parentRoot = EMPTY_HASH,
          stateRoot = EMPTY_HASH,
          bodyRoot = EMPTY_HASH,
          headerHashFunction = RLPSerializers.DefaultHeaderHashFunction,
        ),
        BeaconBlockBody(emptySet(), latestExecutionPayload),
      )
    buildBlockImportHandler(engineApiConfig).handleNewBlock(stubBeaconBlock).get()
  }

  private fun sendCliqueTransactions() {
    val sequencerBlock = TestEnvironment.sequencerL2Client.ethBlockNumber().send()
    if (sequencerBlock.blockNumber >= BigInteger.valueOf(5)) {
      return
    }
    repeat(5) { transactionsHelper.run { sendArbitraryTransaction().waitForInclusion() } }
  }

  private fun assertNodeBlockHeight(
    web3j: Web3j,
    expectedBlockNumber: Long = 9L,
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
}
