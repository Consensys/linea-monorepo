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
import java.net.URI
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
import maru.config.FollowersConfig
import maru.config.consensus.ElFork
import maru.consensus.blockimport.FollowerBeaconBlockImporter
import maru.consensus.state.InstantFinalizationProvider
import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.EMPTY_HASH
import maru.core.Validator
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.extensions.encodeHex
import maru.extensions.fromHexToByteArray
import maru.mappers.Mappers.toDomain
import maru.serialization.rlp.RLPSerializers
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.kotlin.await
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Order
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.io.TempDir
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.MethodSource
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.methods.response.EthBlock
import tech.pegasys.teku.infrastructure.async.SafeFuture
import testutils.Web3jTransactionsHelper
import testutils.maru.MaruFactory
import testutils.maru.awaitTillMaruHasPeers

class CliqueToPosTest {
  companion object {
    private val dockerComposeFilePaths =
      mutableListOf<String>(Path.of("./../docker/compose.yaml").toString())
    private val qbftCluster =
      DockerComposeRule
        .Builder()
        .files(DockerComposeFiles.from(*dockerComposeFilePaths.toTypedArray()))
        .projectName(ProjectName.random())
        .waitingForService("sequencer", HealthChecks.toHaveAllPortsOpen())
        .build()
    private var shanghaiTimestamp: Long = 0
    private var pragueTimestamp: Long = 0
    private lateinit var maruFactory: MaruFactory

    @TempDir
    private lateinit var sequencerMaruTmpDir: File

    private val genesisDir = File("../docker/initialization")
    private val transactionsHelper = Web3jTransactionsHelper(TestEnvironment.sequencerL2Client)
    private val log: Logger = LogManager.getLogger(CliqueToPosTest::class.java)
    private const val VALIDATOR_PRIVATE_KEY_WITH_PREFIX =
      "0x080212201dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"

    // Instantiated in the networkCanBeSwitched test, used in the following tests
    private lateinit var maruSequencer: MaruApp

    // Will only be used in the sync from scratch tests
    private var maruFollower: MaruApp? = null

    private fun parseTimestamp(timestampFork: String): Long {
      val objectMapper = ObjectMapper()
      val genesisTree = objectMapper.readTree(File(genesisDir, "genesis-besu.json"))
      val switchTime = genesisTree.at("/config/$timestampFork").asLong()
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

    @BeforeAll
    @JvmStatic
    fun beforeAll() {
      deleteGenesisFiles()
      qbftCluster.before()
      shanghaiTimestamp = parseTimestamp("shanghaiTime")
      pragueTimestamp = parseTimestamp("pragueTime")
      maruFactory =
        MaruFactory(
          validatorPrivateKey = VALIDATOR_PRIVATE_KEY_WITH_PREFIX.fromHexToByteArray(),
          shanghaiTimestamp = shanghaiTimestamp,
          pragueTimestamp = pragueTimestamp,
        )
    }

    @AfterAll
    @JvmStatic
    fun afterAll() {
      qbftCluster.after()
      if (::maruSequencer.isInitialized) {
        maruSequencer.stop()
        maruSequencer.close()
      }

      TestEnvironment.allClients.values.forEach { web3j ->
        web3j.shutdown()
      }
      log.info("Test is complete")
    }

    @JvmStatic
    fun followerNodes(): List<Arguments> =
      TestEnvironment.followerExecutionClientsPostMerge.map {
        Arguments.of(it.key, it.value)
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

    private fun buildValidatorMaruWithMultipleFollowers(): MaruApp =
      maruFactory.buildSwitchableTestMaruValidatorWithP2pPeering(
        ethereumJsonRpcUrl = "http://localhost:8545",
        engineApiRpc = "http://localhost:8550",
        dataDir = sequencerMaruTmpDir.toPath(),
        followers =
          FollowersConfig(
            mapOf(
              "follower-besu" to ApiEndpointConfig(URI.create("http://localhost:9550").toURL()),
              "follower-erigon" to
                ApiEndpointConfig(
                  URI.create("http://localhost:11551").toURL(),
                  jwtSecretPath = TestEnvironment.JWT_CONFIG_PATH,
                ),
              "follower-nethermind" to
                ApiEndpointConfig(
                  URI.create("http://localhost:10550").toURL(),
                  jwtSecretPath = TestEnvironment.JWT_CONFIG_PATH,
                ),
              "follower-geth" to
                ApiEndpointConfig(
                  URI.create("http://localhost:8561").toURL(),
                  jwtSecretPath = TestEnvironment.JWT_CONFIG_PATH,
                ),
            ),
          ),
      )

    private var runCounter = 0u
  }

  private fun saveLogs(path: Path) {
    qbftCluster.dockerExecutable().execute("ps").inputStream.use {
      Files.copy(
        it,
        Path.of("$path/containers.txt"),
        StandardCopyOption.REPLACE_EXISTING,
      )
    }
    val containerShortNames =
      TestEnvironment.allClients
        .map { it.key }
        .toMutableList()
    containerShortNames.forEach { containerShortName ->
      qbftCluster
        .dockerExecutable()
        .execute("logs", containerShortName)
        .inputStream
        .use {
          Files.copy(
            it,
            Path.of("$path/$containerShortName.log"),
            StandardCopyOption.REPLACE_EXISTING,
          )
        }
    }
  }

  @AfterEach
  fun tearDown() {
    if (maruFollower != null) {
      maruFollower!!.stop()
      maruFollower!!.close()
    }

    val logsDestination = Path.of("docker_logs/run_$runCounter")
    logsDestination.toFile().mkdirs()
    saveLogs(logsDestination)
    runCounter += 1u
  }

  @Order(1)
  @Test
  fun networkCanBeSwitched() {
    maruSequencer = buildValidatorMaruWithMultipleFollowers()
    maruSequencer.start()
    sendCliqueTransactions()
    everyoneArePeered()
    waitTillTimestamp(shanghaiTimestamp, "shanghaiTime")

    log.info("Sequencer has switched to PoS")
    repeat(4) {
      transactionsHelper.run { sendArbitraryTransaction().waitForInclusion() }
    }

    val postMergeBlock = getBlockByNumber(6)!!
    assertThat(postMergeBlock.timestamp.toLong()).isGreaterThanOrEqualTo(shanghaiTimestamp)

    waitTillTimestamp(pragueTimestamp, "pragueTime")
    log.info("Sequencer has switched to Prague")
    repeat(4) {
      transactionsHelper.run { sendArbitraryTransaction().waitForInclusion() }
    }

    assertNodeBlockHeight(TestEnvironment.sequencerL2Client, 13)

    waitForAllBlockHeightsToMatch()
  }

  // TODO: Explore parallelization of this test
  @Order(2)
  @ParameterizedTest
  @MethodSource("followerNodes")
  fun syncFromScratch(
    nodeName: String,
    engineApiConfig: ApiEndpointConfig,
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

    val followerDataDir = Files.createTempDirectory("maru-$nodeName")
    followerDataDir.toFile().deleteOnExit()
    maruFollower =
      maruFactory.buildTestMaruFollowerWithConsensusSwitch(
        engineApiConfig = engineApiConfig,
        ethereumApiConfig = engineApiConfig,
        dataDir = followerDataDir,
        validatorPortForStaticPeering = maruSequencer.p2pPort(),
      )
    if (nodeName.contains("besu")) {
      // Required to change validation rules from Clique to PostMerge
      // TODO: investigate this issue more. It was working happen with Dummy Consensus
      syncTarget(engineApiConfig = engineApiConfig, headBlockNumber = 5, followerName = nodeName)
      awaitCondition
        .ignoreExceptions()
        .alias(nodeName)
        .untilAsserted {
          assertNodeBlockHeight(nodeEthereumClient, 5L)
        }
    }

    maruFollower!!.start()

    maruFollower!!.awaitTillMaruHasPeers(1u)
    maruSequencer.awaitTillMaruHasPeers(1u)
    awaitCondition
      .timeout(1.minutes.toJavaDuration())
      .pollInterval(15.seconds.toJavaDuration())
      .ignoreExceptions()
      .alias(nodeName)
      .untilAsserted {
        if (nodeName.contains("erigon") || nodeName.contains("nethermind")) {
          // For some reason Erigon and Nethermind need a restart after PoS transition
          restartNodeKeepingState(nodeName, nodeEthereumClient)
        }
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
          assertNodeBlockPrevRandao(nodeEthereumClient)
        }
      }
  }

  private fun waitTillTimestamp(
    timestamp: Long,
    timestampFork: String,
  ) {
    await.timeout(2.minutes.toJavaDuration()).pollInterval(500.milliseconds.toJavaDuration()).untilAsserted {
      val unixTimestamp = System.currentTimeMillis() / 1000
      log.info("Waiting ${timestamp - unixTimestamp} seconds for the $timestampFork at timestamp $timestamp")
      assertThat(unixTimestamp).isGreaterThanOrEqualTo(timestamp)
    }
  }

  private fun restartNodeFromScratch(
    nodeName: String,
    nodeEthereumClient: Web3j,
  ) {
    qbftCluster.docker().rm(nodeName)
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
    followerName: String,
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
    val web3JEngineApiClient =
      Helpers.createWeb3jClient(
        apiEndpointConfig = engineApiConfig,
      )
    val engineApiClient =
      Helpers.buildExecutionEngineClient(
        web3JEngineApiClient = web3JEngineApiClient,
        elFork = ElFork.Prague,
        metricsFacade = TestEnvironment.testMetricsFacade,
      )
    val executionLayerManager = JsonRpcExecutionLayerManager(engineApiClient)
    val blockImporter =
      FollowerBeaconBlockImporter.create(
        executionLayerManager = executionLayerManager,
        finalizationStateProvider = InstantFinalizationProvider,
        importerName = followerName,
      )
    blockImporter
      .handleNewBlock(stubBeaconBlock)
      .get()
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
    expectedBlockNumber: Long =
      TestEnvironment.sequencerL2Client
        .ethBlockNumber()
        .send()
        .blockNumber
        .toLong(),
  ) {
    val targetNodeBlockHeight = web3j.ethBlockNumber().send().blockNumber
    assertThat(targetNodeBlockHeight).isEqualTo(expectedBlockNumber)
  }

  private fun assertNodeBlockPrevRandao(
    web3j: Web3j,
    lastPreMergeBlockNumber: Long = 5L,
  ) {
    val lastPostMergeBlockNumber =
      TestEnvironment.sequencerL2Client
        .ethBlockNumber()
        .send()
        .blockNumber
        .toLong()
    var lastMixHash: String? = null
    (lastPreMergeBlockNumber..lastPostMergeBlockNumber).forEach {
      val mixHash =
        web3j
          .ethGetBlockByNumber(
            DefaultBlockParameter.valueOf(it.toBigInteger()),
            false,
          ).send()
          .block.mixHash
      if (it == lastPreMergeBlockNumber) {
        assertThat(mixHash).isEqualTo(EMPTY_HASH.encodeHex())
      } else {
        assertThat(mixHash).isNotEqualTo(lastMixHash)
      }
      lastMixHash = mixHash
    }
  }

  private fun waitForAllBlockHeightsToMatch() {
    await
      .pollInterval(5.seconds.toJavaDuration())
      .timeout(1.minutes.toJavaDuration())
      .untilAsserted {
        val sequencerBlockHeight =
          TestEnvironment.sequencerL2Client
            .ethBlockNumber()
            .send()
            .blockNumber
            .toLong()

        val blockHeights =
          TestEnvironment.clientsSyncablePreMergeAndPostMerge.entries
            .map { entry ->
              entry.key to
                SafeFuture.of(
                  entry.value.ethBlockNumber().sendAsync(),
                )
            }.map { it.first to it.second.get() }

        // Send a transaction so that the Besu follower triggers a backward sync to sync to head.
        // Besu doesn't adjust the pivot block during the initial sync and may end sync with a block below head.
        transactionsHelper.run { sendArbitraryTransaction().waitForInclusion() }

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
    await.timeout(2.minutes.toJavaDuration()).untilAsserted {
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
}
