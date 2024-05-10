package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.Clock
import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.Timer
import java.util.concurrent.Callable
import java.util.concurrent.CompletableFuture
import java.util.function.Function

/**
 * This class can tag a metric with a tag which is unknown at the start of a measured call, but can
 * be derived from its result. Please note that the extraction logic is included in the measurement
 * due to limitations of Micrometer design, thus, this class is not suitable for the precise
 * nanosecond-level measurements. Related issue:
 * https://github.com/micrometer-metrics/micrometer/issues/535
 */
class DynamicTagTimerCapture<Out>(meterRegistry: MeterRegistry, name: String) :
  AbstractTimerCapture<Out>(meterRegistry, name) {
  private var extractor: Function<Out, String>? = null
  private var extractorOnError: Function<Throwable, String>? = null
  private var tagKey: String? = null

  override fun setDescription(description: String): DynamicTagTimerCapture<Out> {
    super.setDescription(description)
    return this
  }

  override fun setTag(tagKey: String, tagValue: String): DynamicTagTimerCapture<Out> {
    throw NoSuchMethodException(
      "If you need to set both value and key, please use ${SimpleTimerCapture::class.qualifiedName}"
    )
  }

  override fun setClock(clock: Clock): DynamicTagTimerCapture<Out> {
    super.setClock(clock)
    return this
  }

  fun setTagKey(tagKey: String?): DynamicTagTimerCapture<Out> {
    this.tagKey = tagKey
    return this
  }

  fun setTagValueExtractor(extractor: Function<Out, String>): DynamicTagTimerCapture<Out> {
    this.extractor = extractor
    return this
  }

  fun setTagValueExtractorOnError(
    onErrorExtractor: Function<Throwable, String>
  ): DynamicTagTimerCapture<Out> {
    this.extractorOnError = onErrorExtractor
    return this
  }

  private fun ensureValidState() {
    assert(extractor != null)
    assert(extractorOnError != null)
    assert(tagKey != null)
  }

  override fun captureTime(f: CompletableFuture<Out>): CompletableFuture<Out> {
    ensureValidState()
    val timerSample = Timer.start(clock)
    f.whenComplete { result: Out?, error: Throwable? ->
      val tagValue: String = result?.let(extractor!!::apply) ?: extractorOnError!!.apply(error!!)
      val timer = timerBuilder.tag(tagKey!!, tagValue).register(meterRegistry)
      timerSample.stop(timer)
    }
    return f
  }

  override fun captureTime(f: Callable<Out>): Out {
    ensureValidState()
    val timerSample = Timer.start(clock)
    val result = f.call()
    val labelValue = extractor!!.apply(result)
    val timer = timerBuilder.tag(tagKey!!, labelValue).register(meterRegistry)
    timerSample.stop(timer)
    return result
  }
}
