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
package maru.executionlayer.manager

import maru.core.ExecutionPayload
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class BlockBuildingResult(
  val executionPayload: ExecutionPayload,
  val resultingBlockHash: ByteArray,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BlockBuildingResult

    if (executionPayload != other.executionPayload) return false
    if (!resultingBlockHash.contentEquals(other.resultingBlockHash)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = executionPayload.hashCode()
    result = 31 * result + resultingBlockHash.contentHashCode()
    return result
  }
}

// Consider switching executionPayloadStatus to enum if it's useful
data class PayloadStatus(
  val executionPayloadStatus: String?,
  val latestValidHash: ByteArray?,
  val validationError: String?,
  val failureCause: Throwable?,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as PayloadStatus

    if (executionPayloadStatus != other.executionPayloadStatus) return false
    if (!latestValidHash.contentEquals(other.latestValidHash)) return false
    if (validationError != other.validationError) return false

    return true
  }

  override fun hashCode(): Int {
    var result = executionPayloadStatus.hashCode()
    result = 31 * result + latestValidHash.contentHashCode()
    result = 31 * result + validationError.hashCode()
    return result
  }
}

data class ForkChoiceUpdatedResult(
  val payloadStatus: PayloadStatus,
  val payloadId: ByteArray?,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ForkChoiceUpdatedResult

    if (payloadStatus != other.payloadStatus) return false
    if (payloadId != null) {
      if (other.payloadId == null) return false
      if (!payloadId.contentEquals(other.payloadId)) return false
    } else if (other.payloadId != null) {
      return false
    }

    return true
  }

  override fun hashCode(): Int {
    var result = payloadStatus.hashCode()
    result = 31 * result + (payloadId?.contentHashCode() ?: 0)
    return result
  }
}

interface ExecutionLayerManager {
  fun setHeadAndStartBlockBuilding(
    headHash: ByteArray,
    safeHash: ByteArray,
    finalizedHash: ByteArray,
  ): SafeFuture<ForkChoiceUpdatedResult>

  fun finishBlockBuilding(): SafeFuture<BlockBuildingResult>

  fun latestBlockHeight(): ULong

  fun setHead(
    headHash: ByteArray,
    safeHash: ByteArray,
    finalizedHash: ByteArray,
  ): SafeFuture<ForkChoiceUpdatedResult>

  fun importBlock(executionPayload: ExecutionPayload): ByteArray
}
