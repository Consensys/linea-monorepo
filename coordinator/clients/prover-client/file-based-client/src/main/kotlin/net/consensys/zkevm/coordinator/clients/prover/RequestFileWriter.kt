package net.consensys.zkevm.coordinator.clients.prover

import com.fasterxml.jackson.databind.ObjectMapper
import io.vertx.core.Vertx
import net.consensys.encodeHex
import net.consensys.linea.async.toSafeFuture
import net.consensys.zkevm.coordinator.clients.GenerateTracesResponse
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import net.consensys.zkevm.domain.BridgeLogsData
import net.consensys.zkevm.domain.ProofIndex
import net.consensys.zkevm.domain.RlpBridgeLogsData
import net.consensys.zkevm.encoding.ExecutionPayloadV1Encoder
import net.consensys.zkevm.encoding.ExecutionPayloadV1RLPEncoderByBesuImplementation
import net.consensys.zkevm.fileio.FileWriter
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import java.util.concurrent.Callable

internal class RequestFileWriter(
  private val vertx: Vertx,
  private val proofRequestFileNameProvider: ProverFileNameProvider,
  private val config: Config,
  private val mapper: ObjectMapper,
  private val log: Logger,
  private val fileWriter: FileWriter = FileWriter(vertx, mapper),
  private val executionPayloadV1Encoder: ExecutionPayloadV1Encoder = ExecutionPayloadV1RLPEncoderByBesuImplementation
) {

  data class Config(
    val requestDirectory: Path,
    val writingInprogressSuffix: String,
    val proverInprogressSuffixPattern: String
  )

  fun write(
    blocksAndLogs: List<Pair<ExecutionPayloadV1, List<BridgeLogsData>>>,
    tracesResponse: GenerateTracesResponse,
    type2StateData: GetZkEVMStateMerkleProofResponse,
    keccakPreviousStateRootHash: String
  ): SafeFuture<Path> {
    val startBlockNumber = blocksAndLogs.first().first.blockNumber.toULong()
    val endBlockNumber = blocksAndLogs.last().first.blockNumber.toULong()
    val requestFilePath = config.requestDirectory
      .resolve(
        proofRequestFileNameProvider.getFileName(
          ProofIndex(
            startBlockNumber = startBlockNumber,
            endBlockNumber = endBlockNumber
          )
        )
      )

    return buildRequest(
      blocksAndLogs,
      tracesResponse,
      type2StateData,
      keccakPreviousStateRootHash
    ).thenCompose { request ->
      writeRequestToFile(requestFilePath, request)
    }
  }

  private fun writeRequestToFile(
    requestFilePath: Path,
    request: FileBasedExecutionProverClient.GetProofRequest
  ): SafeFuture<Path> {
    return fileWriter.write(request, requestFilePath, config.writingInprogressSuffix)
      .thenPeek {
        log.debug("execution proof request created: {}", requestFilePath)
      }
  }

  private fun buildRequest(
    blocksAndLogs: List<Pair<ExecutionPayloadV1, List<BridgeLogsData>>>,
    tracesResponse: GenerateTracesResponse,
    type2StateData: GetZkEVMStateMerkleProofResponse,
    keccakPreviousStateRootHash: String
  ): SafeFuture<FileBasedExecutionProverClient.GetProofRequest> {
    val blocksRlpBridgeLogsDataFuture =
      blocksAndLogs.map { (block, bridgeLogs) ->
        vertx.executeBlocking(
          Callable {
            executionPayloadV1Encoder.encode(block).encodeHex()
          }
        ).toSafeFuture().thenApply { rlp ->
          RlpBridgeLogsData(rlp, bridgeLogs)
        }
      }
    return SafeFuture.collectAll(*blocksRlpBridgeLogsDataFuture.toTypedArray()).thenApply { blocksRlpBridgeLogsData ->
      FileBasedExecutionProverClient.GetProofRequest(
        zkParentStateRootHash = type2StateData.zkParentStateRootHash.toHexString(),
        keccakParentStateRootHash = keccakPreviousStateRootHash,
        conflatedExecutionTracesFile = tracesResponse.tracesFileName,
        tracesEngineVersion = tracesResponse.tracesEngineVersion,
        type2StateManagerVersion = type2StateData.zkStateManagerVersion,
        zkStateMerkleProof = type2StateData.zkStateMerkleProof,
        blocksData = blocksRlpBridgeLogsData
      )
    }
  }
}
