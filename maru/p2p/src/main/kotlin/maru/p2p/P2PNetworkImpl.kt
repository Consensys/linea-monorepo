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
package maru.p2p

import io.libp2p.core.PeerId
import io.libp2p.core.crypto.unmarshalPrivateKey
import java.util.Optional
import java.util.concurrent.TimeUnit
import kotlin.jvm.optionals.getOrNull
import maru.config.P2P
import maru.core.SealedBeaconBlock
import maru.p2p.topics.SealedBlocksTopicHandler
import maru.serialization.Serializer
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
import tech.pegasys.teku.networking.p2p.rpc.RpcStreamController
import tech.pegasys.teku.networking.p2p.network.P2PNetwork as TekuP2PNetwork

class P2PNetworkImpl(
  privateKeyBytes: ByteArray,
  private val p2pConfig: P2P,
  chainId: UInt,
  private val serializer: Serializer<SealedBeaconBlock>,
) : P2PNetwork {
  private val topicIdGenerator = LineaTopicIdGenerator(chainId)
  private val sealedBlocksTopicId = topicIdGenerator.topicId(MessageType.BEACON_BLOCK, Version.V1)
  private val sealedBlocksSubscriptionManager = SubscriptionManager<SealedBeaconBlock>()
  private val sealedBlocksTopicHandler = SealedBlocksTopicHandler(sealedBlocksSubscriptionManager, serializer)

  private fun buildP2PNetwork(
    privateKeyBytes: ByteArray,
    p2pConfig: P2P,
  ): TekuP2PNetwork<Peer> {
    val privateKey = unmarshalPrivateKey(privateKeyBytes)

    return Libp2pNetworkFactory.build(
      privateKey = privateKey,
      ipAddress = p2pConfig.ipAddress,
      port = p2pConfig.port,
      sealedBlocksTopicHandler = sealedBlocksTopicHandler,
      sealedBlocksTopicId = sealedBlocksTopicId,
    )
  }

  private val p2pNetwork: TekuP2PNetwork<Peer> = buildP2PNetwork(privateKeyBytes, p2pConfig)

  private val log: Logger = LogManager.getLogger(this::class.java)
  private val delayedExecutor =
    SafeFuture.delayedExecutor(p2pConfig.reconnectDelay.inWholeMilliseconds, TimeUnit.MILLISECONDS)
  private val staticPeerMap = mutableMapOf<NodeId, MultiaddrPeerAddress>()

  override fun start(): SafeFuture<Unit> =
    p2pNetwork
      .start()
      .thenApply {
        p2pConfig.staticPeers.forEach { peer ->
          p2pNetwork
            .createPeerAddress(peer)
            ?.let { address -> addStaticPeer(address as MultiaddrPeerAddress) }
        }
      }

  override fun stop(): SafeFuture<Unit> = p2pNetwork.stop().thenApply { }

  override fun broadcastMessage(message: Message<*>): SafeFuture<*> =
    when (message.type) {
      MessageType.QBFT -> SafeFuture.completedFuture(Unit) // TODO: Add QBFT messages support later
      MessageType.BEACON_BLOCK -> {
        require(message.payload is SealedBeaconBlock)
        val serializedSealedBeaconBlock = Bytes.wrap(serializer.serialize(message.payload as SealedBeaconBlock))
        p2pNetwork.gossip(topicIdGenerator.topicId(message.type, message.version), serializedSealedBeaconBlock)
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
            log.info("Already connected to peer $peerAddress. Error: ${t.message}")
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
          log.info("Created persistent connection to {}", peerAddress)
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

  // TODO: This is pretty much WIP. This should be addressed with the syncing
  internal fun sendRequest(
    peer: String,
    rpcMethod: MaruRpcMethod,
    request: Bytes,
    responseHandler: MaruRpcResponseHandler,
  ): SafeFuture<RpcStreamController<MaruOutgoingRpcRequestHandler>> {
    val maybePeer =
      p2pNetwork
        .getPeer(LibP2PNodeId(PeerId.fromBase58(peer)))
        .getOrNull()
    return if (maybePeer == null) {
      SafeFuture.failedFuture(IllegalStateException("Peer $peer is not connected!"))
    } else {
      maybePeer.sendRequest(rpcMethod, request, responseHandler)
    }
  }
}
