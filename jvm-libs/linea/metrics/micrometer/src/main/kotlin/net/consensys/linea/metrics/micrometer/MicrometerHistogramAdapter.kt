package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.DistributionSummary
import net.consensys.linea.metrics.Histogram

class MicrometerHistogramAdapter(private val adapter: DistributionSummary, val isRatio: Boolean) : Histogram {
  override fun record(data: Double) {
    if (isRatio) {
      require(data in 0.0..1.0) { "isRatio histogram expects values in [0.0, 1.0], got $data" }
    }
    adapter.record(data)
  }
}
