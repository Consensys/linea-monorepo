package net.consensys.zkevm.coordinator.clients

import com.fasterxml.jackson.databind.node.ArrayNode
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.mapEither
import io.vertx.core.json.JsonObject
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.errors.ErrorResponse
import net.consensys.linea.jsonrpc.BaseJsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.jsonrpc.client.JsonRpcClient
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.util.concurrent.atomic.AtomicInteger

class Type2StateManagerJsonRpcClient(
  private val rpcClient: JsonRpcClient,
  private val config: Config
) : Type2StateManagerClient {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private val objectMapper = jacksonObjectMapper()
  private var id = AtomicInteger(0)

  data class Config(val zkStateManagerVersion: String)

  override fun rollupGetZkEVMStateMerkleProof(
    startBlockNumber: UInt64,
    endBlockNumber: UInt64
  ): SafeFuture<
    Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<Type2StateManagerErrorType>>> {
    if (startBlockNumber > endBlockNumber) {
      throw IllegalArgumentException(
        "endBlockNumber must be greater than startBlockNumber: " +
          "startBlockNumber = $startBlockNumber endBlockNumber = $endBlockNumber"
      )
    }

    val jsonRequest =
      BaseJsonRpcRequest(
        "2.0",
        id.incrementAndGet(),
        "rollup_getZkEVMStateMerkleProofV0",
        listOf(
          JsonObject.of(
            "startBlockNumber",
            startBlockNumber.toULong().toLong(),
            "endBlockNumber",
            endBlockNumber.toULong().toLong(),
            "zkStateManagerVersion",
            config.zkStateManagerVersion
          )
        )
      )

    return rpcClient
      .makeRequest(jsonRequest)
      .map { zkEvmStateMerkleProofResult ->
        zkEvmStateMerkleProofResult.mapEither(
          this::parseZkEVMStateMerkleProofResponse,
          this::mapErrorResponse
        )
      }
      .toSafeFuture()
  }

  private fun mapErrorResponse(
    jsonRpcErrorResponse: JsonRpcErrorResponse
  ): ErrorResponse<Type2StateManagerErrorType> {
    val errorType =
      try {
        Type2StateManagerErrorType.valueOf(
          jsonRpcErrorResponse.error.message.substringBefore('-').trim()
        )
      } catch (_: Exception) {
        log.error(
          "State manager found unrecognised JSON-RPC response error: {}",
          jsonRpcErrorResponse.error
        )
        Type2StateManagerErrorType.UNKNOWN
      }

    return ErrorResponse(
      errorType,
      listOfNotNull(
        jsonRpcErrorResponse.error.message,
        jsonRpcErrorResponse.error.data?.toString()
      )
        .joinToString(": ")
    )
  }

  private fun parseZkEVMStateMerkleProofResponse(
    jsonRpcResponse: JsonRpcSuccessResponse
  ): GetZkEVMStateMerkleProofResponse {
    val json = objectMapper.readTree((jsonRpcResponse.result as JsonObject).toString())
    return GetZkEVMStateMerkleProofResponse(
      zkStateManagerVersion = json.get("zkStateManagerVersion").asText(),
      zkStateMerkleProof = json.get("zkStateMerkleProof") as ArrayNode,
      zkParentStateRootHash = Bytes32.fromHexString(json.get("zkParentStateRootHash").asText())
    )
  }
}
