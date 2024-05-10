package net.consensys.linea.contract

import net.consensys.toULong
import net.consensys.zkevm.coordinator.clients.DataSubmittedEvent
import net.consensys.zkevm.coordinator.clients.RollupSmartContractClient
import org.web3j.abi.EventEncoder
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.methods.request.EthFilter
import org.web3j.protocol.core.methods.response.Log
import tech.pegasys.teku.infrastructure.async.SafeFuture

class RollupSmartContractClientWeb3JImpl(
  val web3jLogsClient: Web3JLogsClient,
  val lineaRollup: LineaRollupAsyncFriendly
) : RollupSmartContractClient {
  override fun findLatestDataSubmittedEmittedEvent(
    startBlockNumberInclusive: Long,
    endBlockNumberInclusive: Long
  ): SafeFuture<DataSubmittedEvent?> {
    val ethFilter =
      EthFilter(
        DefaultBlockParameter.valueOf(startBlockNumberInclusive.toBigInteger()),
        DefaultBlockParameter.valueOf(endBlockNumberInclusive.toBigInteger()),
        lineaRollup.contractAddress
      )
    ethFilter.addSingleTopic(EventEncoder.encode(LineaRollup.DATASUBMITTED_EVENT))

    return web3jLogsClient.getLogs(ethFilter)
      .thenApply { logs ->
        if (logs.isNullOrEmpty()) {
          null
        } else {
          val log: Log = logs.last()
          val dataSubmittedEvent = LineaRollup.getDataSubmittedEventFromLog(log)
          DataSubmittedEvent(
            dataHash = dataSubmittedEvent.dataHash,
            startBlock = dataSubmittedEvent.startBlock.toULong(),
            endBlock = dataSubmittedEvent.endBlock.toULong()
          )
        }
      }
  }

  override fun getMessageRollingHash(messageNumber: Long): SafeFuture<ByteArray> {
    require(messageNumber >= 0) { "messageNumber must be greater than or equal to 0" }
    return SafeFuture.of(lineaRollup.rollingHashes(messageNumber.toBigInteger()).sendAsync())
  }
}
