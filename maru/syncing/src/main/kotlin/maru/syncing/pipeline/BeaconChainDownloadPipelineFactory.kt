/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.pipeline

import maru.consensus.blockimport.SealedBeaconBlockImporter
import maru.extensions.clampedAdd
import maru.p2p.PeerLookup
import maru.p2p.ValidationResult
import maru.p2p.messages.BeaconBlocksByRangeHandler.Companion.MAX_BLOCKS_PER_REQUEST
import org.hyperledger.besu.metrics.BesuMetricCategory
import org.hyperledger.besu.plugin.services.MetricsSystem
import org.hyperledger.besu.services.pipeline.Pipeline
import org.hyperledger.besu.services.pipeline.PipelineBuilder

class BeaconChainDownloadPipelineFactory(
  private val blockImporter: SealedBeaconBlockImporter<ValidationResult>,
  private val metricsSystem: MetricsSystem,
  private val peerLookup: PeerLookup,
  private val downloaderParallelism: Int,
  private val requestSize: UInt,
) {
  init {
    require(requestSize > 0u) { "Request size must be greater than 0" }
    require(requestSize <= MAX_BLOCKS_PER_REQUEST) { "Request size must not be greater than $MAX_BLOCKS_PER_REQUEST" }
  }

  fun createPipeline(
    startBlock: ULong,
    endBlock: ULong,
  ): Pipeline<SyncTargetRange?> {
    check(startBlock <= endBlock) { "Start block ($startBlock) must be less than or equal to end block ($endBlock)" }
    check(endBlock < ULong.MAX_VALUE) { "End block ($endBlock) must be less than ULong max value" }

    val syncTargetRangeSequence = createTargetRangeSequence(startBlock, endBlock)
    val downloadBlocksStep = DownloadBlocksStep(peerLookup)
    val importBlocksStep = ImportBlocksStep(blockImporter)

    return PipelineBuilder
      .createPipelineFrom(
        "blockNumbers",
        syncTargetRangeSequence.iterator(),
        downloaderParallelism,
        metricsSystem.createLabelledCounter(
          BesuMetricCategory.SYNCHRONIZER,
          "chain_download_pipeline_processed_total",
          "Number of entries process by each chain download pipeline stage",
          "step",
          "action",
        ),
        true,
        "importBlocks",
      ).thenProcessAsyncOrdered("downloadBlocks", downloadBlocksStep, downloaderParallelism)
      .andFinishWith("importBlocks", importBlocksStep)
  }

  private fun createTargetRangeSequence(
    startBlock: ULong,
    endBlock: ULong,
  ): Sequence<SyncTargetRange> =
    sequence {
      var currentStart = startBlock
      val maxRangeSize = requestSize.toULong()

      while (currentStart <= endBlock) {
        val nextEnd = currentStart.clampedAdd(maxRangeSize) - 1uL
        val currentEnd = minOf(nextEnd.toULong(), endBlock)
        yield(SyncTargetRange(currentStart, currentEnd))
        currentStart = currentEnd + 1uL
      }
    }
}
