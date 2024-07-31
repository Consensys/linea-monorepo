package net.consensys.zkevm.domain

fun createBatch(
  startBlockNumber: Long,
  endBlockNumber: Long,
  status: Batch.Status = Batch.Status.Proven
): Batch {
  return Batch(
    startBlockNumber = startBlockNumber.toULong(),
    endBlockNumber = endBlockNumber.toULong(),
    status = status
  )
}
