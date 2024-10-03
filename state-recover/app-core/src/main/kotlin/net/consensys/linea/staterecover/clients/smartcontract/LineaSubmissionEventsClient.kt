package net.consensys.linea.staterecover.clients.smartcontract

import net.consensys.eth.EthLog
import net.consensys.eth.EthLogEvent
import net.consensys.linea.BlockInterval
import net.consensys.tuweni.bytes.sliceAsBytes32
import net.consensys.tuweni.bytes.toULong
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture

data class DataSubmittedV3(
  override val startBlockNumber: ULong,
  override val endBlockNumber: ULong,
  val parentShnarf: Bytes32,
  val shnarf: Bytes32,
  val finalStateRootHash: Bytes32
) : BlockInterval {
  constructor(
    startBlockNumber: ULong,
    endBlockNumber: ULong,
    parentShnarf: ByteArray,
    shnarf: ByteArray,
    finalStateRootHash: Bytes32
  ) : this(
    startBlockNumber,
    endBlockNumber,
    Bytes32.wrap(parentShnarf),
    Bytes32.wrap(shnarf),
    Bytes32.wrap(finalStateRootHash)
  )

  companion object {
    fun fromEthLog(ethLog: EthLog): EthLogEvent<DataSubmittedV3> {
      // DataSubmittedV3(
      //  uint256 indexed startBlock,
      //  uint256 indexed endBlock,
      //  bytes32 parentShnarf,
      //  bytes32 indexed shnarf)
      val dataBytes = ethLog.data.toArray()
      return EthLogEvent(
        event = DataSubmittedV3(
          startBlockNumber = ethLog.topics[1].toULong(),
          endBlockNumber = ethLog.topics[2].toULong(),
          parentShnarf = dataBytes.sliceAsBytes32(sliceIndex = 0),
          shnarf = ethLog.topics[3],
          finalStateRootHash = dataBytes.sliceAsBytes32(sliceIndex = 1)
        ),
        log = ethLog
      )
    }
  }
}

data class DataFinalizedV3(
  override val startBlockNumber: ULong,
  override val endBlockNumber: ULong,
  val snarf: Bytes32,
  val parentStateRootHash: Bytes32,
  val finalStateRootHash: Bytes32
) : BlockInterval {
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
          startBlockNumber = ethLog.topics[1].toULong(),
          endBlockNumber = ethLog.topics[2].toULong(),
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
