package net.consensys.zkevm.domain

import build.linea.domain.BlockInterval
import kotlinx.datetime.Instant
import linea.domain.Block
import net.consensys.isSortedBy
import net.consensys.linea.CommonDomainFunctions
import net.consensys.linea.traces.TracesCounters

data class BlocksConflation(
  val blocks: List<Block>,
  val conflationResult: ConflationCalculationResult
) : BlockInterval {
  init {
    require(blocks.isSortedBy { it.number }) { "Blocks list must be sorted by blockNumber" }
  }

  override val startBlockNumber: ULong
    get() = blocks.first().number.toULong()
  override val endBlockNumber: ULong
    get() = blocks.last().number.toULong()
}

data class Batch(
  val startBlockNumber: ULong,
  val endBlockNumber: ULong
) {
  init {
    require(startBlockNumber <= endBlockNumber) {
      "startBlockNumber ($startBlockNumber) must be less than or equal to endBlockNumber ($endBlockNumber)"
    }
  }

  enum class Status {
    Finalized, // Batch is finalized on L1
    Proven // Batch is ready to be sent to L1 to be finalized
  }

  fun intervalString(): String =
    CommonDomainFunctions.blockIntervalString(startBlockNumber, endBlockNumber)

  fun toStringSummary(): String {
    return "Batch(startBlockNumber=$startBlockNumber, endBlockNumber=$endBlockNumber)"
  }
}

enum class ConflationTrigger(val triggerPriority: Int) {
  // Business logic needs priority to pick the trigger in case multiple calculators trigger conflation.
  // TARGET_BLOCK_NUMBER needs to be the highest priority as it is used as conflation, blob and aggregation boundary.
  TARGET_BLOCK_NUMBER(1),
  DATA_LIMIT(2),
  TRACES_LIMIT(3),
  TIME_LIMIT(4),
  BLOCKS_LIMIT(5),
  SWITCH_CUTOFF(6)
}

data class ConflationCalculationResult(
  override val startBlockNumber: ULong,
  override val endBlockNumber: ULong,
  val conflationTrigger: ConflationTrigger,
  val tracesCounters: TracesCounters
) : BlockInterval {
  init {
    require(startBlockNumber <= endBlockNumber) {
      "startBlockNumber ($startBlockNumber) must be less than or equal to endBlockNumber ($endBlockNumber)"
    }
  }
}

data class BlockCounters(
  val blockNumber: ULong,
  val blockTimestamp: Instant,
  val tracesCounters: TracesCounters,
  val blockRLPEncoded: ByteArray
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BlockCounters

    if (blockNumber != other.blockNumber) return false
    if (blockTimestamp != other.blockTimestamp) return false
    if (tracesCounters != other.tracesCounters) return false
    if (!blockRLPEncoded.contentEquals(other.blockRLPEncoded)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = blockNumber.hashCode()
    result = 31 * result + blockTimestamp.hashCode()
    result = 31 * result + tracesCounters.hashCode()
    result = 31 * result + blockRLPEncoded.contentHashCode()
    return result
  }

  override fun toString(): String {
    return "BlockCounters(blockNumber=$blockNumber, " +
      "blockTimestamp=$blockTimestamp, " +
      "tracesCounters=$tracesCounters, " +
      "blockRLPEncoded=${blockRLPEncoded.size}bytes)"
  }
}
