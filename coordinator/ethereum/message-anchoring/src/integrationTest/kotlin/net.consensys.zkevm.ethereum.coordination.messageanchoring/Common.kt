package net.consensys.zkevm.ethereum.coordination.messageanchoring

import org.apache.tuweni.bytes.Bytes32
import java.math.BigInteger

fun createRandomSendMessageEvents(numberOfRandomHashes: ULong): List<SendMessageEvent> {
  return (0UL..numberOfRandomHashes)
    .map { n ->
      SendMessageEvent(
        Bytes32.random(),
        messageNumber = n + 1UL,
        blockNumber = n + 1UL
      )
    }
}

data class L1MessageToSend(
  val recipient: String,
  val fee: BigInteger,
  val calldata: ByteArray,
  val value: BigInteger
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as L1MessageToSend

    if (recipient != other.recipient) return false
    if (fee != other.fee) return false
    if (!calldata.contentEquals(other.calldata)) return false
    return value == other.value
  }

  override fun hashCode(): Int {
    var result = recipient.hashCode()
    result = 31 * result + fee.hashCode()
    result = 31 * result + calldata.contentHashCode()
    result = 31 * result + value.hashCode()
    return result
  }
}
