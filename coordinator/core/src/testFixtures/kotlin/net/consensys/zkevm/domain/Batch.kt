package net.consensys.zkevm.domain

import linea.domain.Batch

fun createBatch(startBlockNumber: Long, endBlockNumber: Long): Batch {
  return Batch(
    startBlockNumber = startBlockNumber.toULong(),
    endBlockNumber = endBlockNumber.toULong(),
  )
}
