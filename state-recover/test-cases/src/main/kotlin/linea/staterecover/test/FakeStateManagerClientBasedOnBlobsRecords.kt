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

open class FakeStateManagerClient(
  private val blocksStateRootHashes: MutableMap<ULong, ByteArray> = mutableMapOf<ULong, ByteArray>(),
  var headBlockNumber: ULong = blocksStateRootHashes.keys.maxOrNull() ?: 0UL
) : StateManagerClientV1 {

  fun setBlockStateRootHash(blockNumber: ULong, stateRootHash: ByteArray) {
    blocksStateRootHashes[blockNumber] = stateRootHash
    headBlockNumber = blocksStateRootHashes.keys.maxOrNull() ?: 0UL
  }

  override fun rollupGetHeadBlockNumber(): SafeFuture<ULong> {
    return SafeFuture.completedFuture(headBlockNumber)
  }

  override fun rollupGetStateMerkleProofWithTypedError(
    blockInterval: BlockInterval
  ): SafeFuture<Result<GetZkEVMStateMerkleProofResponse, ErrorResponse<StateManagerErrorType>>> {
    // For state recovery, we just need the endStateRootHash
    val blockZkStateRootHash = blocksStateRootHashes[blockInterval.endBlockNumber]
    return when {
      blockZkStateRootHash == null ->
        SafeFuture
          .failedFuture(
            RuntimeException(
              "SateRootHash found for block=${blockInterval.endBlockNumber} headBlockNumber=$headBlockNumber"
            )
          )

      else ->
        return SafeFuture.completedFuture(
          Ok(
            GetZkEVMStateMerkleProofResponse(
              zkStateMerkleProof = ArrayNode(null),
              zkParentStateRootHash = ByteArray(32),
              zkEndStateRootHash = blockZkStateRootHash,
              zkStateManagerVersion = "fake-version"
            )
          )
        )
    }
  }
}

class FakeStateManagerClientBasedOnBlobsRecords(
  val blobRecords: List<BlobRecord>
) : FakeStateManagerClient(
  blocksStateRootHashes = blobRecords
    .associate { it.endBlockNumber to it.blobCompressionProof!!.finalStateRootHash }.toMutableMap()
)

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
          null,
//          blockInterval.startBlockNumber.toHexStringUInt256()
          blockInterval.endBlockNumber.toHexStringUInt256()
        )
      )
      .thenCompose { logs ->
        if (logs.isEmpty()) {
          SafeFuture.failedFuture(RuntimeException("No logs found for l2 blocks=$blockInterval"))
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
