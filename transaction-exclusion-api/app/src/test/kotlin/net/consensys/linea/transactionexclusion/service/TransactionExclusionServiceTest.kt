package net.consensys.linea.transactionexclusion.service

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import net.consensys.encodeHex
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.transactionexclusion.ErrorType
import net.consensys.linea.transactionexclusion.RejectedTransactionsRepository
import net.consensys.linea.transactionexclusion.TransactionExclusionError
import net.consensys.linea.transactionexclusion.TransactionExclusionServiceV1
import net.consensys.linea.transactionexclusion.defaultRejectedTransaction
import net.consensys.zkevm.persistence.db.DuplicatedRecordException
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class TransactionExclusionServiceTest {
  private val metricsFacadeMock = mock<MetricsFacade>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
  private lateinit var rejectedTransactionsRepositoryMock: RejectedTransactionsRepository

  @BeforeEach
  fun beforeEach() {
    rejectedTransactionsRepositoryMock = mock<RejectedTransactionsRepository>(
      defaultAnswer = Mockito.RETURNS_DEEP_STUBS
    )
  }

  @Test
  fun saveRejectedTransaction_return_success_result_with_saved_status() {
    whenever(rejectedTransactionsRepositoryMock.findRejectedTransaction(any()))
      .thenReturn(SafeFuture.completedFuture(defaultRejectedTransaction))
    whenever(rejectedTransactionsRepositoryMock.saveRejectedTransaction(any()))
      .thenReturn(SafeFuture.completedFuture(Unit))

    val transactionExclusionService = TransactionExclusionServiceV1Impl(
      repository = rejectedTransactionsRepositoryMock,
      metricsFacade = metricsFacadeMock
    )

    Assertions.assertEquals(
      Ok(TransactionExclusionServiceV1.SaveRejectedTransactionStatus.SAVED),
      transactionExclusionService.saveRejectedTransaction(defaultRejectedTransaction).get()
    )
  }

  @Test
  fun saveRejectedTransaction_return_success_result_with_duplicated_already_saved_status() {
    whenever(rejectedTransactionsRepositoryMock.findRejectedTransaction(any()))
      .thenReturn(SafeFuture.completedFuture(defaultRejectedTransaction))
    whenever(rejectedTransactionsRepositoryMock.saveRejectedTransaction(any()))
      .thenReturn(SafeFuture.failedFuture(DuplicatedRecordException()))

    val transactionExclusionService = TransactionExclusionServiceV1Impl(
      repository = rejectedTransactionsRepositoryMock,
      metricsFacade = metricsFacadeMock
    )

    Assertions.assertEquals(
      Ok(TransactionExclusionServiceV1.SaveRejectedTransactionStatus.DUPLICATE_ALREADY_SAVED_BEFORE),
      transactionExclusionService.saveRejectedTransaction(defaultRejectedTransaction).get()
    )
  }

  @Test
  fun saveRejectedTransaction_return_error_result() {
    whenever(rejectedTransactionsRepositoryMock.findRejectedTransaction(any()))
      .thenReturn(SafeFuture.completedFuture(defaultRejectedTransaction))
    whenever(rejectedTransactionsRepositoryMock.saveRejectedTransaction(any()))
      .thenReturn(SafeFuture.failedFuture(RuntimeException()))

    val transactionExclusionService = TransactionExclusionServiceV1Impl(
      repository = rejectedTransactionsRepositoryMock,
      metricsFacade = metricsFacadeMock
    )

    Assertions.assertEquals(
      Err(TransactionExclusionError(ErrorType.OTHER_ERROR, "")),
      transactionExclusionService.saveRejectedTransaction(defaultRejectedTransaction).get()
    )
  }

  @Test
  fun getTransactionExclusionStatus_return_success_result_with_rejected_txn() {
    whenever(rejectedTransactionsRepositoryMock.findRejectedTransaction(any()))
      .thenReturn(SafeFuture.completedFuture(defaultRejectedTransaction))

    val transactionExclusionService = TransactionExclusionServiceV1Impl(
      repository = rejectedTransactionsRepositoryMock,
      metricsFacade = metricsFacadeMock
    )

    Assertions.assertEquals(
      Ok(defaultRejectedTransaction),
      transactionExclusionService.getTransactionExclusionStatus(
        defaultRejectedTransaction.transactionInfo!!.hash
      ).get()
    )
  }

  @Test
  fun getTransactionExclusionStatus_return_error_result_with_transaction_unavailable() {
    whenever(rejectedTransactionsRepositoryMock.findRejectedTransaction(any()))
      .thenReturn(SafeFuture.completedFuture(null))

    val transactionExclusionService = TransactionExclusionServiceV1Impl(
      repository = rejectedTransactionsRepositoryMock,
      metricsFacade = metricsFacadeMock
    )

    Assertions.assertEquals(
      Err(
        TransactionExclusionError(
          ErrorType.TRANSACTION_UNAVAILABLE,
          "Cannot find the rejected transaction with hash=" +
            defaultRejectedTransaction.transactionInfo!!.hash.encodeHex()
        )
      ),
      transactionExclusionService.getTransactionExclusionStatus(
        defaultRejectedTransaction.transactionInfo!!.hash
      ).get()
    )
  }

  @Test
  fun getTransactionExclusionStatus_return_error_result_with_other_error() {
    whenever(rejectedTransactionsRepositoryMock.findRejectedTransaction(any()))
      .thenReturn(SafeFuture.failedFuture(RuntimeException()))

    val transactionExclusionService = TransactionExclusionServiceV1Impl(
      repository = rejectedTransactionsRepositoryMock,
      metricsFacade = metricsFacadeMock
    )

    Assertions.assertEquals(
      Err(
        TransactionExclusionError(
          ErrorType.OTHER_ERROR,
          ""
        )
      ),
      transactionExclusionService.getTransactionExclusionStatus(
        defaultRejectedTransaction.transactionInfo!!.hash
      ).get()
    )
  }
}
