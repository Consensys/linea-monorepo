/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import io.libp2p.core.PeerId
import io.libp2p.etc.types.fromHex
import java.lang.Thread.sleep
import java.util.concurrent.ExecutionException
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.seconds
import maru.config.P2P
import maru.config.consensus.ElFork
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ConsensusConfig
import maru.consensus.ForkIdHashProvider
import maru.consensus.ForkIdHashProviderImpl
import maru.consensus.ForkIdHasher
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.core.BeaconBlockHeader
import maru.core.BeaconState
import maru.core.SealedBeaconBlock
import maru.core.ext.DataGenerators
import maru.core.ext.metrics.TestMetrics
import maru.crypto.Hashing
import maru.database.BeaconChain
import maru.database.InMemoryBeaconChain
import maru.p2p.messages.Status
import maru.p2p.messages.StatusMessageFactory
import maru.serialization.ForkIdSerializers
import maru.serialization.rlp.RLPSerializers
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatNoException
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.awaitility.Awaitility.await
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.parallel.Execution
import org.junit.jupiter.api.parallel.ExecutionMode
import org.mockito.Mockito.RETURNS_DEEP_STUBS
import org.mockito.Mockito.mock
import org.mockito.kotlin.any
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.eth2.rpc.core.RpcException
import tech.pegasys.teku.networking.eth2.rpc.core.RpcResponseStatus
import tech.pegasys.teku.networking.p2p.libp2p.LibP2PNodeId
import tech.pegasys.teku.networking.p2p.libp2p.MultiaddrPeerAddress
import tech.pegasys.teku.networking.p2p.peer.DisconnectReason
import maru.p2p.ext.DataGenerators as P2P2DataGenerators

@Execution(ExecutionMode.SAME_THREAD)
class P2PTest {
  private val log: Logger = LogManager.getLogger(this.javaClass)

  companion object {
    private val chainId = 1337u

    private const val IPV4 = "127.0.0.1"

    private const val PORT1 = 9234u
    private const val PORT2 = 9235u
    private const val PORT3 = 9236u
    private const val PORT4 = 9237u
    private const val PORT5 = 9238u
    private const val PORT6 = 9239u

    private const val PEER_ID_NODE_1: String = "16Uiu2HAmPRfinavM2jE9BSkCagBGStJ2SEkPPm6fxFVMdCQebzt6"
    private const val PEER_ID_NODE_2: String = "16Uiu2HAmVXtqhevTAJqZucPbR2W4nCMpetrQASgjZpcxDEDaUPPt"
    private const val PEER_ID_NODE_3: String = "16Uiu2HAkzq767a82zfyUz4VLgPbFrxSQBrdmUYxgNDbwgvmjwWo5"

    // TODO: to make these tests reliable it would be good if the ports were not hardcoded, but free ports chosen
    private const val PEER_ADDRESS_NODE_1: String = "/ip4/$IPV4/tcp/$PORT1/p2p/$PEER_ID_NODE_1"
    private const val PEER_ADDRESS_NODE_2: String = "/ip4/$IPV4/tcp/$PORT2/p2p/$PEER_ID_NODE_2"
    private const val PEER_ADDRESS_NODE_3: String = "/ip4/$IPV4/tcp/$PORT3/p2p/$PEER_ID_NODE_3"

    private val key1 = "0802122012c0b113e2b0c37388e2b484112e13f05c92c4471e3ee1dfaa368fa5045325b2".fromHex()
    private val key2 = "0802122100f3d2fffa99dc8906823866d96316492ebf7a8478713a89a58b7385af85b088a1".fromHex()
    private val key3 = "080212204437acb8e84bc346f7640f239da84abe99bc6f97b7855f204e34688d2977fd57".fromHex()
    private val beaconChain = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u))
    private val forkIdHashProvider =
      createForkIdHashProvider()
    private val statusMessageFactory = StatusMessageFactory(beaconChain, forkIdHashProvider)
    private val rpcMethods = createRpcMethods()

    fun createRpcMethods(): RpcMethods {
      val rpcProtocolIdGenerator = LineaRpcProtocolIdGenerator(chainId)
      lateinit var maruPeerManager: MaruPeerManager
      val rpcMethods = RpcMethods(statusMessageFactory, rpcProtocolIdGenerator, { maruPeerManager }, beaconChain)
      val maruPeerFactory =
        DefaultMaruPeerFactory(rpcMethods, statusMessageFactory, P2P(ipAddress = IPV4, port = PORT1))
      maruPeerManager =
        MaruPeerManager(maruPeerFactory = maruPeerFactory, p2pConfig = P2P(ipAddress = IPV4, port = PORT1))
      return rpcMethods
    }

    fun createForkIdHashProvider(): ForkIdHashProvider {
      val consensusConfig: ConsensusConfig =
        QbftConsensusConfig(
          validatorSet =
            setOf(
              DataGenerators.randomValidator(),
              DataGenerators.randomValidator(),
            ),
          elFork = ElFork.Prague,
        )
      val forksSchedule = ForksSchedule(chainId, listOf(ForkSpec(0L, 1, consensusConfig)))

      return ForkIdHashProviderImpl(
        chainId = chainId,
        beaconChain = beaconChain,
        forksSchedule = forksSchedule,
        forkIdHasher = ForkIdHasher(ForkIdSerializers.ForkIdSerializer, Hashing::shortShaHash),
      )
    }

    private fun createP2PNetwork(
      privateKey: ByteArray,
      port: UInt,
      staticPeers: List<String> = emptyList(),
      beaconChain: BeaconChain = Companion.beaconChain,
      statusMessageFactory: StatusMessageFactory = Companion.statusMessageFactory,
      statusUpdate: P2P.StatusUpdateConfig = P2P.StatusUpdateConfig(),
      discovery: P2P.Discovery? = null,
    ): P2PNetworkImpl =
      P2PNetworkImpl(
        privateKeyBytes = privateKey,
        p2pConfig =
          P2P(
            ipAddress = IPV4,
            port = port,
            staticPeers = staticPeers,
            reconnectDelay = 1.seconds,
            statusUpdate = statusUpdate,
            discovery = discovery,
          ),
        chainId = chainId,
        serDe = RLPSerializers.SealedBeaconBlockSerializer,
        metricsFacade = TestMetrics.TestMetricsFacade,
        statusMessageFactory = statusMessageFactory,
        beaconChain = beaconChain,
        metricsSystem = NoOpMetricsSystem(),
        forkIdHashProvider = forkIdHashProvider,
        isBlockImportEnabledProvider = { true },
      )
  }

  @Test
  fun `static peer can be added`() {
    val p2pNetworkImpl1 = createP2PNetwork(privateKey = key1, port = PORT1)
    val p2pNetworkImpl2 = createP2PNetwork(privateKey = key2, port = PORT2)
    try {
      p2pNetworkImpl1.start()

      p2pNetworkImpl2.start()

      p2pNetworkImpl1.addStaticPeer(MultiaddrPeerAddress.fromAddress(PEER_ADDRESS_NODE_2))

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `static peers can be removed`() {
    val p2pNetworkImpl1 = createP2PNetwork(privateKey = key1, port = PORT1)
    val p2pNetworkImpl2 = createP2PNetwork(privateKey = key2, port = PORT2)

    try {
      p2pNetworkImpl1.start()
      p2pNetworkImpl2.start()

      p2pNetworkImpl1.addStaticPeer(MultiaddrPeerAddress.fromAddress(PEER_ADDRESS_NODE_2))

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      p2pNetworkImpl1.removeStaticPeer(MultiaddrPeerAddress.fromAddress(PEER_ADDRESS_NODE_2))

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 0) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 0) }
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `static peers can be configured`() {
    val p2pNetworkImpl1 = createP2PNetwork(privateKey = key1, port = PORT1)
    val p2pNetworkImpl2 = createP2PNetwork(privateKey = key2, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1))
    try {
      p2pNetworkImpl1.start()

      p2pNetworkImpl2.start()

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `static peers reconnect`() {
    val p2pNetworkImpl1 = createP2PNetwork(privateKey = key1, port = PORT1)
    val p2pNetworkImpl2 = createP2PNetwork(privateKey = key2, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1))

    try {
      p2pNetworkImpl1.start()

      p2pNetworkImpl2.start()

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      p2pNetworkImpl1.dropPeer(PEER_ID_NODE_2, DisconnectReason.TOO_MANY_PEERS)

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `two peers can gossip with each other`() {
    val beaconChain2 = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u))
    val p2pNetworkImpl1 = createP2PNetwork(privateKey = key1, port = PORT1)
    val p2pNetworkImpl2 =
      createP2PNetwork(
        privateKey = key2,
        port = PORT2,
        staticPeers = listOf(PEER_ADDRESS_NODE_1),
        beaconChain = beaconChain2,
      )
    try {
      p2pNetworkImpl1.start()

      val blocksReceived = mutableListOf<SealedBeaconBlock>()
      p2pNetworkImpl2.start()
      p2pNetworkImpl2.subscribeToBlocks {
        updateBeaconChainState(beaconChain2, it.beaconBlock.beaconBlockHeader)
        blocksReceived.add(it)
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      val randomBlockMessage1 =
        maru.p2p.ext.DataGenerators
          .randomBlockMessage()
      p2pNetworkImpl1.broadcastMessage(randomBlockMessage1).get()
      val randomBlockMessage2 = P2P2DataGenerators.randomBlockMessage(2UL)
      p2pNetworkImpl1.broadcastMessage(randomBlockMessage2).get()

      awaitUntilAsserted {
        assertThat(blocksReceived).hasSameElementsAs(listOf(randomBlockMessage1.payload, randomBlockMessage2.payload))
      }
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `peer receiving gossip passes message on`() {
    val beaconChain2 = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u))
    val beaconChain3 = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u))
    val p2pNetworkImpl1 = createP2PNetwork(privateKey = key1, port = PORT1, staticPeers = emptyList())
    val p2pNetworkImpl2 =
      createP2PNetwork(
        privateKey = key2,
        port = PORT2,
        staticPeers = listOf(PEER_ADDRESS_NODE_1, PEER_ADDRESS_NODE_3),
        beaconChain = beaconChain2,
      )
    val p2pNetworkImpl3 =
      createP2PNetwork(privateKey = key3, port = PORT3, staticPeers = emptyList(), beaconChain = beaconChain3)
    try {
      p2pNetworkImpl1.start()
      p2pNetworkImpl2.start()
      p2pNetworkImpl2.subscribeToBlocks {
        updateBeaconChainState(beaconChain2, it.beaconBlock.beaconBlockHeader)
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }

      val blockReceived = SafeFuture<SealedBeaconBlock>()
      p2pNetworkImpl3.start()
      p2pNetworkImpl3.subscribeToBlocks {
        updateBeaconChainState(beaconChain3, it.beaconBlock.beaconBlockHeader)
        blockReceived.complete(it)
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }

      awaitUntilAsserted { assertNetworkIsConnectedToPeer(p2pNetworkImpl1, PEER_ID_NODE_2) }
      awaitUntilAsserted { assertNetworkIsConnectedToPeer(p2pNetworkImpl3, PEER_ID_NODE_2) }

      assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1)
      assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 2)
      assertNetworkHasPeers(network = p2pNetworkImpl3, peers = 1)

      sleep(1100L) // to make sure that the peers have communicated that they have subscribed to the topic
      // This sleep can be decreased if the heartbeat is decreased (set to 1s for now, see P2PNetworkFactory) in the GossipRouter

      val randomBlockMessage = P2P2DataGenerators.randomBlockMessage()
      p2pNetworkImpl1.broadcastMessage(randomBlockMessage)

      assertThat(
        blockReceived.get(200, TimeUnit.MILLISECONDS),
      ).isEqualTo(randomBlockMessage.payload)
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
      p2pNetworkImpl3.stop().get()
    }
  }

  private fun updateBeaconChainState(
    beaconChain: BeaconChain,
    beaconBlockHeader: BeaconBlockHeader,
  ) {
    beaconChain
      .newUpdater()
      .putBeaconState(
        beaconChain.getLatestBeaconState().copy(
          latestBeaconBlockHeader = beaconBlockHeader,
        ),
      ).commit()
  }

  @Test
  fun `peer can send a status request`() {
    val p2pNetworkImpl1 = createP2PNetwork(privateKey = key1, port = PORT1)
    val p2pNetworkImpl2 = createP2PNetwork(privateKey = key2, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1))
    try {
      p2pNetworkImpl1.start()
      p2pNetworkImpl2.start()

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      val latestBeaconBlockHeader = beaconChain.getLatestBeaconState().latestBeaconBlockHeader
      val expectedStatus =
        Status(
          forkIdHash = forkIdHashProvider.currentForkIdHash(),
          latestStateRoot = latestBeaconBlockHeader.hash,
          latestBlockNumber = latestBeaconBlockHeader.number,
        )
      val peer1 =
        p2pNetworkImpl2.getPeerLookup().getPeer(LibP2PNodeId(PeerId.fromBase58(PEER_ID_NODE_1)))
          ?: throw IllegalStateException("Peer with ID $PEER_ID_NODE_1 not found in p2pNetworkImpl2")
      val maruPeer1 =
        DefaultMaruPeer(peer1, rpcMethods, statusMessageFactory, p2pConfig = P2P(ipAddress = IPV4, port = PORT1))

      val responseFuture = maruPeer1.sendStatus()

      assertThatNoException().isThrownBy { responseFuture.get(500L, TimeUnit.MILLISECONDS) }
      assertThat(peer1.getStatus()).isEqualTo(expectedStatus)
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `peer send a status request and receive exception when callee throws error`() {
    val p2pNetworkImpl1 =
      createP2PNetwork(
        privateKey = key1,
        port = PORT1,
        statusMessageFactory =
          StatusMessageFactory(
            beaconChain = beaconChain,
            forkIdHashProvider = { throw IllegalStateException("currentForkIdHash exception testing") },
          ),
      )
    val p2pNetworkImpl2 =
      createP2PNetwork(
        privateKey = key2,
        port = PORT2,
        staticPeers = listOf(PEER_ADDRESS_NODE_1),
        statusMessageFactory =
          StatusMessageFactory(
            beaconChain = beaconChain,
            forkIdHashProvider = { throw IllegalStateException("currentForkIdHash exception testing") },
          ),
      )

    try {
      p2pNetworkImpl1.start()
      p2pNetworkImpl2.start()

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      val peer1 =
        p2pNetworkImpl2.getPeerLookup().getPeer(LibP2PNodeId(PeerId.fromBase58(PEER_ID_NODE_1)))
          ?: throw IllegalStateException("Peer with ID $PEER_ID_NODE_1 not found in p2pNetworkImpl2")
      val maruPeer1 =
        DefaultMaruPeer(peer1, rpcMethods, statusMessageFactory, p2pConfig = P2P(ipAddress = IPV4, port = PORT1))

      val responseFuture = maruPeer1.sendStatus()

      assertThatThrownBy { responseFuture.get() }
        .isInstanceOf(ExecutionException::class.java)
        .hasCauseInstanceOf(RpcException::class.java)
        .hasMessageContaining("currentForkIdHash exception testing")
        .matches { (it.cause as RpcException).responseCode == RpcResponseStatus.SERVER_ERROR_CODE }
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `peer can send beacon blocks by range request`() {
    // Set up beacon chain with some blocks
    val testBeaconChain = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u))
    val storedBlocks =
      (0UL..10UL).map { blockNumber ->
        DataGenerators.randomSealedBeaconBlock(number = blockNumber)
      }

    testBeaconChain.newUpdater().use { updater ->
      storedBlocks.forEach { block ->
        updater.putSealedBeaconBlock(block)
      }
      updater.putBeaconState(
        BeaconState(
          latestBeaconBlockHeader = storedBlocks.last().beaconBlock.beaconBlockHeader,
          validators = DataGenerators.randomValidators(),
        ),
      )
      updater.commit()
    }
    val p2pNetworkImpl1 = createP2PNetwork(privateKey = key1, port = PORT1, beaconChain = testBeaconChain)
    val p2pNetworkImpl2 = createP2PNetwork(privateKey = key2, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1))

    try {
      p2pNetworkImpl1.start()
      p2pNetworkImpl2.start()

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      val peer1 =
        p2pNetworkImpl2.getPeerLookup().getPeer(LibP2PNodeId(PeerId.fromBase58(PEER_ID_NODE_1)))
          ?: throw IllegalStateException("Peer with ID $PEER_ID_NODE_1 not found in p2pNetworkImpl2")

      val startBlockNumber = 3UL
      val count = 5UL
      val responseFuture = peer1.sendBeaconBlocksByRange(startBlockNumber, count)

      val response = responseFuture.get(5, TimeUnit.SECONDS)

      val expectedBlocks = storedBlocks.subList(startBlockNumber.toInt(), startBlockNumber.toInt() + count.toInt())
      assertThat(response.blocks).hasSize(5)
      assertThat(response.blocks).isEqualTo(expectedBlocks)
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `peer send a beacon blocks by range request and receive exception when callee throws error`() {
    val p2pNetworkImpl1 = createP2PNetwork(privateKey = key1, port = PORT1, beaconChain = getMockedBeaconChain())
    val p2pNetworkImpl2 = createP2PNetwork(privateKey = key2, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1))
    try {
      p2pNetworkImpl1.start()
      p2pNetworkImpl2.start()

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      val peer1 =
        p2pNetworkImpl2.getPeerLookup().getPeer(LibP2PNodeId(PeerId.fromBase58(PEER_ID_NODE_1)))
          ?: throw IllegalStateException("Peer with ID $PEER_ID_NODE_1 not found in p2pNetworkImpl2")

      val startBlockNumber = 3UL
      val count = 5UL
      val responseFuture = peer1.sendBeaconBlocksByRange(startBlockNumber, count)

      assertThatThrownBy { responseFuture.get() }
        .isInstanceOf(ExecutionException::class.java)
        .hasCauseInstanceOf(RpcException::class.java)
        .hasMessageContaining("Missing sealed beacon block")
        .matches { (it.cause as RpcException).responseCode == RpcResponseStatus.RESOURCE_UNAVAILABLE }
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `should expose discovery enr properly - discovery disabled`() {
    val p2pNetworkImpl1 =
      createP2PNetwork(
        privateKey = key1,
        port = PORT1,
        discovery = null,
      )
    try {
      p2pNetworkImpl1.start()
      val enr = ENR.factory.fromEnr(p2pNetworkImpl1.enr)
      assertThat(enr.tcpAddress.get().port).isEqualTo(PORT1.toInt())
      assertThat(enr.udpAddress.get().port).isEqualTo(PORT1.toInt())
      assertThat(
        enr.udpAddress
          .get()
          .address
          .toString(),
      ).doesNotContain("0.0.0.0")
    } finally {
      p2pNetworkImpl1.stop().get()
    }
  }

  @Test
  fun `should expose discovery enr properly - discovery enabled`() {
    val p2pNetworkImpl1 =
      createP2PNetwork(
        privateKey = key1,
        port = PORT1,
        discovery =
          P2P.Discovery(
            port = PORT2,
            refreshInterval = 1.seconds,
          ),
      )
    try {
      p2pNetworkImpl1.start()
      val enr = ENR.factory.fromEnr(p2pNetworkImpl1.enr)
      assertThat(enr.tcpAddress.get().port).isEqualTo(PORT1.toInt())
      assertThat(enr.udpAddress.get().port).isEqualTo(PORT2.toInt())
      assertThat(
        enr.udpAddress
          .get()
          .address
          .toString(),
      ).doesNotContain("0.0.0.0")
    } finally {
      p2pNetworkImpl1.stop().get()
    }
  }

  @Disabled("FIXME: this test fails 100% of the time. Needs to be fixed")
  fun `peer can be discovered and disconnected peers can be rediscovered`() {
    val refreshInterval = 5.seconds
    val p2pNetworkImpl1 =
      createP2PNetwork(
        privateKey = key1,
        port = PORT1,
        discovery =
          P2P.Discovery(
            port = PORT2,
            refreshInterval = refreshInterval,
          ),
      )
    val p2pNetworkImpl2 =
      createP2PNetwork(
        privateKey = key2,
        port = PORT3,
        beaconChain = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u)),
        discovery =
          P2P.Discovery(
            port = PORT4,
            bootnodes = listOf(p2pNetworkImpl1.enr),
            refreshInterval = refreshInterval,
          ),
      )

    val p2pNetworkImpl3 =
      createP2PNetwork(
        privateKey = key3,
        port = PORT5,
        beaconChain = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u)),
        discovery =
          P2P.Discovery(
            port = PORT6,
            bootnodes = listOf(p2pNetworkImpl1.enr),
            refreshInterval = refreshInterval,
          ),
      )

    try {
      p2pNetworkImpl1.start()
      p2pNetworkImpl2.start()
      p2pNetworkImpl3.start()

      val awaitTimeoutInSeconds = 30L
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl1, PEER_ID_NODE_2)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl2, PEER_ID_NODE_1)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl3, PEER_ID_NODE_2)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl1, PEER_ID_NODE_3)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl2, PEER_ID_NODE_3)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl3, PEER_ID_NODE_1)
      }

      log.info("Dropping peers from p2pNetworkImpl2")
      p2pNetworkImpl2.dropPeer(PEER_ID_NODE_1, DisconnectReason.TOO_MANY_PEERS).get()
      p2pNetworkImpl2.dropPeer(PEER_ID_NODE_3, DisconnectReason.TOO_MANY_PEERS).get()

      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl1, PEER_ID_NODE_2)
      }

      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl2, PEER_ID_NODE_1)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl3, PEER_ID_NODE_2)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl1, PEER_ID_NODE_3)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl2, PEER_ID_NODE_3)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl3, PEER_ID_NODE_1)
      }
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
      p2pNetworkImpl3.stop().get()
    }
  }

  @Disabled("FIXME: this test fails 100% of the time. Needs to be fixed")
  fun `sending status updates updates status`() {
    val refreshInterval = 5.seconds
    val p2pNetworkImpl1 =
      createP2PNetwork(
        privateKey = key1,
        port = PORT1,
        statusUpdate =
          P2P.StatusUpdateConfig(
            refreshInterval = 1.seconds,
            refreshIntervalLeeway = 1.seconds,
            timeout = 1.seconds,
          ),
        discovery =
          P2P.Discovery(
            port = PORT2,
            refreshInterval = refreshInterval,
          ),
        statusMessageFactory = getMockedStatusMessageFactory(),
      )

    val p2pNetworkImpl2 =
      createP2PNetwork(
        privateKey = key2,
        port = PORT3,
        beaconChain = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u)),
        statusUpdate =
          P2P.StatusUpdateConfig(
            refreshInterval = 1.seconds,
            refreshIntervalLeeway = 1.seconds,
            timeout = 1.seconds,
          ),
        discovery =
          P2P.Discovery(
            port = PORT4,
            bootnodes = listOf(p2pNetworkImpl1.enr),
            refreshInterval = refreshInterval,
          ),
        statusMessageFactory = getMockedStatusMessageFactory(),
      )

    try {
      p2pNetworkImpl1.start()
      p2pNetworkImpl2.start()

      val awaitTimeoutInSeconds = 30L
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl1, PEER_ID_NODE_2)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl2, PEER_ID_NODE_1)
      }

      val peer1 = p2pNetworkImpl1.maruPeerManager.getPeers()[0]
      val peer2 = p2pNetworkImpl2.maruPeerManager.getPeers()[0]

      val startTime = System.currentTimeMillis()
      var currentBlockNumber = 0uL
      while (System.currentTimeMillis() < startTime + 13000L) { // max 13 seconds to run this test
        assertNetworkIsConnectedToPeer(p2pNetworkImpl1, PEER_ID_NODE_2)
        assertNetworkIsConnectedToPeer(p2pNetworkImpl2, PEER_ID_NODE_1)
        awaitUntilAsserted { peer1.getStatus()!!.latestBlockNumber == currentBlockNumber }
        awaitUntilAsserted { peer2.getStatus()!!.latestBlockNumber == currentBlockNumber }
        if (peer1.getStatus()!!.latestBlockNumber == 5uL) {
          // we have reached the end of the blocks, so we can stop
          break
        }
        currentBlockNumber++
      }
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  @Test
  fun `peer sending status update too late is disconnected`() {
    val p2pNetworkImpl1 =
      createP2PNetwork(
        privateKey = key1,
        port = PORT1,
        statusUpdate =
          P2P.StatusUpdateConfig(
            refreshInterval = 1.seconds,
            refreshIntervalLeeway = 0.seconds,
            timeout = 1.seconds,
          ),
      )
    // node 2 is initiating the connection and is only sending status updates after 2 seconds,
    // so it should be disconnected by node 1, which expects a status update within 1 second
    // The initial status update works, because node 1 has a timeout of 1 second for the status update
    val p2pNetworkImpl2 =
      createP2PNetwork(
        privateKey = key2,
        port = PORT2,
        staticPeers = listOf(PEER_ADDRESS_NODE_1),
        beaconChain = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u)),
        statusUpdate =
          P2P.StatusUpdateConfig(
            refreshInterval = 2.seconds,
            refreshIntervalLeeway = 1.seconds,
            timeout = 1.seconds,
          ),
      )

    try {
      p2pNetworkImpl1.start()
      p2pNetworkImpl2.start()

      val awaitTimeoutInSeconds = 30L
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl1, PEER_ID_NODE_2)
      }
      awaitUntilAsserted(timeout = awaitTimeoutInSeconds, timeUnit = TimeUnit.SECONDS) {
        assertNetworkIsConnectedToPeer(p2pNetworkImpl2, PEER_ID_NODE_1)
      }

      // check for the next 5 seconds that the peers are at least disconnected twice,
      // because node 2 is reconnecting because of the static connection
      val startTime = System.currentTimeMillis()
      var disconnectCount = 0
      while (System.currentTimeMillis() < startTime + 6000L) {
        if (!p2pNetworkImpl1.isConnected(PEER_ID_NODE_2)) {
          disconnectCount++
          while (!p2pNetworkImpl1.isConnected(PEER_ID_NODE_2)) {
            // wait for the peer to be connected again
            sleep(50L)
          }
        }
      }

      assertThat(disconnectCount).isGreaterThanOrEqualTo(2)
    } finally {
      p2pNetworkImpl1.stop().get()
      p2pNetworkImpl2.stop().get()
    }
  }

  private fun assertNetworkHasPeers(
    network: P2PNetworkImpl,
    peers: Int,
  ) {
    assertThat(network.peerCount).isEqualTo(peers)
  }

  private fun awaitUntilAsserted(
    timeout: Long = 10000L,
    timeUnit: TimeUnit = TimeUnit.MILLISECONDS,
    condition: () -> Unit,
  ) {
    await()
      .timeout(timeout, timeUnit)
      .untilAsserted(condition)
  }

  private fun assertNetworkIsConnectedToPeer(
    p2pNetwork: P2PNetworkImpl,
    peer: String,
  ) {
    assertThat(
      p2pNetwork.getPeerLookup().getPeer(LibP2PNodeId(PeerId.fromBase58(peer))),
    ).isNotNull
  }

  private fun getMockedStatusMessageFactory(): StatusMessageFactory {
    val beaconChain = mock<BeaconChain>()
    val beaconState = mock<BeaconState>()
    val beaconBlockHeader = mock<BeaconBlockHeader>()
    whenever(beaconChain.getLatestBeaconState()).thenReturn(beaconState)
    whenever(beaconState.latestBeaconBlockHeader).thenReturn(beaconBlockHeader)
    whenever(beaconBlockHeader.hash).thenReturn(ByteArray(32))
    whenever(beaconBlockHeader.number).thenReturn(0uL, 1uL, 2uL, 3uL, 4uL, 5uL)
    val forkIdHashProvider = createForkIdHashProvider()

    val statusMessageFactory = StatusMessageFactory(beaconChain, forkIdHashProvider)
    return statusMessageFactory
  }

  private fun getMockedBeaconChain(): BeaconChain {
    val mockedBeaconChain = mock<BeaconChain>(RETURNS_DEEP_STUBS)
    whenever(mockedBeaconChain.getSealedBeaconBlocks(any(), any())).thenThrow(
      IllegalStateException("Missing sealed beacon block"),
    )
    return mockedBeaconChain
  }
}
