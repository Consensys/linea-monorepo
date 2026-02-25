package linea.persistence.ftx

import net.consensys.zkevm.domain.ForcedTransactionRecord
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface ForcedTransactionsDao {
  fun save(ftx: ForcedTransactionRecord): SafeFuture<Unit>
  fun findByNumber(ftxNumber: ULong): SafeFuture<ForcedTransactionRecord?>
  fun list(): SafeFuture<List<ForcedTransactionRecord>>
  fun deleteFtxUpToInclusive(ftxNumber: ULong): SafeFuture<Int>
  fun findHighestForcedTransaction(
    upToSimulatedExecutionBlockNumberInclusive: ULong? = null,
  ): SafeFuture<ForcedTransactionRecord?> {
    return list()
      .thenApply { allFtxs ->
        allFtxs
          .let { ftxs ->
            upToSimulatedExecutionBlockNumberInclusive
              ?.let { endBlock -> ftxs.filter { ftx -> ftx.simulatedExecutionBlockNumber <= endBlock } }
              ?: ftxs
          }
          .maxByOrNull { ftx -> ftx.ftxNumber }
      }
  }

  fun findByStartingNumber(
    ftxStartingNumberInclusive: ULong,
    endSimulatedExecutionBlockNumberInclusive: ULong,
  ): SafeFuture<List<ForcedTransactionRecord>> {
    return list().thenApply { allFtx ->
      allFtx.filter { ftx ->
        ftx.ftxNumber >= ftxStartingNumberInclusive &&
          ftx.simulatedExecutionBlockNumber <= endSimulatedExecutionBlockNumberInclusive
      }
    }
  }
}
