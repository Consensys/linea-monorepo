package net.consensys.zkevm.ethereum.coordination.aggregation

import net.consensys.zkevm.domain.BlobCounters
import net.consensys.zkevm.domain.BlobsToAggregate

class AggregationTriggerCalculatorByBlobLimit(private val maxBlobsPerAggregation: UInt) :
  SyncAggregationTriggerCalculator {

  data class InFlightAggregation(
    val blobsCount: UInt,
    val blobsToAggregate: BlobsToAggregate,
  )

  private var inFlightAggregation: InFlightAggregation? = null

  @Synchronized
  override fun checkAggregationTrigger(blobCounters: BlobCounters): AggregationTrigger? {
    val blobCount = (inFlightAggregation?.blobsCount ?: 0u) + 1u
    return if (blobCount >= maxBlobsPerAggregation) {
      AggregationTrigger(
        aggregationTriggerType = AggregationTriggerType.BLOB_LIMIT,
        aggregation = BlobsToAggregate(
          inFlightAggregation?.blobsToAggregate?.startBlockNumber
            ?: blobCounters.startBlockNumber,
          blobCounters.endBlockNumber,
        ),
      )
    } else {
      null
    }
  }

  @Synchronized
  override fun newBlob(blobCounters: BlobCounters) {
    val blobCount = (inFlightAggregation?.blobsCount ?: 0u) + 1u
    if (blobCount > maxBlobsPerAggregation) {
      throw IllegalArgumentException(
        "Aggregation blob limit overflow: maxBlobsPerAggregation=$maxBlobsPerAggregation " +
          "blobsCount=$blobCount " +
          "blob=${blobCounters.intervalString()}",
      )
    }
    inFlightAggregation = InFlightAggregation(
      blobsCount = blobCount,
      blobsToAggregate = BlobsToAggregate(
        inFlightAggregation?.blobsToAggregate?.startBlockNumber
          ?: blobCounters.startBlockNumber,
        blobCounters.endBlockNumber,
      ),
    )
  }

  @Synchronized
  override fun reset() {
    inFlightAggregation = null
  }
}
