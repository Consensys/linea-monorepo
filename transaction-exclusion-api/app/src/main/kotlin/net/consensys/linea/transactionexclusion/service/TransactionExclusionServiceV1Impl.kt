package net.consensys.linea.transactionexclusion.service

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import kotlinx.datetime.Clock
import net.consensys.linea.metrics.LineaMetricsCategory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.transactionexclusion.ErrorType
import net.consensys.linea.transactionexclusion.RejectedTransaction
import net.consensys.linea.transactionexclusion.RejectedTransactionsRepository
import net.consensys.linea.transactionexclusion.TransactionExclusionError
import net.consensys.linea.transactionexclusion.TransactionExclusionServiceV1
import net.consensys.linea.transactionexclusion.TransactionExclusionServiceV1.SaveRejectedTransactionStatus
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration

class TransactionExclusionServiceV1Impl(
  private val config: Config,
  private val repository: RejectedTransactionsRepository,
  metricsFacade: MetricsFacade,
  private val clock: Clock = Clock.System
) : TransactionExclusionServiceV1 {
  data class Config(
    val rejectedTimestampWithinDuration: Duration
  )

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
    return this.repository.findRejectedTransaction(
      txHash = rejectedTransaction.transactionInfo!!.hash,
      notRejectedBefore = clock.now().minus(config.rejectedTimestampWithinDuration)
    )
      .thenPeek {
        if (it == null) {
          txRejectionCounter.increment()
        }
      }
      .thenCompose {
        this.repository.saveRejectedTransaction(rejectedTransaction)
      }
      .handleComposed { _, error ->
        if (error != null) {
          if (error.cause is DuplicatedRecordException) {
            SafeFuture.completedFuture(
              Ok(SaveRejectedTransactionStatus.DUPLICATE_ALREADY_SAVED_BEFORE)
            )
          } else {
            SafeFuture.completedFuture(
              Err(TransactionExclusionError(ErrorType.SERVER_ERROR, error.cause?.message ?: ""))
            )
          }
        } else {
          SafeFuture.completedFuture(Ok(SaveRejectedTransactionStatus.SAVED))
        }
      }
  }

  override fun getTransactionExclusionStatus(
    txHash: ByteArray
  ): SafeFuture<Result<RejectedTransaction?, TransactionExclusionError>> {
    return this.repository.findRejectedTransaction(
      txHash = txHash,
      notRejectedBefore = clock.now().minus(config.rejectedTimestampWithinDuration)
    )
      .handleComposed { result, error ->
        if (error != null) {
          SafeFuture.completedFuture(
            Err(TransactionExclusionError(ErrorType.SERVER_ERROR, error.message ?: ""))
          )
        } else {
          SafeFuture.completedFuture(Ok(result))
        }
      }
  }
}
