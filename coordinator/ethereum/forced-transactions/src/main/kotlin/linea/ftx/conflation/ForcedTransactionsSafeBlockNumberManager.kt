package linea.ftx.conflation

import linea.contract.events.ForcedTransactionAddedEvent
import net.consensys.zkevm.ethereum.coordination.blockcreation.ConflationSafeBlockNumberProvider
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import kotlin.concurrent.atomics.ExperimentalAtomicApi

/**
 * Provides Safe Block Number (SBN) based on Forced Transactions in flight.
 *
 * SBN = min(simulatedExecutionBlockNumber) - 1 across all FTX records.
 * Returns null when no FTX records exist (unrestricted).
 *
 * Caches the result with periodic refresh to avoid frequent database queries.
 */

@OptIn(ExperimentalAtomicApi::class)
internal class ForcedTransactionsSafeBlockNumberManager : ConflationSafeBlockNumberProvider {
  private val log: Logger = LogManager.getLogger(ForcedTransactionsSafeBlockNumberManager::class.java)
  private var safeBlockNumber: ULong? = 0UL
  private var startUpScanFinished: Boolean = false
  private var firstUnprocessedFtxEventFound: Boolean = false
  private val ftxInSequencerForProcessing: MutableList<ULong> = mutableListOf()

  @Synchronized
  override fun getHighestSafeBlockNumber(): ULong? = safeBlockNumber

  @Synchronized
  fun ftxProcessedBySequencer(ftxNumber: ULong, simulatedExecutionBlockNumber: ULong) {
    log.info(
      "locking conflation: ftxNumber={} at blockNumber={}",
      ftxNumber,
      simulatedExecutionBlockNumber,
    )
    safeBlockNumber = simulatedExecutionBlockNumber

    this.ftxInSequencerForProcessing.removeIf { it <= ftxNumber }
    if (ftxInSequencerForProcessing.isEmpty() && startUpScanFinished) {
      log.info("all ftx sent to sequencer processed, releasing Safe Block Number lock")
      safeBlockNumber = null
    } else {
      safeBlockNumber = simulatedExecutionBlockNumber
      log.info(
        "fxt={} processed at safeBlockNumber={}, ftx in the sequencer {}",
        ftxNumber,
        safeBlockNumber,
        ftxInSequencerForProcessing,
      )
    }
  }

  @Synchronized
  fun ftxSentToSequencer(ftx: List<ForcedTransactionAddedEvent>) {
    this.ftxInSequencerForProcessing.addAll(ftx.map { it.forcedTransactionNumber })
  }

  @Synchronized
  fun lockSafeBlockNumberBeforeSendingToSequencer(headBlockNumber: ULong) {
    if (!startUpScanFinished) {
      log.info(
        "waiting l1 forced transactions scan from start up to finish" +
          " before moving safe block number forward, current safeBlockNumber={}",
        safeBlockNumber,
      )
      return
    }

    if (safeBlockNumber != null && safeBlockNumber!! > 0UL) {
      log.info("conflation already locked at safeBlockNumber={}", safeBlockNumber)
      return
    }

    log.info("locking conflation at latest l2 blockNumber={}", headBlockNumber)
    safeBlockNumber = headBlockNumber
  }

  @Synchronized
  fun unprocessedFtxQueueIsEmpty() {
    // not safe to release lock until we finish start up L1 scan
    if (!startUpScanFinished) {
      return
    }
    log.info("releasing Safe Block Number lock")
    safeBlockNumber = null
  }

  @Synchronized
  fun firstFtxEventFound() {
    firstUnprocessedFtxEventFound = true
  }

  /**
   * Releases Safe Block Number lock after startup
   * it's idempotent
   */
  @Synchronized
  fun caughtUpWithChainHeadAfterStartUp() {
    if (startUpScanFinished) {
      return
    }
    startUpScanFinished = true
    if (this.ftxInSequencerForProcessing.isNotEmpty()) {
      return
    }
    if (safeBlockNumber == 0UL) {
      // still locked at 0 from start, no events on L1, release lock
      log.info("releasing Safe Block Number lock after startup")
      safeBlockNumber = null
    }
  }
}
