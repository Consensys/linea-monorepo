package build.linea.domain

import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds

data class RetryConfig(
  val maxRetries: UInt? = null,
  val timeout: Duration? = null,
  val backoffDelay: Duration = 100.milliseconds,
  val failuresWarningThreshold: UInt = 0u
) {
  val isRetryEnabled: Boolean
    get() = maxRetries != null && maxRetries > 0u || timeout != null && timeout > 0.milliseconds

  init {
    maxRetries?.also {
      require(maxRetries >= failuresWarningThreshold) {
        "maxRetries must be greater or equal than failuresWarningThreshold." +
          " maxRetries=$maxRetries, failuresWarningThreshold=$failuresWarningThreshold"
      }
    }
    timeout?.also {
      require(timeout > 0.milliseconds) { "timeout must be >= 1ms. value=$timeout" }
    }

    require(backoffDelay > 0.milliseconds) { "backoffDelay must be >= 1ms. value=$timeout" }
  }

  companion object {
    val noRetries = RetryConfig(maxRetries = 0u)
  }
}
