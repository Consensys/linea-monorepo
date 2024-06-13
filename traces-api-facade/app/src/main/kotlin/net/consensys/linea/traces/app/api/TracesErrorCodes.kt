package net.consensys.linea.traces.app.api

import net.consensys.linea.ErrorType
import net.consensys.linea.TracesError
import net.consensys.linea.jsonrpc.JsonRpcError

enum class TracesErrorCodes(val code: Int, val message: String) {
  // User' error codes
  INVALID_BLOCK_NUMBERS(-4000, "Invalid block numbers"),
  TRACES_UNAVAILABLE(-4001, "Traces not available"),

  // App/System' error codes
  TRACES_AMBIGUITY(-5001, "Trances Ambiguity: multiple traces found for the same block"),
  TRACES_INVALID_JSON_FORMAT(-5002, "Traces file has invalid json format."),
  TRACES_INVALID_CONTENT(-5003, "Traces file has invalid content.");
  fun toErrorObject(data: Any? = null): JsonRpcError {
    return JsonRpcError(this.code, this.message, data)
  }
}

fun jsonRpcError(appError: TracesError): JsonRpcError {
  return when (appError.errorType) {
    ErrorType.INVALID_BLOCK_NUMBERS_RANGE ->
      TracesErrorCodes.INVALID_BLOCK_NUMBERS.toErrorObject(appError.errorDetail)
    ErrorType.TRACES_UNAVAILABLE ->
      TracesErrorCodes.TRACES_UNAVAILABLE.toErrorObject(appError.errorDetail)
    ErrorType.TRACES_AMBIGUITY ->
      TracesErrorCodes.TRACES_AMBIGUITY.toErrorObject(appError.errorDetail)
    ErrorType.WRONG_JSON_FORMAT ->
      TracesErrorCodes.TRACES_INVALID_JSON_FORMAT.toErrorObject(appError.errorDetail)
    ErrorType.WRONG_JSON_CONTENT ->
      TracesErrorCodes.TRACES_INVALID_CONTENT.toErrorObject(appError.errorDetail)
  }
}
