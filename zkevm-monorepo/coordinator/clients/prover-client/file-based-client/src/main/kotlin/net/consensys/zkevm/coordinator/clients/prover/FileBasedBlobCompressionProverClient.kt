package net.consensys.zkevm.coordinator.clients.prover

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.vertx.core.Vertx
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.BlobCompressionProof
import net.consensys.zkevm.coordinator.clients.BlobCompressionProverClient
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.coordinator.clients.prover.serialization.BlobCompressionProofJsonRequest
import net.consensys.zkevm.coordinator.clients.prover.serialization.BlobCompressionProofJsonResponse
import net.consensys.zkevm.coordinator.clients.prover.serialization.JsonSerialization
import net.consensys.zkevm.domain.BlockInterval
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ConflationCalculationResult
import net.consensys.zkevm.ethereum.coordination.blob.ShnarfResult
import net.consensys.zkevm.fileio.FileMonitor
import net.consensys.zkevm.fileio.FileReader
import net.consensys.zkevm.fileio.FileWriter
import net.consensys.zkevm.fileio.inProgressFilePattern
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.logging.log4j.util.Strings
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import kotlin.io.path.notExists
import kotlin.time.Duration

/**
 * Implementation of interface with the Blob Compression Prover trough Files.
 *
 * Blob Compression Prover will ingest file like
 * path/to/prover-input-dir/<startBlockNumber>-<endBlockNumber>-bcv<blobCalculatorVersion>-ccv<conflationCalculatorVersion>-getZkBlobCompressionProof.json
 *
 * When done prover will output file
 * path/to/prover-output-dir/<startBlockNumber>-<endBlockNumber>-getZkBlobCompressionProof.json
 *
 * So, this class will need to watch the file system and wait for the output proof to be generated
 */
class FileBasedBlobCompressionProverClient(
  private val config: Config,
  private val vertx: Vertx,
  private val fileWriter: FileWriter = FileWriter(vertx, JsonSerialization.proofResponseMapperV1),
  private val fileReader: FileReader<BlobCompressionProofJsonResponse> = FileReader(
    vertx,
    JsonSerialization.proofResponseMapperV1,
    BlobCompressionProofJsonResponse::class.java
  ),
  private val fileMonitor: FileMonitor = FileMonitor(
    vertx,
    FileMonitor.Config(config.pollingInterval, config.timeout)
  ),
  private val compressionProofFileNameProvider: ProofResponseFileNameProvider = CompressionProofFileNameProvider
) : BlobCompressionProverClient {

  init {
    if (config.requestFileDirectory.notExists()) {
      config.requestFileDirectory.toFile().mkdirs()
    }
    if (config.responseFileDirectory.notExists()) {
      config.responseFileDirectory.toFile().mkdirs()
    }
  }

  private val log: Logger = LogManager.getLogger(this::class.java)

  data class Config(
    val requestFileDirectory: Path,
    val responseFileDirectory: Path,
    val inprogressProvingSuffixPattern: String,
    val inprogressRequestFileSuffix: String,
    val pollingInterval: Duration,
    val timeout: Duration,
    val blobCalculatorVersion: String,
    val conflationCalculatorVersion: String
  )

  fun parseResponse(filePath: Path):
    SafeFuture<Result<BlobCompressionProof, ErrorResponse<ProverErrorType>>> {
    return fileReader
      .read(filePath)
      .thenApply {
        when (it) {
          is Ok -> Ok(it.value.toDomainObject())
          is Err -> Err(ErrorResponse(mapFileReaderError(it.error.type), it.error.message))
        }
      }
  }

  private fun getProverRequestIndex(
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): CompressionProverRequestIndex {
    return CompressionProverRequestIndex(
      startBlockNumber,
      endBlockNumber,
      config.blobCalculatorVersion,
      config.conflationCalculatorVersion
    )
  }

  private fun getProverResponseIndex(
    startBlockNumber: ULong,
    endBlockNumber: ULong
  ): CompressionProverResponseIndex {
    return CompressionProverResponseIndex(
      startBlockNumber,
      endBlockNumber
    )
  }

  override fun requestBlobCompressionProof(
    compressedData: ByteArray,
    conflations: List<ConflationCalculationResult>,
    parentStateRootHash: ByteArray,
    finalStateRootHash: ByteArray,
    parentDataHash: ByteArray,
    prevShnarf: ByteArray,
    expectedShnarfResult: ShnarfResult,
    eip4844Enabled: Boolean,
    commitment: ByteArray,
    kzgProofContract: ByteArray,
    kzgProofSideCar: ByteArray
  ): SafeFuture<Result<BlobCompressionProof, ErrorResponse<ProverErrorType>>> {
    val proverRequestIndex = getProverRequestIndex(
      conflations.first().startBlockNumber,
      conflations.last().endBlockNumber
    )
    val proverResponseIndex = getProverResponseIndex(
      conflations.first().startBlockNumber,
      conflations.last().endBlockNumber
    )
    val responseFilePath = config.responseFileDirectory.resolve(
      compressionProofFileNameProvider.getResponseFileName(
        proverResponseIndex.startBlockNumber,
        proverResponseIndex.endBlockNumber
      )
    )
    return fileMonitor.fileExists(responseFilePath)
      .thenCompose { responseFileExists ->
        if (responseFileExists) {
          log.info(
            "compression proof already proven: blob={} reusedResponse={}",
            proverResponseIndex.intervalString(),
            responseFilePath
          )
          parseResponse(responseFilePath)
        } else {
          writeRequest(
            proverRequestIndex = proverRequestIndex,
            compressedData = compressedData,
            conflations = conflations,
            prevShnarf = prevShnarf,
            parentStateRootHash = parentStateRootHash,
            finalStateRootHash = finalStateRootHash,
            parentDataHash = parentDataHash,
            expectedShnarfResult = expectedShnarfResult,
            eip4844Enabled = eip4844Enabled,
            commitment = commitment,
            kzgProofContract = kzgProofContract,
            kzgProofSideCar = kzgProofSideCar
          )
            .thenCompose { fileMonitor.monitor(responseFilePath) }
            .thenCompose {
              when (it) {
                is Ok -> {
                  log.debug("blob compression proof created: ${it.value}")
                  parseResponse(it.value)
                }

                is Err -> {
                  val proverErrorType = mapFileMonitorError(it.error)
                  val errorMessage = if (proverErrorType == ProverErrorType.ResponseNotFound) {
                    "Blob compression proof not found after ${config.timeout.inWholeSeconds}s, " +
                      "blob=${proverRequestIndex.intervalString()}"
                  } else {
                    Strings.EMPTY
                  }
                  SafeFuture.completedFuture(Err(ErrorResponse(proverErrorType, errorMessage)))
                }
              }
            }
        }
      }
  }

  private fun writeRequest(
    proverRequestIndex: CompressionProverRequestIndex,
    compressedData: ByteArray,
    conflations: List<ConflationCalculationResult>,
    prevShnarf: ByteArray,
    parentStateRootHash: ByteArray,
    finalStateRootHash: ByteArray,
    parentDataHash: ByteArray,
    expectedShnarfResult: ShnarfResult,
    eip4844Enabled: Boolean,
    commitment: ByteArray,
    kzgProofContract: ByteArray,
    kzgProofSideCar: ByteArray
  ): SafeFuture<Path> {
    val request = buildRequest(
      compressedData = compressedData,
      conflations = conflations,
      prevShnarf = prevShnarf,
      parentStateRootHash = parentStateRootHash,
      finalStateRootHash = finalStateRootHash,
      parentDataHash = parentDataHash,
      expectedShnarfResult = expectedShnarfResult,
      eip4844Enabled = eip4844Enabled,
      commitment = commitment,
      kzgProofContract = kzgProofContract,
      kzgProofSideCar = kzgProofSideCar
    )
    val requestFilePath = config.requestFileDirectory.resolve(getRequestFileName(proverRequestIndex))
    return fileMonitor.findFile(
      config.requestFileDirectory,
      inProgressFilePattern(requestFilePath.fileName.toString(), config.inprogressProvingSuffixPattern)
    ).thenCompose { alreadyExistingRequest: String? ->
      if (!alreadyExistingRequest.isNullOrEmpty()) {
        log.info(
          "compression proof already requested: blob={} reusingFile={}",
          proverRequestIndex.intervalString(),
          alreadyExistingRequest
        )
        SafeFuture.completedFuture(requestFilePath)
      } else {
        log.debug(
          "requesting compression proof: blob={} fileName={} eip4844Enabled={}",
          proverRequestIndex.intervalString(),
          requestFilePath,
          eip4844Enabled
        )
        fileWriter.write(request, requestFilePath, config.inprogressRequestFileSuffix)
      }
    }
  }

  private fun buildRequest(
    compressedData: ByteArray,
    conflations: List<ConflationCalculationResult>,
    prevShnarf: ByteArray,
    parentStateRootHash: ByteArray,
    finalStateRootHash: ByteArray,
    parentDataHash: ByteArray,
    expectedShnarfResult: ShnarfResult,
    eip4844Enabled: Boolean,
    commitment: ByteArray,
    kzgProofContract: ByteArray,
    kzgProofSideCar: ByteArray
  ): BlobCompressionProofJsonRequest {
    return BlobCompressionProofJsonRequest(
      compressedData = compressedData,
      conflationOrder = BlockIntervals(
        startingBlockNumber = conflations.first().startBlockNumber,
        upperBoundaries = conflations.map { it.endBlockNumber }
      ),
      prevShnarf = prevShnarf,
      parentStateRootHash = parentStateRootHash,
      finalStateRootHash = finalStateRootHash,
      parentDataHash = parentDataHash,
      dataHash = expectedShnarfResult.dataHash,
      snarkHash = expectedShnarfResult.snarkHash,
      expectedX = expectedShnarfResult.expectedX,
      expectedY = expectedShnarfResult.expectedY,
      expectedShnarf = expectedShnarfResult.expectedShnarf,
      eip4844Enabled = eip4844Enabled,
      commitment = commitment,
      kzgProofContract = kzgProofContract,
      kzgProofSidecar = kzgProofSideCar
    )
  }

  data class CompressionProverRequestIndex(
    override val startBlockNumber: ULong,
    override val endBlockNumber: ULong,
    val blobCalculatorVersion: String,
    val conflationCalculatorVersion: String
  ) : BlockInterval

  data class CompressionProverResponseIndex(
    override val startBlockNumber: ULong,
    override val endBlockNumber: ULong
  ) : BlockInterval

  companion object {
    fun getRequestFileName(proverRequestIndex: CompressionProverRequestIndex): String {
      return "${proverRequestIndex.startBlockNumber}-${proverRequestIndex.endBlockNumber}-" +
        "bcv${proverRequestIndex.blobCalculatorVersion}-ccv${proverRequestIndex.conflationCalculatorVersion}-" +
        "getZkBlobCompressionProof.json"
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
