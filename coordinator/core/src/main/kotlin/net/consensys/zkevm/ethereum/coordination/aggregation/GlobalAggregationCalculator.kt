package net.consensys.zkevm.ethereum.coordination.aggregation

import net.consensys.linea.metrics.Counter
import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.domain.BlobsToAggregate
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.stream.Stream

class GlobalAggregationCalculator(
  private var lastBlockNumber: ULong,
  private val syncAggregationTrigger: List<SyncAggregationTriggerCalculator>,
  private val deferredAggregationTrigger: List<DeferredAggregationTriggerCalculator>,
  metricsFacade: MetricsFacade
) : AggregationCalculator {
  private var aggregationHandler: AggregationHandler = AggregationHandler.NOOP_HANDLER
  private val log: Logger = LogManager.getLogger(this::class.java)
  private var inProgressAggregation: BlobsToAggregate? = null
  private val blobsToAggregate: MutableList<BlobCounters> = mutableListOf()
  private val blobsCounter: Counter = metricsFacade.createCounter(
    category = LineaMetricsCategory.AGGREGATION,
    name = "calculator.blobs.accepted",
    description = "Number of blobs accepted by aggregation calculator"
  )
  private val batchesCounter: Counter = metricsFacade.createCounter(
    category = LineaMetricsCategory.AGGREGATION,
    name = "calculator.batches.accepted",
    description = "Number of batches accepted by aggregation calculator"
  )

  init {
    require(syncAggregationTrigger.size + deferredAggregationTrigger.size > 0) { "Specify at least one trigger" }
    deferredAggregationTrigger.forEach { it.onAggregationTrigger(this::handleAggregationTrigger) }
    metricsFacade.createGauge(
      category = LineaMetricsCategory.AGGREGATION,
      name = "proofs.ready",
      description = "Number of proofs pending for aggregation",
      measurementSupplier = {
        blobsToAggregate.size + blobsToAggregate.sumOf { it.numberOfBatches }.toLong()
      }
    )
    metricsFacade.createGauge(
      category = LineaMetricsCategory.AGGREGATION,
      name = "batches.ready",
      description = "Number of batches pending for aggregation",
      measurementSupplier = {
        blobsToAggregate.sumOf { it.numberOfBatches }.toLong()
      }
    )
    metricsFacade.createGauge(
      category = LineaMetricsCategory.AGGREGATION,
      name = "blobs.ready",
      description = "Number of blobs pending for aggregation",
      measurementSupplier = {
        blobsToAggregate.size
      }
    )
  }

  @Synchronized
  override fun newBlob(blobCounters: BlobCounters): SafeFuture<*> {
    ensureBlobIsInOrder(blobCounters.startBlockNumber)
    lastBlockNumber = blobCounters.endBlockNumber

    val aggregationTriggers = syncAggregationTrigger.mapNotNull { it.checkAggregationTrigger(blobCounters) }
    val triggerIncludingCurrentBlob = aggregationTriggers.find { it.includeCurrentBlob }

    if (aggregationTriggers.isNotEmpty()) {
      val aggregationTriggerType =
        if (triggerIncludingCurrentBlob != null) {
          updateInProgressAggregation(blobCounters)
          triggerIncludingCurrentBlob.aggregationTriggerType
        } else {
          aggregationTriggers.first().aggregationTriggerType
        }
      handleAggregationTrigger(aggregationTriggerType)
    }

    return if (triggerIncludingCurrentBlob != null) {
      SafeFuture.completedFuture(Unit)
    } else {
      updateInProgressAggregation(blobCounters)
      SafeFuture.allOf(
        Stream.concat(syncAggregationTrigger.stream(), deferredAggregationTrigger.stream())
          .map { it.newBlob(blobCounters) }
      ).thenApply { }
    }
  }

  @Synchronized
  private fun updateInProgressAggregation(blobCounters: BlobCounters) {
    blobsToAggregate.add(blobCounters)
    blobsCounter.increment()
    batchesCounter.increment(blobCounters.numberOfBatches.toDouble())
    inProgressAggregation = if (inProgressAggregation == null) {
      BlobsToAggregate(
        blobCounters.startBlockNumber,
        blobCounters.endBlockNumber
      )
    } else {
      BlobsToAggregate(
        inProgressAggregation!!.startBlockNumber,
        blobCounters.endBlockNumber
      )
    }
  }

  override fun onAggregation(aggregationHandler: AggregationHandler) {
    this.aggregationHandler = aggregationHandler
  }

  @Synchronized
  internal fun handleAggregationTrigger(aggregationTriggerType: AggregationTriggerType): SafeFuture<Unit> {
    if (inProgressAggregation == null) {
      return SafeFuture.completedFuture(Unit)
    }
    val aggregation = inProgressAggregation!!.copy()
    log.info(
      "new aggregation: trigger={} aggregation={} blobsCount={} batchesCount={} blobs={}",
      aggregationTriggerType.name,
      aggregation.intervalString(),
      blobsToAggregate.size,
      blobsToAggregate.sumOf { it.numberOfBatches },
      blobsToAggregate.map { it.intervalString() }
    )
    reset()
    return aggregationHandler.onAggregation(aggregation)
  }

  private fun reset() {
    blobsToAggregate.clear()
    inProgressAggregation = null
    syncAggregationTrigger.forEach { it.reset() }
    deferredAggregationTrigger.forEach { it.reset() }
  }

  private fun ensureBlobIsInOrder(blockNumber: ULong) {
    if (blockNumber != (lastBlockNumber + 1u)) {
      val error = IllegalArgumentException(
        "Blobs to aggregate must be sequential: lastBlockNumber=$lastBlockNumber, startBlockNumber=$blockNumber for " +
          "new blob"
      )
      log.error(error.message)
      throw error
    }
  }
}
