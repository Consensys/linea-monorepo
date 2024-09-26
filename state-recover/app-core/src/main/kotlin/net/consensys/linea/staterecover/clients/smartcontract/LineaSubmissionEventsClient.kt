package net.consensys.linea.staterecover.clients.smartcontract

import net.consensys.eth.EthLog
import net.consensys.eth.EthLogEvent
import net.consensys.toULong
import net.consensys.tuweni.bytes.sliceAsBytes32
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class DataSubmittedV3(
  val startBlockNumber: ULong,
  val endBlockNumber: ULong,
  val parentShnarf: Bytes32,
  val shnarf: Bytes32
) {
  constructor(
    startBlockNumber: ULong,
    endBlockNumber: ULong,
    parentShnarf: ByteArray,
    shnarf: ByteArray
  ) : this(
    startBlockNumber,
    endBlockNumber,
    Bytes32.wrap(parentShnarf),
    Bytes32.wrap(shnarf)
  )

  companion object {
    fun fromEthLog(ethLog: EthLog): EthLogEvent<DataSubmittedV3> {
      // DataSubmittedV3(
      //  uint256 indexed startBlock,
      //  uint256 indexed endBlock,
      //  bytes32 parentShnarf,
      //  bytes32 indexed shnarf)
      return EthLogEvent(
        event = DataSubmittedV3(
          startBlockNumber = UInt256.fromBytes(ethLog.topics[1]).toBigInteger().toULong(),
          endBlockNumber = UInt256.fromBytes(ethLog.topics[2]).toBigInteger().toULong(),
          parentShnarf = ethLog.data.toArray().sliceAsBytes32(sliceIndex = 0),
          shnarf = ethLog.topics[3]
        ),
        log = ethLog
      )
    }
  }
}

data class DataFinalizedV3(
  val startBlockNumber: ULong,
  val endBlockNumber: ULong,
  val snarf: Bytes32,
  val parentStateRootHash: Bytes32,
  val finalStateRootHash: Bytes32
) {
  companion object {
    fun fromEthLog(ethLog: EthLog): EthLogEvent<DataFinalizedV3> {
      /** event DataFinalizedV3(
       uint256 indexed startBlockNumber,
       uint256 indexed endBlockNumber,
       bytes32 indexed shnarf,
       bytes32 parentStateRootHash,
       bytes32 finalStateRootHash
       );*/
      val dataBytes = ethLog.data.toArray()
      return EthLogEvent(
        event = DataFinalizedV3(
          startBlockNumber = UInt256.fromBytes(ethLog.topics[1]).toBigInteger().toULong(),
          endBlockNumber = UInt256.fromBytes(ethLog.topics[2]).toBigInteger().toULong(),
          snarf = ethLog.topics[3],
          parentStateRootHash = dataBytes.sliceAsBytes32(sliceIndex = 0),
          finalStateRootHash = dataBytes.sliceAsBytes32(sliceIndex = 1)
        ),
        log = ethLog
      )
    }
  }
}

data class FinalizationAndDataEventsV3(
  val dataSubmittedEvents: List<EthLogEvent<DataSubmittedV3>>,
  val dataFinalizedEvent: EthLogEvent<DataFinalizedV3>
)

interface LineaRollupSubmissionEventsClient {
  fun findDataSubmittedV3EventsUtilNextFinalization(
    fromStartL2BlockNumberInclusive: ULong
  ): SafeFuture<FinalizationAndDataEventsV3?>
}
