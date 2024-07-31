package net.consensys.zkevm.persistence.dao.blob

import kotlinx.datetime.Instant
import net.consensys.zkevm.coordinator.clients.BlobCompressionProof
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.BlobStatus
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface BlobsDao {
  fun saveNewBlob(blobRecord: BlobRecord): SafeFuture<Unit>

  fun getConsecutiveBlobsFromBlockNumber(
    startingBlockNumberInclusive: ULong,
    endBlockCreatedBefore: Instant
  ): SafeFuture<List<BlobRecord>>

  fun findBlobByStartBlockNumber(
    startBlockNumber: ULong
  ): SafeFuture<BlobRecord?>

  fun findBlobByEndBlockNumber(
    endBlockNumber: ULong
  ): SafeFuture<BlobRecord?>

  fun updateBlob(
    startingBlockNumber: ULong,
    endBlockNumber: ULong,
    blobStatus: BlobStatus,
    blobCompressionProof: BlobCompressionProof
  ): SafeFuture<Int>

  fun deleteBlobsUpToEndBlockNumber(
    endBlockNumberInclusive: ULong
  ): SafeFuture<Int>

  fun deleteBlobsAfterBlockNumber(
    startingBlockNumberInclusive: ULong
  ): SafeFuture<Int>
}
