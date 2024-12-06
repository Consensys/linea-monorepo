package build.linea.domain

import net.consensys.encodeHex

data class EthLog(
  val removed: Boolean,
  val logIndex: ULong,
  val transactionIndex: ULong,
  val transactionHash: ByteArray,
  val blockHash: ByteArray,
  val blockNumber: ULong,
  val address: ByteArray,
  val data: ByteArray,
  val topics: List<ByteArray>
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as EthLog

    if (removed != other.removed) return false
    if (logIndex != other.logIndex) return false
    if (transactionIndex != other.transactionIndex) return false
    if (!transactionHash.contentEquals(other.transactionHash)) return false
    if (!blockHash.contentEquals(other.blockHash)) return false
    if (blockNumber != other.blockNumber) return false
    if (!address.contentEquals(other.address)) return false
    if (!data.contentEquals(other.data)) return false
    if (topics != other.topics) return false

    return true
  }

  override fun hashCode(): Int {
    var result = removed.hashCode()
    result = 31 * result + logIndex.hashCode()
    result = 31 * result + transactionIndex.hashCode()
    result = 31 * result + transactionHash.contentHashCode()
    result = 31 * result + blockHash.contentHashCode()
    result = 31 * result + blockNumber.hashCode()
    result = 31 * result + address.contentHashCode()
    result = 31 * result + data.contentHashCode()
    result = 31 * result + topics.hashCode()
    return result
  }

  override fun toString(): String {
    return "EthLog(" +
      "removed=$removed, " +
      "logIndex=$logIndex, " +
      "transactionIndex=$transactionIndex, " +
      "transactionHash=${transactionHash.encodeHex()}, " +
      "blockHash=${blockHash.encodeHex()}, " +
      "blockNumber=$blockNumber, " +
      "address=${address.encodeHex()}, " +
      "data=${data.encodeHex()}, " +
      "topics=$topics)"
  }
}

data class EthLogEvent<E>(
  val event: E,
  val log: EthLog
)
