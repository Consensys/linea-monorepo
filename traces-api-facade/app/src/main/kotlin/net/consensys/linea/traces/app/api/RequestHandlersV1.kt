package net.consensys.linea.traces.app.api

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.map
import com.github.michaelbull.result.mapError
import io.vertx.core.Future
import io.vertx.core.json.JsonObject
import io.vertx.ext.auth.User
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.TracesConflationServiceV1
import net.consensys.linea.TracesCountingServiceV1
import net.consensys.linea.TracesError
import net.consensys.linea.VersionedResult
import net.consensys.linea.async.toVertxFuture
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcRequestHandler
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture

private fun parseBlockNumberAndHash(json: JsonObject) = BlockNumberAndHash(
  json.getString("blockNumber").toULong(),
  Bytes32.fromHexString(json.getString("blockHash"))
)

class TracesCounterRequestHandlerV1(private val tracesCountingService: TracesCountingServiceV1) :
  JsonRpcRequestHandler {
  override fun invoke(
    user: User?,
    request: JsonRpcRequest,
    requestJson: JsonObject
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    val block: BlockNumberAndHash
    try {
      if (request.params.isEmpty()) {
        return Future.succeededFuture(Err(JsonRpcErrorResponse.invalidPrams(request.id, "Invalid block parameter!")))
      }
      block = parseBlockNumberAndHash(JsonObject.mapFrom(request.params.first()))
    } catch (e: Exception) {
      return Future.succeededFuture(Err(JsonRpcErrorResponse.invalidPrams(request.id, e.message)))
    }

    return tracesCountingService
      .getBlockTracesCounters(block)
      .thenApply { result ->
        result
          .map {
            val rpcResult =
              JsonObject()
                .put("tracesEngineVersion", it.version)
                .put("blockNumber", block.number.toString())
                .put("blockL1Size", it.result.blockL1Size.toString())
                .put(
                  "tracesCounters",
                  it.result.tracesCounters.mapValues { count -> count.value.toLong() }
                )
            JsonRpcSuccessResponse(request.id, rpcResult)
          }
          .mapError { error -> JsonRpcErrorResponse(request.id, jsonRpcError(error)) }
      }
      .toVertxFuture()
  }
}

abstract class AbstractTracesConflationRequestHandlerV1<T> : JsonRpcRequestHandler {

  abstract fun tracesContent(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<Result<T, TracesError>>

  override fun invoke(
    user: User?,
    request: JsonRpcRequest,
    requestJson: JsonObject
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    val blocks: List<BlockNumberAndHash>
    try {
      if (request.params.isEmpty()) {
        return Future.succeededFuture(Err(JsonRpcErrorResponse.invalidPrams(request.id, "Empty list of blocks!")))
      }
      blocks = request.params.map { blockJson -> parseBlockNumberAndHash(JsonObject.mapFrom(blockJson)) }
    } catch (e: Exception) {
      return Future.succeededFuture(Err(JsonRpcErrorResponse.invalidPrams(request.id, e.message)))
    }

    return tracesContent(blocks)
      .thenApply { result ->
        result
          .map { JsonRpcSuccessResponse(request.id, it) }
          .mapError { error -> JsonRpcErrorResponse(request.id, jsonRpcError(error)) }
      }
      .toVertxFuture()
  }
}

class GenerateConflatedTracesToFileRequestHandlerV1(private val service: TracesConflationServiceV1) :
  AbstractTracesConflationRequestHandlerV1<JsonObject>() {
  override fun tracesContent(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<Result<JsonObject, TracesError>> {
    val blocksSorted = blocks.sortedBy { it.number }
    return service.generateConflatedTracesToFile(blocksSorted)
      .thenApply { result: Result<VersionedResult<String>, TracesError> ->
        result.map {
          JsonObject()
            .put("tracesEngineVersion", it.version)
            .put("startBlockNumber", blocksSorted.first().number.toString())
            .put("endBlockNumber", blocksSorted.last().number.toString())
            .put("conflatedTracesFileName", it.result)
        }
      }
  }
}

class GetConflatedTracesRequestHandlerV1(val service: TracesConflationServiceV1) :
  AbstractTracesConflationRequestHandlerV1<JsonObject>() {

  override fun tracesContent(
    blocks: List<BlockNumberAndHash>
  ): SafeFuture<Result<JsonObject, TracesError>> {
    return service.getConflatedTraces(blocks).thenApply { result: Result<VersionedResult<JsonObject>, TracesError> ->
      result.map {
        JsonObject().put("tracesEngineVersion", it.version).put("conflatedTraces", it.result)
      }
    }
  }
}
