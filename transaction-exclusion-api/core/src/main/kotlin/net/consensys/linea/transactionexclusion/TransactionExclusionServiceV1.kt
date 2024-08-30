package net.consensys.linea.transactionexclusion

import com.github.michaelbull.result.Result
import net.consensys.linea.RejectedTransaction
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface TransactionExclusionServiceV1 {
  fun saveRejectedTransaction(
    rejectedTransaction: RejectedTransaction
  ): SafeFuture<Result<RejectedTransaction, TransactionExclusionError>>

  fun getTransactionExclusionStatus(
    txHash: ByteArray
  ): SafeFuture<Result<RejectedTransaction, TransactionExclusionError>>
}
