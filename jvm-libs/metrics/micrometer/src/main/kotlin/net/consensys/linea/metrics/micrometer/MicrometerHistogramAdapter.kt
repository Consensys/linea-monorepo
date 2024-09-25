package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.DistributionSummary
import net.consensys.linea.metrics.Histogram

class MicrometerHistogramAdapter(private val adapter: DistributionSummary) : Histogram {
  override fun record(var1: Double) {
    adapter.record(var1)
  }

  override fun count(): Long {
    return adapter.count()
  }

  override fun totalAmount(): Double {
    return adapter.totalAmount()
  }

  override fun mean(): Double {
    return adapter.mean()
  }

  override fun max(): Double {
    return adapter.max()
  }
}
