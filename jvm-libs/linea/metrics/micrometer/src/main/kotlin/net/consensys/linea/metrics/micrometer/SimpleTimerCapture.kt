package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.Clock
import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.Timer
import java.util.concurrent.Callable
import java.util.concurrent.CompletableFuture
import java.util.concurrent.ConcurrentHashMap

/**
 * An abstract class which contains everything is needed to make pretty much any Timer capture.
 * Mimicks the Builder pattern and allows captureTime implementation to perform different kinds of
 * captures
 */
class SimpleTimerCapture<T> : AbstractTimerCapture<T> {
  constructor(meterRegistry: MeterRegistry, name: String) : super(meterRegistry, name)
  constructor(
    meterRegistry: MeterRegistry,
    timerBuilder: Timer.Builder
  ) : super(meterRegistry, timerBuilder)

  companion object {
    private val timerCache = ConcurrentHashMap<Timer.Builder, Timer>()
  }

  override fun setDescription(description: String): SimpleTimerCapture<T> {
    super.setDescription(description)
    return this
  }

  override fun setTag(tagKey: String, tagValue: String): SimpleTimerCapture<T> {
    super.setTag(tagKey, tagValue)
    return this
  }

  override fun setClock(clock: Clock): SimpleTimerCapture<T> {
    super.setClock(clock)
    return this
  }

  private fun getOrCreateTimer(): Timer {
    return timerCache.computeIfAbsent(timerBuilder) { builder ->
      builder.register(meterRegistry)
    }
  }

  override fun captureTime(f: CompletableFuture<T>): CompletableFuture<T> {
    val timer = getOrCreateTimer()
    val timerSample = Timer.start(clock)
    f.whenComplete { _, _ -> timerSample.stop(timer) }
    return f
  }

  override fun captureTime(f: Callable<T>): T {
    val timer = getOrCreateTimer()
    val timerSample = Timer.start(clock)
    val result = f.call()
    timerSample.stop(timer)
    return result
  }
}
