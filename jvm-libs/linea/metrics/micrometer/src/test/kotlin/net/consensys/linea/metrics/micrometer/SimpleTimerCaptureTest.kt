package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.MockClock
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.TimeUnit

class SimpleTimerCaptureTest {
  @Test
  fun itExtractsValue_callback() {
    val meterRegistry: MeterRegistry = SimpleMeterRegistry()
    val result =
      SimpleTimerCapture<String>(meterRegistry, "request.counter")
        .setDescription("API request counter")
        .setTag("method", "eth_blockNumber")
        .captureTime {
          Thread.sleep(10)
          "measured_callback_result"
        }

    Assertions.assertThat(result).isEqualTo("measured_callback_result")

    val createdMeter = meterRegistry["request.counter"].timer()
    Assertions.assertThat(createdMeter.count()).isEqualTo(1)
    Assertions.assertThat(createdMeter.id.description).isEqualTo("API request counter")
    Assertions.assertThat(createdMeter.id.tags.size).isEqualTo(1)
    Assertions.assertThat(createdMeter.id.tags[0].key).isEqualTo("method")
    Assertions.assertThat(createdMeter.id.tags[0].value).isEqualTo("eth_blockNumber")
    Assertions.assertThat(createdMeter.totalTime(TimeUnit.MILLISECONDS)).isBetween(10.0, 100.0)
  }

  @Test
  fun itExtractsValue_FutureSuccess() {
    val meterRegistry: MeterRegistry = SimpleMeterRegistry()
    val testClock = MockClock()
    val result =
      SimpleTimerCapture<String>(meterRegistry, "request.counter")
        .setDescription("API request counter")
        .setTag("method", "eth_blockNumber")
        .setClock(testClock)
        .captureTime(SafeFuture.completedFuture("measured_callback_result"))

    Assertions.assertThat(result.get()).isEqualTo("measured_callback_result")

    val createdMeter = meterRegistry["request.counter"].timer()
    Assertions.assertThat(createdMeter.count()).isEqualTo(1)
    Assertions.assertThat(createdMeter.id.description).isEqualTo("API request counter")
    Assertions.assertThat(createdMeter.id.tags.size).isEqualTo(1)
    Assertions.assertThat(createdMeter.id.tags[0].key).isEqualTo("method")
    Assertions.assertThat(createdMeter.id.tags[0].value).isEqualTo("eth_blockNumber")
  }

  @Test
  fun itExtractsValue_FutureFailure() {
    val meterRegistry: MeterRegistry = SimpleMeterRegistry()
    val testClock = MockClock()
    val result =
      SimpleTimerCapture<String>(meterRegistry, "request.counter")
        .setDescription("API request counter")
        .setTag("method", "eth_blockNumber_failure")
        .setClock(testClock)
        .captureTime(SafeFuture.failedFuture(Exception("measured_callback_error")))

    result.finish { error -> Assertions.assertThat(error.message).isEqualTo("measured_callback_error") }

    val createdMeter = meterRegistry["request.counter"].timer()
    Assertions.assertThat(createdMeter.count()).isEqualTo(1)
    Assertions.assertThat(createdMeter.id.description).isEqualTo("API request counter")
    Assertions.assertThat(createdMeter.id.tags.size).isEqualTo(1)
    Assertions.assertThat(createdMeter.id.tags[0].key).isEqualTo("method")
    Assertions.assertThat(createdMeter.id.tags[0].value).isEqualTo("eth_blockNumber_failure")
  }
}
