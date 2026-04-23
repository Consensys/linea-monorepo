package net.consensys.linea.metrics

import java.util.concurrent.Callable
import java.util.concurrent.CompletableFuture
import java.util.concurrent.ConcurrentHashMap
import java.util.function.Supplier

class FakeCounter : Counter {
  @Volatile
  var total: Double = 0.0
    private set

  override fun increment(amount: Double) {
    synchronized(this) { total += amount }
  }

  override fun increment() {
    increment(1.0)
  }

  fun reset() {
    synchronized(this) { total = 0.0 }
  }
}

class FakeTimer : Timer {
  val recordedDurationsNanos: MutableList<Long> = java.util.Collections.synchronizedList(mutableListOf())

  override fun <T> captureTime(f: CompletableFuture<T>): CompletableFuture<T> {
    val start = System.nanoTime()
    return f.whenComplete { _, _ -> recordedDurationsNanos.add(System.nanoTime() - start) }
  }

  override fun <T> captureTime(action: Callable<T>): T {
    val start = System.nanoTime()
    try {
      return action.call()
    } finally {
      recordedDurationsNanos.add(System.nanoTime() - start)
    }
  }
}

class FakeDynamicTagTimer<T>(
  private val tagValueExtractor: (T) -> List<Tag>,
  private val tagValueExtractorOnError: (Throwable) -> List<Tag>,
) : DynamicTagTimer<T> {
  data class Record(val tags: List<Tag>, val durationNanos: Long)

  val records: MutableList<Record> = java.util.Collections.synchronizedList(mutableListOf())

  override fun captureTime(f: CompletableFuture<T>): CompletableFuture<T> {
    val start = System.nanoTime()
    return f.whenComplete { result, error ->
      val tags = if (error != null) tagValueExtractorOnError(error) else tagValueExtractor(result)
      records.add(Record(tags, System.nanoTime() - start))
    }
  }

  override fun captureTime(action: Callable<T>): T {
    val start = System.nanoTime()
    val result = runCatching { action.call() }
    val tags = result.fold(
      onSuccess = { tagValueExtractor(it) },
      onFailure = { tagValueExtractorOnError(it) },
    )
    records.add(Record(tags, System.nanoTime() - start))
    return result.getOrThrow()
  }
}

class FakeCounterFactory : CounterFactory {
  val counters: MutableMap<List<Tag>, FakeCounter> = ConcurrentHashMap()

  override fun create(tags: List<Tag>): Counter = counters.getOrPut(tags) { FakeCounter() }
}

class FakeTimerFactory : TimerFactory {
  val timers: MutableMap<List<Tag>, FakeTimer> = ConcurrentHashMap()

  override fun create(tags: List<Tag>): Timer = timers.getOrPut(tags) { FakeTimer() }
}

/**
 * Fake [MetricsFacade] implementation intended for testing.
 *
 * Every metric registered is captured and exposed via public properties so tests can make
 * assertions on the metrics that the unit under test declared and updated, without pulling
 * in the `micrometer` implementation.
 */
class FakeMetricsFacade : MetricsFacade {

  data class MetricId(val category: String, val name: String) {
    companion object {
      fun of(category: MetricsCategory, name: String): MetricId = MetricId(category.name, name)
    }
  }

  data class GaugeEntry(
    val description: String,
    val tags: List<Tag>,
    val supplier: Supplier<Number>,
  )

  data class HistogramEntry(
    val description: String,
    val tags: List<Tag>,
    val isRatio: Boolean,
    val baseUnit: String?,
    val publishPercentileHistogram: Boolean,
    val histogram: FakeHistogram,
  )

  data class CounterFactoryEntry(
    val description: String,
    val commonTags: List<Tag>,
    val factory: FakeCounterFactory,
  )

  data class TimerFactoryEntry(
    val description: String,
    val commonTags: List<Tag>,
    val factory: FakeTimerFactory,
  )

  data class DynamicTagTimerEntry(
    val description: String,
    val commonTags: List<Tag>,
    val timer: FakeDynamicTagTimer<*>,
  )

  val gauges: MutableMap<MetricId, MutableList<GaugeEntry>> = ConcurrentHashMap()
  val histograms: MutableMap<MetricId, MutableList<HistogramEntry>> = ConcurrentHashMap()
  val counterFactories: MutableMap<MetricId, CounterFactoryEntry> = ConcurrentHashMap()
  val timerFactories: MutableMap<MetricId, TimerFactoryEntry> = ConcurrentHashMap()
  val dynamicTagTimers: MutableMap<MetricId, DynamicTagTimerEntry> = ConcurrentHashMap()

  override fun createGauge(
    category: MetricsCategory,
    name: String,
    description: String,
    measurementSupplier: Supplier<Number>,
    tags: List<Tag>,
  ) {
    gauges
      .computeIfAbsent(MetricId.of(category, name)) { java.util.Collections.synchronizedList(mutableListOf()) }
      .add(GaugeEntry(description, tags, measurementSupplier))
  }

  override fun createHistogram(
    category: MetricsCategory,
    name: String,
    description: String,
    tags: List<Tag>,
    isRatio: Boolean,
    baseUnit: String?,
    publishPercentileHistogram: Boolean,
  ): Histogram {
    val histogram = FakeHistogram()
    histograms
      .computeIfAbsent(MetricId.of(category, name)) { java.util.Collections.synchronizedList(mutableListOf()) }
      .add(HistogramEntry(description, tags, isRatio, baseUnit, publishPercentileHistogram, histogram))
    return histogram
  }

  override fun createTimer(
    category: MetricsCategory,
    name: String,
    description: String,
    tags: List<Tag>,
  ): Timer = createTimerFactory(category, name, description, tags).create()

  override fun <T> createDynamicTagTimer(
    category: MetricsCategory,
    name: String,
    description: String,
    tagValueExtractorOnError: (Throwable) -> List<Tag>,
    tagValueExtractor: (T) -> List<Tag>,
    commonTags: List<Tag>,
  ): DynamicTagTimer<T> {
    val timer = FakeDynamicTagTimer(tagValueExtractor, tagValueExtractorOnError)
    dynamicTagTimers[MetricId.of(category, name)] =
      DynamicTagTimerEntry(description, commonTags, timer)
    return timer
  }

  override fun createCounterFactory(
    category: MetricsCategory,
    name: String,
    description: String,
    commonTags: List<Tag>,
  ): CounterFactory {
    return counterFactories.computeIfAbsent(MetricId.of(category, name)) {
      CounterFactoryEntry(description, commonTags, FakeCounterFactory())
    }.factory
  }

  override fun createTimerFactory(
    category: MetricsCategory,
    name: String,
    description: String,
    commonTags: List<Tag>,
  ): TimerFactory {
    return timerFactories.computeIfAbsent(MetricId.of(category, name)) {
      TimerFactoryEntry(description, commonTags, FakeTimerFactory())
    }.factory
  }

  fun counter(category: MetricsCategory, name: String, tags: List<Tag> = emptyList()): FakeCounter? =
    counterFactories[MetricId.of(category, name)]?.factory?.counters?.get(tags)

  fun timer(category: MetricsCategory, name: String, tags: List<Tag> = emptyList()): FakeTimer? =
    timerFactories[MetricId.of(category, name)]?.factory?.timers?.get(tags)

  fun histograms(category: MetricsCategory, name: String): List<HistogramEntry> =
    histograms[MetricId.of(category, name)]?.toList() ?: emptyList()

  fun gauges(category: MetricsCategory, name: String): List<GaugeEntry> =
    gauges[MetricId.of(category, name)]?.toList() ?: emptyList()
}
