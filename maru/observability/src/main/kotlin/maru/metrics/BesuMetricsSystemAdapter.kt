/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.metrics

import com.google.common.cache.Cache
import io.vertx.core.Vertx
import net.consensys.linea.metrics.MetricsFacade
import org.apache.logging.log4j.LogManager
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem
import org.hyperledger.besu.plugin.services.metrics.Counter
import org.hyperledger.besu.plugin.services.metrics.Histogram
import org.hyperledger.besu.plugin.services.metrics.LabelledMetric
import org.hyperledger.besu.plugin.services.metrics.LabelledSuppliedMetric
import org.hyperledger.besu.plugin.services.metrics.LabelledSuppliedSummary
import org.hyperledger.besu.plugin.services.metrics.OperationTimer
import org.hyperledger.besu.plugin.services.MetricsSystem as BesuMetricsSystem
import org.hyperledger.besu.plugin.services.metrics.MetricCategory as BesuMetricCategory

class BesuMetricsSystemAdapter(
  val metricsFacade: MetricsFacade,
  val vertx: Vertx,
  val enabledMaruCategories: Set<BesuMetricCategory> =
    MaruMetricsCategory.entries
      .map {
        BesuMetricsCategoryAdapter.from(it)
      }.toSet(),
) : BesuMetricsSystem {
  private val logger = LogManager.getLogger(BesuMetricsSystemAdapter::class.java)
  private val noOpMetricsSystem = NoOpMetricsSystem()

  override fun createLabelledCounter(
    category: BesuMetricCategory,
    name: String,
    help: String,
    vararg labelNames: String,
  ): LabelledMetric<Counter> =
    LabelledCounterAdapter(
      metricsFacade = metricsFacade,
      lineaMetricsCategory = BesuMetricsCategoryAdapter.toLineaMetricsCategory(category),
      name = name,
      description = help,
      labelNames = labelNames.toList(),
    )

  override fun createLabelledSuppliedCounter(
    category: BesuMetricCategory,
    name: String,
    help: String,
    vararg labelNames: String,
  ): LabelledSuppliedMetric =
    LabelledSuppliedCounterAdapter(
      metricsFacade = metricsFacade,
      lineaMetricsCategory = BesuMetricsCategoryAdapter.toLineaMetricsCategory(category),
      name = name,
      description = help,
      labelNames = labelNames.toList(),
      vertx = vertx,
    )

  override fun createLabelledSuppliedGauge(
    category: BesuMetricCategory,
    name: String,
    help: String,
    vararg labelNames: String,
  ): LabelledSuppliedMetric =
    LabelledSuppliedGaugeAdapter(
      metricsFacade = metricsFacade,
      lineaMetricsCategory = BesuMetricsCategoryAdapter.toLineaMetricsCategory(category),
      name = name,
      description = help,
      labelNames = labelNames.toList(),
    )

  override fun createLabelledTimer(
    category: BesuMetricCategory,
    name: String,
    help: String,
    vararg labelNames: String,
  ): LabelledMetric<OperationTimer> =
    LabelledTimerAdapter(
      metricsFacade = metricsFacade,
      lineaMetricsCategory = BesuMetricsCategoryAdapter.toLineaMetricsCategory(category),
      labelNames = labelNames.toList(),
      name = name,
      description = help,
    )

  override fun createSimpleLabelledTimer(
    category: BesuMetricCategory,
    name: String,
    help: String,
    vararg labelNames: String,
  ): LabelledMetric<OperationTimer> =
    createLabelledTimer(
      category = category,
      name = name,
      help = help,
      labelNames = labelNames,
    )

  override fun createLabelledHistogram(
    category: BesuMetricCategory,
    name: String,
    help: String,
    buckets: DoubleArray,
    vararg labelNames: String,
  ): LabelledMetric<Histogram> =
    LabelledHistogramAdapter(
      metricsFacade = metricsFacade,
      lineaMetricsCategory = BesuMetricsCategoryAdapter.toLineaMetricsCategory(category),
      name = name,
      description = help,
      labelNames = labelNames.toList(),
    )

  override fun createLabelledSuppliedSummary(
    category: BesuMetricCategory,
    name: String,
    help: String,
    vararg labelNames: String,
  ): LabelledSuppliedSummary =
    LabelledSuppliedSummaryAdapter(
      metricsFacade = metricsFacade,
      lineaMetricsCategory = BesuMetricsCategoryAdapter.toLineaMetricsCategory(category),
      name = name,
      description = help,
      labelNames = labelNames.toList(),
      vertx = vertx,
    )

  override fun createGuavaCacheCollector(
    category: BesuMetricCategory,
    name: String,
    cache: Cache<*, *>,
  ) {
    logger.warn("Guava cache collector is not supported in Maru metrics system. Category: $category, Name: $name")
    noOpMetricsSystem.createGuavaCacheCollector(
      category,
      name,
      cache,
    )
  }

  override fun getEnabledCategories(): Set<BesuMetricCategory> = enabledMaruCategories
}
