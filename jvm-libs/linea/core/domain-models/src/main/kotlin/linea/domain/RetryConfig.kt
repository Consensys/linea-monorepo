package linea.domain

import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds

data class RetryConfig(
  val maxRetries: UInt? = null,
  val timeout: Duration? = null,
  val exceptionConsumerDelay: Duration? = null,
  val backoffDelay: Duration = 100.milliseconds,
  val failuresWarningThreshold: UInt = 0u,
) {
  val isRetryDisabled = maxRetries == 0u || timeout == 0.milliseconds
  val isRetryEnabled: Boolean = !isRetryDisabled

  init {
    maxRetries?.also {
      require(maxRetries >= failuresWarningThreshold) {
        "maxRetries must be greater or equal than failuresWarningThreshold." +
          " maxRetries=$maxRetries, failuresWarningThreshold=$failuresWarningThreshold"
      }
    }

    timeout?.also {
      require(timeout >= 1.milliseconds) { "timeout must be >= 1ms. value=$timeout" }
    }

    require(backoffDelay >= 1.milliseconds) { "backoffDelay must be >= 1ms. value=$backoffDelay" }

    exceptionConsumerDelay?.also {
      require(exceptionConsumerDelay >= 1.milliseconds) {
        "exceptionConsumerDelay must be >= 1ms. value=$exceptionConsumerDelay"
      }
    }
  }

  companion object {
    val noRetries = RetryConfig(maxRetries = 0u)
    fun endlessRetry(backoffDelay: Duration, failuresWarningThreshold: UInt) = RetryConfig(
      maxRetries = null,
      timeout = null,
      backoffDelay = backoffDelay,
      failuresWarningThreshold = failuresWarningThreshold,
    )
  }
}
