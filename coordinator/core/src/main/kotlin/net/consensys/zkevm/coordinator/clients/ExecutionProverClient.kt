package net.consensys.zkevm.coordinator.clients

import net.consensys.zkevm.domain.BlockInterval
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class GetProofResponse(
  override val startBlockNumber: ULong,
  override val endBlockNumber: ULong
) : BlockInterval

interface ExecutionProverClient {
  /**
   * Creates a batch execution proof request and returns a future that will be completed when the proof is ready.
   */
  fun requestBatchExecutionProof(
    blocks: List<ExecutionPayloadV1>,
    tracesResponse: GenerateTracesResponse,
    type2StateData: GetZkEVMStateMerkleProofResponse
  ): SafeFuture<GetProofResponse>
}
