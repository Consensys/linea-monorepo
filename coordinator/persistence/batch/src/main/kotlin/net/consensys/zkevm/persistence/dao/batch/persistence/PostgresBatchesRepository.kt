package net.consensys.zkevm.persistence.dao.batch.persistence

import net.consensys.zkevm.domain.Batch
import net.consensys.zkevm.persistence.BatchesRepository
import tech.pegasys.teku.infrastructure.async.SafeFuture

/**
 * WARNING: Existing mappings should not chane. Otherwise, can break production New One can be added
 * though.
 */
fun batchStatusToDbValue(status: Batch.Status): Int {
  // using manual mapping to catch errors at compile time instead of runtime
  return when (status) {
    Batch.Status.Finalized -> 1
    Batch.Status.Proven -> 2
  }
}

class PostgresBatchesRepository(
  private val batchesDao: BatchesDao
) : BatchesRepository {
  override fun saveNewBatch(batch: Batch): SafeFuture<Unit> {
    return batchesDao.saveNewBatch(batch)
  }

  override fun findHighestConsecutiveEndBlockNumberFromBlockNumber(
    startingBlockNumberInclusive: Long
  ): SafeFuture<Long?> {
    return batchesDao.findHighestConsecutiveEndBlockNumberFromBlockNumber(startingBlockNumberInclusive)
  }

  override fun deleteBatchesUpToEndBlockNumber(
    endBlockNumberInclusive: Long
  ): SafeFuture<Int> {
    return batchesDao.deleteBatchesUpToEndBlockNumber(endBlockNumberInclusive)
  }

  override fun deleteBatchesAfterBlockNumber(
    startingBlockNumberInclusive: Long
  ): SafeFuture<Int> {
    return batchesDao.deleteBatchesAfterBlockNumber(startingBlockNumberInclusive)
  }
}
