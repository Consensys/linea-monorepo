package net.consensys.zkevm.persistence.dao.blob

import linea.domain.BlobRecord
import net.consensys.zkevm.persistence.db.PersistenceRetryer
import net.consensys.zkevm.persistence.db.RetryingDaoBase
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Instant

class RetryingBlobsPostgresDao(
  delegate: BlobsPostgresDao,
  persistenceRetryer: PersistenceRetryer,
) : RetryingDaoBase<BlobsPostgresDao>(delegate, persistenceRetryer), BlobsDao {
  override fun saveNewBlob(blobRecord: BlobRecord): SafeFuture<Unit> =
    retrying { delegate.saveNewBlob(blobRecord) }

  override fun getConsecutiveBlobsFromBlockNumber(
    startingBlockNumberInclusive: ULong,
    endBlockCreatedBefore: Instant,
  ): SafeFuture<List<BlobRecord>> =
    retrying { delegate.getConsecutiveBlobsFromBlockNumber(startingBlockNumberInclusive, endBlockCreatedBefore) }

  override fun findBlobByStartBlockNumber(startBlockNumber: ULong): SafeFuture<BlobRecord?> =
    retrying { delegate.findBlobByStartBlockNumber(startBlockNumber) }

  override fun findBlobByEndBlockNumber(endBlockNumber: ULong): SafeFuture<BlobRecord?> =
    retrying { delegate.findBlobByEndBlockNumber(endBlockNumber) }

  override fun deleteBlobsUpToEndBlockNumber(endBlockNumberInclusive: ULong): SafeFuture<Int> =
    retrying { delegate.deleteBlobsUpToEndBlockNumber(endBlockNumberInclusive) }

  override fun deleteBlobsAfterBlockNumber(startingBlockNumberInclusive: ULong): SafeFuture<Int> =
    retrying { delegate.deleteBlobsAfterBlockNumber(startingBlockNumberInclusive) }
}
