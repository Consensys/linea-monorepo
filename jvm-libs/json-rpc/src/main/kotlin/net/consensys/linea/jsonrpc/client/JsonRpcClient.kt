package net.consensys.linea.jsonrpc.client

import com.github.michaelbull.result.Result
import io.vertx.core.Future
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse

fun identityMapper(value: Any?): Any? = value

interface JsonRpcClient {
  fun makeRequest(
    request: JsonRpcRequest,
    resultMapper: (Any?) -> Any? = ::identityMapper
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>>
}
