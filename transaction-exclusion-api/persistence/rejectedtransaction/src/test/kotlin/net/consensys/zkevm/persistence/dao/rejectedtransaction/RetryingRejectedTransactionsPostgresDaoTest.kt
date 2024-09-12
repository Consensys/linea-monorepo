package net.consensys.zkevm.persistence.dao.rejectedtransaction

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.decodeHex
import net.consensys.linea.transactionexclusion.ModuleOverflow
import net.consensys.linea.transactionexclusion.RejectedTransaction
import net.consensys.linea.transactionexclusion.TransactionInfo
import net.consensys.zkevm.persistence.db.PersistenceRetryer
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito.mock
import org.mockito.Mockito.times
import org.mockito.kotlin.eq
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.milliseconds

@ExtendWith(VertxExtension::class)
class RetryingRejectedTransactionsPostgresDaoTest {
  private lateinit var retryingRejectedTransactionsPostgresDao: RetryingRejectedTransactionsPostgresDao
  private val delegateRejectedTransactionsDao = mock<RejectedTransactionsPostgresDao>()
  private val now = Clock.System.now()
  private val rejectedTransaction = RejectedTransaction(
    txRejectionStage = RejectedTransaction.Stage.SEQUENCER,
    timestamp = Instant.parse("2024-08-31T09:18:51Z"),
    blockNumber = 10000UL,
    transactionRLP =
    (
      "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c4" +
        "96d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8"
      )
      .decodeHex(),
    reasonMessage = "Transaction line count for module ADD=402 is above the limit 70",
    overflows = listOf(
      ModuleOverflow(
        module = "ADD",
        count = 402,
        limit = 70
      )
    ),
    transactionInfo = TransactionInfo(
      hash = "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7".decodeHex(),
      from = "0x4d144d7b9c96b26361d6ac74dd1d8267edca4fc2".decodeHex(),
      to = "0x1195cf65f83b3a5768f3c496d3a05ad6412c64b3".decodeHex(),
      nonce = 100UL
    )
  )

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    retryingRejectedTransactionsPostgresDao = RetryingRejectedTransactionsPostgresDao(
      delegate = delegateRejectedTransactionsDao,
      PersistenceRetryer(
        vertx = vertx,
        PersistenceRetryer.Config(
          backoffDelay = 1.milliseconds
        )
      )
    )

    whenever(delegateRejectedTransactionsDao.saveNewRejectedTransaction(eq(rejectedTransaction)))
      .thenReturn(SafeFuture.completedFuture(Unit))

    whenever(
      delegateRejectedTransactionsDao.findRejectedTransactionByTxHash(eq(rejectedTransaction.transactionInfo!!.hash))
    )
      .thenReturn(SafeFuture.completedFuture(null))

    whenever(delegateRejectedTransactionsDao.deleteRejectedTransactionsBeforeTimestamp(eq(now)))
      .thenReturn(SafeFuture.completedFuture(0))
  }

  @Test
  fun `retrying rejected transactions dao should delegate all queries to standard dao`() {
    retryingRejectedTransactionsPostgresDao.saveNewRejectedTransaction(rejectedTransaction)
    verify(delegateRejectedTransactionsDao, times(1)).saveNewRejectedTransaction(eq(rejectedTransaction))

    retryingRejectedTransactionsPostgresDao.findRejectedTransactionByTxHash(
      rejectedTransaction.transactionInfo!!.hash
    )
    verify(delegateRejectedTransactionsDao, times(1)).findRejectedTransactionByTxHash(
      eq(
        rejectedTransaction.transactionInfo!!.hash
      )
    )

    retryingRejectedTransactionsPostgresDao.deleteRejectedTransactionsBeforeTimestamp(now)
    verify(delegateRejectedTransactionsDao, times(1)).deleteRejectedTransactionsBeforeTimestamp(now)
  }
}
