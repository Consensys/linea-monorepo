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
abstract class AbstractTimerCapture<T>(
  protected val meterRegistry: MeterRegistry,
  protected val timerBuilder: Timer.Builder,
  protected val clock: Clock,
) : TimerCapture<T>
