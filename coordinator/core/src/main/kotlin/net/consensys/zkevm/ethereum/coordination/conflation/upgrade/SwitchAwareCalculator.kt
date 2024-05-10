package net.consensys.zkevm.ethereum.coordination.conflation.upgrade

import net.consensys.zkevm.domain.BlockCounters
import net.consensys.zkevm.domain.ConflationCalculationResult
import net.consensys.zkevm.ethereum.coordination.conflation.BlobCreationHandler
import net.consensys.zkevm.ethereum.coordination.conflation.NOOP_BLOB_HANDLER
import net.consensys.zkevm.ethereum.coordination.conflation.NOOP_CONSUMER
import net.consensys.zkevm.ethereum.coordination.conflation.TracesConflationCalculator
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.math.max

@Deprecated("We may use it for 4844 switch, but the switch procedure is yet unknown")
class SwitchAwareCalculator(
  private val oldCalculator: TracesConflationCalculator,
  private val newCalculatorProvider: (lastBlockNumber: ULong) -> TracesConflationCalculator,
  private val switchBlockNumber: ULong?
) : TracesConflationCalculator {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private lateinit var newCalculator: TracesConflationCalculator
  private var switchBlock: ULong? = null
  override val lastBlockNumber: ULong
    get() = if (this::newCalculator.isInitialized && newCalculator.lastBlockNumber > switchBlock!!) {
      newCalculator.lastBlockNumber
    } else {
      oldCalculator.lastBlockNumber
    }
  private var conflationConsumer: ((ConflationCalculationResult) -> SafeFuture<*>)? = NOOP_CONSUMER
  private var blobHandler: BlobCreationHandler? = NOOP_BLOB_HANDLER

  private fun initializeNewCalculator(blockCounters: BlockCounters) {
    if (!this::newCalculator.isInitialized &&
      switchBlockNumber != null
    ) {
      switchBlock = switchBlockNumber
      log.debug("Initializing new calculator on block {}", switchBlock)
      newCalculator = newCalculatorProvider(
        max(
          blockCounters.blockNumber - 1UL,
          switchBlockNumber - 1UL
        )
      )
      if (conflationConsumer != NOOP_CONSUMER) {
        newCalculator.onConflatedBatch(conflationConsumer!!)
      }
      if (blobHandler != NOOP_BLOB_HANDLER) {
        newCalculator.onBlobCreation(blobHandler!!)
      }
    }
  }

  override fun newBlock(blockCounters: BlockCounters) {
    initializeNewCalculator(blockCounters)
    log.debug(
      "New block {} with switch block number {}",
      blockCounters.blockNumber,
      switchBlockNumber
    )
    when {
      switchBlockNumber == null -> {
        log.debug("Block number {} without switch event. Sending to old calculator.", blockCounters.blockNumber)
        oldCalculator.newBlock(blockCounters)
      }

      blockCounters.blockNumber < switchBlockNumber -> {
        log.debug(
          "Block number {} before switch block number {} found. Sending to old calculator",
          blockCounters.blockNumber,
          switchBlockNumber
        )
        oldCalculator.newBlock(blockCounters)
      }

      blockCounters.blockNumber == switchBlockNumber -> {
        log.debug(
          "Switch block number {} found. Sending new block to both new and old calculator",
          blockCounters.blockNumber
        )
        oldCalculator.newBlock(blockCounters) // To create conflation with end block number M - 1
        newCalculator.newBlock(blockCounters) // To create conflation with start block number M
      }

      blockCounters.blockNumber > switchBlockNumber -> {
        log.debug(
          "Block number {} after the switch event on block {}." +
            " Sending to new calculator.",
          blockCounters.blockNumber,
          switchBlockNumber
        )
        newCalculator.newBlock(blockCounters)
      }
      else -> {
        log.error(
          "Unexpected case! {}, {}, {}",
          blockCounters.blockNumber,
          switchBlockNumber,
          blockCounters
        )
        throw IllegalStateException("This is unexpected!")
      }
    }
  }

  override fun onConflatedBatch(conflationConsumer: (ConflationCalculationResult) -> SafeFuture<*>) {
    oldCalculator.onConflatedBatch(conflationConsumer)
    this.conflationConsumer = conflationConsumer
  }

  override fun onBlobCreation(blobHandler: BlobCreationHandler) {
    this.blobHandler = blobHandler
  }
}
