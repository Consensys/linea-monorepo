package net.consensys.zkevm.ethereum.submission

import org.apache.logging.log4j.Logger

fun logUnhandledError(
  log: Logger,
  errorOrigin: String,
  error: Throwable
) {
  if (error.message != null) {
    log.error(
      "Error from {}: errorMessage={}",
      errorOrigin,
      error.message
    )
  } else {
    log.error("Error from {}: ", errorOrigin, error)
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
  isEthCall: Boolean = false
) {
  val ethMethod = if (isEthCall) "eth_call" else "eth_sendRawTransaction"

  log.error(
    logMessage,
    ethMethod,
    intervalString,
    error.message,
    error
  )
}
