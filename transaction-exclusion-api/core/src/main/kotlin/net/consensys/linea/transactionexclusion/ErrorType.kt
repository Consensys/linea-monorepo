package net.consensys.linea.transactionexclusion

/** For simplicity, placing all error codes into single enum */
enum class ErrorType {
  TRANSACTION_UNAVAILABLE,
  OTHER_ERROR
}

data class TransactionExclusionError(val errorType: ErrorType, val errorDetail: String)