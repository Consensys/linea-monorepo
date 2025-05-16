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

import io.libp2p.core.ConnectionHandler
import io.libp2p.core.Host
import io.libp2p.core.PeerId
import io.libp2p.core.crypto.PrivKey
import io.libp2p.core.dsl.host
import io.libp2p.core.multiformats.Multiaddr
import io.libp2p.core.mux.StreamMuxerProtocol
import io.libp2p.etc.types.seconds
import io.libp2p.pubsub.PubsubApiImpl
import io.libp2p.pubsub.gossip.Gossip
import io.libp2p.pubsub.gossip.GossipPeerScoreParams
import io.libp2p.pubsub.gossip.GossipScoreParams
import io.libp2p.pubsub.gossip.GossipTopicsScoreParams
import io.libp2p.pubsub.gossip.builders.GossipParamsBuilder
import io.libp2p.pubsub.gossip.builders.GossipRouterBuilder
import io.libp2p.security.secio.SecIoSecureChannel
import io.libp2p.transport.tcp.TcpTransport
import java.util.Optional
import kotlin.random.Random
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem
import pubsub.pb.Rpc
import tech.pegasys.teku.infrastructure.async.AsyncRunnerFactory
import tech.pegasys.teku.infrastructure.async.MetricTrackingExecutorFactory
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.networking.p2p.libp2p.LibP2PNetwork
import tech.pegasys.teku.networking.p2p.libp2p.LibP2PNodeId
import tech.pegasys.teku.networking.p2p.libp2p.PeerManager
import tech.pegasys.teku.networking.p2p.libp2p.gossip.GossipTopicHandlers
import tech.pegasys.teku.networking.p2p.libp2p.gossip.LibP2PGossipNetwork
import tech.pegasys.teku.networking.p2p.libp2p.gossip.PreparedPubsubMessage
import tech.pegasys.teku.networking.p2p.libp2p.rpc.RpcHandler
import tech.pegasys.teku.networking.p2p.network.P2PNetwork
import tech.pegasys.teku.networking.p2p.network.PeerHandler
import tech.pegasys.teku.networking.p2p.peer.Peer
import tech.pegasys.teku.networking.p2p.reputation.ReputationManager

object P2PNetworkFactory {
  fun build(
    privateKey: PrivKey,
    port: String,
    ipAddress: String,
  ): P2PNetwork<Peer> {
    val addr = Multiaddr("/ip4/$ipAddress/tcp/$port")

    return setupP2PNetwork(privateKey, addr)
  }

  private fun setupP2PNetwork(
    privateKey: PrivKey,
    ipv4Address: Multiaddr,
  ): P2PNetwork<Peer> {
    val rpcMethod = MaruRpcMethod()
    val gossipTopicHandlers = GossipTopicHandlers() // TODO: add handlers for topics as needed

    val gossipParams = GossipParamsBuilder().heartbeatInterval(1.seconds).build()
    val gossipRouterBuilder =
      GossipRouterBuilder().apply {
        params = gossipParams
        scoreParams = GossipScoreParams(GossipPeerScoreParams(), GossipTopicsScoreParams())
        messageFactory = { getMessageFactory(it, gossipTopicHandlers) }
      }
    val gossipRouter = gossipRouterBuilder.build()
    val pubsubApiImpl = PubsubApiImpl(gossipRouter)
    val gossip = Gossip(gossipRouter, pubsubApiImpl)

    val metricsSystem = NoOpMetricsSystem()
    val publisherApi = gossip.createPublisher(privateKey, Random.nextLong())
    val gossipNetwork = LibP2PGossipNetwork(metricsSystem, gossip, publisherApi, gossipTopicHandlers)

    val peerId = PeerId.fromPubKey(privateKey.publicKey())
    val libP2PNodeId = LibP2PNodeId(peerId)

    val metricTrackingExecutorFactory = MetricTrackingExecutorFactory(metricsSystem)
    val asyncRunner = AsyncRunnerFactory.createDefault(metricTrackingExecutorFactory).create("maru", 2)
    val rpcHandler = RpcHandler(asyncRunner, rpcMethod)

    val peerManager =
      PeerManager(
        metricsSystem,
        ReputationManager.NOOP,
        listOf<PeerHandler>(MaruPeerHandler()),
        listOf(rpcHandler),
        { _ -> 50.0 }, // TODO: I guess we need a scoring function here
      )

    val host =
      createHost(
        privateKey = privateKey,
        connectionHandlers = listOf(gossip, peerManager),
        gossip = gossip,
        rpcHandler = rpcHandler,
        ipv4Address = ipv4Address,
      )

    val advertisedAddresses = listOf(ipv4Address)

    return LibP2PNetwork(
      privateKey,
      libP2PNodeId,
      host,
      peerManager,
      advertisedAddresses,
      gossipNetwork,
      listOf(1),
    )
  }

  private fun getMessageFactory(
    msg: Rpc.Message,
    gossipTopicHandlers: GossipTopicHandlers,
  ): PreparedPubsubMessage {
    val arrivalTimestamp = Optional.empty<UInt64>()
    val topic = msg.getTopicIDs(0)
    val payload = Bytes.wrap(msg.data.toByteArray())

    val preparedMessage =
      gossipTopicHandlers
        .getHandlerForTopic(topic)
        .map { handler -> handler.prepareMessage(payload, arrivalTimestamp) }
        .orElse(MaruPreparedGossipMessage(payload, arrivalTimestamp))

    return PreparedPubsubMessage(msg, preparedMessage)
  }

  private fun createHost(
    privateKey: PrivKey,
    connectionHandlers: List<ConnectionHandler>,
    gossip: Gossip,
    rpcHandler: RpcHandler<MaruOutgoingRpcRequestHandler, Bytes, MaruRpcResponseHandler>,
    ipv4Address: Multiaddr,
  ): Host =
    host {
      protocols {
        +gossip
        +rpcHandler
      }
      network {
        listen(ipv4Address.toString())
      }
      transports {
        add(::TcpTransport)
      }
      identity {
        factory = { privateKey }
      }
      secureChannels {
        add { localKey, muxerProtocols -> SecIoSecureChannel(localKey, muxerProtocols) }
      }
      connectionHandlers {
        connectionHandlers.forEach { handler ->
          add(handler)
        }
      }
      muxers {
        add(StreamMuxerProtocol.Mplex)
      }
    }
}
