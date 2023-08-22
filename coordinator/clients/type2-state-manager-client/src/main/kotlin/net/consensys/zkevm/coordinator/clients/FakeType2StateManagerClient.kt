package net.consensys.zkevm.coordinator.clients

import com.fasterxml.jackson.databind.node.ArrayNode
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.nio.file.Path

class FakeType2StateManagerClient(fakeResponseFilePath: Path) : Type2StateManagerClient {
  private val response: GetZkEVMStateMerkleProofResponse
  private val objectMapper = jacksonObjectMapper()
  init {
    val json = objectMapper.readTree(fakeResponseFilePath.toFile())
    response =
      GetZkEVMStateMerkleProofResponse(
        zkStateManagerVersion = json.get("zkStateManagerVersion").asText(),
        zkStateMerkleProof = json.get("zkStateMerkleProof") as ArrayNode,
        zkParentStateRootHash =
        Bytes32.fromHexString(json.get("zkParentStateRootHash").asText())
      )
  }
  override fun rollupGetZkEVMStateMerkleProof(
    startBlockNumber: UInt64,
    endBlockNumber: UInt64
  ): SafeFuture<
    Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<Type2StateManagerErrorType>>> {
    // clone the JSON zkProof just in case client mutate the response;
    val responseClone = response.copy(zkStateMerkleProof = response.zkStateMerkleProof.deepCopy())
    return SafeFuture.completedFuture(Ok(responseClone))
  }
}
