package net.consensys.zkevm.coordinator.clients.prover

import com.fasterxml.jackson.databind.ObjectMapper
import io.vertx.core.Vertx
import net.consensys.linea.CommonDomainFunctions.batchIntervalString
import net.consensys.linea.async.toSafeFuture
import net.consensys.zkevm.coordinator.clients.FileBasedProverClient
import net.consensys.zkevm.coordinator.clients.GenerateTracesResponse
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.Hash
import org.hyperledger.besu.datatypes.Wei
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.BlockBody
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder
import org.hyperledger.besu.ethereum.core.Difficulty
import org.hyperledger.besu.ethereum.core.encoding.TransactionDecoder
import org.hyperledger.besu.ethereum.mainnet.BodyValidation
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.evm.log.LogsBloomFilter
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import kotlin.io.path.Path
import kotlin.io.path.absolutePathString

internal class RequestFileWriter(
  private val vertx: Vertx,
  private val fileNamesProvider: ProverFilesNameProvider,
  private val config: Config,
  private val mapper: ObjectMapper,
  private val log: Logger
) {
  data class Config(
    val requestDirectory: Path,
    val writingInprogressSuffix: String,
    val proverInprogressSuffixPattern: String
  )

  private fun hasProvingInProgress(
    requestFilePath: Path,
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): SafeFuture<Boolean> {
    return vertx.fileSystem()
      .readDir(
        config.requestDirectory.toString(),
        requestFilePath.fileName.toString() + config.proverInprogressSuffixPattern
      )
      .map { files ->
        when {
          files.isEmpty() -> false
          else -> {
            log.info(
              "Proving already in progress for batch={}: File in progress {}",
              batchIntervalString(startBlockNumber, endBlockNumber),
              files
            )
            true
          }
        }
      }
      .toSafeFuture()
  }

  fun write(
    blocksAndLogs: List<Pair<ExecutionPayloadV1, List<FileBasedProverClient.BridgeLogsData>>>,
    tracesResponse: GenerateTracesResponse,
    type2StateData: GetZkEVMStateMerkleProofResponse,
    keccakPreviousStateRootHash: String
  ): SafeFuture<Path> {
    val startBlockNumber = blocksAndLogs.first().first.blockNumber.toULong()
    val endBlockNumber = blocksAndLogs.last().first.blockNumber.toULong()
    val requestFilePath = buildRequestFilePath(startBlockNumber, endBlockNumber)

    return hasProvingInProgress(
      requestFilePath,
      startBlockNumber,
      endBlockNumber
    )
      .thenCompose { hasProvingInProgress ->
        when {
          hasProvingInProgress -> SafeFuture.completedFuture(requestFilePath)
          else -> {
            val request = buildRequest(
              blocksAndLogs,
              tracesResponse,
              type2StateData,
              keccakPreviousStateRootHash
            )
            writeRequestToFile(requestFilePath, request)
          }
        }
      }
  }

  private fun writeRequestToFile(
    requestFilePath: Path,
    request: FileBasedProverClient.GetProofRequest
  ): SafeFuture<Path> {
    val requestWriteInprogressFilePath = Path(requestFilePath.absolutePathString() + config.writingInprogressSuffix)
    return vertx
      .executeBlocking {
        try {
          val inprogressFile = requestWriteInprogressFilePath.toFile()
          mapper.writeValue(inprogressFile, request)
          inprogressFile.renameTo(requestFilePath.toFile())
        } catch (t: Throwable) {
          it.fail(t)
        }
        it.complete(requestFilePath)
      }
      .toSafeFuture()
  }

  private fun buildRequest(
    blocksAndLogs: List<Pair<ExecutionPayloadV1, List<FileBasedProverClient.BridgeLogsData>>>,
    tracesResponse: GenerateTracesResponse,
    type2StateData: GetZkEVMStateMerkleProofResponse,
    keccakPreviousStateRootHash: String
  ): FileBasedProverClient.GetProofRequest {
    val blocksRlpBridgeLogsData =
      blocksAndLogs.map {
        val block = it.first
        val bridgeLogs = it.second
        val parsedTransactions = block.transactions.map(TransactionDecoder::decodeOpaqueBytes)
        val parsedBody = BlockBody(parsedTransactions, emptyList())
        val blockHeader =
          BlockHeaderBuilder.create()
            .parentHash(Hash.wrap(block.parentHash))
            .ommersHash(Hash.EMPTY_LIST_HASH)
            .coinbase(Address.wrap(block.feeRecipient.wrappedBytes))
            .stateRoot(Hash.wrap(block.stateRoot))
            .transactionsRoot(BodyValidation.transactionsRoot(parsedBody.transactions))
            .receiptsRoot(Hash.wrap(block.receiptsRoot))
            .logsBloom(LogsBloomFilter(block.logsBloom))
            .difficulty(Difficulty.ZERO)
            .number(block.blockNumber.longValue())
            .gasLimit(block.gasLimit.longValue())
            .gasUsed(block.gasLimit.longValue())
            .timestamp(block.timestamp.longValue())
            .extraData(block.extraData)
            .baseFee(Wei.wrap(block.baseFeePerGas.toBytes()))
            .mixHash(Hash.wrap(block.prevRandao))
            .nonce(0)
            .blockHeaderFunctions(MainnetBlockHeaderFunctions())
            .buildBlockHeader()
        val rlp = Block(blockHeader, parsedBody).toRlp().toHexString()
        FileBasedProverClient.RlpBridgeLogsData(rlp, bridgeLogs)
      }

    return FileBasedProverClient.GetProofRequest(
      zkParentStateRootHash = type2StateData.zkParentStateRootHash?.toHexString(),
      keccakParentStateRootHash = keccakPreviousStateRootHash,
      conflatedExecutionTracesFile = tracesResponse.tracesFileName,
      tracesEngineVersion = tracesResponse.tracesEngineVersion,
      type2StateManagerVersion = type2StateData.zkStateManagerVersion,
      zkStateMerkleProof = type2StateData.zkStateMerkleProof,
      blocksData = blocksRlpBridgeLogsData
    )
  }

  private fun buildRequestFilePath(
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): Path = config.requestDirectory.resolve(fileNamesProvider.getRequestFileName(startBlockNumber, endBlockNumber))
}
