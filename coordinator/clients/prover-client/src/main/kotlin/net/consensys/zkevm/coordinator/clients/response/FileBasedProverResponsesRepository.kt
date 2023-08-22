package net.consensys.zkevm.coordinator.clients.response

import com.fasterxml.jackson.databind.ObjectMapper
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.coordinator.clients.prover.ProofResponseFileNameProvider
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.notExists

class FileBasedProverResponsesRepository(
  private val config: Config,
  private val mapper: ObjectMapper,
  private val proofFileNameSupplier: ProofResponseFileNameProvider
) : ProverResponsesRepository {
  private val log: Logger = LogManager.getLogger(this::class.java)

  init {
    if (config.responseDirectory.notExists()) {
      config.responseDirectory.toFile().mkdirs()
    }
  }

  data class Config(val responseDirectory: Path, val proverVersion: String)

  private fun outputFileNameForBlockInterval(
    proverResponseIndex: ProverResponsesRepository.ProverResponseIndex
  ): Path {
    val proofFileName = proofFileNameSupplier.getResponseFileName(
      proverResponseIndex.startBlockNumber.longValue().toULong(),
      proverResponseIndex.endBlockNumber.longValue().toULong()
    )
    return config.responseDirectory.resolve(proofFileName)
  }

  override fun find(
    index: ProverResponsesRepository.ProverResponseIndex
  ): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>> {
    val outputFile = outputFileNameForBlockInterval(index)
    log.trace("Polling for file {}", outputFile)
    val result =
      if (outputFile.exists()) {
        parseResponseFile(outputFile)
      } else {
        Err(
          ErrorResponse(
            ProverErrorType.ResponseNotFound,
            "Response file '$outputFile' wasn't found in the repo"
          )
        )
      }
    return SafeFuture.completedFuture(result)
  }

  private fun parseResponseFile(
    outputFile: Path
  ): Result<GetProofResponse, ErrorResponse<ProverErrorType>> {
    return runCatching {
      val proofResponse = mapper.readValue(outputFile.toFile(), GetProofResponse::class.java)!!
      // prover uses the commit hash it's version. Overriding to configured version
      val sanitizedProofResponse = if (proofResponse.proverVersion != config.proverVersion) {
        log.warn(
          "Inconsistent prover versions: expected={}, returned={}",
          config.proverVersion,
          proofResponse.proverVersion
        )
        proofResponse.copy(proverVersion = config.proverVersion)
      } else {
        proofResponse
      }
      Ok(sanitizedProofResponse)
    }
      .recoverCatching { getProofResponseException ->
        runCatching {
          Err(
            mapper.readValue(outputFile.toFile(), ErrorResponse::class.java)!!
              as ErrorResponse<ProverErrorType>
          )
        }
          .getOrElse { throw getProofResponseException }
      }
      .getOrElse {
        Err(
          ErrorResponse(
            ProverErrorType.ParseError,
            "Failed to parse $outputFile, ${it.message}"
          )
        )
      }
  }
}
