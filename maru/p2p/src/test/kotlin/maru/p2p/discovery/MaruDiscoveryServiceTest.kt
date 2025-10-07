/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.discovery

import java.net.InetAddress
import java.net.InetSocketAddress
import java.util.Optional
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration
import linea.kotlin.decodeHex
import linea.kotlin.toULong
import maru.config.P2PConfig
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ConsensusConfig
import maru.consensus.ElFork
import maru.consensus.ForkId
import maru.consensus.ForkIdHashManagerImpl
import maru.consensus.ForkIdHasher
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.core.ext.DataGenerators
import maru.crypto.Hashing
import maru.database.InMemoryBeaconChain
import maru.database.InMemoryP2PState
import maru.database.P2PState
import maru.p2p.discovery.MaruDiscoveryService.Companion.FORK_ID_HASH_FIELD_NAME
import maru.p2p.discovery.MaruDiscoveryService.Companion.convertSafeNodeRecordToDiscoveryPeer
import maru.p2p.discovery.MaruDiscoveryService.Companion.isValidNodeRecord
import maru.serialization.ForkIdSerializer
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.crypto.SECP256K1
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.kotlin.await
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
    // private val IPV4 = NetworkHelper.listIpsV4(excludeLoopback = true).first()
    // Tests seem to fail when using IP address of the machine... ¯\_(ツ)_/¯
    private val IPV4 = "127.0.0.1"

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
    val forkSpec = ForkSpec(0UL, 1u, consensusConfig)
    val forksSchedule = ForksSchedule(chainId = chainId, forks = listOf(forkSpec))

    private val forkIdHashProvider =
      ForkIdHashManagerImpl(
        chainId = chainId,
        beaconChain = beaconChain,
        forksSchedule = forksSchedule,
        forkIdHasher = ForkIdHasher(ForkIdSerializer, Hashing::shortShaHash),
      )

    val otherForkSpec = ForkSpec(1UL, 1u, consensusConfig)
  }

  private lateinit var p2PState: P2PState
  private lateinit var service: MaruDiscoveryService

  private val keyPair = SECP256K1.KeyPair.random()
  private val publicKey = Functions.deriveCompressedPublicKeyFromPrivate(keyPair.secretKey())
  private val dummyAddr = Optional.of(InetSocketAddress(InetAddress.getByName("1.1.1.1"), 1234))

  private fun createValidNodeRecord(
    forkIdHash: ByteArray? = forkIdHashProvider.currentHash(),
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
      P2PConfig(
        ipAddress = "127.0.0.1",
        port = 9001u,
        discovery =
          P2PConfig.Discovery(
            port = 9000u,
            bootnodes = listOf(),
            refreshInterval = 10.seconds,
          ),
      )
    p2PState = InMemoryP2PState()
    service =
      MaruDiscoveryService(
        privateKeyBytes = keyPair.secretKey().bytesArray(),
        p2pConfig = p2pConfig,
        forkIdHashManager = forkIdHashProvider,
        p2PState = p2PState,
      )
  }

  @Test
  fun `converts node record with valid forkId`() {
    val node = createValidNodeRecord()

    val peer = convertSafeNodeRecordToDiscoveryPeer(node)

    assertEquals(publicKey, peer.publicKey)
    assertEquals(dummyAddr.get(), peer.nodeAddress)
    assertEquals(Bytes.wrap(forkIdHashProvider.currentHash()), peer.forkIdBytes)
  }

  @Test
  fun `seq number updates on local record updates`() {
    val sequenceNumberAfterInitialization = 1uL
    assertThat(p2PState.getLocalNodeRecordSequenceNumber()).isEqualTo(sequenceNumberAfterInitialization)
    service.start()
    assertThat(
      service
        .getLocalNodeRecord()
        .seq
        .toBigInteger()
        .toULong(),
    ).isEqualTo(sequenceNumberAfterInitialization)

    service.updateForkIdHash(Bytes.wrap("update 1".toByteArray()))
    assertThat(p2PState.getLocalNodeRecordSequenceNumber()).isEqualTo(sequenceNumberAfterInitialization + 1uL)
    assertThat(
      service
        .getLocalNodeRecord()
        .seq
        .toBigInteger()
        .toULong(),
    ).isEqualTo(sequenceNumberAfterInitialization + 1uL)

    service.updateForkIdHash(Bytes.wrap("update 2".toByteArray()))
    assertThat(p2PState.getLocalNodeRecordSequenceNumber()).isEqualTo(sequenceNumberAfterInitialization + 2uL)
    assertThat(
      service
        .getLocalNodeRecord()
        .seq
        .toBigInteger()
        .toULong(),
    ).isEqualTo(sequenceNumberAfterInitialization + 2uL)

    service.stop()
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
        ForkIdHasher(ForkIdSerializer, Hashing::shortShaHash).hash(differentForkId),
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
          P2PConfig(
            ipAddress = IPV4,
            port = PORT1,
            discovery =
              P2PConfig.Discovery(
                port = PORT2,
                bootnodes = emptyList(),
                refreshInterval = 10.seconds,
              ),
          ),
        forkIdHashManager = forkIdHashProvider,
        p2PState = InMemoryP2PState(),
      )

    val discoveryService2 =
      MaruDiscoveryService(
        privateKeyBytes = key2,
        p2pConfig =
          P2PConfig(
            ipAddress = IPV4,
            port = PORT3,
            discovery =
              P2PConfig.Discovery(
                port = PORT4,
                bootnodes = listOf(bootnode.getLocalNodeRecord().asEnr()),
                refreshInterval = 500.milliseconds,
              ),
          ),
        forkIdHashManager = forkIdHashProvider,
        p2PState = InMemoryP2PState(),
      )

    val discoveryService3 =
      MaruDiscoveryService(
        privateKeyBytes = key3,
        p2pConfig =
          P2PConfig(
            ipAddress = IPV4,
            port = PORT5,
            discovery =
              P2PConfig.Discovery(
                port = PORT6,
                bootnodes = listOf(bootnode.getLocalNodeRecord().asEnr()),
                refreshInterval = 500.milliseconds,
              ),
          ),
        forkIdHashManager = forkIdHashProvider,
        p2PState = InMemoryP2PState(),
      )

    try {
      bootnode.start()
      discoveryService2.start()
      discoveryService3.start()

      await
        .timeout(15.seconds.toJavaDuration())
        .untilAsserted {
          val foundPeers =
            discoveryService2
              .searchForPeers()
              .join()

          foundPeersContains(foundPeers, bootnode, discoveryService3)
        }

      await
        .timeout(10.seconds.toJavaDuration())
        .untilAsserted {
          val foundPeers =
            discoveryService3
              .searchForPeers()
              .join()
          foundPeersContains(foundPeers, bootnode, discoveryService2)
        }
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
  fun `isValidNodeRecord returns true for valid node record`() {
    val node = createValidNodeRecord()

    val result = isValidNodeRecord(forkIdHashProvider, node)

    assertTrue(result)
  }

  @Test
  fun `isValidNodeRecord returns false when forkId field is missing`() {
    val node = createValidNodeRecord(forkIdHash = null)

    val result = isValidNodeRecord(forkIdHashProvider, node)

    assertThat(result).isFalse()
  }

  @Test
  fun `isValidNodeRecord returns false when address is missing`() {
    val node = createValidNodeRecord(tcpAddress = Optional.empty())

    val result = isValidNodeRecord(forkIdHashProvider, node)

    assertThat(result).isFalse()
  }
}
