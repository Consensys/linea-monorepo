package net.consensys.linea.metrics.micrometer

import io.micrometer.core.instrument.ImmutableTag
import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.metrics.Tag
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class MicrometerMetricsFacadeTest {
  private lateinit var meterRegistry: MeterRegistry
  private lateinit var metricsFacade: MetricsFacade

  @BeforeEach
  fun beforeEach() {
    meterRegistry = SimpleMeterRegistry()
    metricsFacade = MicrometerMetricsFacade(meterRegistry, "linea.test")
  }

  @Test
  fun `createGauge creates gauge with specified parameters`() {
    var metricMeasureValue = 0L
    val expectedTags = listOf(Tag("key1", "value1"), Tag("key2", "value2"))
    metricsFacade.createGauge(
      category = LineaMetricsCategory.BATCH,
      name = "some.metric",
      description = "This is a test metric",
      measurementSupplier = { metricMeasureValue },
      tags = expectedTags
    )
    metricMeasureValue = 13L
    val createdGauge = meterRegistry.find("linea.test.batch.some.metric").gauge()
    assertThat(createdGauge).isNotNull
    assertThat(createdGauge!!.value()).isEqualTo(13.0)
    metricMeasureValue = 2L
    assertThat(createdGauge.value()).isEqualTo(2.0)
    assertThat(createdGauge.id.tags).isEqualTo(listOf(ImmutableTag("key1", "value1"), ImmutableTag("key2", "value2")))
    assertThat(createdGauge.id.description).isEqualTo("This is a test metric")
  }

  @Test
  fun `createCounter creates counter with specified parameters`() {
    val expectedTags = listOf(Tag("key1", "value1"), Tag("key2", "value2"))
    val counter = metricsFacade.createCounter(
      category = LineaMetricsCategory.BATCH,
      name = "some.metric",
      description = "This is a test metric",
      tags = expectedTags
    )
    val createdCounter = meterRegistry.find("linea.test.batch.some.metric").counter()
    assertThat(createdCounter!!.count()).isEqualTo(0.0)
    assertThat(createdCounter).isNotNull
    counter.increment(13.0)
    assertThat(createdCounter.count()).isEqualTo(13.0)

    counter.increment(2.0)
    assertThat(createdCounter.count()).isEqualTo(15.0)
    counter.increment()
    assertThat(createdCounter.count()).isEqualTo(16.0)
    counter.increment(0.5)
    assertThat(createdCounter.count()).isEqualTo(16.5)
    assertThat(createdCounter.id.tags).isEqualTo(listOf(ImmutableTag("key1", "value1"), ImmutableTag("key2", "value2")))
    assertThat(createdCounter.id.description).isEqualTo("This is a test metric")
  }
}
