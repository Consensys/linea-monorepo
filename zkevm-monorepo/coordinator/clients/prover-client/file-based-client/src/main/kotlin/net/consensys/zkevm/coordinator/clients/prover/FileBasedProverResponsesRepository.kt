package net.consensys.zkevm.coordinator.clients.prover

import com.fasterxml.jackson.databind.ObjectMapper
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.mapError
import io.vertx.core.Vertx
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.fileio.FileMonitor
import net.consensys.zkevm.fileio.FileReader
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import kotlin.io.path.notExists

class FileBasedProverResponsesRepository(
  private val vertx: Vertx,
  private val config: Config,
  private val mapper: ObjectMapper,
  private val proofFileNameSupplier: ProofResponseFileNameProvider,
  private val fileMonitor: FileMonitor,
  private val responseReader: FileReader<GetProofResponse> = FileReader(vertx, mapper, GetProofResponse::class.java),
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
    proverResponseIndex: ProverResponsesRepository.ProverResponseIndex
  ): Path {
    val proofFileName = proofFileNameSupplier.getResponseFileName(
      proverResponseIndex.startBlockNumber,
      proverResponseIndex.endBlockNumber
    )
    return config.responseDirectory.resolve(proofFileName)
  }

  override fun find(
    index: ProverResponsesRepository.ProverResponseIndex
  ): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>> {
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
    index: ProverResponsesRepository.ProverResponseIndex
  ): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>> {
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
  ): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>> {
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
