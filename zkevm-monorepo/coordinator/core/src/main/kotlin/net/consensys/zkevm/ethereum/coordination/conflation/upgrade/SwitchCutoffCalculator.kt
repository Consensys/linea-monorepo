package net.consensys.zkevm.ethereum.coordination.conflation.upgrade

import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationTrigger
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCounters
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

@Deprecated("We may use it for 4844 switch, but the switch procedure is yet unknown")
class SwitchCutoffCalculator(private val switchBlockNumber: ULong?) : ConflationCalculator {
  private val log: Logger = LogManager.getLogger(this::class.java)
  override val id: String = ConflationTrigger.SWITCH_CUTOFF.name
  override fun checkOverflow(blockCounters: BlockCounters): ConflationCalculator.OverflowTrigger? {
    return when {
      blockCounters.blockNumber == switchBlockNumber ->
        ConflationCalculator.OverflowTrigger(
          ConflationTrigger.SWITCH_CUTOFF,
          false
        )

      switchBlockNumber == null -> null
      blockCounters.blockNumber > switchBlockNumber ->
        throw IllegalStateException(
          "checkOverflow with block number ${blockCounters.blockNumber} is not expected " +
            "after block number $switchBlockNumber!"
        )

      else -> null
    }
  }

  override fun appendBlock(blockCounters: BlockCounters) {
    log.debug("Block {} was received", blockCounters.blockNumber)
  }

  override fun reset() {
    log.debug("Reset was called")
  }

  override fun copyCountersTo(counters: ConflationCounters) {
    log.debug("copyCountersTo was called")
  }
}
