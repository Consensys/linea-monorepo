package net.consensys.zkevm.ethereum.coordination.conflation

import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationTrigger
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

/**
 * Conflation calculator that triggers on TTD (Total Terminal Difficulty) hard fork.
 * When a block reaches the configured TTD, it becomes the last block in the current conflation.
 */
class TtdHardForkConflationCalculator(
  private val totalTerminalDifficulty: ULong,
  initialTotalDifficulty: ULong,
  private val log: Logger = LogManager.getLogger(TtdHardForkConflationCalculator::class.java),
) : ConflationCalculator {

  override val id: String = "TTD_HARD_FORK"

  private var ttdReached = initialTotalDifficulty >= totalTerminalDifficulty

  override fun checkOverflow(blockCounters: BlockCounters): ConflationCalculator.OverflowTrigger? {
    if (!ttdReached && blockCounters.totalDifficulty >= totalTerminalDifficulty) {
      log.info(
        "TTD reached at block {} with total difficulty {}. TTD threshold: {}",
        blockCounters.blockNumber,
        blockCounters.totalDifficulty,
        totalTerminalDifficulty,
      )

      ttdReached = true
      return ConflationCalculator.OverflowTrigger(
        trigger = ConflationTrigger.HARD_FORK,
        singleBlockOverSized = false,
      )
    }

    return null
  }

  override fun appendBlock(blockCounters: BlockCounters) {
    // No state to track - TTD is checked in checkOverflow
  }

  override fun reset() {
    // Note: We don't reset ttdReached as TTD is a one-time event
  }

  override fun copyCountersTo(counters: ConflationCounters) {
    // No counters to copy - this calculator doesn't track conflation data
  }
}
