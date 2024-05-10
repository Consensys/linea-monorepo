package net.consensys.zkevm.coordinator.clients.prover

import com.fasterxml.jackson.databind.ObjectMapper
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.vertx.core.Vertx
import net.consensys.encodeHex
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.ProofAggregationClient
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.coordinator.clients.prover.serialization.JsonSerialization.proofResponseMapperV1
import net.consensys.zkevm.coordinator.clients.prover.serialization.ProofToFinalizeJsonResponse
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.ProofsToAggregate
import net.consensys.zkevm.fileio.FileMonitor
import net.consensys.zkevm.fileio.FileReader
import net.consensys.zkevm.fileio.FileWriter
import org.apache.logging.log4j.util.Strings
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import java.security.MessageDigest
import java.security.NoSuchAlgorithmException
import java.util.*
import kotlin.io.path.notExists
import kotlin.time.Duration

class FileBasedProofAggregationClient(
  private val vertx: Vertx,
  private val config: Config,
  private val mapper: ObjectMapper = proofResponseMapperV1,
  private val proofAggregationFileNameProviderV2: ProofResponseFileNameProvider =
    AggregationProofResponseFileNameProviderV2,
  private val proofAggregationFileNameProviderV3: ProofResponseFileNameProviderV2 =
    AggregationProofResponseFileNameProviderV3,

  private val executionProofFileNameProvider: ProofResponseFileNameProvider = ExecutionProofResponseFileNameProvider,
  private val compressionProofFileNameProvider: ProofResponseFileNameProvider = CompressionProofFileNameProvider,
  private val fileWriter: FileWriter = FileWriter(vertx, mapper),
  private val fileReader: FileReader<ProofToFinalizeJsonResponse> = FileReader(
    vertx,
    mapper,
    ProofToFinalizeJsonResponse::class.java
  ),
  private val fileMonitor: FileMonitor = FileMonitor(
    vertx,
    FileMonitor.Config(config.responseFilePollingInterval, config.responseFileMonitorTimeout)
  )
) : ProofAggregationClient {

  init {
    if (config.requestFileDirectory.notExists()) {
      config.requestFileDirectory.toFile().mkdirs()
    }
    if (config.responseFileDirectory.notExists()) {
      config.responseFileDirectory.toFile().mkdirs()
    }
  }

  data class Config(
    val requestFileDirectory: Path,
    val responseFileDirectory: Path,
    val responseFilePollingInterval: Duration,
    val responseFileMonitorTimeout: Duration,
    val inProgressRequestFileSuffix: String,
    val proverInProgressSuffixPattern: String
  )

  data class Request(
    val executionProofs: List<String>,
    val compressionProofs: List<String>,
    val parentAggregationLastBlockTimestamp: Long,
    val parentAggregationLastL1RollingHashMessageNumber: Long,
    val parentAggregationLastL1RollingHash: String
  )

  override fun getAggregatedProof(
    aggregation: ProofsToAggregate
  ): SafeFuture<Result<ProofToFinalize, ErrorResponse<ProverErrorType>>> {
    val blockInterval = aggregation.getStartEndBlockInterval()
    val request = buildRequest(aggregation)
    val responseFilePathV2 = config.responseFileDirectory.resolve(
      proofAggregationFileNameProviderV2.getResponseFileName(
        blockInterval.startBlockNumber,
        blockInterval.endBlockNumber
      )
    )
    val responseFilePathV3 = config.responseFileDirectory.resolve(
      proofAggregationFileNameProviderV3.getResponseFileName(
        blockInterval.startBlockNumber,
        blockInterval.endBlockNumber,
        getRequestHash(request)
      )
    )
    val possibleResponseFiles = listOf(responseFilePathV2, responseFilePathV3)

    return SafeFuture.collectAll(possibleResponseFiles.map { fileMonitor.fileExists(it) }.stream())
      .thenCompose { responseFileExistList ->
        val existingResponseFileIndex = responseFileExistList.indexOf(true)
        if (existingResponseFileIndex >= 0) {
          parseResponse(possibleResponseFiles[existingResponseFileIndex])
        } else {
          val requestFilePath = config.requestFileDirectory
            .resolve(getZkAggregatedProofRequestFileName(request, aggregation))
          writeRequest(request = request, requestFilePath = requestFilePath)
            .thenCompose { fileMonitor.monitorFiles(possibleResponseFiles) }
            .thenCompose {
              when (it) {
                is Ok -> parseResponse(it.value)
                is Err -> SafeFuture.completedFuture(
                  Err(ErrorResponse(mapFileMonitorError(it.error), Strings.EMPTY))
                )
              }
            }
        }
      }
  }

  private fun writeRequest(request: Request, requestFilePath: Path): SafeFuture<Path> {
    val provingInProgress = fileMonitor.fileExists(
      config.requestFileDirectory,
      requestFilePath.fileName.toString() + config.proverInProgressSuffixPattern
    )
    val requestFileWritingDoneOrInProgress = fileWriter.writingDoneOrInProgress(
      requestFilePath,
      config.inProgressRequestFileSuffix
    )
    return SafeFuture.collectAll(provingInProgress, requestFileWritingDoneOrInProgress)
      .thenCompose {
        if (it.contains(true)) {
          SafeFuture.completedFuture(requestFilePath)
        } else {
          fileWriter.write(request, requestFilePath, config.inProgressRequestFileSuffix)
        }
      }
  }

  internal fun buildRequest(proofsToAggregate: ProofsToAggregate): Request {
    val executionProofs = proofsToAggregate.executionProofs
      .toIntervalList()
      .zip(proofsToAggregate.executionVersion)
      .map {
        val blockInterval = it.first
        executionProofFileNameProvider.getResponseFileName(
          blockInterval.startBlockNumber,
          blockInterval.endBlockNumber
        )
      }

    val compressionProofs = proofsToAggregate.compressionProofs
      .toIntervalList()
      .map { blockInterval ->
        compressionProofFileNameProvider.getResponseFileName(
          blockInterval.startBlockNumber,
          blockInterval.endBlockNumber
        )
      }

    return Request(
      executionProofs = executionProofs,
      compressionProofs = compressionProofs,
      parentAggregationLastBlockTimestamp = proofsToAggregate.parentAggregationLastBlockTimestamp.epochSeconds,
      parentAggregationLastL1RollingHashMessageNumber =
      proofsToAggregate.parentAggregationLastL1RollingHashMessageNumber.toLong(),
      parentAggregationLastL1RollingHash = proofsToAggregate.parentAggregationLastL1RollingHash.encodeHex()
    )
  }

  private fun getRequestHash(request: Request): String {
    val contentBytes = (request.compressionProofs + request.executionProofs).joinToString().toByteArray()
    return HexFormat.of().formatHex(sha256(contentBytes))
  }

  internal fun getZkAggregatedProofRequestFileName(
    request: Request,
    proofsToAggregate: ProofsToAggregate
  ): String {
    val startEndBlockInterval = proofsToAggregate.getStartEndBlockInterval()
    val contentHash = getRequestHash(request)
    return "${startEndBlockInterval.startBlockNumber}-${startEndBlockInterval.endBlockNumber}-$contentHash" +
      "-getZkAggregatedProof.json"
  }

  private fun parseResponse(filePath: Path):
    SafeFuture<Result<ProofToFinalize, ErrorResponse<ProverErrorType>>> {
    return fileReader
      .read(filePath)
      .thenApply {
        when (it) {
          is Ok -> Ok(it.value.toDomainObject())
          is Err -> Err(ErrorResponse(mapFileReaderError(it.error.type), it.error.message))
        }
      }
  }

  companion object {
    fun sha256(input: ByteArray): ByteArray {
      try {
        val digest = MessageDigest.getInstance("SHA-256")
        return digest.digest(input)
      } catch (e: NoSuchAlgorithmException) {
        throw RuntimeException("Couldn't find a SHA-256 provider", e)
      }
    }

    private fun mapFileMonitorError(error: FileMonitor.ErrorType): ProverErrorType {
      return when (error) {
        FileMonitor.ErrorType.TIMED_OUT -> ProverErrorType.ResponseNotFound
      }
    }

    private fun mapFileReaderError(error: FileReader.ErrorType): ProverErrorType {
      return when (error) {
        FileReader.ErrorType.PARSING_ERROR -> ProverErrorType.ParseError
      }
    }
  }
}
