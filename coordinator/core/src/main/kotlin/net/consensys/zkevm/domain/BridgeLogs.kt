package net.consensys.zkevm.domain

data class RlpBridgeLogsData(val rlp: String, val bridgeLogs: List<BridgeLogsData>)

data class BridgeLogsData(
  val removed: Boolean,
  val logIndex: String,
  val transactionIndex: String,
  val transactionHash: String,
  val blockHash: String,
  val blockNumber: String,
  val address: String,
  val data: String,
  val topics: List<String>
)

data class L2RollingHashUpdatedEvent(val messageNumber: ULong, val messageRollingHash: ByteArray) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as L2RollingHashUpdatedEvent

    if (messageNumber != other.messageNumber) return false
    if (!messageRollingHash.contentEquals(other.messageRollingHash)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = messageNumber.hashCode()
    result = 31 * result + messageRollingHash.contentHashCode()
    return result
  }
}
