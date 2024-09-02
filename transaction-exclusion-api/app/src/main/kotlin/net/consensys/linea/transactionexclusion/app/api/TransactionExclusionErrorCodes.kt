package net.consensys.linea.transactionexclusion.app.api

import net.consensys.linea.jsonrpc.JsonRpcError
import net.consensys.linea.transactionexclusion.ErrorType
import net.consensys.linea.transactionexclusion.TransactionExclusionError

enum class TransactionExclusionErrorCodes(val code: Int, val message: String) {
  // User' error codes
  TRANSACTION_UNAVAILABLE(-4000, "Rejected transaction not available"),
  TRANSACTION_DUPLICATED(-4001, "Rejected transaction was already persisted"),

  // App/System' error codes
  OTHER_ERROR(-5001, "Unidentified error was occurred");

  fun toErrorObject(data: Any? = null): JsonRpcError {
    return JsonRpcError(this.code, this.message, data)
  }
}

fun jsonRpcError(appError: TransactionExclusionError): JsonRpcError {
  return when (appError.errorType) {
    ErrorType.TRANSACTION_UNAVAILABLE ->
      TransactionExclusionErrorCodes.TRANSACTION_UNAVAILABLE.toErrorObject(appError.errorDetail)
    ErrorType.TRANSACTION_DUPLICATED ->
      TransactionExclusionErrorCodes.TRANSACTION_DUPLICATED.toErrorObject(appError.errorDetail)
    ErrorType.OTHER_ERROR ->
      TransactionExclusionErrorCodes.OTHER_ERROR.toErrorObject(appError.errorDetail)
  }
}
