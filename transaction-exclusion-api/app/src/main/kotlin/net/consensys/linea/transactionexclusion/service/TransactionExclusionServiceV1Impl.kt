package net.consensys.linea.transactionexclusion.service

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import net.consensys.encodeHex
import net.consensys.linea.RejectedTransaction
import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.transactionexclusion.ErrorType
import net.consensys.linea.transactionexclusion.RejectedTransactionsRepository
import net.consensys.linea.transactionexclusion.TransactionExclusionError
import net.consensys.linea.transactionexclusion.TransactionExclusionServiceV1
import net.consensys.linea.transactionexclusion.TransactionExclusionServiceV1.SaveRejectedTransactionStatus
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

class TransactionExclusionServiceV1Impl(
  private val repository: RejectedTransactionsRepository,
  metricsFacade: MetricsFacade
) : TransactionExclusionServiceV1 {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private val txRejectionCounter = metricsFacade.createCounter(
    LineaMetricsCategory.TX_EXCLUSION_API,
    "transactions.rejected",
    "Counter of rejected transactions reported to Transaction Exclusion API service"
  )

  override fun saveRejectedTransaction(
    rejectedTransaction: RejectedTransaction
  ): SafeFuture<
    Result<SaveRejectedTransactionStatus, TransactionExclusionError>
    > {
    return this.repository.findRejectedTransaction(rejectedTransaction.transactionInfo!!.hash)
      .thenApply {
        if (it == null) {
          txRejectionCounter.increment()
        }
      }
      .thenCompose {
        this.repository.saveRejectedTransaction(rejectedTransaction)
          .handleComposed { _, error ->
            if (error == null) {
              SafeFuture.completedFuture(Ok(SaveRejectedTransactionStatus.SAVED))
            } else {
              if (error is DuplicatedRecordException) {
                SafeFuture.completedFuture(
                  Ok(SaveRejectedTransactionStatus.DUPLICATE_ALREADY_SAVED_BEFORE)
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
