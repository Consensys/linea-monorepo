package net.consensys.zkevm.ethereum.submission

import org.apache.logging.log4j.Level
import org.apache.logging.log4j.Logger

fun getErrorAppropriateLevel(error: Throwable, isEthCall: Boolean): Level {
  return when {
    error.message.isNullOrEmpty() -> Level.ERROR
    (
      error.message!!.contains("replacement transaction underpriced", ignoreCase = true) ||
        error.message!!.contains("already known", ignoreCase = true) ||
        error.message!!.contains("Transaction receipt was not generated after", ignoreCase = true) ||
        error.message!!.contains("Shutting down", ignoreCase = true) ||
        error.message!!.contains("Known transaction", ignoreCase = true) ||
        error.message!!.contains("header not found", ignoreCase = true) ||
        error.message!!.contains("Block not found", ignoreCase = true)
      ) -> {
      Level.DEBUG
    }
    error.message!!.contains("Nonce too low", ignoreCase = true) -> Level.INFO

    (
      error.message!!.contains("Contract Call has been reverted by the EVM with the reason", ignoreCase = true)
      ) -> {
      if (isEthCall) Level.INFO else Level.ERROR
    }

    else -> if (isEthCall) Level.WARN else Level.ERROR
  }
}

/**
 * Logs submission errors
 * 3 parameters to the message:
 * "{ETH_CALL_TYPE} for blob submission failed: blob={INTERVAL_STING} errorMessage={ERROR_MESSAGE}"
 */
fun logSubmissionError(
  log: Logger,
  logMessage: String,
  intervalString: String,
  error: Throwable,
  isEthCall: Boolean = false,
  logLevel: Level = getErrorAppropriateLevel(error, isEthCall)
) {
  val ethMethod = if (isEthCall) "eth_call" else "eth_sendRawTransaction"
  val errorToLog = if (logLevel == Level.WARN || logLevel == Level.ERROR) {
    error
  } else {
    null
  }

  log.log(
    logLevel,
    logMessage,
    ethMethod,
    intervalString,
    error.message,
    errorToLog
  )
}
