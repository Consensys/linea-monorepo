package linea.contract.events

import linea.domain.EthLog
import linea.domain.EthLogEvent
import linea.kotlin.sliceOf32
import linea.kotlin.toULongFromLast8Bytes
import kotlin.time.Instant

data class FinalizedStateUpdatedEvent(
  val blockNumber: ULong,
  val timestamp: Instant,
  val messageNumber: ULong,
  val forcedTransactionNumber: ULong,
) {
  companion object {
    const val topic = "0x32e016ccc5c33419c35caa94023fdeb75143da613fb2ac738ab736404c09fc5d"
    fun fromEthLog(ethLog: EthLog): EthLogEvent<FinalizedStateUpdatedEvent> {
      /**event FinalizedStateUpdated(
       uint256 indexed blockNumber,
       uint256 timestamp,
       uint256 messageNumber,
       uint256 forcedTransactionNumber
       );*/
      val dataBytes = ethLog.data
      return EthLogEvent(
        event = FinalizedStateUpdatedEvent(
          blockNumber = ethLog.topics[1].toULongFromLast8Bytes(),
          timestamp = Instant.fromEpochSeconds(dataBytes.sliceOf32(sliceNumber = 0).toULongFromLast8Bytes().toLong()),
          messageNumber = dataBytes.sliceOf32(sliceNumber = 1).toULongFromLast8Bytes(),
          forcedTransactionNumber = dataBytes.sliceOf32(sliceNumber = 2).toULongFromLast8Bytes(),
        ),
        log = ethLog,
      )
    }
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as FinalizedStateUpdatedEvent

    if (blockNumber != other.blockNumber) return false
    if (timestamp != other.timestamp) return false
    if (messageNumber != other.messageNumber) return false
    if (forcedTransactionNumber != other.forcedTransactionNumber) return false

    return true
  }

  override fun hashCode(): Int {
    var result = blockNumber.hashCode()
    result = 31 * result + timestamp.hashCode()
    result = 31 * result + messageNumber.hashCode()
    result = 31 * result + forcedTransactionNumber.hashCode()
    return result
  }

  override fun toString(): String {
    return "FinalizedStateUpdatedEvent(" +
      "blockNumber=$blockNumber, " +
      "timestamp=$timestamp, " +
      "messageNumber=$messageNumber, " +
      "forcedTransactionNumber=$forcedTransactionNumber)"
  }
}
