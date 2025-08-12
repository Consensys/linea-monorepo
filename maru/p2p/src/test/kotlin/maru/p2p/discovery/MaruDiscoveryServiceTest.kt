/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.discovery

import java.lang.Thread.sleep
import java.net.InetAddress
import java.net.InetSocketAddress
import java.util.Optional
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import linea.kotlin.decodeHex
import maru.config.P2P
import maru.config.consensus.ElFork
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ConsensusConfig
import maru.consensus.ForkId
import maru.consensus.ForkIdHashProviderImpl
import maru.consensus.ForkIdHasher
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.core.ext.DataGenerators
import maru.crypto.Hashing
import maru.database.InMemoryBeaconChain
import maru.p2p.discovery.MaruDiscoveryService.Companion.FORK_ID_HASH_FIELD_NAME
import maru.p2p.getBootnodeEnrString
import maru.serialization.ForkIdSerializers
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.crypto.SECP256K1
import org.assertj.core.api.Assertions.assertThat
import org.ethereum.beacon.discovery.schema.IdentitySchemaInterpreter
import org.ethereum.beacon.discovery.schema.NodeRecord
import org.ethereum.beacon.discovery.schema.NodeRecordBuilder
import org.ethereum.beacon.discovery.schema.NodeRecordFactory
import org.ethereum.beacon.discovery.util.Functions
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class MaruDiscoveryServiceTest {
  companion object {
    private const val IPV4 = "127.0.0.1"

    private const val PORT1 = 9334u
    private const val PORT2 = 9335u
    private const val PORT3 = 9336u
    private const val PORT4 = 9337u
    private const val PORT5 = 9338u
    private const val PORT6 = 9339u

    private val key1 = "0x12c0b113e2b0c37388e2b484112e13f05c92c4471e3ee1dfaa368fa5045325b2".decodeHex()
    private val key2 = "0xf3d2fffa99dc8906823866d96316492ebf7a8478713a89a58b7385af85b088a1".decodeHex()
    private val key3 = "0x4437acb8e84bc346f7640f239da84abe99bc6f97b7855f204e34688d2977fd57".decodeHex()

    private val chainId = 1337u
    private val beaconChain = InMemoryBeaconChain(DataGenerators.randomBeaconState(number = 0u, timestamp = 0u))
    val consensusConfig: ConsensusConfig =
      QbftConsensusConfig(
        validatorSet =
          setOf(
            DataGenerators.randomValidator(),
            DataGenerators.randomValidator(),
          ),
        elFork = ElFork.Prague,
      )
    val forkSpec = ForkSpec(0L, 1, consensusConfig)
    val forksSchedule = ForksSchedule(chainId = chainId, forks = listOf(forkSpec))

    private val forkIdHashProvider =
      ForkIdHashProviderImpl(
        chainId = chainId,
        beaconChain = beaconChain,
        forksSchedule = forksSchedule,
        forkIdHasher = ForkIdHasher(ForkIdSerializers.ForkIdSerializer, Hashing::shortShaHash),
      )

    val otherForkSpec = ForkSpec(1L, 1, consensusConfig)
  }

  private lateinit var service: MaruDiscoveryService

  private val keyPair = SECP256K1.KeyPair.random()
  private val publicKey = Functions.deriveCompressedPublicKeyFromPrivate(keyPair.secretKey())
  private val dummyAddr = Optional.of(InetSocketAddress(InetAddress.getByName("1.1.1.1"), 1234))

  private fun createValidNodeRecord(
    forkIdHash: ByteArray? = forkIdHashProvider.currentForkIdHash(),
    tcpAddress: Optional<InetSocketAddress> = dummyAddr,
  ): NodeRecord {
    val nrBuilder =
      NodeRecordBuilder()
        .nodeRecordFactory(NodeRecordFactory(IdentitySchemaInterpreter.V4))
        .seq(1)
        .secretKey(keyPair.secretKey())
    if (forkIdHash != null) {
      nrBuilder.customField(FORK_ID_HASH_FIELD_NAME, Bytes.wrap(forkIdHash))
    }
    if (tcpAddress.isPresent) {
      nrBuilder.address(tcpAddress.get().address.hostAddress, tcpAddress.get().port)
    }

    return nrBuilder.build()
  }

  @BeforeEach
  fun setUp() {
    val p2pConfig =
      P2P(
        ipAddress = "127.0.0.1",
        port = 9001u,
        discovery =
          P2P.Discovery(
            port = 9000u,
            bootnodes = listOf(),
            refreshInterval = 10.seconds,
          ),
      )
    service =
      MaruDiscoveryService(
        privateKeyBytes = keyPair.secretKey().bytesArray(),
        p2pConfig = p2pConfig,
        forkIdHashProvider = forkIdHashProvider,
      )
  }

  @Test
  fun `converts node record with valid forkId`() {
    val node = createValidNodeRecord()

    val peer =
      service.run {
        val method = this::class.java.getDeclaredMethod("convertSafeNodeRecordToDiscoveryPeer", NodeRecord::class.java)
        method.isAccessible = true
        method.invoke(this, node) as MaruDiscoveryPeer
      }

    assertEquals(publicKey, peer.publicKey)
    assertEquals(dummyAddr.get(), peer.nodeAddress)
    assertEquals(Bytes.wrap(forkIdHashProvider.currentForkIdHash()), peer.forkIdBytes)
  }

  @Test
  fun `updateForkIdHash updates local`() {
    val localNodeRecordBefore = service.getLocalNodeRecord()

    val differentForkId =
      ForkId(
        chainId = chainId + 2u,
        forkSpec = otherForkSpec,
        genesisRootHash = ByteArray(32),
      )
    val differentForkIdHash =
      Bytes.wrap(
        ForkIdHasher(ForkIdSerializers.ForkIdSerializer, Hashing::shortShaHash).hash(differentForkId),
      )
    service.updateForkIdHash(differentForkIdHash)

    val localNodeRecordAfter = service.getLocalNodeRecord()
    val actual = localNodeRecordAfter.get(FORK_ID_HASH_FIELD_NAME)
    assertThat(actual).isNotEqualTo(localNodeRecordBefore.get(FORK_ID_HASH_FIELD_NAME))
    assertThat(actual).isEqualTo(differentForkIdHash)
  }

  @Test
  fun `discovery finds nodes`() {
    val bootnode =
      MaruDiscoveryService(
        privateKeyBytes = key1,
        p2pConfig =
          P2P(
            ipAddress = IPV4,
            port = PORT1,
            discovery =
              P2P.Discovery(
                port = PORT2,
                bootnodes = emptyList(),
                refreshInterval = 10.seconds,
              ),
          ),
        forkIdHashProvider = forkIdHashProvider,
      )

    val enrString =
      getBootnodeEnrString(
        key1,
        IPV4,
        PORT2.toInt(),
        PORT1.toInt(),
      )

    val discoveryService2 =
      MaruDiscoveryService(
        privateKeyBytes = key2,
        p2pConfig =
          P2P(
            ipAddress = IPV4,
            port = PORT3,
            discovery =
              P2P.Discovery(
                port = PORT4,
                bootnodes = listOf(enrString),
                refreshInterval = 500.milliseconds,
              ),
          ),
        forkIdHashProvider = forkIdHashProvider,
      )

    val discoveryService3 =
      MaruDiscoveryService(
        privateKeyBytes = key3,
        p2pConfig =
          P2P(
            ipAddress = IPV4,
            port = PORT5,
            discovery =
              P2P.Discovery(
                port = PORT6,
                bootnodes = listOf(enrString),
                refreshInterval = 500.milliseconds,
              ),
          ),
        forkIdHashProvider = forkIdHashProvider,
      )

    try {
      bootnode.start()
      discoveryService2.start()
      discoveryService3.start()

      // make sure services are started and bootnodes have been pinged
      sleep(1000)

      discoveryService2
        .searchForPeers()
        .thenAccept { foundPeers ->
          foundPeersContains(foundPeers, bootnode, discoveryService3)
        }.join()
      discoveryService3
        .searchForPeers()
        .thenAccept { foundPeers ->
          foundPeersContains(foundPeers, bootnode, discoveryService2)
        }.join()
    } finally {
      bootnode.stop()
      discoveryService2.stop()
      discoveryService3.stop()
    }
  }

  private fun foundPeersContains(
    foundPeers: Collection<MaruDiscoveryPeer>,
    vararg nodes: MaruDiscoveryService,
  ) {
    nodes.forEach { node -> assertThat(foundPeers.any { it.nodeId == node.getLocalNodeRecord().nodeId }).isTrue }
  }

  @Test
  fun `checkNodeRecord returns true for valid node record`() {
    val node = createValidNodeRecord()

    val result =
      service.run {
        val method = this::class.java.getDeclaredMethod("checkNodeRecord", NodeRecord::class.java)
        method.isAccessible = true
        method.invoke(this, node) as Boolean
      }

    assertTrue(result)
  }

  @Test
  fun `checkNodeRecord returns false when forkId field is missing`() {
    val node = createValidNodeRecord(forkIdHash = null)

    val result =
      service.run {
        val method = this::class.java.getDeclaredMethod("checkNodeRecord", NodeRecord::class.java)
        method.isAccessible = true
        method.invoke(this, node) as Boolean
      }

    assertThat(result).isFalse()
  }

  @Test
  fun `checkNodeRecord returns false when address is missing`() {
    val node = createValidNodeRecord(tcpAddress = Optional.empty())

    val result =
      service.run {
        val method = this::class.java.getDeclaredMethod("checkNodeRecord", NodeRecord::class.java)
        method.isAccessible = true
        method.invoke(this, node) as Boolean
      }

    assertThat(result).isFalse()
  }
}
