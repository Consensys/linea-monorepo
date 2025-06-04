package net.consensys.linea.metrics

import io.vertx.core.Future
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.Callable
import java.util.concurrent.CompletableFuture
import java.util.function.Function
import java.util.function.Supplier

data class Tag(val key: String, val value: String)

interface MetricsCategory {
  val name: String
}

interface Counter {
  fun increment(amount: Double)

  fun increment()
}

interface CounterProvider {
  fun withTags(tags: List<Tag>): Counter
}

interface Histogram {
  fun record(data: Double)
}

interface TimerProvider {
  fun withTags(tags: List<Tag>): Timer
}

interface Timer {
  fun <T> captureTime(f: CompletableFuture<T>): CompletableFuture<T>
  fun <T> captureTime(action: Callable<T>): T
  fun <T> captureTime(f: SafeFuture<T>): SafeFuture<T> {
    captureTime(f.toCompletableFuture())
    return f
  }
  fun <T> captureTime(f: Future<T>): Future<T> {
    captureTime(f.toCompletionStage().toCompletableFuture())
    return f
  }
}

interface TimerCapture<T> {
  fun captureTime(f: CompletableFuture<T>): CompletableFuture<T>
  fun captureTime(action: Callable<T>): T
  fun captureTime(f: SafeFuture<T>): SafeFuture<T> {
    captureTime(f.toCompletableFuture())
    return f
  }
  fun captureTime(f: Future<T>): Future<T> {
    captureTime(f.toCompletionStage().toCompletableFuture())
    return f
  }
}

interface MetricsFacade {
  fun createGauge(
    category: MetricsCategory,
    name: String,
    description: String,
    measurementSupplier: Supplier<Number>,
    tags: List<Tag> = emptyList(),
  )

  fun createCounter(
    category: MetricsCategory,
    name: String,
    description: String,
    tags: List<Tag> = emptyList(),
  ): Counter

  fun createHistogram(
    category: MetricsCategory,
    name: String,
    description: String,
    tags: List<Tag> = emptyList(),
    isRatio: Boolean = false,
    baseUnit: String? = null,
  ): Histogram

  fun <T> createSimpleTimer(
    category: MetricsCategory,
    name: String,
    description: String,
    tags: List<Tag> = emptyList(),
  ): TimerCapture<T>

  fun <T> createDynamicTagTimer(
    category: MetricsCategory,
    name: String,
    description: String,
    tagKey: String,
    tagValueExtractorOnError: Function<Throwable, String>,
    tagValueExtractor: Function<T, String>,
  ): TimerCapture<T>

  fun createCounterProvider(
    category: MetricsCategory,
    name: String,
    description: String,
    commonTags: List<Tag> = emptyList(),
  ): CounterProvider

  fun createTimerProvider(
    category: MetricsCategory,
    name: String,
    description: String,
    commonTags: List<Tag> = emptyList(),
  ): TimerProvider
}

class FakeHistogram : Histogram {
  val records = mutableListOf<Double>()
  override fun record(data: Double) {
    records.add(data)
  }
}
