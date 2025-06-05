package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.MockClock
import io.micrometer.core.instrument.Timer
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.time.Duration
import java.util.concurrent.TimeUnit

internal class DynamicTagTimerCaptureTest {
  @Test
  fun itExtractsValueAndDoesntAlterResult_callback() {
    val meterRegistry: MeterRegistry = SimpleMeterRegistry()
    val testClock = MockClock()
    val timerBuilder = Timer.builder("request.counter")
    timerBuilder.description("API request counter")
    val result =
      DynamicTagTimerCaptureBuilder<String>(
        meterRegistry = meterRegistry,
        wrappedTimerBuilder = timerBuilder,
        clock = testClock,
      ).setDynamicTagValueExtractor { "eth_blockNumber" }
        .setDynamicTagValueExtractorOnError { "eth_blockNumber_failure" }
        .setDynamicTagKey("method")
        .build()
        .captureTime {
          testClock.add(Duration.ofSeconds(1))
          "measured_callback_result"
        }
    assertThat(result).isEqualTo("measured_callback_result")
    val createdMeter = meterRegistry["request.counter"].timer()

    assertThat(createdMeter.count()).isEqualTo(1)
    assertThat(createdMeter.totalTime(TimeUnit.SECONDS)).isEqualTo(1.0)
    assertThat(createdMeter.id.description).isEqualTo("API request counter")
    assertThat(createdMeter.id.tags.size).isEqualTo(1)
    assertThat(createdMeter.id.tags[0].key).isEqualTo("method")
    assertThat(createdMeter.id.tags[0].value).isEqualTo("eth_blockNumber")
    assertThat(createdMeter.totalTime(TimeUnit.SECONDS)).isEqualTo(1.0)
  }

  @Test
  fun itExtractsValue_FutureFailure() {
    val meterRegistry: MeterRegistry = SimpleMeterRegistry()
    val testClock = MockClock()
    val timerBuilder = Timer.builder("request.counter")
      .description("API request counter")
    val result =
      DynamicTagTimerCaptureBuilder<String>(
        meterRegistry = meterRegistry,
        wrappedTimerBuilder = timerBuilder,
        clock = testClock,
      ).setDynamicTagValueExtractor { "eth_blockNumber" }
        .setDynamicTagValueExtractorOnError { "eth_blockNumber_failure" }
        .setDynamicTagKey("method")
        .build()
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
