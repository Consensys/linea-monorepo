package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.DistributionSummary
import net.consensys.linea.metrics.Histogram

class MicrometerHistogramAdapter(private val adapter: DistributionSummary) : Histogram {
  override fun record(data: Double) {
    adapter.record(data)
  }
}
