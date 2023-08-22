package net.consensys.linea.forkchoicestate.api

import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.vertx.core.Future
import io.vertx.core.json.JsonObject
import io.vertx.ext.auth.User
import net.consensys.linea.forkchoicestate.ForkChoiceStateProvider
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcRequestHandler
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse

class GetForkChoiceStateHandler(private val controller: ForkChoiceStateProvider) :
  JsonRpcRequestHandler {
  override fun invoke(
    user: User?,
    request: JsonRpcRequest,
    requestJson: JsonObject
  ): Future<Result<JsonRpcSuccessResponse, JsonRpcErrorResponse>> {
    return Future.succeededFuture(
      Ok(JsonRpcSuccessResponse(request.id, controller.getLatestForkChoiceState()))
    )
  }
}
