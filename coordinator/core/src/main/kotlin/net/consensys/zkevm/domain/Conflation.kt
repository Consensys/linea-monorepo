package net.consensys.zkevm.domain

import kotlinx.datetime.Instant
import net.consensys.isSortedBy
import net.consensys.linea.CommonDomainFunctions
import net.consensys.linea.traces.TracesCounters
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1

/**
 * Represents a block interval, with inclusive start and end block numbers
 * @property startBlockNumber start block number, inclusive
 * @property endBlockNumber end block number, inclusive
 */
interface BlockInterval {
  val startBlockNumber: ULong
  val endBlockNumber: ULong
  val blocksRange: ULongRange
    get() = startBlockNumber..endBlockNumber
  fun intervalString(): String = CommonDomainFunctions.blockIntervalString(startBlockNumber, endBlockNumber)

  companion object {
    fun between(
      startBlockNumber: ULong,
      endBlockNumber: ULong
    ): BlockInterval {
      return BlockIntervalData(startBlockNumber, endBlockNumber)
    }
  }
}

fun List<BlockInterval>.toBlockIntervalsString(): String {
  return this.joinToString(
    separator = ", ",
    prefix = "[",
    postfix = "]$size",
    transform = BlockInterval::intervalString
  )
}

fun <T : BlockInterval> List<T>.filterOutWithEndBlockNumberBefore(
  endBlockNumberInclusive: ULong
): List<T> {
  return this.filter { int -> int.endBlockNumber > endBlockNumberInclusive }
}

fun assertConsecutiveIntervals(intervals: List<BlockInterval>) {
  require(intervals.isSortedBy { it.startBlockNumber }) { "Intervals must be sorted by startBlockNumber" }
  require(intervals.zipWithNext().all { (a, b) -> a.endBlockNumber + 1u == b.startBlockNumber }) {
    "Intervals must be consecutive: intervals=${intervals.toBlockIntervalsString()}"
  }
}

data class BlocksConflation(
  val blocks: List<ExecutionPayloadV1>,
  val conflationResult: ConflationCalculationResult
) {
  init {
    require(blocks.isSortedBy { it.blockNumber }) { "Blocks list must be sorted by blockNumber" }
  }
}

data class Batch(
  val startBlockNumber: ULong,
  val endBlockNumber: ULong,
  val status: Status = Status.Proven
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
    return "Batch(startBlockNumber=$startBlockNumber, endBlockNumber=$endBlockNumber, status=$status)"
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

/**
 * Represents a block interval
 * @property startBlockNumber starting block number inclusive
 * @property endBlockNumber ending block number inclusive
 */
data class BlockIntervalData(
  override val startBlockNumber: ULong,
  override val endBlockNumber: ULong
) : BlockInterval

/**
 * Data class that represents sequential blocks intervals for either Conflations, Blobs or Aggregations.
 * Example:
 *  conflations: [100..110], [111..120], [121..130] --> BlockIntervals(100, [110, 120, 130])
 *  Blobs with
 *   Blob1 2 conflations above: [100..110], [111..120]
 *   Blob2 1 conflations:  [121..130]
 *   --> BlockIntervals(100, [120, 130])
 */
data class BlockIntervals(
  val startingBlockNumber: ULong,
  val upperBoundaries: List<ULong>
) {
  // This default constructor is to avoid the parse error when deserializing
  constructor() : this(0UL, listOf())

  fun toIntervalList(): List<BlockInterval> {
    var previousBlockNumber = startingBlockNumber
    val intervals = mutableListOf<BlockInterval>()
    upperBoundaries.forEach {
      intervals.add(BlockIntervalData(previousBlockNumber, it))
      previousBlockNumber = it + 1u
    }
    return intervals
  }

  fun toBlockInterval(): BlockInterval {
    return BlockIntervalData(startingBlockNumber, upperBoundaries.last())
  }
}

fun List<BlockInterval>.toBlockIntervals(): BlockIntervals {
  require(isNotEmpty()) { "BlockIntervals list must not be empty" }
  return BlockIntervals(
    startingBlockNumber = first().startBlockNumber,
    upperBoundaries = map { it.endBlockNumber }
  )
}
