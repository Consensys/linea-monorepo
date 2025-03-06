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
package maru.executionlayer.client

import java.util.Optional
import maru.executionlayer.manager.BlockMetadata
import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceUpdatedResult
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV1
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadStatusV1
import tech.pegasys.teku.ethereum.executionclient.schema.Response
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes8

class Web3jJsonRpcExecutionLayerClient(
  private val web3jEngineClient: Web3JExecutionEngineClient,
  private val web3jEthereumApiClient: Web3JClient,
) : ExecutionLayerClient {
  override fun getLatestBlockMetadata(): SafeFuture<BlockMetadata> =
    SafeFuture
      .of(
        web3jEthereumApiClient.eth1Web3j
          .ethGetBlockByNumber(DefaultBlockParameter.valueOf("latest"), false)
          .sendAsync()
          .minimalCompletionStage(),
      ).thenApply {
        val block = it.block
        BlockMetadata(
          block.number
            .toLong()
            .toULong(),
          Bytes32.fromHexString(block.hash).toArray(),
          block.timestamp.toLong(),
        )
      }

  override fun getPayload(payloadId: Bytes8): SafeFuture<Response<ExecutionPayloadV1>> =
    web3jEngineClient.getPayloadV1(payloadId)

  override fun newPayload(executionPayload: ExecutionPayloadV1): SafeFuture<Response<PayloadStatusV1>> =
    web3jEngineClient.newPayloadV1(executionPayload)

  override fun forkChoiceUpdate(
    forkChoiceState: ForkChoiceStateV1,
    payloadAttributes: PayloadAttributesV1?,
  ): SafeFuture<Response<ForkChoiceUpdatedResult>> =
    web3jEngineClient.forkChoiceUpdatedV1(forkChoiceState, Optional.ofNullable(payloadAttributes))
}
