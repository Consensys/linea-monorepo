package linea.ftx.conflation

import linea.forcedtx.ForcedTransactionInclusionStatus
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationTrigger
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculator
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCounters
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.util.Queue

/**
 * Deferred trigger calculator that creates conflation boundaries at FTX execution blocks
 * that require invalidity proofs.
 *
 * Reads from processedFtxQueue (shared with AggregationCalculatorByForcedTransaction)
 * to detect processed FTXs that require invalidity proofs and triggers FORCED_TRANSACTION
 * conflation at the FTX execution block to isolate that block for invalidity proof generation.
 *
 * Rules:
 * - Trigger conflation at ftx.blockNumber for every processed FTX
 *
 * Note: This calculator reads from the queue without consuming.
 * AggregationCalculatorByForcedTransaction is responsible for consuming items.
 *
 * This ensures that:
 * 1. Processed FTXs are isolated in their own conflation for invalidity proof generation
 * 2. Mixed invalidity proofs from consecutive simulated execution blocks are avoided
 * 3. The conflation before the FTX execution block is sealed when that block is processed
 */
class ConflationCalculatorByForcedTransaction(
  private val processedFtxQueue: Queue<ForcedTransactionInclusionStatus>,
  private val log: Logger = LogManager.getLogger(ConflationCalculatorByForcedTransaction::class.java),
) : ConflationCalculator {
  override val id: String = ConflationTrigger.FORCED_TRANSACTION.name

  // Track pending trigger blocks keyed by the FTX execution block number.
  private val pendingTriggerBlocks = mutableSetOf<ULong>()

  @Volatile
  private var lastBlockNumber: ULong = 0UL

  @Synchronized
  override fun checkOverflow(blockCounters: BlockCounters): ConflationCalculator.OverflowTrigger? {
    // First, read all available processed FTXs from the queue (without consuming)
    readProcessedFtxs()

    // Check if this block should trigger conflation
    return if (pendingTriggerBlocks.contains(blockCounters.blockNumber)) {
      log.info(
        "FTX conflation overflow detected at block {} (will seal before processing this FTX execution block)",
        blockCounters.blockNumber,
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
    // Read all available FTX statuses from the queue WITHOUT consuming
    // they cannot be consumed, it's AggregationCalculatorByForcedTransaction responsibility to consume them
    val processedFtxs = processedFtxQueue.toList()

    if (processedFtxs.isEmpty()) {
      return
    }

    val newTriggerBlocks = processedFtxs
      .map { ftx ->
        ftx.blockNumber
      }
      .toSet()

    // Only update if we have new trigger blocks not already tracked
    val actuallyNewBlocks = newTriggerBlocks - pendingTriggerBlocks
    pendingTriggerBlocks.addAll(actuallyNewBlocks)
    if (actuallyNewBlocks.isNotEmpty()) {
      val newProcessedFtxs = processedFtxs.filter {
        actuallyNewBlocks.contains(it.blockNumber)
      }
      log.info(
        "added {} new FTX conflation trigger blocks for {} processed FTXs. ftxs={} total pending triggers: {}",
        actuallyNewBlocks.size,
        newProcessedFtxs.size,
        newProcessedFtxs.map { "ftx=${it.ftxNumber} result=${it.inclusionResult}" },
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
