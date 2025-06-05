package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.Counter as MicrometerCounter
import net.consensys.linea.metrics.Counter as LineaCounter

class MicrometerCounterAdapter(private val adaptee: MicrometerCounter) : LineaCounter {
  override fun increment(amount: Double) {
    adaptee.increment(amount)
  }

  override fun increment() {
    adaptee.increment()
  }
}
