package build.linea.staterecover.clients

import build.linea.domain.BlockInterval
import build.linea.domain.EthLog
import build.linea.domain.EthLogEvent
import net.consensys.encodeHex
import net.consensys.sliceOf32
import net.consensys.toULongFromLast8Bytes
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class DataSubmittedV3(
  val parentShnarf: ByteArray,
  val shnarf: ByteArray,
  val finalStateRootHash: ByteArray
) {
  companion object {
    fun fromEthLog(ethLog: EthLog): EthLogEvent<DataSubmittedV3> {
      // DataSubmittedV3(bytes32 parentShnarf, bytes32 indexed shnarf, bytes32 finalStateRootHash);
      return EthLogEvent(
        event = DataSubmittedV3(
          parentShnarf = ethLog.data.sliceOf32(0),
          shnarf = ethLog.topics[1],
          finalStateRootHash = ethLog.data.sliceOf32(1)
        ),
        log = ethLog
      )
    }
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as DataSubmittedV3

    if (!parentShnarf.contentEquals(other.parentShnarf)) return false
    if (!shnarf.contentEquals(other.shnarf)) return false
    if (!finalStateRootHash.contentEquals(other.finalStateRootHash)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = parentShnarf.contentHashCode()
    result = 31 * result + shnarf.contentHashCode()
    result = 31 * result + finalStateRootHash.contentHashCode()
    return result
  }

  override fun toString(): String {
    return "DataSubmittedV3(" +
      "parentShnarf=${parentShnarf.encodeHex()}," +
      " shnarf=${shnarf.encodeHex()}," +
      " finalStateRootHash=${finalStateRootHash.encodeHex()})"
  }
}

data class DataFinalizedV3(
  override val startBlockNumber: ULong,
  override val endBlockNumber: ULong,
  val snarf: ByteArray,
  val parentStateRootHash: ByteArray,
  val finalStateRootHash: ByteArray
) : BlockInterval {
  companion object {
    fun fromEthLog(ethLog: EthLog): EthLogEvent<DataFinalizedV3> {
      /**event DataFinalizedV3(
       uint256 indexed startBlockNumber,
       uint256 indexed endBlockNumber,
       bytes32 indexed shnarf,
       bytes32 parentStateRootHash,
       bytes32 finalStateRootHash
       );*/
      val dataBytes = ethLog.data
      return EthLogEvent(
        event = DataFinalizedV3(
          startBlockNumber = ethLog.topics[1].toULongFromLast8Bytes(),
          endBlockNumber = ethLog.topics[2].toULongFromLast8Bytes(),
          snarf = ethLog.topics[3],
          parentStateRootHash = dataBytes.sliceOf32(sliceNumber = 0),
          finalStateRootHash = dataBytes.sliceOf32(sliceNumber = 1)
        ),
        log = ethLog
      )
    }
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as DataFinalizedV3

    if (startBlockNumber != other.startBlockNumber) return false
    if (endBlockNumber != other.endBlockNumber) return false
    if (!snarf.contentEquals(other.snarf)) return false
    if (!parentStateRootHash.contentEquals(other.parentStateRootHash)) return false
    if (!finalStateRootHash.contentEquals(other.finalStateRootHash)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = startBlockNumber.hashCode()
    result = 31 * result + endBlockNumber.hashCode()
    result = 31 * result + snarf.contentHashCode()
    result = 31 * result + parentStateRootHash.contentHashCode()
    result = 31 * result + finalStateRootHash.contentHashCode()
    return result
  }

  override fun toString(): String {
    return "DataFinalizedV3(" +
      "startBlockNumber=$startBlockNumber, " +
      "endBlockNumber=$endBlockNumber, " +
      "snarf=${snarf.encodeHex()}, " +
      "parentStateRootHash=${parentStateRootHash.encodeHex()}, " +
      "finalStateRootHash=${finalStateRootHash.encodeHex()})"
  }
}

data class FinalizationAndDataEventsV3(
  val dataSubmittedEvents: List<EthLogEvent<DataSubmittedV3>>,
  val dataFinalizedEvent: EthLogEvent<DataFinalizedV3>
)

interface LineaRollupSubmissionEventsClient {
  fun findDataFinalizedEventByStartBlockNumber(
    blockNumber: ULong
  ): SafeFuture<EthLogEvent<DataFinalizedV3>?>

//  fun findDataFinalizedEventByEndBlockNumber(
//    blockNumber: ULong
//  ): SafeFuture<EthLogEvent<DataFinalizedV3>?>

  fun findDataFinalizedEventContainingBlock(
    blockNumber: ULong
  ): SafeFuture<EthLogEvent<DataFinalizedV3>?>

  fun findDataSubmittedV3EventsUtilNextFinalization(
    l2StartBlockNumberInclusive: ULong
  ): SafeFuture<FinalizationAndDataEventsV3?>
}
