package net.consensys.linea.consensus

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.github.michaelbull.result.mapBoth
import io.vertx.core.json.JsonObject
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.forkchoicestate.ForkChoiceStateInfoV0
import net.consensys.linea.jsonrpc.BaseJsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.jsonrpc.client.JsonRpcClient
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.atomic.AtomicLong

interface ForkChoiceStateClient {
  fun getForkChoiceState(): SafeFuture<ForkChoiceStateInfoV0>
}

class ForkChoiceStateJsonRpcClient(
  private val jsonRpcClient: JsonRpcClient,
  private val objectMapper: ObjectMapper = jacksonObjectMapper()
) : ForkChoiceStateClient {
  private val idCounter = AtomicLong(0)
  private val requestTemplate =
    BaseJsonRpcRequest("2.0", 0, "linea_getForkChoiceState", emptyList<Any>())

  override fun getForkChoiceState(): SafeFuture<ForkChoiceStateInfoV0> {
    return jsonRpcClient
      .makeRequest(requestTemplate.copy(id = idCounter.getAndIncrement()))
      .toSafeFuture()
      .thenCompose() { result ->
        result.mapBoth(
          { SafeFuture.completedFuture(mapResult(it)) },
          { SafeFuture.failedFuture(it.error.asException()) }
        )
      }
  }

  private fun mapResult(response: JsonRpcSuccessResponse): ForkChoiceStateInfoV0 {
    val result = response.result as JsonObject
    return objectMapper.readValue(result.encode(), ForkChoiceStateInfoV0::class.java)
  }
}
