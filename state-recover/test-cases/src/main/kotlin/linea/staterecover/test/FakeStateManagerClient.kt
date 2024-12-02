package linea.staterecover.test

import build.linea.clients.GetZkEVMStateMerkleProofResponse
import build.linea.clients.StateManagerClientV1
import build.linea.clients.StateManagerErrorType
import build.linea.domain.BlockInterval
import com.fasterxml.jackson.databind.node.ArrayNode
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import linea.EthLogsSearcher
import linea.staterecover.DataFinalizedV3
import net.consensys.linea.BlockParameter
import net.consensys.linea.errors.ErrorResponse
import net.consensys.toHexStringUInt256
import net.consensys.zkevm.domain.BlobRecord
import tech.pegasys.teku.infrastructure.async.SafeFuture

class FakeStateManagerClient(
  val blobRecords: List<BlobRecord>,
  var headBlockNumber: ULong = blobRecords.last().endBlockNumber
) : StateManagerClientV1 {

  override fun rollupGetHeadBlockNumber(): SafeFuture<ULong> {
    return SafeFuture.completedFuture(headBlockNumber)
  }

  override fun rollupGetStateMerkleProofWithTypedError(
    blockInterval: BlockInterval
  ): SafeFuture<Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<StateManagerErrorType>>> {
    val parentStateRootHash = blobRecords.find { it.startBlockNumber == blockInterval.startBlockNumber }
      ?.blobCompressionProof?.parentStateRootHash
    val finalStateRootHash = blobRecords.find { it.endBlockNumber == blockInterval.endBlockNumber }
      ?.blobCompressionProof?.finalStateRootHash
    return when {
      parentStateRootHash == null ->
        SafeFuture
          .failedFuture(RuntimeException("Blob record not found for block: ${blockInterval.startBlockNumber}"))

      finalStateRootHash == null ->
        SafeFuture
          .failedFuture(RuntimeException("Blob record not found for block: ${blockInterval.endBlockNumber}"))

      else ->
        return SafeFuture.completedFuture(
          Ok(
            GetZkEVMStateMerkleProofResponse(
              zkStateMerkleProof = ArrayNode(null),
              zkParentStateRootHash = parentStateRootHash,
              zkEndStateRootHash = finalStateRootHash,
              zkStateManagerVersion = "fake-version"
            )
          )
        )
    }
  }
}

class FakeStateManagerClientReadFromL1(
  val headBlockNumber: ULong,
  val logsSearcher: EthLogsSearcher,
  val contracAddress: String
) : StateManagerClientV1 {

  override fun rollupGetHeadBlockNumber(): SafeFuture<ULong> {
    return SafeFuture.completedFuture(headBlockNumber)
  }

  override fun rollupGetStateMerkleProofWithTypedError(
    blockInterval: BlockInterval
  ): SafeFuture<Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<StateManagerErrorType>>> {
    return logsSearcher
      .getLogs(
        fromBlock = BlockParameter.Tag.EARLIEST,
        toBlock = BlockParameter.Tag.FINALIZED,
        address = contracAddress,
        topics = listOf(
          DataFinalizedV3.topic,
          blockInterval.startBlockNumber.toHexStringUInt256(),
          blockInterval.endBlockNumber.toHexStringUInt256()
        )
      )
      .thenCompose { logs ->
        if (logs.isEmpty()) {
          SafeFuture.failedFuture(RuntimeException("No logs found for block interval: $blockInterval"))
        } else {
          val logEvent = DataFinalizedV3.fromEthLog(logs.first())
          SafeFuture.completedFuture(
            Ok(
              GetZkEVMStateMerkleProofResponse(
                zkStateMerkleProof = ArrayNode(null),
                zkParentStateRootHash = logEvent.event.parentStateRootHash,
                zkEndStateRootHash = logEvent.event.finalStateRootHash,
                zkStateManagerVersion = "fake-version"
              )
            )
          )
        }
      }
  }
}
