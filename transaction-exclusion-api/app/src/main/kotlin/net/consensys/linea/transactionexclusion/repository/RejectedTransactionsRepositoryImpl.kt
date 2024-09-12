package net.consensys.linea.transactionexclusion.repository

import kotlinx.datetime.Instant
import net.consensys.linea.transactionexclusion.RejectedTransaction
import net.consensys.linea.transactionexclusion.RejectedTransactionsRepository
import net.consensys.zkevm.persistence.dao.rejectedtransaction.RejectedTransactionsDao
import tech.pegasys.teku.infrastructure.async.SafeFuture

class RejectedTransactionsRepositoryImpl(
  private val rejectedTransactionsDao: RejectedTransactionsDao
) : RejectedTransactionsRepository {
  override fun findRejectedTransaction(txHash: ByteArray): SafeFuture<RejectedTransaction?> {
    return rejectedTransactionsDao.findRejectedTransactionByTxHash(txHash)
  }

  override fun saveRejectedTransaction(rejectedTransaction: RejectedTransaction): SafeFuture<Unit> {
    return rejectedTransactionsDao.saveNewRejectedTransaction(rejectedTransaction)
  }

  override fun deleteRejectedTransaction(timestamp: Instant): SafeFuture<Int> {
    return rejectedTransactionsDao.deleteRejectedTransactionsBeforeTimestamp(timestamp)
  }
}
