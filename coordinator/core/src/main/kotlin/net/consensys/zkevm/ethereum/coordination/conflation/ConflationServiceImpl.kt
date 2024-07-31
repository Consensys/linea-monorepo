package net.consensys.zkevm.ethereum.coordination.conflation

import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.BlocksConflation
import net.consensys.zkevm.domain.ConflationCalculationResult
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.PriorityBlockingQueue
import java.util.concurrent.TimeUnit

class ConflationServiceImpl(
  private val calculator: TracesConflationCalculator,
  metricsFacade: MetricsFacade
) :
  ConflationService {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private var listener: ConflationHandler = ConflationHandler { SafeFuture.completedFuture<Unit>(null) }
  private val blocksInProgress: MutableList<ExecutionPayloadV1> = mutableListOf()

  data class PayloadAndBlockCounters(
    val executionPayload: ExecutionPayloadV1,
    val blockCounters: BlockCounters
  ) : Comparable<PayloadAndBlockCounters> {
    override fun compareTo(other: PayloadAndBlockCounters): Int {
      return this.executionPayload.blockNumber.compareTo(other.executionPayload.blockNumber)
    }
  }

  internal val blocksToConflate = PriorityBlockingQueue<PayloadAndBlockCounters>()

  private val blocksCounter = metricsFacade.createCounter(
    LineaMetricsCategory.CONFLATION,
    "blocks.imported",
    "New blocks arriving to conflation service counter"
  )

  init {
    metricsFacade.createGauge(
      category = LineaMetricsCategory.CONFLATION,
      name = "inprogress.blocks",
      description = "Number of blocks in progress of conflation",
      measurementSupplier = { blocksInProgress.size }
    )
    metricsFacade.createGauge(
      category = LineaMetricsCategory.CONFLATION,
      name = "queue.size",
      description = "Number of blocks in conflation queue",
      measurementSupplier = { blocksToConflate.size }
    )
    calculator.onConflatedBatch(this::handleConflation)
  }

  @Synchronized
  internal fun handleConflation(conflation: ConflationCalculationResult): SafeFuture<*> {
    log.debug(
      "new conflation: batch={} trigger={} tracesCounters={} blocksNumbers={}",
      conflation.intervalString(),
      conflation.conflationTrigger,
      conflation.tracesCounters,
      conflation.blocksRange.joinToString(",", "[", "]") { it.toString() }
    )
    val blocksToConflate =
      blocksInProgress
        .filter { it.blockNumber.toULong() in conflation.blocksRange }
        .sortedBy { it.blockNumber }
    blocksInProgress.removeAll(blocksToConflate)

    return listener.handleConflatedBatch(BlocksConflation(blocksToConflate, conflation))
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
    blocksCounter.increment()
    log.trace(
      "newBlock={} calculatorLastBlockNumber={} blocksToConflateSize={} blocksInProgressSize={}",
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
      nextAvailableBlock = blocksToConflate.poll(100, TimeUnit.MILLISECONDS)
      log.trace(
        "block {} removed from conflation queue and sent to calculator",
        nextAvailableBlock?.executionPayload?.blockNumber
      )
      calculator.newBlock(nextAvailableBlock.blockCounters)
      nextBlockNumberToConflate = nextAvailableBlock.executionPayload.blockNumber.toULong() + 1u
      nextAvailableBlock = blocksToConflate.peek()
    }
  }

  override fun onConflatedBatch(consumer: ConflationHandler) {
    this.listener = consumer
  }
}
