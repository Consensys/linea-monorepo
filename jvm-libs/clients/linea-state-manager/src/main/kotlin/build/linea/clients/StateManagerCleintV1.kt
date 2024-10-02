package net.consensys.zkevm.coordinator.clients

import com.fasterxml.jackson.databind.node.ArrayNode
import com.github.michaelbull.result.Result
import net.consensys.linea.BlockInterval
import net.consensys.linea.clients.AsyncClient
import net.consensys.linea.clients.ClientError
import net.consensys.linea.clients.unwrapResultMonad
import net.consensys.linea.errors.ErrorResponse
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture

enum class StateManagerErrorType : ClientError {
  UNKNOWN,
  UNSUPPORTED_VERSION,
  BLOCK_MISSING_IN_CHAIN
}

typealias GetZkEVMStateMerkleProofRequest = BlockInterval

data class GetZkEVMStateMerkleProofResponse(
  val zkStateMerkleProof: ArrayNode,
  val zkParentStateRootHash: Bytes32,
  val zkEndStateRootHash: Bytes32,
  val zkStateManagerVersion: String
)

interface StateManagerClientV1 :
  AsyncClient<
    GetZkEVMStateMerkleProofRequest,
    GetZkEVMStateMerkleProofResponse
    > {
  fun rollupGetZkEVMStateMerkleProof(
    blockInterval: BlockInterval
  ): SafeFuture<Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<StateManagerErrorType>>>

  override fun makeRequest(request: GetZkEVMStateMerkleProofRequest):
    SafeFuture<GetZkEVMStateMerkleProofResponse> {
    return rollupGetZkEVMStateMerkleProof(request).unwrapResultMonad()
  }
}
