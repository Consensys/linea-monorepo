package net.consensys.zkevm.ethereum.coordination.aggregation

import linea.domain.InvalidityProofIndex
import linea.persistence.ForcedTransactionsDao
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface InvalidityProofProvider {
  /**
   * Returns invalidity-proof indexes for the FTXs that fall inside an aggregation.
   *
   * @param ftxStartingNumber the lowest FTX number to include (inclusive); the FTXs strictly
   *  before this one are accounted for by the parent aggregation.
   * @param aggregationEndBlockNumber the highest L2 block number an FTX's
   *  `simulatedExecutionBlockNumber` may have to still be considered part of this aggregation
   *  (inclusive). FTXs whose simulated execution block falls beyond this end block belong to
   *  the next aggregation.
   */
  fun getInvalidityProofs(
    ftxStartingNumber: ULong,
    aggregationEndBlockNumber: ULong,
  ): SafeFuture<List<InvalidityProofIndex>>
}

class InvalidityProofProviderImpl(
  private val forcedTransactionsDao: ForcedTransactionsDao,
) : InvalidityProofProvider {
  override fun getInvalidityProofs(
    ftxStartingNumber: ULong,
    aggregationEndBlockNumber: ULong,
  ): SafeFuture<List<InvalidityProofIndex>> {
    return forcedTransactionsDao
      .findByStartingNumber(
        ftxStartingNumberInclusive = ftxStartingNumber,
        endSimulatedExecutionBlockNumberInclusive = aggregationEndBlockNumber,
      ).thenApply { forcedTransactions ->
        forcedTransactions.map { forcedTransaction ->
          InvalidityProofIndex(
            ftxNumber = forcedTransaction.ftxNumber,
            simulatedExecutionBlockNumber = forcedTransaction.simulatedExecutionBlockNumber,
            startBlockTimestamp = forcedTransaction.simulatedExecutionBlockTimestamp,
          )
        }
      }
  }
}
