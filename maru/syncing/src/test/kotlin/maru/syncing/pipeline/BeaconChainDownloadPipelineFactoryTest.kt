/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.pipeline

import java.util.concurrent.ExecutorService
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import maru.consensus.blockimport.SealedBeaconBlockImporter
import maru.core.SealedBeaconBlock
import maru.core.ext.DataGenerators
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
import tech.pegasys.teku.infrastructure.async.SafeFuture

class BeaconChainDownloadPipelineFactoryTest {
  private lateinit var blockImporter: SealedBeaconBlockImporter<ValidationResult>
  private lateinit var peerLookup: PeerLookup
  private lateinit var factory: BeaconChainDownloadPipelineFactory
  private lateinit var executorService: ExecutorService

  @BeforeEach
  fun setUp() {
    blockImporter = mock()
    peerLookup = mock()
    executorService = Executors.newCachedThreadPool()
    factory =
      BeaconChainDownloadPipelineFactory(
        blockImporter = blockImporter,
        metricsSystem = NoOpMetricsSystem(),
        peerLookup = peerLookup,
        downloaderParallelism = 1,
        requestSize = 10u,
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
    rangeResponses[100uL to 10uL] = (100uL..109uL).map { DataGenerators.randomSealedBeaconBlock(it) }
    rangeResponses[110uL to 10uL] = (110uL..119uL).map { DataGenerators.randomSealedBeaconBlock(it) }
    rangeResponses[120uL to 6uL] = (120uL..125uL).map { DataGenerators.randomSealedBeaconBlock(it) }

    rangeResponses.forEach { (range, blocks) ->
      val response = mock<BeaconBlocksByRangeResponse>()
      whenever(response.blocks).thenReturn(blocks)
      whenever(peer.sendBeaconBlocksByRange(range.first, range.second)).thenReturn(SafeFuture.completedFuture(response))
    }

    whenever(blockImporter.importBlock(any())).thenReturn(
      SafeFuture.completedFuture(ValidationResult.Companion.Valid),
    )

    val pipeline = factory.createPipeline(100uL, 125uL)
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

    val blocks = listOf(DataGenerators.randomSealedBeaconBlock(42uL))
    val response = mock<BeaconBlocksByRangeResponse>()
    whenever(response.blocks).thenReturn(blocks)
    whenever(peer.sendBeaconBlocksByRange(42uL, 1uL)).thenReturn(SafeFuture.completedFuture(response))

    whenever(blockImporter.importBlock(any())).thenReturn(
      SafeFuture.completedFuture(ValidationResult.Companion.Valid),
    )

    val pipeline = factory.createPipeline(42uL, 42uL)
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
        downloaderParallelism = 1,
        requestSize = 100u,
      )

    val peer = mock<MaruPeer>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    // Create blocks for range [0, 50]
    val blocks = (0uL..50uL).map { DataGenerators.randomSealedBeaconBlock(it) }
    val response = mock<BeaconBlocksByRangeResponse>()
    whenever(response.blocks).thenReturn(blocks)
    whenever(peer.sendBeaconBlocksByRange(0uL, 51uL)).thenReturn(SafeFuture.completedFuture(response))

    whenever(blockImporter.importBlock(any())).thenReturn(
      SafeFuture.completedFuture(ValidationResult.Companion.Valid),
    )

    val pipeline = largeRequestSizeFactory.createPipeline(0uL, 50uL)
    val completionFuture = pipeline.start(executorService)

    completionFuture.get(5, TimeUnit.SECONDS)

    // Should make only one request since request size (100) is larger than range
    verify(peer).sendBeaconBlocksByRange(0uL, 51uL)
  }

  @Test
  fun `factory creates multiple independent pipelines`() {
    val pipeline1 = factory.createPipeline(0uL, 100uL)
    val pipeline2 = factory.createPipeline(200uL, 300uL)

    assertThat(pipeline1).isNotNull()
    assertThat(pipeline2).isNotNull()
    assertThat(pipeline1).isNotSameAs(pipeline2)
  }

  @Test
  fun `createPipeline throws when startBlock is greater than endBlock`() {
    assertThatThrownBy {
      factory.createPipeline(100uL, 50uL)
    }.isInstanceOf(IllegalStateException::class.java)
      .hasMessageContaining("Start block (100) must be less than or equal to end block (50)")
  }

  @Test
  fun `factory construction throws when requestSize is zero`() {
    assertThatThrownBy {
      BeaconChainDownloadPipelineFactory(
        blockImporter = blockImporter,
        metricsSystem = NoOpMetricsSystem(),
        peerLookup = peerLookup,
        downloaderParallelism = 2,
        requestSize = 0u,
      )
    }.isInstanceOf(IllegalArgumentException::class.java)
      .hasMessageContaining("Request size must be greater than 0")
  }

  @Test
  fun `pipeline handles ranges near ULong MAX_VALUE without overflow`() {
    val peer = mock<MaruPeer>()
    whenever(peerLookup.getPeers()).thenReturn(listOf(peer))

    // Test with a range very close to ULong.MAX_VALUE
    val startBlock = ULong.MAX_VALUE - 20uL
    val endBlock = ULong.MAX_VALUE - 1uL

    // The expected ranges with request size 10
    val response1 = mock<BeaconBlocksByRangeResponse>()
    whenever(response1.blocks).thenReturn(emptyList())
    whenever(peer.sendBeaconBlocksByRange(startBlock, 10uL)).thenReturn(SafeFuture.completedFuture(response1))

    val response2 = mock<BeaconBlocksByRangeResponse>()
    whenever(response2.blocks).thenReturn(emptyList())
    whenever(
      peer.sendBeaconBlocksByRange(ULong.MAX_VALUE - 10uL, 10uL),
    ).thenReturn(SafeFuture.completedFuture(response2))

    val pipeline = factory.createPipeline(startBlock, endBlock)
    val completionFuture = pipeline.start(executorService)

    // Should complete without overflow errors
    completionFuture.get(5, TimeUnit.SECONDS)
    assertThat(completionFuture).isCompleted
  }
}
