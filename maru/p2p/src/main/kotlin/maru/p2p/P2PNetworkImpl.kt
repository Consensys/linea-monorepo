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
import io.libp2p.core.crypto.unmarshalPrivateKey
import java.util.Optional
import java.util.concurrent.TimeUnit
import kotlin.jvm.optionals.getOrElse
import kotlin.jvm.optionals.getOrNull
import maru.config.P2P
import maru.consensus.ForkIdHashProvider
import maru.core.SealedBeaconBlock
import maru.database.BeaconChain
import maru.metrics.MaruMetricsCategory
import maru.p2p.discovery.MaruDiscoveryService
import maru.p2p.messages.StatusMessageFactory
import maru.p2p.topics.TopicHandlerWithInOrderDelivering
import maru.serialization.SerDe
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.metrics.Tag
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.p2p.libp2p.LibP2PNodeId
import tech.pegasys.teku.networking.p2p.libp2p.MultiaddrPeerAddress
import tech.pegasys.teku.networking.p2p.libp2p.PeerAlreadyConnectedException
import tech.pegasys.teku.networking.p2p.network.PeerAddress
import tech.pegasys.teku.networking.p2p.peer.DisconnectReason
import tech.pegasys.teku.networking.p2p.peer.NodeId
import tech.pegasys.teku.networking.p2p.peer.Peer
import org.hyperledger.besu.plugin.services.MetricsSystem as BesuMetricsSystem

class P2PNetworkImpl(
  privateKeyBytes: ByteArray,
  private val p2pConfig: P2P,
  private val chainId: UInt,
  private val serDe: SerDe<SealedBeaconBlock>,
  private val metricsFacade: MetricsFacade,
  private val metricsSystem: BesuMetricsSystem,
  private val statusMessageFactory: StatusMessageFactory,
  private val beaconChain: BeaconChain,
  private val forkIdHashProvider: ForkIdHashProvider,
  nextExpectedBeaconBlockNumber: ULong,
) : P2PNetwork {
  lateinit var maruPeerManager: MaruPeerManager
  private val topicIdGenerator = LineaMessageIdGenerator(chainId)
  private val sealedBlocksTopicId = topicIdGenerator.id(GossipMessageType.BEACON_BLOCK.name, Version.V1)
  private val sealedBlocksSubscriptionManager = SubscriptionManager<SealedBeaconBlock>()
  private val sealedBlocksTopicHandler =
    TopicHandlerWithInOrderDelivering(
      initialExpectedSequenceNumber = nextExpectedBeaconBlockNumber,
      subscriptionManager = sealedBlocksSubscriptionManager,
      sequenceNumberExtractor = { it.beaconBlock.beaconBlockHeader.number },
      deserializer = serDe,
      topicId = sealedBlocksTopicId,
    )
  private val broadcastMessageCounterFactory =
    metricsFacade.createCounterFactory(
      category = MaruMetricsCategory.P2P_NETWORK,
      name = "message.broadcast.counter",
      description = "Count of messages broadcasted over the P2P network",
    )

  private fun buildP2PNetwork(
    privateKeyBytes: ByteArray,
    p2pConfig: P2P,
    besuMetricsSystem: BesuMetricsSystem,
  ): TekuLibP2PNetwork {
    val privateKey = unmarshalPrivateKey(privateKeyBytes)
    val rpcIdGenerator = LineaRpcProtocolIdGenerator(chainId)

    val rpcMethods = RpcMethods(statusMessageFactory, rpcIdGenerator, { maruPeerManager }, beaconChain)
    maruPeerManager =
      MaruPeerManager(
        maruPeerFactory = DefaultMaruPeerFactory(rpcMethods, statusMessageFactory),
        p2pConfig = p2pConfig,
      )

    return Libp2pNetworkFactory(LINEA_DOMAIN).build(
      privateKey = privateKey,
      ipAddress = p2pConfig.ipAddress,
      port = p2pConfig.port,
      sealedBlocksTopicHandler = sealedBlocksTopicHandler,
      sealedBlocksTopicId = sealedBlocksTopicId,
      rpcMethods = rpcMethods.all(),
      maruPeerManager = maruPeerManager,
      metricsSystem = besuMetricsSystem,
    )
  }

  private val builtNetwork: TekuLibP2PNetwork = buildP2PNetwork(privateKeyBytes, p2pConfig, metricsSystem)
  internal val p2pNetwork = builtNetwork.p2PNetwork
  private val discoveryService: MaruDiscoveryService? =
    p2pConfig.discovery?.let {
      MaruDiscoveryService(
        privateKeyBytes =
          privateKeyBytes
            .slice(
              (privateKeyBytes.size - 32).rangeTo(privateKeyBytes.size - 1),
            ).toByteArray(),
        p2pConfig = p2pConfig,
        forkIdHashProvider = forkIdHashProvider,
        metricsSystem = metricsSystem,
      )
    }

  // TODO: We need to call the updateForkId method on the discovery service when the forkId changes internal
  private val peerLookup = builtNetwork.peerLookup
  private val log: Logger = LogManager.getLogger(this::javaClass)
  private val delayedExecutor =
    SafeFuture.delayedExecutor(p2pConfig.reconnectDelay.inWholeMilliseconds, TimeUnit.MILLISECONDS)
  private val staticPeerMap = mutableMapOf<NodeId, MultiaddrPeerAddress>()

  override val nodeId: String = p2pNetwork.nodeId.toBase58()
  override val discoveryAddresses: List<String> = p2pNetwork.discoveryAddresses.getOrElse { emptyList() }
  override val enr: String? = p2pNetwork.enr.getOrNull()
  override val nodeAddresses: List<String> = p2pNetwork.nodeAddresses

  override fun start(): SafeFuture<Unit> =
    p2pNetwork
      .start()
      .thenApply {
        log.info(
          "Starting P2P network. port=$port, nodeId=${
            p2pNetwork.nodeId
          }",
        )
        p2pConfig.staticPeers.forEach { peer ->
          p2pNetwork
            .createPeerAddress(peer)
            ?.let { address -> addStaticPeer(address as MultiaddrPeerAddress) }
        }
      }.thenPeek {
        discoveryService?.start()
        maruPeerManager.start(discoveryService, p2pNetwork)
        metricsFacade.createGauge(
          category = MaruMetricsCategory.P2P_NETWORK,
          name = "peer.count",
          description = "Number of peers connected to the P2P network",
          measurementSupplier = { peerCount.toLong() },
        )
      }

  override fun stop(): SafeFuture<Unit> {
    val pmStop = maruPeerManager.stop()
    discoveryService?.stop()
    val p2pStop = p2pNetwork.stop()
    return SafeFuture.allOf(p2pStop, pmStop).thenApply {}
  }

  override fun broadcastMessage(message: Message<*, GossipMessageType>): SafeFuture<*> {
    broadcastMessageCounterFactory
      .create(
        listOf(
          Tag("message.type", message.type.name),
          Tag("message.version", message.version.name),
        ),
      ).increment()
    return when (message.type) {
      GossipMessageType.QBFT -> SafeFuture.completedFuture(Unit) // TODO: Add QBFT messages support later
      GossipMessageType.BEACON_BLOCK -> {
        require(message.payload is SealedBeaconBlock)
        val serializedSealedBeaconBlock = Bytes.wrap(serDe.serialize(message.payload as SealedBeaconBlock))
        p2pNetwork.gossip(topicIdGenerator.id(message.type.name, message.version), serializedSealedBeaconBlock)
      }
    }
  }

  override fun subscribeToBlocks(subscriber: SealedBeaconBlockHandler<ValidationResult>): Int {
    val subscriptionManagerHadSubscriptions = sealedBlocksSubscriptionManager.hasSubscriptions()

    return sealedBlocksSubscriptionManager.subscribeToBlocks(subscriber::handleSealedBlock).also {
      if (!subscriptionManagerHadSubscriptions) {
        p2pNetwork.subscribe(sealedBlocksTopicId, sealedBlocksTopicHandler)
      }
    }
  }

  /**
   * Unsubscribes the handler with a given subscriptionId from sealed block handling.
   * Note that it's impossible to unsubscribe from a topic on LibP2P level, so the messages will still be received and
   * handled by LibP2P, but not processed by Maru
   */
  override fun unsubscribeFromBlocks(subscriptionId: Int) = sealedBlocksSubscriptionManager.unsubscribe(subscriptionId)

  fun addStaticPeer(peerAddress: MultiaddrPeerAddress) {
    if (peerAddress.id == p2pNetwork.nodeId) { // Don't connect to self
      return
    }
    synchronized(this) {
      if (staticPeerMap.containsKey(peerAddress.id)) {
        return
      }
      staticPeerMap[peerAddress.id] = peerAddress
    }
    maintainPersistentConnection(peerAddress)
  }

  fun removeStaticPeer(peerAddress: PeerAddress) {
    synchronized(this) {
      staticPeerMap.remove(peerAddress.id)
      p2pNetwork.getPeer(peerAddress.id).ifPresent { peer -> peer.disconnectImmediately(Optional.empty(), true) }
    }
  }

  private fun maintainPersistentConnection(peerAddress: MultiaddrPeerAddress): SafeFuture<Unit> =
    p2pNetwork
      .connect(peerAddress)
      .whenComplete { peer: Peer?, t: Throwable? ->
        if (t != null) {
          if (t is PeerAlreadyConnectedException) {
            log.debug("Already connected to peer {}. Error: {}", peerAddress, t.message)
            reconnectWhenDisconnected(peer!!, peerAddress)
          } else {
            log.trace(
              "Failed to connect to peer {}, retrying after {} ms. Error: {}",
              peerAddress,
              p2pConfig.reconnectDelay,
              t.message,
            )
            SafeFuture
              .runAsync({ maintainPersistentConnection(peerAddress) }, delayedExecutor)
          }
        } else {
          log.debug("Created persistent connection to {}", peerAddress)
          reconnectWhenDisconnected(peer!!, peerAddress)
        }
      }.thenApply {}

  private fun reconnectWhenDisconnected(
    peer: Peer,
    peerAddress: MultiaddrPeerAddress,
  ) {
    peer.subscribeDisconnect { _: Optional<DisconnectReason>, _: Boolean ->
      if (staticPeerMap.containsKey(peerAddress.id)) {
        SafeFuture.runAsync({ maintainPersistentConnection(peerAddress) }, delayedExecutor)
      }
    }
  }

  override val port
    get(): UInt =
      builtNetwork.host.network.transports
        .first()
        .listenAddresses()
        .first()
        .components
        .last()
        .stringValue!!
        .toUInt()

  internal val peerCount: Int
    get() = p2pNetwork.peerCount

  internal fun isConnected(peer: String): Boolean {
    val peerAddress =
      PeerAddress(
        LibP2PNodeId(
          PeerId.fromBase58(
            peer,
          ),
        ),
      )
    return p2pNetwork.isConnected(peerAddress)
  }

  internal fun dropPeer(
    peer: String,
    reason: DisconnectReason,
  ): SafeFuture<Unit> {
    val maybePeer =
      p2pNetwork
        .getPeer(LibP2PNodeId(PeerId.fromBase58(peer)))
        .getOrNull()
    return if (maybePeer == null) {
      log.warn("Trying to disconnect from peer {}, but there's no connection to it!", peer)
      SafeFuture.completedFuture(Unit)
    } else {
      maybePeer.disconnectCleanly(reason).thenApply { }
    }
  }

  override fun getPeers(): List<PeerInfo> = peerLookup.getPeers().map { it.toPeerInfo() }

  override fun getPeer(peerId: String): PeerInfo? =
    peerLookup.getPeer(LibP2PNodeId(PeerId.fromBase58(peerId)))?.toPeerInfo()

  override fun getPeerLookup(): PeerLookup = peerLookup
}
