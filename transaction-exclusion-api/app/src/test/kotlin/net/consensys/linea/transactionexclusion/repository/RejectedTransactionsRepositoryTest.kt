package net.consensys.linea.transactionexclusion.repository

import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.linea.transactionexclusion.test.defaultRejectedTransaction
import net.consensys.zkevm.persistence.dao.rejectedtransaction.RejectedTransactionsPostgresDao
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class RejectedTransactionsRepositoryTest {
  private val rejectedTransactionsDaoMock = mock<RejectedTransactionsPostgresDao>(
    defaultAnswer = Mockito.RETURNS_DEEP_STUBS
  ).also {
    whenever(it.findRejectedTransactionByTxHash(any(), any()))
      .thenReturn(SafeFuture.completedFuture(defaultRejectedTransaction))
    whenever(it.saveNewRejectedTransaction(any()))
      .thenReturn(SafeFuture.completedFuture(Unit))
    whenever(it.deleteRejectedTransactions(any()))
      .thenReturn(SafeFuture.completedFuture(1))
  }
  private val rejectedTransactionsRepository = RejectedTransactionsRepositoryImpl(
    rejectedTransactionsDao = rejectedTransactionsDaoMock
  )

  @Test
  fun findRejectedTransaction_should_return_rejected_transaction() {
    Assertions.assertEquals(
      defaultRejectedTransaction,
      rejectedTransactionsRepository.findRejectedTransaction(
        defaultRejectedTransaction.transactionInfo!!.hash,
        Instant.DISTANT_PAST
      ).get()
    )
  }

  @Test
  fun saveRejectedTransaction_should_return_unit() {
    Assertions.assertEquals(
      Unit,
      rejectedTransactionsRepository.saveRejectedTransaction(
        defaultRejectedTransaction
      ).get()
    )
  }

  @Test
  fun deleteRejectedTransaction_should_return_one() {
    Assertions.assertEquals(
      1,
      rejectedTransactionsRepository.deleteRejectedTransactions(
        Clock.System.now()
      ).get()
    )
  }
}
