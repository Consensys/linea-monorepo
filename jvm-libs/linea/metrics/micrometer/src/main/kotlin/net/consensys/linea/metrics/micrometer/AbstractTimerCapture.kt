package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.Clock
import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.Timer
import net.consensys.linea.metrics.TimerCapture

/**
 * An abstract class which contains everything is needed to make pretty much any Timer capture.
 * Mimicks the Builder pattern and allows captureTime implementation to perform different kinds of
 * captures TODO: In order to improve performance, TimerCapture instances can be cached into a
 * thread safe Map just like it's made in MeterRegistry class
 */
abstract class AbstractTimerCapture<T> : TimerCapture<T> {
  protected val meterRegistry: MeterRegistry
  protected var timerBuilder: Timer.Builder
  protected var clock = Clock.SYSTEM

  constructor(meterRegistry: MeterRegistry, name: String) {
    this.meterRegistry = meterRegistry
    timerBuilder = Timer.builder(name)
  }

  constructor(meterRegistry: MeterRegistry, timerBuilder: Timer.Builder) {
    this.meterRegistry = meterRegistry
    this.timerBuilder = timerBuilder
  }

  open fun setDescription(description: String): AbstractTimerCapture<T> {
    timerBuilder.description(description)
    return this
  }

  open fun setTag(tagKey: String, tagValue: String): AbstractTimerCapture<T> {
    timerBuilder.tag(tagKey, tagValue)
    return this
  }

  open fun setClock(clock: Clock): AbstractTimerCapture<T> {
    this.clock = clock
    return this
  }
}
