package net.consensys.linea.transactionexclusion

import kotlinx.datetime.Instant
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface RejectedTransactionsRepository {
  fun findRejectedTransaction(txHash: ByteArray, notRejectedBefore: Instant): SafeFuture<RejectedTransaction?>
  fun saveRejectedTransaction(rejectedTransaction: RejectedTransaction): SafeFuture<Unit>
  fun deleteRejectedTransactions(createdBefore: Instant): SafeFuture<Int>
}
