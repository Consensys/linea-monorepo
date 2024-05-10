package net.consensys.zkevm.coordinator.clients

import tech.pegasys.teku.infrastructure.async.SafeFuture

data class DataSubmittedEvent(
  val dataHash: ByteArray,
  val startBlock: ULong,
  val endBlock: ULong
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as DataSubmittedEvent

    if (!dataHash.contentEquals(other.dataHash)) return false
    if (startBlock != other.startBlock) return false
    if (endBlock != other.endBlock) return false

    return true
  }

  override fun hashCode(): Int {
    var result = dataHash.contentHashCode()
    result = 31 * result + startBlock.hashCode()
    result = 31 * result + endBlock.hashCode()
    return result
  }
}

interface RollupSmartContractClient {
  // using Long instead of ULong because is impossible stub interfaces with ULong/UInt
  fun findLatestDataSubmittedEmittedEvent(
    startBlockNumberInclusive: Long,
    endBlockNumberInclusive: Long
  ): SafeFuture<DataSubmittedEvent?>

  fun getMessageRollingHash(messageNumber: Long): SafeFuture<ByteArray>
}
