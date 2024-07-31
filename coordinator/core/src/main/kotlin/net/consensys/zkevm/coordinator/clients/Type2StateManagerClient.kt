package net.consensys.zkevm.coordinator.clients

import com.fasterxml.jackson.databind.node.ArrayNode
import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

enum class Type2StateManagerErrorType {
  UNKNOWN,
  UNSUPPORTED_VERSION,
  BLOCK_MISSING_IN_CHAIN
}

data class GetZkEVMStateMerkleProofResponse(
  val zkStateMerkleProof: ArrayNode,
  val zkParentStateRootHash: Bytes32,
  val zkEndStateRootHash: Bytes32,
  val zkStateManagerVersion: String
)

interface Type2StateManagerClient {
  fun rollupGetZkEVMStateMerkleProof(
    startBlockNumber: UInt64,
    endBlockNumber: UInt64
  ): SafeFuture<Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<Type2StateManagerErrorType>>>
}
