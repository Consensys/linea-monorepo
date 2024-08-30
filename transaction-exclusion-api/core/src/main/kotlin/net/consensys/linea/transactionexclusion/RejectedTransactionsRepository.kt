package net.consensys.linea.transactionexclusion

import kotlinx.datetime.Instant
import net.consensys.linea.RejectedTransaction
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface RejectedTransactionsRepository {
  fun findRejectedTransaction(txHash: ByteArray): SafeFuture<RejectedTransaction?>
  fun saveRejectedTransaction(rejectedTransaction: RejectedTransaction): SafeFuture<Unit>
  fun deleteRejectedTransaction(timestamp: Instant): SafeFuture<Int>
}
