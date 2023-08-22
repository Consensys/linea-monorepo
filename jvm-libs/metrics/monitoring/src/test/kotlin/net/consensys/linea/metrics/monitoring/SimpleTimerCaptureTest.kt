package net.consensys.linea.metrics.monitoring

import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.MockClock
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import org.assertj.core.api.Assertions.assertThat
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

    assertThat(result).isEqualTo("measured_callback_result")

    val createdMeter = meterRegistry["request.counter"].timer()
    assertThat(createdMeter.count()).isEqualTo(1)
    assertThat(createdMeter.id.description).isEqualTo("API request counter")
    assertThat(createdMeter.id.tags.size).isEqualTo(1)
    assertThat(createdMeter.id.tags[0].key).isEqualTo("method")
    assertThat(createdMeter.id.tags[0].value).isEqualTo("eth_blockNumber")
    assertThat(createdMeter.totalTime(TimeUnit.MILLISECONDS)).isBetween(10.0, 100.0)
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

    assertThat(result.get()).isEqualTo("measured_callback_result")

    val createdMeter = meterRegistry["request.counter"].timer()
    assertThat(createdMeter.count()).isEqualTo(1)
    assertThat(createdMeter.id.description).isEqualTo("API request counter")
    assertThat(createdMeter.id.tags.size).isEqualTo(1)
    assertThat(createdMeter.id.tags[0].key).isEqualTo("method")
    assertThat(createdMeter.id.tags[0].value).isEqualTo("eth_blockNumber")
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

    result.finish { error -> assertThat(error.message).isEqualTo("measured_callback_error") }

    val createdMeter = meterRegistry["request.counter"].timer()
    assertThat(createdMeter.count()).isEqualTo(1)
    assertThat(createdMeter.id.description).isEqualTo("API request counter")
    assertThat(createdMeter.id.tags.size).isEqualTo(1)
    assertThat(createdMeter.id.tags[0].key).isEqualTo("method")
    assertThat(createdMeter.id.tags[0].value).isEqualTo("eth_blockNumber_failure")
  }
}
