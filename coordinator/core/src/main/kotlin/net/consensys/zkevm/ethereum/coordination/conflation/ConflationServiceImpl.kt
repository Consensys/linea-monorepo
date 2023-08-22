package net.consensys.zkevm.ethereum.coordination.conflation

import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.PriorityBlockingQueue

class ConflationServiceImpl(
  private val calculator: TracesConflationCalculator
) :
  ConflationService {

  private val log: Logger = LogManager.getLogger(this::class.java)
  private var listener: (BlocksConflation) -> SafeFuture<*> = { SafeFuture.completedFuture(null) }
  private var blocksInProgress: MutableList<ExecutionPayloadV1> = mutableListOf()

  data class PayloadAndBlockCounters(
    val executionPayload: ExecutionPayloadV1,
    val blockCounters: BlockCounters
  ) : Comparable<PayloadAndBlockCounters> {
    override fun compareTo(other: PayloadAndBlockCounters): Int {
      return this.executionPayload.blockNumber.compareTo(other.executionPayload.blockNumber)
    }
  }

  internal val blocksToConflate = PriorityBlockingQueue<PayloadAndBlockCounters>()

  init {
    calculator.onConflatedBatch(this::handleConflation)
  }

  @Synchronized
  internal fun handleConflation(conflation: ConflationCalculationResult): SafeFuture<*> {
    log.info(
      "new conflation: batch={}, trigger={}, dataL1Size={} bytes, tracesCounters={}, blocksNumbers={}",
      conflation.intervalString(),
      conflation.conflationTrigger,
      conflation.dataL1Size,
      conflation.tracesCounters,
      conflation.blocksRange.joinToString(",", "[", "]") { it.toString() }
    )
    val blocksToConflate =
      blocksInProgress
        .filter { it.blockNumber.toULong() in conflation.blocksRange }
        .sortedBy { it.blockNumber }
    blocksInProgress.removeAll(blocksToConflate)

    return listener.invoke(BlocksConflation(blocksToConflate, conflation))
      .whenException { th ->
        log.error(
          "Conflation listener failed: batch={} errorMessage={}",
          conflation.intervalString(),
          th.message,
          th
        )
      }
  }

  @Synchronized
  override fun newBlock(block: ExecutionPayloadV1, blockCounters: BlockCounters) {
    require(block.blockNumber.toULong() == blockCounters.blockNumber) {
      "Payload blockNumber ${block.blockNumber} does not match blockCounters.blockNumber=${blockCounters.blockNumber}"
    }
    log.trace(
      "newBlock={}, calculatorLastBlockNumber={}, blocksToConflateSize={}, blocksInProgressSize={}",
      block.blockNumber,
      calculator.lastBlockNumber,
      blocksToConflate.size,
      blocksInProgress.size
    )
    blocksToConflate.add(PayloadAndBlockCounters(block, blockCounters))
    blocksInProgress.add(block)
    log.trace("block {} added to conflation queue", block.blockNumber)
    sendBlocksInOrderToTracesCounter()
  }

  private fun sendBlocksInOrderToTracesCounter() {
    var nextBlockNumberToConflate = calculator.lastBlockNumber + 1u
    var nextAvailableBlock = blocksToConflate.peek()

    while (nextAvailableBlock?.executionPayload?.blockNumber?.toULong() == nextBlockNumberToConflate) {
      nextAvailableBlock = blocksToConflate.poll(100, java.util.concurrent.TimeUnit.MILLISECONDS)
      log.trace(
        "block {} removed from conflation queue and sent to calculator",
        nextAvailableBlock?.executionPayload?.blockNumber
      )
      calculator.newBlock(nextAvailableBlock.blockCounters)
      nextBlockNumberToConflate = nextAvailableBlock.executionPayload.blockNumber.toULong() + 1u
      nextAvailableBlock = blocksToConflate.peek()
    }
  }

  override fun onConflatedBatch(consumer: (BlocksConflation) -> SafeFuture<*>) {
    this.listener = consumer
  }
}
