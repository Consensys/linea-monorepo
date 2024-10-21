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

sealed interface StateManagerRequest
sealed class GetChainHeadRequest() : StateManagerRequest
data class GetStateMerkleProofRequest(
  val blockInterval: BlockInterval
) : StateManagerRequest, BlockInterval by blockInterval

sealed interface StateManagerResponse
data class GetZkEVMStateMerkleProofResponse(
  val zkStateMerkleProof: ArrayNode,
  val zkParentStateRootHash: Bytes32,
  val zkEndStateRootHash: Bytes32,
  val zkStateManagerVersion: String
) : StateManagerResponse

// Type alias dedicated for each method
typealias StateManagerClientToGetStateMerkleProofV0 =
  AsyncClient<GetStateMerkleProofRequest, GetZkEVMStateMerkleProofResponse>

typealias StateManagerClientToGetChainHeadV1 =
  AsyncClient<GetChainHeadRequest, ULong>

interface StateManagerClientV1 {
  /**
   * Get the head block number of the chain.
   * @return GetZkEVMStateMerkleProofResponse
   * @throws ClientException with errorType StateManagerErrorType when know error occurs
   */
  fun rollupGetStateMerkleProof(
    blockInterval: BlockInterval
  ): SafeFuture<GetZkEVMStateMerkleProofResponse> = rollupGetStateMerkleProofWithTypedError(blockInterval)
    .unwrapResultMonad()

  fun rollupGetStateMerkleProofWithTypedError(
    blockInterval: BlockInterval
  ): SafeFuture<Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<StateManagerErrorType>>>

  fun rollupGetHeadBlockNumber(): SafeFuture<ULong>
}
