package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.MockClock
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.Future
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
    val result =
      DynamicTagTimerCapture<String>(meterRegistry, "request.counter")
        .setTagValueExtractor { "eth_blockNumber" }
        .setTagValueExtractorOnError { "eth_blockNumber_failure" }
        .setDescription("API request counter")
        .setTagKey("method")
        .setClock(testClock)
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
  fun itExtractsValue_FutureSuccess() {
    val meterRegistry: MeterRegistry = SimpleMeterRegistry()
    val testClock = MockClock()
    val future: SafeFuture<String> = SafeFuture()
    val result =
      DynamicTagTimerCapture<String>(meterRegistry, "request.counter")
        .setTagValueExtractor { "eth_blockNumber" }
        .setTagValueExtractorOnError { "eth_blockNumber_failure" }
        .setDescription("API request counter")
        .setTagKey("method")
        .setClock(testClock)
        .captureTime(Future.succeededFuture("measured_callback_result"))
    future.complete("measured_callback_result")

    assertThat(result.toCompletionStage().toCompletableFuture().get())
      .isEqualTo("measured_callback_result")

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
      DynamicTagTimerCapture<String>(meterRegistry, "request.counter")
        .setTagValueExtractor { "eth_blockNumber" }
        .setTagValueExtractorOnError { "eth_blockNumber_failure" }
        .setDescription("API request counter")
        .setTagKey("method")
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
