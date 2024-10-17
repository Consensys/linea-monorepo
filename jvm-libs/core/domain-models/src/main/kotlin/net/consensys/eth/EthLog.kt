package net.consensys.eth

import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32

data class EthLog(
  val removed: Boolean,
  val logIndex: ULong,
  val transactionIndex: ULong,
  val transactionHash: Bytes32,
  val blockHash: Bytes32,
  val blockNumber: ULong,
  val address: Bytes,
  val data: Bytes,
  val topics: List<Bytes32>
) {
  /**
   * Just a handy constructor
   */
  constructor(
    removed: Boolean,
    logIndex: ULong,
    transactionIndex: ULong,
    transactionHash: ByteArray,
    blockHash: ByteArray,
    blockNumber: ULong,
    address: ByteArray,
    data: ByteArray,
    topics: List<ByteArray>
  ) : this(
    removed,
    logIndex,
    transactionIndex,
    Bytes32.wrap(transactionHash),
    Bytes32.wrap(blockHash),
    blockNumber,
    Bytes.wrap(address),
    Bytes.wrap(data),
    topics.map { Bytes32.wrap(it) }
  )
}

data class EthLogEvent<E>(
  val event: E,
  val log: EthLog
)
