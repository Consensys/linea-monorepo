package linea.persistence.conflation

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import linea.domain.createBlobRecord
import linea.kotlin.trimToSecondPrecision
import net.consensys.FakeFixedClock
import net.consensys.zkevm.persistence.db.PersistenceRetryer
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito.mock
import org.mockito.kotlin.any
import org.mockito.kotlin.eq
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.milliseconds

@ExtendWith(VertxExtension::class)
class RetryingBlobsPostgresDaoTest {
  private lateinit var retryingBatchesPostgresDao: RetryingBlobsPostgresDao
  private val delegateBlobsDao = mock<BlobsPostgresDao>()

  private val fakeClock = FakeFixedClock()
  private val now = fakeClock.now().trimToSecondPrecision()
  private val blobRecord = createBlobRecord(
    startBlockNumber = 0U,
    endBlockNumber = 10U,
    startBlockTime = now,
  )

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    retryingBatchesPostgresDao = RetryingBlobsPostgresDao(
      delegate = delegateBlobsDao,
      PersistenceRetryer(
        vertx = vertx,
        PersistenceRetryer.Config(
          backoffDelay = 1.milliseconds,
        ),
      ),
    )

    whenever(delegateBlobsDao.saveNewBlob(eq(blobRecord)))
      .thenReturn(SafeFuture.completedFuture(Unit))

    whenever(delegateBlobsDao.getConsecutiveBlobsFromBlockNumber(any(), eq(now)))
      .thenReturn(SafeFuture.completedFuture(listOf(blobRecord)))

    whenever(delegateBlobsDao.findBlobByStartBlockNumber(any()))
      .thenReturn(SafeFuture.completedFuture(blobRecord))

    whenever(delegateBlobsDao.findBlobByEndBlockNumber(any()))
      .thenReturn(SafeFuture.completedFuture(blobRecord))

    whenever(delegateBlobsDao.deleteBlobsUpToEndBlockNumber(any()))
      .thenReturn(SafeFuture.completedFuture(1))

    whenever(delegateBlobsDao.deleteBlobsAfterBlockNumber(any()))
      .thenReturn(SafeFuture.completedFuture(1))
  }

  @Test
  fun `retrying batches dao should delegate all queries to standard dao`() {
    retryingBatchesPostgresDao.saveNewBlob(blobRecord)
    verify(delegateBlobsDao, times(1)).saveNewBlob(eq(blobRecord))

    retryingBatchesPostgresDao.getConsecutiveBlobsFromBlockNumber(0U, now)
    verify(delegateBlobsDao, times(1)).getConsecutiveBlobsFromBlockNumber(any(), eq(now))

    retryingBatchesPostgresDao.findBlobByStartBlockNumber(0U)
    verify(delegateBlobsDao, times(1)).findBlobByStartBlockNumber(any())

    retryingBatchesPostgresDao.findBlobByEndBlockNumber(10U)
    verify(delegateBlobsDao, times(1)).findBlobByEndBlockNumber(any())

    retryingBatchesPostgresDao.deleteBlobsUpToEndBlockNumber(10U)
    verify(delegateBlobsDao, times(1)).deleteBlobsUpToEndBlockNumber(any())

    retryingBatchesPostgresDao.deleteBlobsAfterBlockNumber(0U)
    verify(delegateBlobsDao, times(1)).deleteBlobsAfterBlockNumber(any())
  }
}
