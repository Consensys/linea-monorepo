package net.consensys.zkevm.coordinator.clients.prover

import com.fasterxml.jackson.annotation.JsonProperty
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.annotation.JsonDeserialize
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.mapError
import io.vertx.core.Vertx
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.domain.ProofIndex
import net.consensys.zkevm.fileio.FileMonitor
import net.consensys.zkevm.fileio.FileReader
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.ethereum.executionclient.serialization.Bytes32Deserializer
import tech.pegasys.teku.ethereum.executionclient.serialization.BytesDeserializer
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import java.time.Instant
import kotlin.io.path.notExists

data class ExecutionProofResponse(
  @JsonDeserialize(using = BytesDeserializer::class) val proof: Bytes,
  val verifierIndex: Long,
  @JsonProperty("parentStateRootHash")
  @JsonDeserialize(using = Bytes32Deserializer::class)
  val zkParentStateRootHash: Bytes32,
  val blocksData: List<BlockData>
) {
  data class BlockData(
    /** Type 2 State Manager RootHash */
    @JsonProperty("rootHash")
    @JsonDeserialize(using = Bytes32Deserializer::class)
    val zkRootHash: Bytes32,
    /** Unix timestamp in seconds */
    val timestamp: Instant,
    val rlpEncodedTransactions: List<String>,
    val batchReceptionIndices: List<UShort>,
    /** L2->L1 Message service Smart Contract Logs abi encoded */
    val l2ToL1MsgHashes: List<String>,
    @JsonDeserialize(using = BytesDeserializer::class)
    val fromAddresses: Bytes
  )
}

class FileBasedProverResponsesRepository(
  private val vertx: Vertx,
  private val config: Config,
  private val mapper: ObjectMapper,
  private val proofResponseFileNameProvider: ProverFileNameProvider,
  private val fileMonitor: FileMonitor,
  private val responseReader: FileReader<ExecutionProofResponse> = FileReader(
    vertx,
    mapper,
    ExecutionProofResponse::class.java
  ),
  @Suppress("UNCHECKED_CAST")
  private val errorReader: FileReader<ErrorResponse<ProverErrorType>> =
    FileReader(vertx, mapper, ErrorResponse::class.java) as FileReader<ErrorResponse<ProverErrorType>>
) : ProverResponsesRepository {
  private val log: Logger = LogManager.getLogger(this::class.java)

  init {
    if (config.responseDirectory.notExists()) {
      config.responseDirectory.toFile().mkdirs()
    }
  }

  data class Config(val responseDirectory: Path)

  private fun outputFileNameForBlockInterval(
    proverResponseIndex: ProofIndex
  ): Path {
    val proofFileName = proofResponseFileNameProvider.getFileName(
      ProofIndex(
        startBlockNumber = proverResponseIndex.startBlockNumber,
        endBlockNumber = proverResponseIndex.endBlockNumber
      )
    )
    return config.responseDirectory.resolve(proofFileName)
  }

  override fun find(
    index: ProofIndex
  ): SafeFuture<Result<ExecutionProofResponse, ErrorResponse<ProverErrorType>>> {
    val outputFile = outputFileNameForBlockInterval(index)
    log.trace("Polling for file {}", outputFile)
    return fileMonitor.fileExists(outputFile).thenCompose {
      if (it == true) {
        parseResponseFile(outputFile)
      } else {
        SafeFuture.completedFuture(
          Err(
            ErrorResponse(
              ProverErrorType.ResponseNotFound,
              "Response file '$outputFile' wasn't found in the repo"
            )
          )
        )
      }
    }
  }

  override fun monitor(
    index: ProofIndex
  ): SafeFuture<Result<ExecutionProofResponse, ErrorResponse<ProverErrorType>>> {
    val outputFile = outputFileNameForBlockInterval(index)
    return fileMonitor.monitor(outputFile).thenCompose {
      when (it) {
        is Ok -> parseResponseFile(it.value)
        is Err -> {
          when (it.error) {
            FileMonitor.ErrorType.TIMED_OUT -> SafeFuture.completedFuture(
              Err(
                ErrorResponse(ProverErrorType.ResponseTimeout, "Monitoring timed out")
              )
            )
          }
        }
      }
    }
  }

  private fun parseResponseFile(
    outputFile: Path
  ): SafeFuture<Result<ExecutionProofResponse, ErrorResponse<ProverErrorType>>> {
    return responseReader.read(outputFile).thenApply { result ->
      result
        .mapError { error ->
          ErrorResponse(
            ProverErrorType.ParseError,
            "Failed to parse $outputFile, ${error.message}"
          )
        }
    }.thenCompose { responseResult ->
      when (responseResult) {
        is Ok -> SafeFuture.completedFuture(Ok(responseResult.value))
        is Err -> errorReader.read(outputFile).thenApply {
          when (it) {
            is Ok -> Err(it.value)
            is Err -> Err(
              ErrorResponse(
                ProverErrorType.ParseError,
                "Failed to parse $outputFile, ${it.error.message}"
              )
            )
          }
        }
      }
    }
  }
}
