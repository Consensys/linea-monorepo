package net.consensys.zkevm.ethereum.coordination.conflation

import net.consensys.linea.BlockNumberAndHash
import net.consensys.zkevm.coordinator.clients.GenerateTracesResponse
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class BlocksTracesConflated(
  val tracesResponse: GenerateTracesResponse,
  val zkStateTraces: GetZkEVMStateMerkleProofResponse
)

interface TracesConflationCoordinator {
  fun conflateExecutionTraces(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<BlocksTracesConflated>
}
