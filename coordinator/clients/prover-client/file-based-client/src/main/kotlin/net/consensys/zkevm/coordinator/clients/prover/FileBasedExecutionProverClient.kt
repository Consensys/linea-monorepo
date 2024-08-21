package net.consensys.zkevm.coordinator.clients.prover

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.node.ArrayNode
import io.vertx.core.Vertx
import net.consensys.linea.CommonDomainFunctions.blockIntervalString
import net.consensys.zkevm.coordinator.clients.ExecutionProverClient
import net.consensys.zkevm.coordinator.clients.GenerateTracesResponse
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import net.consensys.zkevm.coordinator.clients.L2MessageServiceLogsClient
import net.consensys.zkevm.coordinator.clients.prover.serialization.JsonSerialization
import net.consensys.zkevm.domain.ProofIndex
import net.consensys.zkevm.domain.RlpBridgeLogsData
import net.consensys.zkevm.fileio.FileMonitor
import net.consensys.zkevm.fileio.inProgressFilePattern
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.nio.file.Path
import kotlin.io.path.notExists
import kotlin.time.Duration

/**
 * Implementation of interface with the Prover trough Files.
 *
 * Prover will ingest file like
 * path/to/prover-input-dir/<startBlockNumber>-<endBlockNumber>-getZkProof.json
 *
 * When done prover will output file
 * path/to/prover-output-dir/<startBlockNumber>-<endBlockNumber>-proof.json
 *
 * So, this class will need to watch the file system and wait for the output proof to be generated
 */
class FileBasedExecutionProverClient(
  private val config: Config,
  private val l2MessageServiceLogsClient: L2MessageServiceLogsClient,
  private val vertx: Vertx,
  private val l2Web3jClient: Web3j,
  private val mapper: ObjectMapper = JsonSerialization.proofResponseMapperV1,
  private val executionProofRequestFileNameProvider: ProverFileNameProvider =
    ExecutionProofRequestFileNameProvider(
      tracesVersion = config.tracesVersion,
      stateManagerVersion = config.stateManagerVersion
    ),
  private val executionProofResponseFileNameProvider: ProverFileNameProvider = ExecutionProofResponseFileNameProvider,
  private val fileMonitor: FileMonitor = FileMonitor(
    vertx,
    FileMonitor.Config(config.pollingInterval, config.timeout)
  ),
  val proverResponsesRepository: FileBasedProverResponsesRepository = FileBasedProverResponsesRepository(
    FileBasedProverResponsesRepository.Config(config.responseDirectory),
    proofResponseFileNameProvider = executionProofResponseFileNameProvider,
    fileMonitor
  )
) : ExecutionProverClient {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private val requestFileWriter: RequestFileWriter = RequestFileWriter(
    vertx = vertx,
    config = RequestFileWriter.Config(
      requestDirectory = config.requestDirectory,
      writingInprogressSuffix = ".coordinator_writing_inprogress",
      proverInprogressSuffixPattern = config.inprogressProvingSuffixPattern
    ),
    mapper = mapper,
    proofRequestFileNameProvider = executionProofRequestFileNameProvider,
    log = log
  )

  init {
    if (config.requestDirectory.notExists()) {
      val dirCreated = config.requestDirectory.toFile().mkdirs()
      if (!dirCreated) {
        log.error("Failed to create prover request directory {}!", config.requestDirectory)
      }
    }
    if (config.responseDirectory.notExists()) {
      val dirCreated = config.responseDirectory.toFile().mkdirs()
      if (!dirCreated) {
        log.error("Failed to create prover response directory {}!", config.responseDirectory)
      }
    }
  }

  data class Config(
    val requestDirectory: Path,
    val responseDirectory: Path,
    val inprogressProvingSuffixPattern: String,
    val pollingInterval: Duration,
    val timeout: Duration,
    val tracesVersion: String,
    val stateManagerVersion: String
  )

  internal data class GetProofRequest(
    val zkParentStateRootHash: String?,
    val keccakParentStateRootHash: String,
    val conflatedExecutionTracesFile: String,
    val tracesEngineVersion: String,
    val type2StateManagerVersion: String?,
    val zkStateMerkleProof: ArrayNode,
    val blocksData: List<RlpBridgeLogsData>
  )

  internal inner class ResponseFileMonitor(
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ) {
    private val proverResponseIndex = ProofIndex(
      startBlockNumber,
      endBlockNumber
    )

    fun findResponse(): SafeFuture<Unit> {
      return proverResponsesRepository.find(proverResponseIndex)
    }

    fun monitor(): SafeFuture<Unit> {
      return proverResponsesRepository.monitor(proverResponseIndex)
    }
  }

  private fun getPreviousBlockKeccakStateRootHash(blockNumber: Long): SafeFuture<String> {
    return SafeFuture.of(
      l2Web3jClient
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf(BigInteger.valueOf(blockNumber - 1)),
          false
        )
        .sendAsync()
    )
      .thenApply { previousBlock -> previousBlock.block.stateRoot }
  }

  private fun isRequestAlreadyExistingOrProvingInProgress(
    requestFilePath: Path,
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): SafeFuture<Boolean> {
    return fileMonitor.fileExists(
      config.requestDirectory,
      inProgressFilePattern(requestFilePath.fileName.toString(), config.inprogressProvingSuffixPattern)
    ).thenApply {
      if (it == true) {
        log.info(
          "Request file exists or proving already in progress for batch={}: requestFile={}",
          blockIntervalString(startBlockNumber, endBlockNumber),
          requestFilePath.fileName
        )
      }
      it
    }
  }

  private fun buildRequestFilePath(
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): Path = config.requestDirectory.resolve(
    executionProofRequestFileNameProvider.getFileName(
      ProofIndex(
        startBlockNumber = startBlockNumber,
        endBlockNumber = endBlockNumber
      )
    )
  )

  override fun requestBatchExecutionProof(
    blocks: List<ExecutionPayloadV1>,
    tracesResponse: GenerateTracesResponse,
    type2StateData: GetZkEVMStateMerkleProofResponse
  ): SafeFuture<GetProofResponse> {
    val startBlockNumber = blocks.first().blockNumber.toULong()
    val endBlockNumber = blocks.last().blockNumber.toULong()
    val responseMonitor = ResponseFileMonitor(startBlockNumber, endBlockNumber)
    val requestFilePath = buildRequestFilePath(startBlockNumber, endBlockNumber)

    // Check if the request is already proven. If so, return it.
    // This happens when coordinator is restarted and the request is already proven.
    return responseMonitor.findResponse().handleComposed { _, throwable ->
      if (throwable == null) {
        log.debug(
          "execution proof already proven: batch={} reusedResponse={}",
          blockIntervalString(startBlockNumber, endBlockNumber),
          executionProofResponseFileNameProvider.getFileName(
            ProofIndex(
              startBlockNumber = startBlockNumber,
              endBlockNumber = endBlockNumber
            )
          )
        )
        SafeFuture.completedFuture(GetProofResponse(startBlockNumber, endBlockNumber))
      } else {
        isRequestAlreadyExistingOrProvingInProgress(
          requestFilePath = requestFilePath,
          startBlockNumber = startBlockNumber,
          endBlockNumber = endBlockNumber
        ).thenCompose { requestAlreadyExistingOrProvingInProgress ->
          when {
            requestAlreadyExistingOrProvingInProgress -> SafeFuture.completedFuture(requestFilePath)
            else -> {
              val bridgeLogsSfList =
                blocks.map { block ->
                  l2MessageServiceLogsClient.getBridgeLogs(
                    blockNumber = block.blockNumber.longValue()
                  )
                }
              SafeFuture.collectAll(bridgeLogsSfList.stream())
                .thenApply { blocksLogs ->
                  blocks.zip(blocksLogs)
                }.thenComposeCombined(
                  getPreviousBlockKeccakStateRootHash(blocks.first().blockNumber.longValue())
                ) { bundledBlocks, previousKeccakStateRootHash ->
                  requestFileWriter.write(
                    bundledBlocks,
                    tracesResponse,
                    type2StateData,
                    previousKeccakStateRootHash
                  )
                }
            }
          }
            .thenCompose {
              responseMonitor
                .monitor()
                .thenApply {
                  GetProofResponse(startBlockNumber, endBlockNumber)
                }
            }
        }
      }
    }
  }
}
