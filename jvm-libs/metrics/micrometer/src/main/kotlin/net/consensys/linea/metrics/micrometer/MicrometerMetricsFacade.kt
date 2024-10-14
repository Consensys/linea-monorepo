package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.DistributionSummary
import io.micrometer.core.instrument.Gauge
import io.micrometer.core.instrument.MeterRegistry
import net.consensys.linea.metrics.Counter
import net.consensys.linea.metrics.Histogram
import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.metrics.Tag
import java.util.function.Supplier
import io.micrometer.core.instrument.Counter as MicrometerCounter

class MicrometerMetricsFacade(private val registry: MeterRegistry, private val metricsPrefix: String) : MetricsFacade {
  companion object {
    private val validBaseUnits = listOf(
      "seconds",
      "minutes",
      "hours"
    )

    fun requireValidMicrometerName(name: String) {
      require(name.lowercase().trim() == name && name.all { it.isLetterOrDigit() || it == '.' }) {
        "$name must adhere to Micrometer naming convention!"
      }
    }
    fun requireValidBaseUnit(baseUnit: String) {
      require(validBaseUnits.contains(baseUnit))
    }
  }

  init {
    requireValidMicrometerName(metricsPrefix)
  }

  override fun createGauge(
    category: LineaMetricsCategory,
    name: String,
    description: String,
    measurementSupplier: Supplier<Number>,
    tags: List<Tag>
  ) {
    requireValidMicrometerName(category.toString())
    requireValidMicrometerName(name)
    val builder = Gauge.builder(metricHandle(category, name), measurementSupplier)
    if (tags.isNotEmpty()) {
      val flatTags = tags.flatMap {
        requireValidMicrometerName(it.key)
        listOf(it.key, it.value)
      }
      builder.tags(*flatTags.toTypedArray())
    }
    builder.description(description)
    builder.register(registry)
  }

  override fun createCounter(
    category: LineaMetricsCategory,
    name: String,
    description: String,
    tags: List<Tag>
  ): Counter {
    requireValidMicrometerName(category.toString())
    requireValidMicrometerName(name)
    val builder = MicrometerCounter.builder(metricHandle(category, name))
    if (tags.isNotEmpty()) {
      val flatTags = tags.flatMap {
        requireValidMicrometerName(it.key)
        listOf(it.key, it.value)
      }
      builder.tags(*flatTags.toTypedArray())
    }
    builder.description(description)
    return MicrometerCounterAdapter(builder.register(registry))
  }

  override fun createHistogram(
    category: LineaMetricsCategory,
    name: String,
    description: String,
    tags: List<Tag>,
    baseUnit: String
  ): Histogram {
    requireValidMicrometerName(category.toString())
    requireValidMicrometerName(name)
    requireValidBaseUnit(baseUnit)
    val distributionSummaryBuilder = DistributionSummary.builder(metricHandle(category, name))
    if (tags.isNotEmpty()) {
      val flatTags = tags.flatMap {
        requireValidMicrometerName(it.key)
        listOf(it.key, it.value)
      }
      distributionSummaryBuilder.tags(*flatTags.toTypedArray())
    }
    distributionSummaryBuilder.description(description)
    distributionSummaryBuilder.baseUnit(baseUnit)
    return MicrometerHistogramAdapter(distributionSummaryBuilder.register(registry))
  }

  private fun metricHandle(category: LineaMetricsCategory, metricName: String): String {
    return "$metricsPrefix.$category.$metricName"
  }
}
