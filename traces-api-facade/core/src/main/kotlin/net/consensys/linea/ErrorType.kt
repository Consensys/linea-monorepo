package net.consensys.linea

/** For simplicity, placing all error codes into single enum */
enum class ErrorType {
  INVALID_BLOCK_NUMBERS_RANGE,
  TRACES_UNAVAILABLE,
  TRACES_AMBIGUITY,
  WRONG_JSON_FORMAT,
  WRONG_JSON_CONTENT
}

data class TracesError(val errorType: ErrorType, val errorDetail: String)
