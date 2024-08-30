package net.consensys.linea.transactionexclusion.service

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import net.consensys.encodeHex
import net.consensys.linea.ErrorType
import net.consensys.linea.RejectedTransaction
import net.consensys.linea.RejectedTransactionsRepository
import net.consensys.linea.TransactionExclusionError
import net.consensys.linea.TransactionExclusionServiceV1
import net.consensys.linea.TransactionInfo
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.core.encoding.TransactionDecoder
import tech.pegasys.teku.infrastructure.async.SafeFuture

class TransactionExclusionServiceV1Impl(
  private val repository: RejectedTransactionsRepository
) : TransactionExclusionServiceV1 {
  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun saveRejectedTransaction(
    rejectedTransaction: RejectedTransaction
  ): SafeFuture<Result<RejectedTransaction, TransactionExclusionError>> {
    val rejectedTransactionWithTxInfo = rejectedTransaction.also {
      TransactionDecoder.decodeOpaqueBytes(
        Bytes.wrap(it.transactionRLP)
      ).run {
        it.transactionInfo = TransactionInfo(
          hash = this.hash.toArray(),
          to = this.to.get().toArray(),
          from = this.sender.toArray(),
          nonce = this.nonce.toULong()
        )
      }
    }

    return this.repository.saveRejectedTransaction(rejectedTransactionWithTxInfo)
      .handleComposed { _, error ->
        if (error == null) {
          SafeFuture.completedFuture(Ok(rejectedTransactionWithTxInfo))
        } else {
          if (error is DuplicatedRecordException) {
            SafeFuture.completedFuture(
              Err(
                TransactionExclusionError(
                  ErrorType.TRANSACTION_DUPLICATED,
                  error.message!!
                )
              )
            )
          } else {
            SafeFuture.completedFuture(
              Err(
                TransactionExclusionError(
                  ErrorType.OTHER_ERROR,
                  error.message ?: ""
                )
              )
            )
          }
        }
      }
  }

  override fun getTransactionExclusionStatus(
    txHash: ByteArray
  ): SafeFuture<Result<RejectedTransaction, TransactionExclusionError>> {
    return this.repository.findRejectedTransaction(txHash).thenApply {
      if (it == null) {
        Err(
          TransactionExclusionError(
            ErrorType.TRANSACTION_UNAVAILABLE,
            "Cannot find the rejected transaction with hash=${txHash.encodeHex()}"
          )
        )
      } else {
        Ok(it)
      }
    }
  }
}
