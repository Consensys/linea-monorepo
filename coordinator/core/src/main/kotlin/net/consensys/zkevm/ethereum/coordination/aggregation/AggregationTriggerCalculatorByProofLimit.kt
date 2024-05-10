package net.consensys.zkevm.ethereum.coordination.aggregation

import net.consensys.zkevm.domain.BlobCounters
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.lang.IllegalStateException

class AggregationTriggerCalculatorByProofLimit(private val maxProofsPerAggregation: UInt) :
  SyncAggregationTriggerCalculator {

  private var proofsCount = 0u

  private fun countProofs(blobCounters: BlobCounters): UInt {
    return blobCounters.numberOfBatches + 1u
  }

  @Synchronized
  override fun checkAggregationTrigger(blobCounters: BlobCounters): AggregationTrigger? {
    val newProofsCount = proofsCount + countProofs(blobCounters)
    return if (newProofsCount > maxProofsPerAggregation) {
      AggregationTrigger(aggregationTriggerType = AggregationTriggerType.PROOF_LIMIT, includeCurrentBlob = false)
    } else if (newProofsCount == maxProofsPerAggregation) {
      AggregationTrigger(aggregationTriggerType = AggregationTriggerType.PROOF_LIMIT, includeCurrentBlob = true)
    } else {
      null
    }
  }

  @Synchronized
  override fun newBlob(blobCounters: BlobCounters): SafeFuture<*> {
    return if (proofsCount + countProofs(blobCounters) >= maxProofsPerAggregation) {
      SafeFuture.failedFuture<Unit>(
        IllegalStateException("Proof count already overflowed, should have been reset before this new blob")
      )
    } else {
      proofsCount += countProofs(blobCounters)
      SafeFuture.completedFuture(Unit)
    }
  }

  @Synchronized
  override fun reset() {
    proofsCount = 0u
  }
}
