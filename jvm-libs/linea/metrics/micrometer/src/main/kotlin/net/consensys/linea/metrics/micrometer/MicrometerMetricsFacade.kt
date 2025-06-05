package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.DistributionSummary
import io.micrometer.core.instrument.Gauge
import io.micrometer.core.instrument.MeterRegistry
import net.consensys.linea.metrics.Counter
import net.consensys.linea.metrics.Histogram
import net.consensys.linea.metrics.MetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.metrics.Tag
import net.consensys.linea.metrics.TimerCapture
import java.util.function.Function
import java.util.function.Supplier
import io.micrometer.core.instrument.Counter as MicrometerCounter
import io.micrometer.core.instrument.Timer as MicrometerTimer

class MicrometerMetricsFacade(
  private val registry: MeterRegistry,
  private val metricsPrefix: String? = null,
  private val allMetricsCommonTags: List<Tag> = emptyList(),
) : MetricsFacade {
  companion object {
    private val validBaseUnits = listOf(
      "seconds",
      "minutes",
      "hours",
    )

    fun requireValidBaseUnit(baseUnit: String) {
      require(validBaseUnits.contains(baseUnit))
    }
  }

  init {
    metricsPrefix?.requireValidMicrometerName()
    allMetricsCommonTags.forEach { it.requireValidMicrometerName() }
  }

  private val allMetricsCommonMicrometerTags = allMetricsCommonTags.toMicrometerTag()

  private fun metricHandle(category: MetricsCategory, metricName: String): String {
    val prefixName = if (metricsPrefix == null) "" else "$metricsPrefix."
    return "$prefixName${category.toValidMicrometerName()}.$metricName"
  }

  override fun createGauge(
    category: MetricsCategory,
    name: String,
    description: String,
    measurementSupplier: Supplier<Number>,
    tags: List<Tag>,
  ) {
    category.toValidMicrometerName().requireValidMicrometerName()
    name.requireValidMicrometerName()
    tags.forEach { it.requireValidMicrometerName() }
    val builder = Gauge.builder(metricHandle(category, name), measurementSupplier)
      .description(description)
      .tags(allMetricsCommonMicrometerTags)
      .tags(tags.toMicrometerTag())
    builder.register(registry)
  }

  override fun createCounter(
    category: MetricsCategory,
    name: String,
    description: String,
    tags: List<Tag>,
  ): Counter {
    category.toValidMicrometerName().requireValidMicrometerName()
    name.requireValidMicrometerName()
    tags.forEach { it.requireValidMicrometerName() }
    val builder = MicrometerCounter.builder(metricHandle(category, name))
      .description(description)
      .tags(allMetricsCommonMicrometerTags)
      .tags(tags.toMicrometerTag())
    return MicrometerCounterAdapter(builder.register(registry))
  }

  override fun createHistogram(
    category: MetricsCategory,
    name: String,
    description: String,
    tags: List<Tag>,
    isRatio: Boolean,
    baseUnit: String?,
  ): Histogram {
    category.toValidMicrometerName().requireValidMicrometerName()
    name.requireValidMicrometerName()
    tags.forEach { it.requireValidMicrometerName() }
    if (baseUnit != null) requireValidBaseUnit(baseUnit)
    val distributionSummaryBuilder = DistributionSummary.builder(metricHandle(category, name))
      .description(description)
      .baseUnit(baseUnit)
      .tags(allMetricsCommonMicrometerTags)
      .tags(tags.toMicrometerTag())
    if (isRatio) {
      distributionSummaryBuilder.scale(100.0)
      distributionSummaryBuilder.maximumExpectedValue(100.0)
    }
    return MicrometerHistogramAdapter(distributionSummaryBuilder.register(registry))
  }

  override fun <T> createSimpleTimer(
    category: MetricsCategory,
    name: String,
    description: String,
    tags: List<Tag>,
  ): TimerCapture<T> {
    category.toValidMicrometerName().requireValidMicrometerName()
    name.requireValidMicrometerName()
    val builder = MicrometerTimer.builder(metricHandle(category, name))
      .description(description)
      .tags(allMetricsCommonMicrometerTags)
      .tags(tags.toMicrometerTag())

    return SimpleTimerCapture(registry, builder)
  }

  override fun <T> createDynamicTagTimer(
    category: MetricsCategory,
    name: String,
    description: String,
    tags: List<Tag>,
    dynamicTagKey: String,
    dynamicTagValueExtractorOnError: Function<Throwable, String>,
    dynamicTagValueExtractor: Function<T, String>,
  ): TimerCapture<T> {
    category.toValidMicrometerName().requireValidMicrometerName()
    name.requireValidMicrometerName()
    dynamicTagKey.requireValidMicrometerName()
    val builder = MicrometerTimer.builder(metricHandle(category, name))
      .description(description)
      .tags(allMetricsCommonMicrometerTags)
      .tags(tags.toMicrometerTag())
    val dynamicTagTimerCapture = DynamicTagTimerCaptureBuilder<T>(
      meterRegistry = registry,
      wrappedTimerBuilder = builder,
    )
      .setDynamicTagKey(dynamicTagKey)
      .setDynamicTagValueExtractor(dynamicTagValueExtractor)
      .setDynamicTagValueExtractorOnError(dynamicTagValueExtractorOnError)
      .build()
    return dynamicTagTimerCapture
  }
}
