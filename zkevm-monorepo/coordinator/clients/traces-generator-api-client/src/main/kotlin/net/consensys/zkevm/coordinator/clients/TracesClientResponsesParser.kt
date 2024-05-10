package net.consensys.zkevm.coordinator.clients

import com.github.michaelbull.result.getOrElse
import com.github.michaelbull.result.getOrThrow
import com.github.michaelbull.result.onFailure
import com.github.michaelbull.result.runCatching
import io.vertx.core.json.JsonObject
import net.consensys.linea.errors.ErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcSuccessResponse
import net.consensys.linea.traces.TracesCounters
import net.consensys.linea.traces.TracingModule
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

object TracesClientResponsesParser {
  private val log: Logger = LogManager.getLogger(this::class.java)

  internal fun mapErrorResponse(
    jsonRpcErrorResponse: JsonRpcErrorResponse
  ): ErrorResponse<TracesServiceErrorType> {
    val errorType: TracesServiceErrorType = runCatching {
      TracesServiceErrorType.valueOf(jsonRpcErrorResponse.error.message.substringBefore(':'))
    }.getOrElse { TracesServiceErrorType.UNKNOWN_ERROR }

    return ErrorResponse(errorType, jsonRpcErrorResponse.error.message)
  }

  internal fun parseTracesCounterResponse(
    jsonRpcResponse: JsonRpcSuccessResponse
  ): GetTracesCountersResponse {
    val result = jsonRpcResponse.result as JsonObject

    return GetTracesCountersResponse(
      result.getString("blockL1Size").toUInt(),
      result.getJsonObject("tracesCounters").let(this::parseTracesCounters),
      result.getString("tracesEngineVersion")
    )
  }

  internal fun parseTracesCounters(tracesCounters: JsonObject): TracesCounters {
    val expectedModules = TracingModule.values().map { it.name }.toSet()
    val evmModulesInResponse = tracesCounters.map.keys.toSet()
    val modulesMissing = expectedModules - evmModulesInResponse
    val unExpectedModules = evmModulesInResponse - expectedModules
    val error =
      if (modulesMissing.isNotEmpty()) {
        "Traces counters response is missing modules: ${modulesMissing.joinToString(",")}"
      } else if (unExpectedModules.isNotEmpty()) {
        "Traces counters has unsupported modules: ${unExpectedModules.joinToString(",")}"
      } else {
        null
      }
    if (error != null) {
      log.error(error)
      throw IllegalStateException(error)
    }

    val traces = mutableMapOf<TracingModule, UInt>()
    for (traceModule in TracingModule.values()) {
      val counterValue = tracesCounters.getString(traceModule.name)
      traces[traceModule] =
        runCatching { counterValue.toUInt() }
          .onFailure {
            log.error(
              "Failed to parse Evm module ${traceModule.name}='$counterValue' to UInt. errorMessage={}",
              it.message,
              it
            )
          }
          .getOrThrow()
    }

    return traces
  }

  internal fun parseConflatedTracesToFileResponse(
    jsonRpcResponse: JsonRpcSuccessResponse
  ): GenerateTracesResponse {
    val result = jsonRpcResponse.result as JsonObject
    return GenerateTracesResponse(
      result.getString("conflatedTracesFileName"),
      result.getString("tracesEngineVersion")
    )
  }
}
