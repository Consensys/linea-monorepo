package net.consensys.zkevm.ethereum.coordination.conflation

import kotlinx.datetime.Instant
import net.consensys.linea.CommonDomainFunctions.batchIntervalString
import net.consensys.linea.isSortedBy
import net.consensys.linea.traces.TracesCounters
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.toULong
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

data class BlocksConflation(
  val blocks: List<ExecutionPayloadV1>,
  val conflationResult: ConflationCalculationResult
) {
  init {
    require(blocks.isSortedBy { it.blockNumber }) { "Blocks list must be sorted by blockNumber" }
  }
}

data class Batch(
  val startBlockNumber: UInt64,
  val endBlockNumber: UInt64,
  val proverResponse: GetProofResponse,
  val status: Status = Status.Pending
) {
  init {
    require(startBlockNumber <= endBlockNumber) {
      "startBlockNumber ($startBlockNumber) must be less than or equal to endBlockNumber ($endBlockNumber)"
    }
  }
  enum class Status {
    Finalized, // Batch is finalized on L1
    Pending // Batch is ready to be sent to L1 to be finalized
  }

  fun intervalString(): String = batchIntervalString(startBlockNumber.toULong(), endBlockNumber.toULong())

  fun toStringSummary(): String {
    return "Batch(startBlockNumber=$startBlockNumber, endBlockNumber=$endBlockNumber, status=$status)"
  }
}

enum class ConflationTrigger {
  TRACES_LIMIT,
  DATA_LIMIT,
  TIME_LIMIT,
  BLOCKS_LIMIT
}
data class ConflationCalculationResult(
  val startBlockNumber: ULong,
  val endBlockNumber: ULong,
  val tracesCounters: TracesCounters,
  val dataL1Size: UInt,
  val conflationTrigger: ConflationTrigger
) {
  init {
    require(startBlockNumber <= endBlockNumber) {
      "startBlockNumber ($startBlockNumber) must be less than or equal to endBlockNumber ($endBlockNumber)"
    }
  }
  val blocksRange: ULongRange = startBlockNumber..endBlockNumber

  fun intervalString(): String {
    return "[$startBlockNumber..$endBlockNumber](${endBlockNumber - startBlockNumber + 1u})"
  }
}

data class BlockCounters(
  val blockNumber: ULong,
  val blockTimestamp: Instant,
  val tracesCounters: TracesCounters,
  val l1DataSize: UInt
)

interface TracesConflationCalculator {
  val lastBlockNumber: ULong
  fun newBlock(blockCounters: BlockCounters)
  fun onConflatedBatch(conflationConsumer: (ConflationCalculationResult) -> SafeFuture<*>)
  fun getConflationInProgress(): ConflationCalculationResult?
}

interface ConflationService {
  fun newBlock(block: ExecutionPayloadV1, blockCounters: BlockCounters)
  fun onConflatedBatch(consumer: (BlocksConflation) -> SafeFuture<*>)
}
