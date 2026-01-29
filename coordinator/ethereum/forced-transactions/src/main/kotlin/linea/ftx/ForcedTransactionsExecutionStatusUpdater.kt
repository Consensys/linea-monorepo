package linea.ftx

import linea.contract.events.ForcedTransactionAddedEvent
import linea.forcedtx.ForcedTransactionsClient
import linea.persistence.ftx.ForcedTransactionsDao
import net.consensys.zkevm.domain.ForcedTransactionRecord
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ProcessedTransactionsFilter {
  fun isAlreadyProcessed(ftx: ForcedTransactionAddedEvent): SafeFuture<Boolean>
  fun filterOutAlreadyProcessed(
    ftxs: List<ForcedTransactionAddedEvent>,
  ): SafeFuture<List<ForcedTransactionAddedEvent>> {
    val unprocessed = ftxs.map {
      this.isAlreadyProcessed(it)
        .thenApply { alreadyProcessed ->
          if (alreadyProcessed) {
            null
          } else {
            it
          }
        }
    }

    return SafeFuture
      .collectAll(unprocessed.stream())
      .thenApply { it.filterNotNull() }
  }
}

/**
 * Responsible for getting Forced Transactions status from the sequencer and update local DB.
 */
class ForcedTransactionsStatusUpdater(
  val dao: ForcedTransactionsDao,
  val ftxClient: ForcedTransactionsClient,
) : ProcessedTransactionsFilter {
  override fun isAlreadyProcessed(ftx: ForcedTransactionAddedEvent): SafeFuture<Boolean> {
    // 1 check local db, if not present, check sequencer and update DB if processed by sequencer
    return dao
      .findByNumber(ftxNumber = ftx.forcedTransactionNumber)
      .thenCompose { dbRecord ->
        if (dbRecord != null) {
          SafeFuture.completedFuture(true)
        } else {
          checkStatusAndUpdateLocalDb(ftx)
        }
      }
  }

  private fun checkStatusAndUpdateLocalDb(ftx: ForcedTransactionAddedEvent): SafeFuture<Boolean> {
    return ftxClient
      .lineaFindForcedTransactionStatus(ftx.forcedTransactionNumber)
      .thenCompose { ftxStatus ->
        if (ftxStatus == null) {
          SafeFuture.completedFuture(false)
        } else {
          val record = ForcedTransactionRecord(
            ftxNumber = ftx.forcedTransactionNumber,
            inclusionResult = ftxStatus.inclusionResult,
            simulatedExecutionBlockNumber = ftxStatus.blockNumber,
            simulatedExecutionBlockTimestamp = ftxStatus.blockTimestamp,
          )
          dao
            .save(record)
            .thenApply { true }
        }
      }
  }
}
