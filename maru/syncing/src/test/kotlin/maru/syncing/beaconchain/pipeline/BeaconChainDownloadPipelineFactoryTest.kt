/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.beaconchain.pipeline

import java.util.concurrent.ExecutorService
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import maru.consensus.blockimport.SealedBeaconBlockImporter
import maru.core.SealedBeaconBlock
import maru.core.ext.DataGenerators.randomSealedBeaconBlock
import maru.p2p.MaruPeer
import maru.p2p.PeerLookup
import maru.p2p.ValidationResult
import maru.p2p.messages.BeaconBlocksByRangeResponse
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.Mockito.times
import org.mockito.kotlin.any
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture.completedFuture

class BeaconChainDownloadPipelineFactoryTest {
  private lateinit var blockImporter: SealedBeaconBlockImporter<ValidationResult>
  private lateinit var peerLookup: PeerLookup
  private lateinit var factory: BeaconChainDownloadPipelineFactory
  private lateinit var executorService: ExecutorService
  private lateinit var syncTargetProvider: () -> ULong

  @BeforeEach
  fun setUp() {
    blockImporter = mock()
    peerLookup = mock()
    executorService = Executors.newCachedThreadPool()
    syncTargetProvider = mock()
    factory =
      BeaconChainDownloadPipelineFactory(
        blockImporter = blockImporter,
        metricsSystem = NoOpMetricsSystem(),
        peerLookup = peerLookup,
        config = BeaconChainDownloadPipelineFactory.Config(),
        syncTargetProvider = syncTargetProvider,
      )
  }

  @AfterEach
  fun tearDown() {
    executorService.shutdownNow()
  }

  @Test
  fun `pipeline processes blocks in correct ranges`() {
    val peer = mock<MaruPeer>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    val rangeResponses = mutableMapOf<Pair<ULong, ULong>, List<SealedBeaconBlock>>()

    // Ranges: [100, 109], [110, 119], [120, 125]
    rangeResponses[100uL to 10uL] = (100uL..109uL).map { randomSealedBeaconBlock(it) }
    rangeResponses[110uL to 10uL] = (110uL..119uL).map { randomSealedBeaconBlock(it) }
    rangeResponses[120uL to 6uL] = (120uL..125uL).map { randomSealedBeaconBlock(it) }

    rangeResponses.forEach { (range, blocks) ->
      val response = mock<BeaconBlocksByRangeResponse>()
      whenever(response.blocks).thenReturn(blocks)
      whenever(peer.sendBeaconBlocksByRange(range.first, range.second)).thenReturn(completedFuture(response))
    }

    whenever(blockImporter.importBlock(any())).thenReturn(
      completedFuture(ValidationResult.Companion.Valid),
    )
    whenever(syncTargetProvider.invoke()).thenReturn(125uL)

    val pipeline = factory.createPipeline(100uL)
    val completionFuture = pipeline.start(executorService)

    // Wait for completion
    completionFuture.get(5, TimeUnit.SECONDS)

    // Verify all blocks were imported
    val numberOfImportedBlocks = 125 - 100 + 1 // Total blocks from 100 to 125 inclusive
    verify(blockImporter, times(numberOfImportedBlocks)).importBlock(any())
  }

  @Test
  fun `pipeline adapts to increased syncTarget during execution`() {
    val peer = mock<MaruPeer>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    val rangeResponses = mutableMapOf<Pair<ULong, ULong>, List<SealedBeaconBlock>>()

    // Ranges: [100, 109], [110, 119], [120, 125]
    rangeResponses[100uL to 10uL] = (100uL..109uL).map { randomSealedBeaconBlock(it) }
    rangeResponses[110uL to 10uL] = (110uL..119uL).map { randomSealedBeaconBlock(it) }
    rangeResponses[120uL to 6uL] = (120uL..125uL).map { randomSealedBeaconBlock(it) }

    rangeResponses.forEach { (range, blocks) ->
      val response = mock<BeaconBlocksByRangeResponse>()
      whenever(response.blocks).thenReturn(blocks)
      whenever(peer.sendBeaconBlocksByRange(range.first, range.second)).thenReturn(completedFuture(response))
    }

    whenever(blockImporter.importBlock(any())).thenReturn(
      completedFuture(ValidationResult.Companion.Valid),
    )
    // the initial sync target is 119, but we will change it to 125 during execution
    whenever(syncTargetProvider.invoke()).thenReturn(119uL, 125uL, 125uL)

    val pipeline = factory.createPipeline(100uL)
    val completionFuture = pipeline.start(executorService)

    // Wait for completion
    completionFuture.get(5, TimeUnit.SECONDS)

    // Verify all blocks were imported
    val numberOfImportedBlocks = 125 - 100 + 1 // Total blocks from 100 to 125 inclusive
    verify(blockImporter, times(numberOfImportedBlocks)).importBlock(any())
  }

  @Test
  fun `pipeline handles single block range`() {
    val peer = mock<MaruPeer>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    val blocks = listOf(randomSealedBeaconBlock(42uL))
    val response = mock<BeaconBlocksByRangeResponse>()
    whenever(response.blocks).thenReturn(blocks)
    whenever(peer.sendBeaconBlocksByRange(42uL, 1uL)).thenReturn(completedFuture(response))

    whenever(blockImporter.importBlock(any())).thenReturn(
      completedFuture(ValidationResult.Companion.Valid),
    )
    whenever(syncTargetProvider.invoke()).thenReturn(42uL)

    val pipeline = factory.createPipeline(42uL)
    val completionFuture = pipeline.start(executorService)

    completionFuture.get(5, TimeUnit.SECONDS)

    verify(peer).sendBeaconBlocksByRange(42uL, 1uL)
    verify(blockImporter).importBlock(blocks[0])
  }

  @Test
  fun `pipeline with large request size processes correct ranges`() {
    val largeRequestSizeFactory =
      BeaconChainDownloadPipelineFactory(
        blockImporter = blockImporter,
        metricsSystem = NoOpMetricsSystem(),
        peerLookup = peerLookup,
        config = BeaconChainDownloadPipelineFactory.Config(blocksBatchSize = 100u),
        syncTargetProvider = { 50uL },
      )

    val peer = mock<MaruPeer>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    // Create blocks for range [0, 50]
    val blocks = (0uL..50uL).map { randomSealedBeaconBlock(it) }
    val response = mock<BeaconBlocksByRangeResponse>()
    whenever(response.blocks).thenReturn(blocks)
    whenever(peer.sendBeaconBlocksByRange(0uL, 51uL)).thenReturn(completedFuture(response))

    whenever(blockImporter.importBlock(any())).thenReturn(
      completedFuture(ValidationResult.Companion.Valid),
    )

    val pipeline = largeRequestSizeFactory.createPipeline(0uL)
    val completionFuture = pipeline.start(executorService)

    completionFuture.get(5, TimeUnit.SECONDS)

    // Should make only one request since request size (100) is larger than range
    verify(peer).sendBeaconBlocksByRange(0uL, 51uL)
  }

  @Test
  fun `factory construction throws when requestSize is zero`() {
    assertThatThrownBy {
      BeaconChainDownloadPipelineFactory(
        blockImporter = blockImporter,
        metricsSystem = NoOpMetricsSystem(),
        peerLookup = peerLookup,
        config = BeaconChainDownloadPipelineFactory.Config(blocksBatchSize = 0u),
        syncTargetProvider = { 0uL },
      )
    }.isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("Request size must be greater than 0")
  }

  @Test
  fun `pipeline handles ranges near ULong MAX_VALUE without overflow`() {
    val peer = mock<MaruPeer>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    val block1 = randomSealedBeaconBlock(ULong.MAX_VALUE - 2uL)
    val block2 = randomSealedBeaconBlock(ULong.MAX_VALUE - 1uL)

    // Test with a range very close to ULong.MAX_VALUE
    val startBlock = ULong.MAX_VALUE - 2uL

    // The expected ranges with request size 2
    whenever(
      peer.sendBeaconBlocksByRange(startBlock, 2uL),
    ).thenReturn(completedFuture(BeaconBlocksByRangeResponse(listOf(block1))))

    whenever(
      peer.sendBeaconBlocksByRange(startBlock + 1uL, 1uL),
    ).thenReturn(completedFuture(BeaconBlocksByRangeResponse(listOf(block2))))

    whenever(blockImporter.importBlock(any())).thenReturn(completedFuture(ValidationResult.Companion.Valid))
    whenever(syncTargetProvider.invoke()).thenReturn(ULong.MAX_VALUE - 1uL)

    val pipeline = factory.createPipeline(startBlock)
    val completionFuture = pipeline.start(executorService)

    // Should complete without overflow errors
    completionFuture.get(5, TimeUnit.SECONDS)
    assertThat(completionFuture).isCompleted
  }
}
