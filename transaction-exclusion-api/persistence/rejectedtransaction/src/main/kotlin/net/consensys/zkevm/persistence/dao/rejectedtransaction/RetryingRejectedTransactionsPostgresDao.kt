package net.consensys.zkevm.persistence.dao.rejectedtransaction

import kotlinx.datetime.Instant
import net.consensys.linea.transactionexclusion.RejectedTransaction
import net.consensys.zkevm.persistence.db.PersistenceRetryer
import tech.pegasys.teku.infrastructure.async.SafeFuture

class RetryingRejectedTransactionsPostgresDao(
  private val delegate: RejectedTransactionsPostgresDao,
  private val persistenceRetryer: PersistenceRetryer
) : RejectedTransactionsDao {
  override fun saveNewRejectedTransaction(rejectedTransaction: RejectedTransaction): SafeFuture<Unit> {
    return persistenceRetryer.retryQuery({ delegate.saveNewRejectedTransaction(rejectedTransaction) })
  }

  override fun findRejectedTransactionByTxHash(txHash: ByteArray): SafeFuture<RejectedTransaction?> {
    return persistenceRetryer.retryQuery({ delegate.findRejectedTransactionByTxHash(txHash) })
  }

  override fun deleteRejectedTransactionsBeforeTimestamp(timestamp: Instant): SafeFuture<Int> {
    return persistenceRetryer.retryQuery({ delegate.deleteRejectedTransactionsBeforeTimestamp(timestamp) })
  }
}
