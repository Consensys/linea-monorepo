package linea.ftx.conflation

import linea.domain.BlockCounters
import linea.domain.ConflationTrigger
import linea.forcedtx.ForcedTransactionInclusionStatus
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCounters
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.util.Queue
import kotlin.collections.map

/**
 * Trigger calculator that creates conflation boundaries AT FTX execution blocks.
 *
 * Consumes from a dedicated processedFtxQueue (not shared with the aggregation calculator)
 * and triggers FORCED_TRANSACTION conflation at ftx.blockNumber for every processed FTX,
 * regardless of inclusion result. This makes each FTX execution block the first block
 * of a new blob.
 *
 * Rules:
 * - Trigger conflation overflow at ftx.blockNumber for every processed FTX
 *
 * The queue is dedicated to this calculator (populated by onFtxProcessed callback).
 * Entries are consumed via poll() so there is no race with the aggregation calculator.
 */
class ConflationCalculatorByForcedTransaction(
  private val processedFtxQueue: Queue<ForcedTransactionInclusionStatus>,
  private val log: Logger = LogManager.getLogger(ConflationCalculatorByForcedTransaction::class.java),
) : ConflationCalculator {
  override val id: String = ConflationTrigger.FORCED_TRANSACTION.name

  // Track pending trigger blocks (ftx.blockNumber) for all processed FTXs
  private val pendingTriggerBlocks = mutableSetOf<ULong>()

  @Volatile
  private var lastBlockNumber: ULong = 0UL

  @Synchronized
  override fun checkOverflow(blockCounters: BlockCounters): ConflationCalculator.OverflowTrigger? {
    readProcessedFtxs()

    log.debug(
      "checking ftx conflation trigger: blockNumber={} pendingTriggers={}",
      blockCounters.blockNumber,
      pendingTriggerBlocks.sorted(),
    )

    // Check if this block should trigger conflation
    return if (pendingTriggerBlocks.contains(blockCounters.blockNumber)) {
      log.info(
        "FTX conflation trigger at block={} pendingTriggers={}",
        blockCounters.blockNumber,
        pendingTriggerBlocks.sorted(),
      )
      ConflationCalculator.OverflowTrigger(
        trigger = ConflationTrigger.FORCED_TRANSACTION,
        singleBlockOverSized = false,
      )
    } else {
      null
    }
  }

  private fun readProcessedFtxs() {
    // Consume from dedicated queue — safe because this queue is not shared with the aggregation calculator.
    val newFtxs = mutableListOf<ForcedTransactionInclusionStatus>()
    while (true) {
      newFtxs.add(processedFtxQueue.poll() ?: break)
    }

    if (newFtxs.isEmpty()) {
      return
    }

    val triggerBlocks = newFtxs.map { it.blockNumber }.toSet()
    val actuallyNew = triggerBlocks - pendingTriggerBlocks
    pendingTriggerBlocks.addAll(triggerBlocks)
    if (actuallyNew.isNotEmpty()) {
      log.info(
        "appended new conflation triggers {} for ftxs={}, all pending triggers={}",
        actuallyNew,
        newFtxs.filter { it.blockNumber in actuallyNew }.map { it.toStringShortForLogging() },
        pendingTriggerBlocks.sorted(),
      )
    }
  }

  @Synchronized
  override fun appendBlock(blockCounters: BlockCounters) {
    pendingTriggerBlocks.removeIf { it <= blockCounters.blockNumber }
    lastBlockNumber = blockCounters.blockNumber
  }

  @Synchronized
  override fun reset() {
    // Don't clear pendingTriggerBlocks on reset - they persist until processed
    log.trace("FTX conflation calculator reset at block {}", lastBlockNumber)
  }

  override fun copyCountersTo(counters: ConflationCounters) {
    // No counters to copy
  }
}
