package net.consensys.zkevm.ethereum.coordination.aggregation

import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.domain.BlobsToAggregate
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface AggregationHandler {
  fun onAggregation(blobsToAggregate: BlobsToAggregate): SafeFuture<Unit>
  companion object {
    internal val NOOP_HANDLER: AggregationHandler = AggregationHandler { SafeFuture.completedFuture(Unit) }
  }
}
enum class AggregationTriggerType {
  TIME_LIMIT,
  PROOF_LIMIT
}

data class AggregationTrigger(
  val aggregationTriggerType: AggregationTriggerType,
  val includeCurrentBlob: Boolean
)

fun interface AggregationTriggerHandler {
  fun onAggregationTrigger(aggregationTriggerType: AggregationTriggerType): SafeFuture<Unit>
  companion object {
    internal val NOOP_HANDLER: AggregationTriggerHandler =
      AggregationTriggerHandler { SafeFuture.completedFuture(Unit) }
  }
}

/**
 * The aggregation proof can be triggered by 2 conditions:
 *  - Number of batches - if number of all batches combined for all subsequent blobs above a given threshold;
 *  - Finalization deadline: the time elapsed between first block’s timestamp in the first submitted blob and
 *     current time is greater than a finalization deadline configured.
 *
 *   Sum(blob.numberOfBatches) < prover_batches_limit
 *   blob.firstBlockTimestamp < current_time - submission_deadline;
 */
interface AggregationCalculator {
  fun newBlob(blobCounters: BlobCounters): SafeFuture<*>
  fun onAggregation(aggregationHandler: AggregationHandler)
}

interface AggregationTriggerCalculator {
  fun newBlob(blobCounters: BlobCounters): SafeFuture<*>
  fun reset()
}

interface SyncAggregationTriggerCalculator : AggregationTriggerCalculator {
  fun checkAggregationTrigger(blobCounters: BlobCounters): AggregationTrigger?
}

interface DeferredAggregationTriggerCalculator : AggregationTriggerCalculator {
  fun onAggregationTrigger(aggregationTriggerHandler: AggregationTriggerHandler)
}
