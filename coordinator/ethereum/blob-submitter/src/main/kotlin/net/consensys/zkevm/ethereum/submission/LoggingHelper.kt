package net.consensys.zkevm.ethereum.submission

import org.apache.logging.log4j.Logger

const val insufficientGasFeeRegexStr =
  "(max fee per (blob )?gas less than block (blob gas|base) fee:" +
    " address 0x[a-fA-F0-9]{40},? (blobGasFeeCap|maxFeePerGas): [0-9]+," +
    " (blobBaseFee|baseFee): [0-9]+ \\(supplied gas [0-9]+\\))"

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
  val errorMessage = if (isEthCall) {
    error.message?.let {
      Regex(insufficientGasFeeRegexStr).find(it)?.groupValues?.first()
    } ?: error.message
  } else {
    error.message
  }

  log.error(
    logMessage,
    ethMethod,
    intervalString,
    errorMessage,
    error
  )
}
