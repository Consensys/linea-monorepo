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

import java.lang.Thread.sleep
import java.util.concurrent.TimeUnit
import maru.config.P2P
import maru.core.SealedBeaconBlock
import maru.core.ext.DataGenerators
import maru.serialization.rlp.RLPSerializers
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatNoException
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.parallel.Execution
import org.junit.jupiter.api.parallel.ExecutionMode
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.p2p.libp2p.MultiaddrPeerAddress
import tech.pegasys.teku.networking.p2p.peer.DisconnectReason

@Execution(ExecutionMode.SAME_THREAD)
class P2PTest {
  companion object {
    private val chainId = 1337u

    private const val IPV4 = "127.0.0.1"

    private const val PORT1 = "9234"
    private const val PORT2 = "9235"
    private const val PORT3 = "9236"

    private const val PRIVATE_KEY1: String =
      "0x0802122012c0b113e2b0c37388e2b484112e13f05c92c4471e3ee1dfaa368fa5045325b2"
    private const val PRIVATE_KEY2: String =
      "0x0802122100f3d2fffa99dc8906823866d96316492ebf7a8478713a89a58b7385af85b088a1"
    private const val PRIVATE_KEY3: String =
      "0x080212204437acb8e84bc346f7640f239da84abe99bc6f97b7855f204e34688d2977fd57"

    private const val PEER_ID_NODE_1: String = "16Uiu2HAmPRfinavM2jE9BSkCagBGStJ2SEkPPm6fxFVMdCQebzt6"
    private const val PEER_ID_NODE_2: String = "16Uiu2HAmVXtqhevTAJqZucPbR2W4nCMpetrQASgjZpcxDEDaUPPt"
    private const val PEER_ID_NODE_3: String = "16Uiu2HAkzq767a82zfyUz4VLgPbFrxSQBrdmUYxgNDbwgvmjwWo5"

    // TODO: to make these tests reliable it would be good if the ports were not hardcoded, but free ports chosen
    private const val PEER_ADDRESS_NODE_1: String = "/ip4/$IPV4/tcp/$PORT1/p2p/$PEER_ID_NODE_1"
    private const val PEER_ADDRESS_NODE_2: String = "/ip4/$IPV4/tcp/$PORT2/p2p/$PEER_ID_NODE_2"
    private const val PEER_ADDRESS_NODE_3: String = "/ip4/$IPV4/tcp/$PORT3/p2p/$PEER_ID_NODE_3"

    private val key1 = Bytes.fromHexString(PRIVATE_KEY1).toArray()
    private val key2 = Bytes.fromHexString(PRIVATE_KEY2).toArray()
    private val key3 = Bytes.fromHexString(PRIVATE_KEY3).toArray()
  }

  @Test
  fun `static peer can be added`() {
    val p2PNetworkImpl1 =
      P2PNetworkImpl(
        privateKeyBytes = key1,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT1, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    val p2pNetworkImpl2 =
      P2PNetworkImpl(
        privateKeyBytes = key2,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT2, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    try {
      p2PNetworkImpl1.start()

      p2pNetworkImpl2.start()

      p2PNetworkImpl1.addStaticPeer(MultiaddrPeerAddress.fromAddress(PEER_ADDRESS_NODE_2))

      awaitUntilAsserted { assertNetworkHasPeers(network = p2PNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }
    } finally {
      p2PNetworkImpl1.stop()
      p2pNetworkImpl2.stop()
    }
  }

  @Test
  fun `static peers can be removed`() {
    val p2PNetworkImpl1 =
      P2PNetworkImpl(
        privateKeyBytes = key1,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT1, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    val p2pNetworkImpl2 =
      P2PNetworkImpl(
        privateKeyBytes = key2,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT2, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    try {
      p2PNetworkImpl1.start()

      p2pNetworkImpl2.start()

      p2PNetworkImpl1.addStaticPeer(MultiaddrPeerAddress.fromAddress(PEER_ADDRESS_NODE_2))

      awaitUntilAsserted { assertNetworkHasPeers(network = p2PNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      p2PNetworkImpl1.removeStaticPeer(MultiaddrPeerAddress.fromAddress(PEER_ADDRESS_NODE_2))

      awaitUntilAsserted { assertNetworkHasPeers(network = p2PNetworkImpl1, peers = 0) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 0) }
    } finally {
      p2PNetworkImpl1.stop()
      p2pNetworkImpl2.stop()
    }
  }

  @Test
  fun `static peers can be configured`() {
    val p2PNetworkImpl1 =
      P2PNetworkImpl(
        privateKeyBytes = key1,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT1, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    val p2pNetworkImpl2 =
      P2PNetworkImpl(
        privateKeyBytes = key2,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1)),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    try {
      p2PNetworkImpl1.start()

      p2pNetworkImpl2.start()

      awaitUntilAsserted { assertNetworkHasPeers(network = p2PNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }
    } finally {
      p2PNetworkImpl1.stop()
      p2pNetworkImpl2.stop()
    }
  }

  @Test
  fun `static peers reconnect`() {
    val p2PNetworkImpl1 =
      P2PNetworkImpl(
        privateKeyBytes = key1,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT1, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    val p2pNetworkImpl2 =
      P2PNetworkImpl(
        privateKeyBytes = key2,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1)),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    try {
      p2PNetworkImpl1.start()

      p2pNetworkImpl2.start()

      awaitUntilAsserted { assertNetworkHasPeers(network = p2PNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      p2PNetworkImpl1.dropPeer(PEER_ID_NODE_2, DisconnectReason.TOO_MANY_PEERS)

      awaitUntilAsserted { assertNetworkHasPeers(network = p2PNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }
    } finally {
      p2PNetworkImpl1.stop()
      p2pNetworkImpl2.stop()
    }
  }

  @Test
  fun `two peers can gossip with each other`() {
    val p2pNetworkImpl1 =
      P2PNetworkImpl(
        privateKeyBytes = key1,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT1, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    val p2pNetworkImpl2 =
      P2PNetworkImpl(
        privateKeyBytes = key2,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1)),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    try {
      p2pNetworkImpl1.start()

      val blocksReceived = mutableListOf<SealedBeaconBlock>()
      p2pNetworkImpl2.start()
      p2pNetworkImpl2.subscribeToBlocks {
        blocksReceived.add(it)
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetworkImpl2, peers = 1) }

      val randomBlockMessage1 = DataGenerators.randomBlockMessage()
      p2pNetworkImpl1.broadcastMessage(randomBlockMessage1).get()
      val randomBlockMessage2 = DataGenerators.randomBlockMessage()
      p2pNetworkImpl1.broadcastMessage(randomBlockMessage2).get()

      awaitUntilAsserted {
        assertThat(blocksReceived).hasSameElementsAs(listOf(randomBlockMessage1.payload, randomBlockMessage2.payload))
      }
    } finally {
      p2pNetworkImpl1.stop()
      p2pNetworkImpl2.stop()
    }
  }

  @Test
  fun `peer receiving gossip passes message on`() {
    val p2PNetworkImpl1 =
      P2PNetworkImpl(
        privateKeyBytes = key1,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT1, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    val p2PNetworkImpl2 =
      P2PNetworkImpl(
        privateKeyBytes = key2,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1, PEER_ADDRESS_NODE_3)),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    val p2PNetworkImpl3 =
      P2PNetworkImpl(
        privateKeyBytes = key3,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT3, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    try {
      p2PNetworkImpl1.start()

      p2PNetworkImpl2.start()
      p2PNetworkImpl2.subscribeToBlocks { SafeFuture.completedFuture(ValidationResult.Companion.Valid) }

      val blockReceived = SafeFuture<SealedBeaconBlock>()
      p2PNetworkImpl3.start()
      p2PNetworkImpl3.subscribeToBlocks {
        blockReceived.complete(it)
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }

      awaitUntilAsserted { assertNetworkIsConnectedToPeer(p2PNetworkImpl1, PEER_ID_NODE_2) }
      awaitUntilAsserted { assertNetworkIsConnectedToPeer(p2PNetworkImpl3, PEER_ID_NODE_2) }

      assertNetworkHasPeers(network = p2PNetworkImpl1, peers = 1)
      assertNetworkHasPeers(network = p2PNetworkImpl2, peers = 2)
      assertNetworkHasPeers(network = p2PNetworkImpl3, peers = 1)

      sleep(1100L) // to make sure that the peers have communicated that they have subscribed to the topic
      // This sleep can be decreased if the heartbeat is decreased (set to 1s for now, see P2PNetworkFactory) in the GossipRouter

      val randomBlockMessage = DataGenerators.randomBlockMessage()
      p2PNetworkImpl1.broadcastMessage(randomBlockMessage)

      assertThat(
        blockReceived.get(100, TimeUnit.MILLISECONDS),
      ).isEqualTo(randomBlockMessage.payload)
    } finally {
      p2PNetworkImpl1.stop()
      p2PNetworkImpl2.stop()
      p2PNetworkImpl3.stop()
    }
  }

  @Test
  fun `peer can send a request`() {
    val p2PNetworkImpl1 =
      P2PNetworkImpl(
        privateKeyBytes = key1,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT1, staticPeers = emptyList()),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    val p2pManagerImpl2 =
      P2PNetworkImpl(
        privateKeyBytes = key2,
        p2pConfig = P2P(ipAddress = IPV4, port = PORT2, staticPeers = listOf(PEER_ADDRESS_NODE_1)),
        chainId = chainId,
        serializer = RLPSerializers.SealedBeaconBlockSerializer,
      )
    try {
      p2PNetworkImpl1.start()

      p2pManagerImpl2.start()

      awaitUntilAsserted { assertNetworkHasPeers(network = p2PNetworkImpl1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pManagerImpl2, peers = 1) }

      val request = Bytes.wrap(byteArrayOf(0, 0, 1, 2, 3, 4))
      val maruRpcResponseHandler = MaruRpcResponseHandler()
      val responseFuture = p2pManagerImpl2.sendRequest(PEER_ID_NODE_1, MaruRpcMethod(), request, maruRpcResponseHandler)
      responseFuture.thenPeek {
        it.rpcStream.closeWriteStream()
      }

      assertThatNoException().isThrownBy { responseFuture.get(500L, TimeUnit.MILLISECONDS) }
      assertThat(
        maruRpcResponseHandler.response().get(500L, TimeUnit.MILLISECONDS),
      ).isEqualTo(request.reverse())
    } finally {
      p2PNetworkImpl1.stop()
      p2pManagerImpl2.stop()
    }
  }

  private fun assertNetworkHasPeers(
    network: P2PNetworkImpl,
    peers: Int,
  ) {
    assertThat(network.peerCount).isEqualTo(peers)
  }

  private fun awaitUntilAsserted(
    timeout: Long = 6000L,
    timeUnit: TimeUnit = TimeUnit.MILLISECONDS,
    condition: () -> Unit,
  ) {
    await()
      .timeout(timeout, timeUnit)
      .untilAsserted(condition)
  }

  private fun assertNetworkIsConnectedToPeer(
    p2pNetwork3: P2PNetworkImpl,
    peer: String,
  ) {
    assertThat(
      p2pNetwork3.isConnected(peer),
    ).isTrue()
  }
}
