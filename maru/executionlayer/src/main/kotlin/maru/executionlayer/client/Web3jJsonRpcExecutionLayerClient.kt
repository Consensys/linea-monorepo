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
import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV3
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceUpdatedResult
import tech.pegasys.teku.ethereum.executionclient.schema.GetPayloadV3Response
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV3
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadStatusV1
import tech.pegasys.teku.ethereum.executionclient.schema.Response
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes8

class Web3jJsonRpcExecutionLayerClient(
  private val web3jEngineClient: Web3JExecutionEngineClient,
  private val web3jClient: Web3JClient,
) : ExecutionLayerClient {
  override fun getLatestBlockMetadata(): SafeFuture<BlockNumberAndHash> =
    SafeFuture
      .of(
        web3jClient.eth1Web3j
          .ethGetBlockByNumber(DefaultBlockParameter.valueOf("latest"), false)
          .sendAsync()
          .minimalCompletionStage(),
      ).thenApply {
        BlockNumberAndHash(
          it.block.number
            .toLong()
            .toULong(),
          Bytes32.fromHexString(it.block.hash),
        )
      }

  override fun getPayload(payloadId: Bytes8): SafeFuture<Response<GetPayloadV3Response>> =
    web3jEngineClient.getPayloadV3(payloadId)

  override fun newPayload(executionPayload: ExecutionPayloadV3): SafeFuture<Response<PayloadStatusV1>> =
    web3jEngineClient.newPayloadV3(executionPayload, emptyList(), Bytes32.ZERO)

  override fun forkChoiceUpdate(
    forkChoiceState: ForkChoiceStateV1,
    payloadAttributes: PayloadAttributesV3?,
  ): SafeFuture<Response<ForkChoiceUpdatedResult>> =
    web3jEngineClient.forkChoiceUpdatedV3(forkChoiceState, Optional.ofNullable(payloadAttributes))
}
