package build.linea.clients

import com.fasterxml.jackson.databind.JsonNode
import com.fasterxml.jackson.databind.node.ArrayNode
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.mapEither
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject
import net.consensys.fromHexString
import net.consensys.linea.BlockInterval
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.errors.ErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequestListParams
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.jsonrpc.client.JsonRpcClient
import net.consensys.linea.jsonrpc.client.JsonRpcRequestRetryer
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import net.consensys.linea.jsonrpc.client.toPrimitiveOrJacksonJsonNode
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import net.consensys.zkevm.coordinator.clients.StateManagerClientV1
import net.consensys.zkevm.coordinator.clients.StateManagerErrorType
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicInteger

class StateManagerV1JsonRpcClient(
  private val rpcClient: JsonRpcClient,
  private val config: Config
) : StateManagerClientV1 {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private val objectMapper = jacksonObjectMapper()
  private var id = AtomicInteger(0)

  data class Config(
    val zkStateManagerVersion: String,
    val requestRetry: RequestRetryConfig
  )

  constructor(
    vertx: Vertx,
    rpcClient: JsonRpcClient,
    config: Config,
    log: Logger = LogManager.getLogger(StateManagerV1JsonRpcClient::class.java)
  ) : this(
    rpcClient = JsonRpcRequestRetryer(
      vertx,
      rpcClient,
      config = JsonRpcRequestRetryer.Config(
        methodsToRetry = retryableMethods,
        requestRetry = config.requestRetry
      ),
      log = log
    ),
    config = config
  )

  override fun rollupGetHeadBlockNumber(): SafeFuture<ULong> {
    val jsonRequest =
      JsonRpcRequestListParams(
        jsonrpc = "2.0",
        id = id.incrementAndGet(),
        method = "rollup_getZkEVMBlockNumber",
        params = listOf()
      )

    return rpcClient
      .makeRequest(jsonRequest).toSafeFuture()
      .thenApply { responseResult ->
        when (responseResult) {
          is Ok -> {
            ULong.fromHexString(responseResult.value.result as String)
          }

          is Err -> {
            throw responseResult.error.error.asException()
          }
        }
      }
  }

  override fun rollupGetStateMerkleProofWithTypedError(
    blockInterval: BlockInterval
  ): SafeFuture<Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<StateManagerErrorType>>> {
    val jsonRequest =
      JsonRpcRequestListParams(
        jsonrpc = "2.0",
        id = id.incrementAndGet(),
        method = "rollup_getZkEVMStateMerkleProofV0",
        params = listOf(
          JsonObject.of(
            "startBlockNumber",
            blockInterval.startBlockNumber.toLong(),
            "endBlockNumber",
            blockInterval.endBlockNumber.toLong(),
            "zkStateManagerVersion",
            config.zkStateManagerVersion
          )
        )
      )

    return rpcClient
      .makeRequest(
        request = jsonRequest,
        resultMapper = ::toPrimitiveOrJacksonJsonNode
      ).toSafeFuture()
      .thenApply { responseResult ->
        responseResult.mapEither(this::parseZkEVMStateMerkleProofResponse, this::mapErrorResponse)
      }
  }

  private fun mapErrorResponse(
    jsonRpcErrorResponse: JsonRpcErrorResponse
  ): ErrorResponse<StateManagerErrorType> {
    val errorType =
      try {
        StateManagerErrorType.valueOf(
          jsonRpcErrorResponse.error.message.substringBefore('-').trim()
        )
      } catch (_: Exception) {
        log.error(
          "State manager found unrecognised JSON-RPC response error: {}",
          jsonRpcErrorResponse.error
        )
        StateManagerErrorType.UNKNOWN
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
    val json = jsonRpcResponse.result as JsonNode

    return GetZkEVMStateMerkleProofResponse(
      zkStateManagerVersion = json.get("zkStateManagerVersion").asText(),
      zkStateMerkleProof = json.get("zkStateMerkleProof") as ArrayNode,
      zkParentStateRootHash = Bytes32.fromHexString(json.get("zkParentStateRootHash").asText()),
      zkEndStateRootHash = Bytes32.fromHexString(json.get("zkEndStateRootHash").asText())
    )
  }

  companion object {
    internal val retryableMethods = setOf("rollup_getZkEVMStateMerkleProofV0")
  }
}
