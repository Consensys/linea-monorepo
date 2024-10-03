package net.consensys.linea.staterecover.clients.smartcontract

import net.consensys.encodeHex
import net.consensys.eth.EthLog
import net.consensys.eth.EthLogEvent
import net.consensys.linea.BlockParameter
import net.consensys.linea.contract.LineaRollupV6
import net.consensys.linea.contract.Web3JLogsClient
import net.consensys.linea.contract.l1.toWeb3j
import net.consensys.linea.web3j.domainmappers.toDomain
import net.consensys.toBigInteger
import org.apache.tuweni.units.bigints.UInt256
import org.web3j.abi.EventEncoder
import org.web3j.protocol.core.methods.request.EthFilter
import org.web3j.protocol.core.methods.response.Log
import tech.pegasys.teku.infrastructure.async.SafeFuture

class LineaSubmissionEventsClientWeb3jIpml(
  private val logsClient: Web3JLogsClient,
  private val smartContractAddress: String
) : LineaRollupSubmissionEventsClient {

  override fun findDataFinalizedEventContainingBlock(blockNumber: ULong): SafeFuture<EthLogEvent<DataFinalizedV3>?> {
    TODO("Not yet implemented")
  }

  override fun findDataFinalizedEventByStartBlockNumber(blockNumber: ULong): SafeFuture<EthLogEvent<DataFinalizedV3>?> {
    return findDataFinalizedV3Event(
      fromL1BlockNumber = BlockParameter.Tag.EARLIEST,
      toL1BlockNumber = BlockParameter.Tag.FINALIZED,
      startBlockNumber = blockNumber
    )
  }

  override fun findDataSubmittedV3EventsUtilNextFinalization(
    l2StartBlockNumberInclusive: ULong
  ): SafeFuture<FinalizationAndDataEventsV3?> {
    return findDataFinalizedV3Event(startBlockNumber = l2StartBlockNumberInclusive)
      .thenCompose { finalizationEvent ->
        finalizationEvent
          ?.let {
            findAggregationDataSubmittedV3Events(it)
              .thenApply { dataSubmittedEvents -> FinalizationAndDataEventsV3(dataSubmittedEvents, it) }
          }
          ?: SafeFuture.completedFuture(null)
      }
  }

  private fun findDataFinalizedV3Event(
    fromL1BlockNumber: BlockParameter = BlockParameter.Tag.EARLIEST,
    toL1BlockNumber: BlockParameter = BlockParameter.Tag.LATEST,
    startBlockNumber: ULong? = null,
    endBlockNumber: ULong? = null
  ): SafeFuture<EthLogEvent<DataFinalizedV3>?> {
    assert(startBlockNumber != null || endBlockNumber != null) {
      "Either startBlockNumber or endBlockNumber must be provided"
    }

    val ethFilter =
      EthFilter(
        fromL1BlockNumber.toWeb3j(),
        toL1BlockNumber.toWeb3j(),
        smartContractAddress
      ).apply {
        /**
         event DataFinalizedV3(
         uint256 indexed startBlockNumber,
         uint256 indexed endBlockNumber,
         bytes32 indexed shnarf,
         bytes32 parentStateRootHash,
         bytes32 finalStateRootHash
         );
         */
        addSingleTopic(EventEncoder.encode(LineaRollupV6.DATAFINALIZEDV3_EVENT))
        addSingleTopic(startBlockNumber?.let { UInt256.valueOf(it.toBigInteger()).toHexString() })
        addSingleTopic(endBlockNumber?.let { UInt256.valueOf(it.toBigInteger()).toHexString() })
      }

    return logsClient
      .getLogs(ethFilter)
      .thenApply(Companion::parseDataFinalizedV2)
      .thenCompose { finalizedEvents ->
        if (finalizedEvents.size > 1) {
          // just a safety check
          // this should never happen: Finalization events shall be sequential and deterministic
          val errorMessage =
            "More than one DataFinalizedV3 event found for startBlockNumber=$startBlockNumber events=$finalizedEvents"
          SafeFuture.failedFuture(IllegalStateException(errorMessage))
        } else {
          SafeFuture.completedFuture(finalizedEvents.firstOrNull())
        }
      }
  }

  private fun findAggregationDataSubmittedV3Events(
    finalizationEvent: EthLogEvent<DataFinalizedV3>
  ): SafeFuture<List<EthLogEvent<DataSubmittedV3>>> {
    val dataEvents = mutableListOf<EthLogEvent<DataSubmittedV3>>()
    val futureResult = SafeFuture<List<EthLogEvent<DataSubmittedV3>>>()
    fun fetchParentDataSubmission(dataSubmission: EthLogEvent<DataSubmittedV3>) {
      dataEvents.add(dataSubmission)
      if (dataSubmission.event.startBlockNumber == finalizationEvent.event.startBlockNumber) {
        futureResult.complete(dataEvents.sortedBy { it.event.startBlockNumber })
      } else {
        getDataSubmittedV3EventByShnarf(
          fromL1BlockParameter = BlockParameter.Tag.EARLIEST,
          tol1BlockParameter = BlockParameter.fromNumber(dataSubmission.log.blockNumber.toLong()),
          shnarf = dataSubmission.event.parentShnarf.toArray()
        ).thenPeek(::fetchParentDataSubmission)
      }
    }

    getDataSubmittedV3EventByShnarf(
      fromL1BlockParameter = BlockParameter.Tag.EARLIEST,
      tol1BlockParameter = BlockParameter.fromNumber(finalizationEvent.log.blockNumber.toLong()),
      shnarf = finalizationEvent.event.snarf.toArray()
    ).thenPeek(::fetchParentDataSubmission)

    return futureResult
  }

  private fun getDataSubmittedV3EventByShnarf(
    fromL1BlockParameter: BlockParameter = BlockParameter.Tag.EARLIEST,
    tol1BlockParameter: BlockParameter = BlockParameter.Tag.LATEST,
    shnarf: ByteArray
  ): SafeFuture<EthLogEvent<DataSubmittedV3>> {
    return findDataSubmittedV3EventByShnarf(fromL1BlockParameter, tol1BlockParameter, shnarf)
      .thenApply { event ->
        event ?: throw IllegalStateException("DataSubmittedV3 event not found for shnarf=${shnarf.encodeHex()}")
      }
  }

  private fun findDataSubmittedV3EventByShnarf(
    fromL1BlockParameter: BlockParameter = BlockParameter.Tag.EARLIEST,
    tol1BlockParameter: BlockParameter = BlockParameter.Tag.LATEST,
    shnarf: ByteArray
  ): SafeFuture<EthLogEvent<DataSubmittedV3>?> {
    val ethFilter =
      EthFilter(
        fromL1BlockParameter.toWeb3j(),
        tol1BlockParameter.toWeb3j(),
        smartContractAddress
      ).apply {
        // DataSubmittedV3(uint256 indexed startBlock, uint256 indexed endBlock, bytes32 parentShnarf, bytes32 indexed shnarf)
        addSingleTopic(EventEncoder.encode(LineaRollupV6.DATASUBMITTEDV3_EVENT))
        addSingleTopic(null) // startBlockNumber
        addSingleTopic(null) // endBlockNumber
        addSingleTopic(shnarf.encodeHex()) // shnarf
      }

    return logsClient
      .getLogs(ethFilter)
      .thenApply(Companion::parseDataSubmittedV3)
      .thenApply { events ->
        if (events.size > 1) {
          // just a safety check
          // this should never happen: having more than blob with the same shnarf
          val errorMessage =
            "More than one DataSubmittedV3 event found with shnarf=${shnarf.encodeHex()} events=$events"
          throw IllegalStateException(errorMessage)
        } else {
          events.firstOrNull()
        }
      }
  }

  companion object {
    private fun <T> parseEvents(parser: (EthLog) -> EthLogEvent<T>): (List<Log>) -> List<EthLogEvent<T>> {
      return { logs -> logs.map { web3jLog -> parser(web3jLog.toDomain()) } }
    }

    private fun parseDataSubmittedV3(logs: List<Log>): List<EthLogEvent<DataSubmittedV3>> {
      return logs.map { log -> DataSubmittedV3.fromEthLog(log.toDomain()) }
    }

    private fun parseDataFinalizedV2(logs: List<Log>): List<EthLogEvent<DataFinalizedV3>> {
      return logs.map { log -> DataFinalizedV3.fromEthLog(log.toDomain()) }
    }
  }
}
