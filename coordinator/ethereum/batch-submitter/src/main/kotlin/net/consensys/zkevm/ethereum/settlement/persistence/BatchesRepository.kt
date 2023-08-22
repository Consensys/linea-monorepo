package net.consensys.zkevm.ethereum.settlement.persistence

import kotlinx.datetime.Instant
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64

class DuplicatedBatchException(message: String? = null, cause: Throwable? = null) : Exception(message, cause)

interface BatchesRepository {
  fun saveNewBatch(batch: Batch): SafeFuture<Unit>

  fun getConsecutiveBatchesFromBlockNumber(
    startingBlockNumberInclusive: UInt64,
    endBlockCreatedBefore: Instant
  ): SafeFuture<List<Batch>>

  fun findHighestConsecutiveBatchByStatus(status: Batch.Status): SafeFuture<Batch?>
  fun findBatchWithHighestEndBlockNumberByStatus(status: Batch.Status): SafeFuture<Batch?>

  fun setBatchStatusUpToEndBlockNumber(
    endBlockNumberInclusive: UInt64,
    currentStatus: Batch.Status,
    newStatus: Batch.Status
  ): SafeFuture<Int>
}
