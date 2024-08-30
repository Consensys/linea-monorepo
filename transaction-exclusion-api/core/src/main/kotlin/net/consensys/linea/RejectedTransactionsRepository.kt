package net.consensys.linea

import kotlinx.datetime.Instant
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface RejectedTransactionsRepository {
  fun findRejectedTransaction(txHash: ByteArray): SafeFuture<RejectedTransaction?>
  fun saveRejectedTransaction(rejectedTransaction: RejectedTransaction): SafeFuture<Unit>
  fun deleteRejectedTransaction(timestamp: Instant): SafeFuture<Int>
}
