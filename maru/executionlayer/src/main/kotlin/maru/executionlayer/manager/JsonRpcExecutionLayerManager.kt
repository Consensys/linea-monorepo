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
import maru.executionlayer.client.ExecutionLayerEngineApiClient
import maru.mappers.Mappers.toDomain
import maru.mappers.Mappers.toPayloadAttributesV1
import org.apache.logging.log4j.LogManager
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes8

class JsonRpcExecutionLayerManager(
  private val executionLayerEngineApiClient: ExecutionLayerEngineApiClient,
) : ExecutionLayerManager {
  private val log = LogManager.getLogger(this.javaClass)

  private var payloadId: ByteArray? = null

  override fun setHeadAndStartBlockBuilding(
    headHash: ByteArray,
    safeHash: ByteArray,
    finalizedHash: ByteArray,
    nextBlockTimestamp: Long,
    feeRecipient: ByteArray,
  ): SafeFuture<ForkChoiceUpdatedResult> {
    log.debug(
      "Trying to create a new block with timestamp={}",
      nextBlockTimestamp,
    )
    val payloadAttributes =
      PayloadAttributes(
        timestamp = nextBlockTimestamp,
        suggestedFeeRecipient = feeRecipient,
      )
    log.debug("Starting block building with payloadAttributes={}", payloadAttributes)
    return forkChoiceUpdate(headHash, safeHash, finalizedHash, payloadAttributes).thenPeek {
      log.debug("Setting payload Id, nextBlockTimestamp={}", nextBlockTimestamp)
      payloadId = it.payloadId
    }
  }

  override fun finishBlockBuilding(): SafeFuture<ExecutionPayload> {
    if (payloadId == null) {
      return SafeFuture.failedFuture(
        IllegalStateException(
          "finishBlockBuilding is called before setHeadAndStartBlockBuilding was completed",
        ),
      )
    }
    return executionLayerEngineApiClient.getPayload(Bytes8(Bytes.wrap(payloadId!!))).thenApply { payloadResponse ->
      if (payloadResponse.isSuccess) {
        payloadResponse.payload
      } else {
        throw IllegalStateException("engine_getPayload request failed! Cause: " + payloadResponse.errorMessage)
      }
    }
  }

  override fun setHead(
    headHash: ByteArray,
    safeHash: ByteArray,
    finalizedHash: ByteArray,
  ): SafeFuture<ForkChoiceUpdatedResult> =
    forkChoiceUpdate(
      headHash = headHash,
      safeHash = safeHash,
      finalizedHash = finalizedHash,
      payloadAttributes = null,
    )

  private fun forkChoiceUpdate(
    headHash: ByteArray,
    safeHash: ByteArray,
    finalizedHash: ByteArray,
    payloadAttributes: PayloadAttributes?,
  ): SafeFuture<ForkChoiceUpdatedResult> =
    executionLayerEngineApiClient
      .forkChoiceUpdate(
        ForkChoiceStateV1(
          Bytes32.wrap(headHash),
          Bytes32.wrap(safeHash),
          Bytes32.wrap(finalizedHash),
        ),
        payloadAttributes?.toPayloadAttributesV1(),
      ).thenApply { response ->
        log.debug("Forkchoice update response with payload attributes {}", response)
        if (response.isFailure) {
          throw IllegalStateException(
            "forkChoiceUpdate request failed! nextBlockTimestamp=${
              payloadAttributes?.timestamp
            } " + response.errorMessage,
          )
        } else {
          response.payload.toDomain()
        }
      }

  override fun newPayload(executionPayload: ExecutionPayload): SafeFuture<PayloadStatus> =
    executionLayerEngineApiClient.newPayload(executionPayload).thenApply { payloadStatusResponse ->
      if (payloadStatusResponse.isSuccess) {
        if (payloadStatusResponse.payload == null) {
          throw IllegalStateException(
            "engine_newPayload request failed! blockNumber=${executionPayload.blockNumber} " + "response=" +
              payloadStatusResponse,
          )
        }
        log.debug("Unsetting payload id, after importing blockNumber={}", executionPayload.blockNumber)
        payloadId = null // Not necessary, but it helps to reinforce the order of calls
        payloadStatusResponse.payload.asInternalExecutionPayload().toDomain()
      } else {
        throw IllegalStateException("engine_newPayload request failed! Cause: " + payloadStatusResponse.errorMessage)
      }
    }
}
