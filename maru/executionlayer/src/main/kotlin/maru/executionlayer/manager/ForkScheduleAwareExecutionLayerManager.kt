/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.executionlayer.manager

import kotlinx.datetime.Clock
import maru.config.consensus.ElFork
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ForksSchedule
import maru.core.ExecutionPayload
import tech.pegasys.teku.infrastructure.async.SafeFuture

class ForkScheduleAwareExecutionLayerManager(
  val forksSchedule: ForksSchedule,
  val executionLayerManagerMap: Map<ElFork, ExecutionLayerManager>,
  val clock: Clock = Clock.System,
) : ExecutionLayerManager {
  init {
    ElFork.entries.forEach {
      require(executionLayerManagerMap.containsKey(it)) { "No execution layer manager provided for $it" }
    }
  }

  internal fun getCurrentElFork(): ElFork {
    val currentTimestamp = clock.now().epochSeconds
    val forkSpec = forksSchedule.getForkByTimestamp(currentTimestamp)
    return when (val configuration = forkSpec.configuration) {
      is QbftConsensusConfig -> configuration.elFork
      else -> throw IllegalStateException("No ELFork configured found for $configuration")
    }
  }

  private fun getExecutionLayerManager(): ExecutionLayerManager = executionLayerManagerMap.getValue(getCurrentElFork())

  override fun setHeadAndStartBlockBuilding(
    headHash: ByteArray,
    safeHash: ByteArray,
    finalizedHash: ByteArray,
    nextBlockTimestamp: Long,
    feeRecipient: ByteArray,
    prevRandao: ByteArray,
  ): SafeFuture<ForkChoiceUpdatedResult> =
    getExecutionLayerManager().setHeadAndStartBlockBuilding(
      headHash = headHash,
      safeHash = safeHash,
      finalizedHash = finalizedHash,
      nextBlockTimestamp = nextBlockTimestamp,
      feeRecipient = feeRecipient,
      prevRandao = prevRandao,
    )

  override fun finishBlockBuilding(): SafeFuture<ExecutionPayload> = getExecutionLayerManager().finishBlockBuilding()

  override fun setHead(
    headHash: ByteArray,
    safeHash: ByteArray,
    finalizedHash: ByteArray,
  ): SafeFuture<ForkChoiceUpdatedResult> =
    getExecutionLayerManager().setHead(
      headHash = headHash,
      safeHash = safeHash,
      finalizedHash = finalizedHash,
    )

  override fun newPayload(executionPayload: ExecutionPayload): SafeFuture<PayloadStatus> =
    getExecutionLayerManager().newPayload(executionPayload)

  override fun getLatestBlockHash(): SafeFuture<ByteArray> = getExecutionLayerManager().getLatestBlockHash()

  override fun isOnline(): SafeFuture<Boolean> = getExecutionLayerManager().isOnline()
}
