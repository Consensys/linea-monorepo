package net.consensys.zkevm.ethereum.coordination.aggregation

import java.util.concurrent.ConcurrentSkipListSet

/**
 * Target end block numbers for [AggregationTriggerCalculatorByTargetBlockNumbers], implementing `() -> List<ULong>`.
 *
 * Initialized with `proof-aggregation.target-end-blocks`. Additional ends are registered at runtime via
 * [registerAggregationEndBlockInclusive] when conflation closes on
 * [net.consensys.zkevm.domain.ConflationTrigger.HARD_FORK] (e.g. timestamp-based hard forks).
 *
 * In-memory only: coordinator restart does not replay registered block targets; pending blobs may still be
 * aggregated by proof-limit or deadline triggers.
 */
class HardForkAggregationTargetEndBlocks(
  configuredTargetEndBlocks: List<ULong> = emptyList(),
) : () -> List<ULong> {
  private val endBlockNumbers = ConcurrentSkipListSet(compareBy<ULong> { it })

  init {
    endBlockNumbers.addAll(configuredTargetEndBlocks)
  }

  fun registerAggregationEndBlockInclusive(endBlockNumber: ULong) {
    endBlockNumbers.add(endBlockNumber)
  }

  override fun invoke(): List<ULong> = endBlockNumbers.toList()
}
