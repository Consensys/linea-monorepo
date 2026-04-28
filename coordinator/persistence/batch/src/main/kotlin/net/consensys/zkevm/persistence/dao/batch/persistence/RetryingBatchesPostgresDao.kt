package net.consensys.zkevm.persistence.dao.batch.persistence

import linea.domain.Batch
import net.consensys.zkevm.persistence.db.PersistenceRetryer
import net.consensys.zkevm.persistence.db.RetryingDaoBase
import tech.pegasys.teku.infrastructure.async.SafeFuture

class RetryingBatchesPostgresDao(
  delegate: BatchesPostgresDao,
  persistenceRetryer: PersistenceRetryer,
) : RetryingDaoBase<BatchesPostgresDao>(delegate, persistenceRetryer), BatchesDao {
  override fun saveNewBatch(batch: Batch): SafeFuture<Unit> =
    retrying { delegate.saveNewBatch(batch) }

  override fun findHighestConsecutiveEndBlockNumberFromBlockNumber(
    startingBlockNumberInclusive: Long,
  ): SafeFuture<Long?> =
    retrying { delegate.findHighestConsecutiveEndBlockNumberFromBlockNumber(startingBlockNumberInclusive) }

  override fun deleteBatchesUpToEndBlockNumber(endBlockNumberInclusive: Long): SafeFuture<Int> =
    retrying { delegate.deleteBatchesUpToEndBlockNumber(endBlockNumberInclusive) }

  override fun deleteBatchesAfterBlockNumber(startingBlockNumberInclusive: Long): SafeFuture<Int> =
    retrying { delegate.deleteBatchesAfterBlockNumber(startingBlockNumberInclusive) }
}
