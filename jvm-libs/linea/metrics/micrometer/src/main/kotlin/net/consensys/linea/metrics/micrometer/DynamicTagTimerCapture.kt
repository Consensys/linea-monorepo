package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.Clock
import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.Timer
import java.util.concurrent.Callable
import java.util.concurrent.CompletableFuture
import java.util.function.Function
import kotlin.concurrent.timer

class DynamicTagTimerCaptureBuilder<T>(
  private val meterRegistry: MeterRegistry,
  private val wrappedTimerBuilder: Timer.Builder,
  private val clock: Clock = Clock.SYSTEM,
) {
  private lateinit var extractor: Function<T, String>
  private lateinit var extractorOnError: Function<Throwable, String>
  private lateinit var tagKey: String

  fun setDynamicTagKey(tagKey: String): DynamicTagTimerCaptureBuilder<T> {
    this.tagKey = tagKey
    return this
  }

  fun setDynamicTagValueExtractor(extractor: Function<T, String>): DynamicTagTimerCaptureBuilder<T> {
    this.extractor = extractor
    return this
  }

  fun setDynamicTagValueExtractorOnError(
    onErrorExtractor: Function<Throwable, String>,
  ): DynamicTagTimerCaptureBuilder<T> {
    this.extractorOnError = onErrorExtractor
    return this
  }

  private fun validate() {
    require(::extractor.isInitialized)
    require(::extractorOnError.isInitialized)
    require(::tagKey.isInitialized)
  }

  fun build(): DynamicTagTimerCapture<T> {
    validate()
    return DynamicTagTimerCapture(
      meterRegistry = meterRegistry,
      clock = clock,
      timerBuilder = wrappedTimerBuilder,
      extractor = extractor,
      extractorOnError = extractorOnError,
      tagKey = tagKey,
    )
  }
}

/**
 * This class can tag a metric with a tag which is unknown at the start of a measured call, but can
 * be derived from its result. Please note that the extraction logic is included in the measurement
 * due to limitations of Micrometer design, thus, this class is not suitable for the precise
 * nanosecond-level measurements. Related issue:
 * https://github.com/micrometer-metrics/micrometer/issues/535
 */
class DynamicTagTimerCapture<T>(
  meterRegistry: MeterRegistry,
  clock: Clock,
  timerBuilder: Timer.Builder,
  private val extractor: Function<T, String>,
  private val extractorOnError: Function<Throwable, String>,
  private val tagKey: String,
) : AbstractTimerCapture<T>(meterRegistry, timerBuilder, clock) {
  override fun captureTime(f: CompletableFuture<T>): CompletableFuture<T> {
    val timerSample = Timer.start(clock)
    f.whenComplete { result: T?, error: Throwable? ->
      val tagValue: String = result?.let(extractor::apply) ?: extractorOnError.apply(error!!)
      val timer = timerBuilder.tag(tagKey, tagValue).register(meterRegistry)
      timerSample.stop(timer)
    }
    return f
  }

  override fun captureTime(action: Callable<T>): T {
    val timerSample = Timer.start(clock)
    val result = action.call()
    val labelValue = extractor.apply(result)
    val timer = timerBuilder.tag(tagKey, labelValue).register(meterRegistry)
    timerSample.stop(timer)
    return result
  }
}
