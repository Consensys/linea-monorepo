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
 * Deferred trigger calculator that creates conflation boundaries at FTX execution blocks
 * that require invalidity proofs.
 *
 * Reads from processedFtxQueue (shared with AggregationCalculatorByForcedTransaction)
 * to detect processed FTXs with non-Included results and triggers FORCED_TRANSACTION
 * conflation at (blockNumber - 1) to isolate the FTX block for invalidity proof generation.
 *
 * Rules:
 * - Trigger conflation at (ftx.blockNumber - 1) when inclusionResult != Included
 * - Do NOT trigger if inclusionResult == Included (successfully executed FTXs)
 *
 * Note: This calculator has its own dedicated queue (not shared with the aggregation calculator)
 * and consumes entries via poll(). This prevents the aggregation calculator from draining
 * entries before the conflation calculator reads them.
 *
 * This ensures that:
 * 1. Failed FTXs are isolated in their own conflation for invalidity proof generation
 * 2. Successfully included FTXs don't create unnecessary conflation boundaries
 * 3. The conflation before the failed FTX block is sealed at (blockNumber - 1)
 */
class ConflationCalculatorByForcedTransaction(
  private val processedFtxQueue: Queue<ForcedTransactionInclusionStatus>,
  private val log: Logger = LogManager.getLogger(ConflationCalculatorByForcedTransaction::class.java),
) : ConflationCalculator {
  override val id: String = ConflationTrigger.FORCED_TRANSACTION.name

  // Track pending trigger blocks (blockNumber) for non-included FTXs
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

    val newTriggerBlocks = newFtxs.map { it.blockNumber }.toSet()
    pendingTriggerBlocks.addAll(newTriggerBlocks)
    logNewFtxConflationTriggers(newTriggerBlocks, newFtxs)
  }

  fun logNewFtxConflationTriggers(
    actuallyNewBlocks: Set<ULong>,
    processedFtxs: List<ForcedTransactionInclusionStatus>,
  ) {
    if (actuallyNewBlocks.isNotEmpty()) {
      val newFailedFtxs = processedFtxs.filter { actuallyNewBlocks.contains(it.blockNumber) }
      log.info(
        "appended new conflation triggers {} of non-included ftxs={}, all pending triggers {}",
        actuallyNewBlocks,
        newFailedFtxs.map { it.toStringShortForLogging() },
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
