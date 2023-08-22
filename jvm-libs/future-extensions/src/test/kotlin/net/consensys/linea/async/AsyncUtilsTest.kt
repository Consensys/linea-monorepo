package net.consensys.linea.async

import io.vertx.core.Vertx
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ExecutionException
import java.util.concurrent.atomic.AtomicInteger
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

class AsyncUtilsTest {
  private val vertx = Vertx.vertx()

  @Test
  fun `retryWithInterval should resolve when predicate is satisfied`() {
    var callCount = 0
    val result =
      retryWithInterval(2, 20.milliseconds, vertx, { result -> result == "hello" }) {
        callCount++
        SafeFuture.completedFuture("hello")
      }
        .get()

    assertThat(callCount).isEqualTo(1)
    assertThat(result).isEqualTo("hello")
  }

  @Test
  fun `retryWithInterval should work with promises that return null values`() {
    var callCount = 0
    val result =
      retryWithInterval(2, 20.milliseconds, vertx) {
        callCount++
        SafeFuture.completedFuture<Unit>(null)
      }
        .get()

    assertThat(callCount).isEqualTo(1)
    assertThat(result).isNull()
  }

  @Test
  fun `retryWithInterval should retry and resolve when predicate is satisfied`() {
    var callCount = AtomicInteger(0)
    val result =
      retryWithInterval(4, 20.milliseconds, vertx, { result -> result == "2" }) {
        SafeFuture.completedFuture("${callCount.incrementAndGet()}")
      }
        .get()

    assertThat(callCount.get()).isEqualTo(2)
    assertThat(result).isEqualTo("2")
  }

  @Test
  fun `retryWithInterval should retry and reject when predicate is not satisfied until maxRetries is reached`() {
    var callCount = 0
    val future =
      retryWithInterval(2, 20.milliseconds, vertx, { result -> result == "false" }) {
        callCount++
        SafeFuture.completedFuture("true")
      }

    val exception = assertThrows<ExecutionException> { future.get() }

    assertThat(callCount).isEqualTo(3)
    assertThat(exception.message).endsWith("RetriedExecutionException: Stop condition wasn't met after 2 retries.")
  }

  @Test
  fun `retryWithInterval should retry when promise fails and resolve once resolves and predicate is true`() {
    var callCount = 0
    val result =
      retryWithInterval(5, 20.milliseconds, vertx, { it == "Result" }) {
        callCount++
        if (callCount < 3) {
          SafeFuture.failedFuture(Exception("Upss"))
        } else {
          SafeFuture.completedFuture("Result")
        }
      }
        .get()

    assertThat(callCount).isEqualTo(3)
    assertThat(result).isEqualTo("Result")
  }

  @Test
  fun `retryWithInterval should return failed future if retries are exhausted and predicate is never true`() {
    val expectedException = Exception("Retries exhausted")
    val result =
      retryWithInterval(retries = 3, interval = 20.milliseconds, vertx, { _ -> false }) {
        SafeFuture.failedFuture<Unit>(expectedException)
      }

    val exception = assertThrows<ExecutionException> { result.get() }
    assertThat(exception.cause).isEqualTo(expectedException)
  }

  @Test
  fun `retryWithInterval should retry after initialDelay`() {
    val startTime = System.currentTimeMillis()
    var firstDelay = 0L
    var callCount = 0
    val result =
      retryWithInterval(retries = 3, initialDelay = 1.seconds, interval = 20.milliseconds, vertx) {
        callCount++
        firstDelay = if (firstDelay == 0L) { System.currentTimeMillis() - startTime } else firstDelay
        SafeFuture.failedFuture<Unit>(Exception("Upss"))
      }
    val exception = assertThrows<ExecutionException> { result.get() }
    assertThat(firstDelay).isGreaterThanOrEqualTo(1000)
    assertThat(callCount).isEqualTo(4)
  }
}
