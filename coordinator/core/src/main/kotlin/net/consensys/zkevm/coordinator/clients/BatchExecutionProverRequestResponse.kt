package net.consensys.zkevm.coordinator.clients

import build.linea.clients.GetZkEVMStateMerkleProofResponse
import build.linea.domain.BlockInterval
import net.consensys.zkevm.toULong
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1

data class BatchExecutionProofRequestV1(
  val blocks: List<ExecutionPayloadV1>,
  val tracesResponse: GenerateTracesResponse,
  val type2StateData: GetZkEVMStateMerkleProofResponse
) : BlockInterval {
  override val startBlockNumber: ULong
    get() = blocks.first().blockNumber.toULong()
  override val endBlockNumber: ULong
    get() = blocks.last().blockNumber.toULong()
}

data class BatchExecutionProofResponse(
  override val startBlockNumber: ULong,
  override val endBlockNumber: ULong
) : BlockInterval
