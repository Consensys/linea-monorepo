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
import net.consensys.zkevm.domain.ProofIndex
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.ProofsToAggregate
import net.consensys.zkevm.ethereum.crypto.HashFunction
import net.consensys.zkevm.ethereum.crypto.Sha256HashFunction
import net.consensys.zkevm.fileio.FileMonitor
import net.consensys.zkevm.fileio.FileReader
import net.consensys.zkevm.fileio.FileWriter
import net.consensys.zkevm.fileio.inProgressFilePattern
import org.apache.logging.log4j.util.Strings
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import kotlin.io.path.notExists
import kotlin.time.Duration

class FileBasedProofAggregationClient(
  private val vertx: Vertx,
  private val config: Config,
  private val mapper: ObjectMapper = proofResponseMapperV1,
  private val proofAggregationResponseFileNameProvider: ProverFileNameProvider =
    AggregationProofFileNameProvider,
  private val proofAggregationRequestFileNameProvider: ProverFileNameProvider =
    AggregationProofFileNameProvider,
  private val executionProofResponseFileNameProvider: ProverFileNameProvider =
    ExecutionProofResponseFileNameProvider,
  private val compressionProofResponseFileNameProvider: ProverFileNameProvider =
    CompressionProofResponseFileNameProvider,
  private val fileWriter: FileWriter = FileWriter(vertx, mapper),
  private val fileReader: FileReader<ProofToFinalizeJsonResponse> = FileReader(
    vertx,
    mapper,
    ProofToFinalizeJsonResponse::class.java
  ),
  private val fileMonitor: FileMonitor = FileMonitor(
    vertx,
    FileMonitor.Config(config.responseFilePollingInterval, config.responseFileMonitorTimeout)
  ),
  private val hashFunction: HashFunction = Sha256HashFunction()
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
    val inprogressRequestFileSuffix: String,
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
    val responseFilePath = config.responseFileDirectory.resolve(
      proofAggregationResponseFileNameProvider.getFileName(
        ProofIndex(
          startBlockNumber = blockInterval.startBlockNumber,
          endBlockNumber = blockInterval.endBlockNumber,
          hash = getRequestHash(request)
        )
      )
    )

    return fileMonitor.fileExists(responseFilePath)
      .thenCompose { responseExists ->
        if (responseExists) {
          parseResponse(responseFilePath)
        } else {
          val requestFilePath = config.requestFileDirectory
            .resolve(getZkAggregatedProofRequestFileName(request, aggregation))
          writeRequest(request = request, requestFilePath = requestFilePath)
            .thenCompose { fileMonitor.monitor(responseFilePath) }
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
    return fileMonitor.fileExists(
      config.requestFileDirectory,
      inProgressFilePattern(requestFilePath.fileName.toString(), config.proverInProgressSuffixPattern)
    ).thenCompose {
      if (it) {
        SafeFuture.completedFuture(requestFilePath)
      } else {
        fileWriter.write(request, requestFilePath, config.inprogressRequestFileSuffix)
      }
    }
  }

  internal fun buildRequest(proofsToAggregate: ProofsToAggregate): Request {
    val executionProofs = proofsToAggregate.executionProofs
      .toIntervalList()
      .map { blockInterval ->
        executionProofResponseFileNameProvider.getFileName(
          ProofIndex(
            startBlockNumber = blockInterval.startBlockNumber,
            endBlockNumber = blockInterval.endBlockNumber
          )
        )
      }

    val compressionProofs = proofsToAggregate.compressionProofIndexes
      .map {
        compressionProofResponseFileNameProvider.getFileName(
          ProofIndex(
            startBlockNumber = it.startBlockNumber,
            endBlockNumber = it.endBlockNumber,
            hash = it.hash
          )
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

  internal fun getZkAggregatedProofRequestFileName(
    request: Request,
    proofsToAggregate: ProofsToAggregate
  ): String {
    val startEndBlockInterval = proofsToAggregate.getStartEndBlockInterval()
    val contentHash = getRequestHash(request)
    return proofAggregationRequestFileNameProvider.getFileName(
      ProofIndex(
        startBlockNumber = startEndBlockInterval.startBlockNumber,
        endBlockNumber = startEndBlockInterval.endBlockNumber,
        hash = contentHash
      )
    )
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

  private fun getRequestHash(request: Request): ByteArray {
    val contentBytes = (request.compressionProofs + request.executionProofs).joinToString().toByteArray()
    return hashFunction.hash(contentBytes)
  }

  companion object {
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
