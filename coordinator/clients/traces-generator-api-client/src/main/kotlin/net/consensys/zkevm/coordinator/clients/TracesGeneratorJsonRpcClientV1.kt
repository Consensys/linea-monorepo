package net.consensys.zkevm.coordinator.clients

import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.mapEither
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.async.RetriedExecutionException
import net.consensys.linea.async.retryWithInterval
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.errors.ErrorResponse
import net.consensys.linea.jsonrpc.BaseJsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.jsonrpc.client.JsonRpcClient
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.atomic.AtomicReference
import kotlin.time.Duration

class TracesGeneratorJsonRpcClientV1(
  private val vertx: Vertx,
  private val rpcClient: JsonRpcClient,
  private val config: Config

) :
  TracesCountersClientV1, TracesConflationClientV1 {

  data class Config(
    val requestMaxRetries: Int,
    val requestRetryInterval: Duration
  )

  private var id = AtomicInteger(0)

  override fun rollupGetTracesCounters(
    block: BlockNumberAndHash
  ): SafeFuture<Result<GetTracesCountersResponse, ErrorResponse<TracesServiceErrorType>>> {
    val jsonRequest =
      BaseJsonRpcRequest(
        "2.0",
        id.incrementAndGet(),
        "rollup_getBlockTracesCountersV1",
        listOf(
          mapOf(
            "blockNumber" to block.number.toString(),
            "blockHash" to block.hash.toHexString()
          )
        )
      )

    return retryRequest(jsonRequest, TracesClientResponsesParser::parseTracesCounterResponse)
  }

  private fun <S, E> responseIsOk(result: Result<S, E>): Boolean {
    return result is Ok
  }

  override fun rollupGenerateConflatedTracesToFile(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<Result<GenerateTracesResponse, ErrorResponse<TracesServiceErrorType>>> {
    // TODO: validate list of blocks
    // 1 - does not have repeated/duplicated pairs
    // 2 - blocks numbers are consecutive

    val jsonRequest =
      BaseJsonRpcRequest(
        "2.0",
        id.incrementAndGet(),
        "rollup_generateConflatedTracesToFileV1",
        blocks.map { block ->
          JsonObject.of(
            "blockNumber",
            block.number.toString(),
            "blockHash",
            block.hash.toHexString()
          )
        }
      )

    return retryRequest(jsonRequest, TracesClientResponsesParser::parseConflatedTracesToFileResponse)
  }

  private fun <T> retryRequest(
    jsonRequest: BaseJsonRpcRequest,
    responseParser: (jsonRpcResponse: JsonRpcSuccessResponse) -> T
  ): SafeFuture<Result<T, ErrorResponse<TracesServiceErrorType>>> {
    val lastResult = AtomicReference<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>>()
    return retryWithInterval(
      config.requestMaxRetries + 1,
      config.requestRetryInterval,
      vertx,
      { result: Result<JsonRpcSuccessResponse, JsonRpcErrorResponse> ->
        lastResult.set(result)
        responseIsOk(result)
      }
    ) { rpcClient.makeRequest(jsonRequest).toSafeFuture() }
      .exceptionallyCompose { th ->
        if (th is RetriedExecutionException) {
          SafeFuture.completedFuture(lastResult.get()!!)
        } else {
          SafeFuture.failedFuture(th)
        }
      }
      .thenApply { responseResult ->
        responseResult.mapEither(responseParser, TracesClientResponsesParser::mapErrorResponse)
      }
  }
}
