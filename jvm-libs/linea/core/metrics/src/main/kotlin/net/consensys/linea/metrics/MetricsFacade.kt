package net.consensys.linea.metrics

import java.util.function.Supplier

data class Tag(val key: String, val value: String)

enum class LineaMetricsCategory {
  CONFLATION,
  BATCH,
  BLOB,
  AGGREGATION,
  GAS_PRICE_CAP,
  TX_EXCLUSION_API;

  override fun toString(): String {
    return this.name.replace('_', '.').lowercase()
  }
}

interface Counter {
  fun increment(amount: Double)

  fun increment()
}

interface MetricsFacade {
  fun createGauge(
    category: LineaMetricsCategory,
    name: String,
    description: String,
    measurementSupplier: Supplier<Number>,
    tags: List<Tag> = emptyList()
  )

  fun createCounter(
    category: LineaMetricsCategory,
    name: String,
    description: String,
    tags: List<Tag> = emptyList()
  ): Counter
}
