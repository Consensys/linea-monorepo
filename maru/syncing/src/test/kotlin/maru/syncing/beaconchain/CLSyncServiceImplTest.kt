/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.beaconchain

import java.net.ServerSocket
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean
import kotlin.random.Random
import maru.config.P2P
import maru.config.consensus.ElFork
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ConsensusConfig
import maru.consensus.ForkIdHashProvider
import maru.consensus.ForkIdHasher
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.StaticValidatorProvider
import maru.consensus.qbft.DelayedQbftBlockCreator
import maru.core.BeaconBlock
import maru.core.BeaconBlockHeader
import maru.core.BeaconState
import maru.core.Seal
import maru.core.SealedBeaconBlock
import maru.core.Validator
import maru.core.ext.DataGenerators
import maru.core.ext.metrics.TestMetrics.TestMetricsFacade
import maru.core.ext.metrics.TestMetrics.TestMetricsSystemAdapter
import maru.crypto.Hashing
import maru.database.BeaconChain
import maru.database.InMemoryBeaconChain
import maru.extensions.fromHexToByteArray
import maru.p2p.P2PNetworkImpl
import maru.p2p.PeerLookup
import maru.p2p.messages.StatusMessageFactory
import maru.serialization.ForkIdSerializers
import maru.serialization.rlp.RLPSerializers
import maru.syncing.beaconchain.pipeline.BeaconChainDownloadPipelineFactory.Config
import net.consensys.linea.metrics.Counter
import net.consensys.linea.metrics.MetricsFacade
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.hyperledger.besu.crypto.KeyPair
import org.hyperledger.besu.crypto.SignatureAlgorithm
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory
import org.hyperledger.besu.ethereum.core.Util
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.Mockito.times
import org.mockito.kotlin.any
import org.mockito.kotlin.doAnswer
import org.mockito.kotlin.never
import org.mockito.kotlin.reset
import org.mockito.kotlin.spy
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.networking.p2p.libp2p.MultiaddrPeerAddress

class CLSyncServiceImplTest {
  companion object {
    private const val CHAIN_ID = 1337u
    private const val IPV4 = "127.0.0.1"
    private const val PEER_ID_NODE_2: String = "16Uiu2HAmVXtqhevTAJqZucPbR2W4nCMpetrQASgjZpcxDEDaUPPt"

    private val key1 = "0x0802122012c0b113e2b0c37388e2b484112e13f05c92c4471e3ee1dfaa368fa5045325b2".fromHexToByteArray()
    private val key2 =
      "0x0802122100f3d2fffa99dc8906823866d96316492ebf7a8478713a89a58b7385af85b088a1"
        .fromHexToByteArray()
  }

  private var port1: UInt = 0u
  private var port2: UInt = 0u
  private lateinit var signatureAlgorithm: SignatureAlgorithm
  private lateinit var keypair: KeyPair
  private lateinit var beaconChain1: BeaconChain
  private lateinit var beaconChain2: BeaconChain
  private lateinit var validators: Set<Validator>
  private lateinit var p2pNetwork1: P2PNetworkImpl
  private lateinit var p2pNetwork2: P2PNetworkImpl
  private lateinit var clSyncService: CLSyncServiceImpl
  private lateinit var peerLookup: PeerLookup

  @BeforeEach
  fun setUp() {
    signatureAlgorithm = SignatureAlgorithmFactory.getInstance()
    keypair = signatureAlgorithm.generateKeyPair()
    validators = setOf(Validator(Util.publicKeyToAddress(keypair.publicKey).toArray()))

    val genesisTimestamp = DataGenerators.randomTimestamp()
    val (genesisBeaconState, genesisBeaconBlock) = genesisState(genesisTimestamp, validators)
    beaconChain1 = spy(InMemoryBeaconChain(genesisBeaconState, genesisBeaconBlock))
    beaconChain2 = spy(InMemoryBeaconChain(genesisBeaconState, genesisBeaconBlock))

    port1 = findFreePort()
    port2 = findFreePort()
    p2pNetwork1 = createNetwork(beaconChain1, key1, port1)
    p2pNetwork2 = createNetwork(beaconChain2, key2, port2)

    createBlocks(
      beaconChain = beaconChain2,
      genesisBeaconBlock = genesisBeaconBlock,
      genesisTimestamp = genesisBeaconBlock.beaconBlock.beaconBlockHeader.timestamp,
      validators = validators,
      signatureAlgorithm = signatureAlgorithm,
      keypair = keypair,
    )
    peerLookup = spy(p2pNetwork1.getPeerLookup())
    clSyncService =
      CLSyncServiceImpl(
        beaconChain = beaconChain1,
        executorService = Executors.newCachedThreadPool(),
        validatorProvider = StaticValidatorProvider(validators),
        allowEmptyBlocks = true,
        peerLookup = peerLookup,
        besuMetrics = TestMetricsSystemAdapter,
        metricsFacade = TestMetricsFacade,
        pipelineConfig = Config(blocksBatchSize = 10u, blocksParallelism = 1u),
      )

    try {
      p2pNetwork1.start()
      p2pNetwork2.start()
      p2pNetwork1.addStaticPeer(createPeerAddress(port2))

      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetwork1, peers = 1) }
      awaitUntilAsserted { assertNetworkHasPeers(network = p2pNetwork2, peers = 1) }
    } catch (e: Exception) {
      p2pNetwork1.stop()
      p2pNetwork2.stop()
      throw IllegalStateException("Failed to start P2P networks", e)
    }
  }

  @AfterEach
  fun tearDown() {
    p2pNetwork1.stop()
    p2pNetwork2.stop()
  }

  @Test
  fun `chain sync downloads and imports blocks from another node`() {
    val synced = AtomicBoolean(false)
    clSyncService.setSyncTarget(100uL)
    clSyncService.onSyncComplete { synced.set(true) }
    awaitUntilAsserted { assertThat(synced).isTrue() }
    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(100uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getLatestBeaconState())
    for (i in 1uL..100uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }
  }

  @Test
  fun `sync target can be updated ahead of current target and continue downloading`() {
    // sync to block 50
    val synced = AtomicBoolean(false)
    clSyncService.setSyncTarget(50uL)
    clSyncService.onSyncComplete { synced.set(true) }
    awaitUntilAsserted { assertThat(synced).isTrue() }

    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(50uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getBeaconState(50uL))
    for (i in 1uL..50uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }

    // update sync target to 100
    synced.set(false)
    clSyncService.setSyncTarget(100uL)
    awaitUntilAsserted { assertThat(synced).isTrue() }

    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(100uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getLatestBeaconState())
    for (i in 1uL..100uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }
  }

  @Test
  fun `chain sync download restarts on errors`() {
    val peerLookup = spy(p2pNetwork1.getPeerLookup())

    // Fail the first two calls to getPeers() to simulate failure getting peers and not downloading blocks
    var retries = 0
    doAnswer {
      if (retries < 2) {
        retries++
        throw IllegalStateException("Simulated failure for testing")
      } else {
        it.callRealMethod()
      }
    }.whenever(peerLookup).getPeers()

    val metricsFacade = mock(MetricsFacade::class.java)
    val retriesCounter = mock(Counter::class.java)
    whenever(metricsFacade.createCounter(any(), any(), any(), any())).thenReturn(retriesCounter)

    val synced = AtomicBoolean(false)
    clSyncService.setSyncTarget(100uL)
    clSyncService.onSyncComplete { synced.set(true) }
    awaitUntilAsserted { assertThat(synced).isTrue() }

    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(100uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getLatestBeaconState())
    for (i in 0uL..100uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }

    verify(retriesCounter, times(retries)).increment()
  }

  @Test
  fun `sync target set to same target returns immediately`() {
    // sync to block 50
    val synced = AtomicBoolean(false)
    clSyncService.setSyncTarget(50uL)
    clSyncService.onSyncComplete { synced.set(true) }
    awaitUntilAsserted { assertThat(synced).isTrue() }

    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(50uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getBeaconState(50uL))
    for (i in 1uL..50uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }

    // update the sync target to 50 again
    synced.set(false)
    reset(beaconChain1, beaconChain2)
    clSyncService.setSyncTarget(50uL)
    awaitUntilAsserted { assertThat(synced).isTrue() }

    // Ensure the state remains unchanged
    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(50uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getBeaconState(50uL))
    for (i in 1uL..50uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }

    // Verify that beaconChain for node1 was not updated, and there are no reads on node2 beaconChain
    verify(beaconChain1, never()).newUpdater()
    verify(beaconChain2, never()).getSealedBeaconBlocks(any(), any())
  }

  @Test
  fun `sync target set to older target returns immediately`() {
    // sync to block 50
    val synced = AtomicBoolean(false)
    clSyncService.setSyncTarget(50uL)
    clSyncService.onSyncComplete { synced.set(true) }
    awaitUntilAsserted { assertThat(synced).isTrue() }

    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(50uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getBeaconState(50uL))
    for (i in 1uL..50uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }

    // update the sync target to 40, should return immediately
    synced.set(false)
    reset(beaconChain1, beaconChain2)
    clSyncService.setSyncTarget(40uL)
    awaitUntilAsserted { assertThat(synced).isTrue() }

    // Ensure the state remains unchanged
    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(50uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getBeaconState(50uL))
    for (i in 1uL..50uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }

    // Verify that beaconChain for node1 was not updated, and there are no reads on node2 beaconChain
    verify(beaconChain1, never()).newUpdater()
    verify(beaconChain2, never()).getSealedBeaconBlocks(any(), any())
  }

  @Test
  fun `chain sync continues to download after stopped is called`() {
    // sync to block 50
    val synced = AtomicBoolean(false)
    clSyncService.setSyncTarget(50uL)
    clSyncService.onSyncComplete { synced.set(true) }
    awaitUntilAsserted { assertThat(synced).isTrue() }

    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(50uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getBeaconState(50uL))
    for (i in 1uL..50uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }

    clSyncService.stop()

    // update sync target to 100
    synced.set(false)
    clSyncService.setSyncTarget(100uL)
    awaitUntilAsserted { assertThat(synced).isTrue() }

    assertThat(beaconChain1.getLatestBeaconState().latestBeaconBlockHeader.number).isEqualTo(100uL)
    assertThat(beaconChain1.getLatestBeaconState()).isEqualTo(beaconChain2.getLatestBeaconState())
    for (i in 1uL..100uL) {
      assertThat(beaconChain1.getSealedBeaconBlock(i)).isEqualTo(beaconChain2.getSealedBeaconBlock(i))
      assertThat(beaconChain1.getBeaconState(i)).isEqualTo(beaconChain2.getBeaconState(i))
    }
  }

  @Test
  fun `onSyncComplete handler is called only once per sync`() {
    var callCount = 0
    clSyncService.onSyncComplete { callCount++ }
    clSyncService.setSyncTarget(50uL)
    awaitUntilAsserted { assertThat(callCount).isEqualTo(1) }
  }

  @Test
  fun `multiple onSyncComplete handlers are called`() {
    var handler1Called = false
    var handler2Called = false
    clSyncService.onSyncComplete { handler1Called = true }
    clSyncService.onSyncComplete { handler2Called = true }

    clSyncService.setSyncTarget(50uL)
    awaitUntilAsserted {
      assertThat(handler1Called).isTrue()
      assertThat(handler2Called).isTrue()
    }
  }

  private fun createPeerAddress(port: UInt): MultiaddrPeerAddress =
    MultiaddrPeerAddress.fromAddress("/ip4/$IPV4/tcp/$port/p2p/$PEER_ID_NODE_2")

  private fun genesisState(
    genesisTimestamp: ULong,
    validators: Set<Validator>,
  ): Pair<BeaconState, SealedBeaconBlock> {
    val genesisBeaconBlockHeader =
      BeaconBlockHeader(
        number = 0uL,
        round = 0u,
        timestamp = genesisTimestamp,
        proposer = validators.first(),
        parentRoot = Random.nextBytes(32),
        stateRoot = Random.nextBytes(32),
        bodyRoot = Random.nextBytes(32),
        headerHashFunction = RLPSerializers.DefaultHeaderHashFunction,
      )
    val genesisBeaconState =
      BeaconState(
        latestBeaconBlockHeader = genesisBeaconBlockHeader,
        validators = validators,
      )

    val genesisBeaconBlock =
      SealedBeaconBlock(
        beaconBlock =
          BeaconBlock(
            beaconBlockHeader = genesisBeaconBlockHeader,
            beaconBlockBody = DataGenerators.randomBeaconBlockBody(),
          ),
        commitSeals = setOf(Seal(Random.nextBytes(96))),
      )
    return Pair(genesisBeaconState, genesisBeaconBlock)
  }

  private fun createNetwork(
    beaconChain: BeaconChain,
    key: ByteArray,
    port: UInt,
  ): P2PNetworkImpl {
    val forkIdHashProvider = createForkIdHashProvider(beaconChain)
    val statusMessageFactory = StatusMessageFactory(beaconChain, forkIdHashProvider)
    val p2pNetworkImpl =
      P2PNetworkImpl(
        privateKeyBytes = key,
        p2pConfig =
          P2P(
            ipAddress = IPV4,
            port = port,
            staticPeers = emptyList(),
          ),
        chainId = CHAIN_ID,
        serDe = RLPSerializers.SealedBeaconBlockSerializer,
        metricsFacade = TestMetricsFacade,
        statusMessageFactory = statusMessageFactory,
        beaconChain = beaconChain,
        metricsSystem = TestMetricsSystemAdapter,
        forkIdHashProvider = forkIdHashProvider,
        isBlockImportEnabledProvider = { true },
      )
    return p2pNetworkImpl
  }

  private fun createBlocks(
    beaconChain: BeaconChain,
    genesisBeaconBlock: SealedBeaconBlock,
    genesisTimestamp: ULong,
    validators: Set<Validator>,
    signatureAlgorithm: SignatureAlgorithm,
    keypair: KeyPair,
  ) {
    val updater = beaconChain.newUpdater()
    var parentSealedBeaconBlock = genesisBeaconBlock
    for (i in 1uL..100uL) {
      val beaconBlock =
        DelayedQbftBlockCreator.createBeaconBlock(
          parentSealedBeaconBlock = parentSealedBeaconBlock,
          executionPayload = DataGenerators.randomExecutionPayload(),
          round = 0,
          timestamp = genesisTimestamp + i,
          proposer = validators.first().address,
          validators = validators,
        )
      val seal = signatureAlgorithm.sign(Bytes32.wrap(beaconBlock.beaconBlockHeader.hash), keypair)
      val sealedBlock =
        SealedBeaconBlock(
          beaconBlock = beaconBlock,
          setOf(Seal(seal.encodedBytes().toArray())),
        )
      val beaconState =
        BeaconState(
          latestBeaconBlockHeader = beaconBlock.beaconBlockHeader,
          validators = validators,
        )
      updater.putSealedBeaconBlock(sealedBlock)
      updater.putBeaconState(beaconState)
      parentSealedBeaconBlock = sealedBlock
    }
    updater.commit()
  }

  fun createForkIdHashProvider(beaconChain: BeaconChain): ForkIdHashProvider {
    val consensusConfig: ConsensusConfig =
      QbftConsensusConfig(
        validatorSet =
          setOf(
            DataGenerators.randomValidator(),
            DataGenerators.randomValidator(),
          ),
        elFork = ElFork.Prague,
      )
    val forksSchedule = ForksSchedule(CHAIN_ID, listOf(ForkSpec(0L, 1, consensusConfig)))

    return ForkIdHashProvider(
      chainId = CHAIN_ID,
      beaconChain = beaconChain,
      forksSchedule = forksSchedule,
      forkIdHasher = ForkIdHasher(ForkIdSerializers.ForkIdSerializer, Hashing::shortShaHash),
    )
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

  private fun assertNetworkHasPeers(
    network: P2PNetworkImpl,
    peers: Int,
  ) {
    assertThat(network.getPeers().count()).isEqualTo(peers)
  }

  private fun findFreePort(): UInt =
    runCatching {
      ServerSocket(0).use { socket ->
        socket.reuseAddress = true
        socket.localPort.toUInt()
      }
    }.getOrElse {
      throw IllegalStateException("Could not find a free port", it)
    }
}
