package net.consensys.linea.jsonrpc

import com.github.michaelbull.result.Result
import com.github.michaelbull.result.getOrElse
import io.vertx.core.Future

fun Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>.unfoldSuccessOrThrowError(): JsonRpcSuccessResponse {
  return this.getOrElse { jsonRpcError -> throw jsonRpcError.error.asException() }
}

// ugly indentation to please the linter :(
fun Future<
  Result<
    JsonRpcSuccessResponse,
    JsonRpcErrorResponse
    >
  >.unfoldSuccessOrThrowError(): Future<JsonRpcSuccessResponse> {
  return this.map { result -> result.unfoldSuccessOrThrowError() }
}
