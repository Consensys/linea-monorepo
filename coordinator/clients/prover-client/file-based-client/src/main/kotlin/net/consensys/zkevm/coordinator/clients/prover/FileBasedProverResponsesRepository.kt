package net.consensys.zkevm.coordinator.clients.prover

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.ProverErrorType
import net.consensys.zkevm.domain.ProofIndex
import net.consensys.zkevm.fileio.FileMonitor
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import kotlin.io.path.notExists

class FileBasedProverResponsesRepository(
  private val config: Config,
  private val proofResponseFileNameProvider: ProverFileNameProvider,
  private val fileMonitor: FileMonitor
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
  ): SafeFuture<Unit> {
    val outputFile = outputFileNameForBlockInterval(index)
    log.trace("Polling for file {}", outputFile)
    return fileMonitor.fileExists(outputFile).thenCompose {
      if (it == true) {
        SafeFuture.completedFuture(Unit)
      } else {
        SafeFuture.failedFuture(
          ErrorResponse(
            ProverErrorType.ResponseNotFound,
            "Response file '$outputFile' wasn't found in the repo"
          ).asException()
        )
      }
    }
  }

  override fun monitor(
    index: ProofIndex
  ): SafeFuture<Unit> {
    val outputFile = outputFileNameForBlockInterval(index)
    return fileMonitor.monitor(outputFile).thenCompose {
      when (it) {
        is Ok -> SafeFuture.completedFuture(Unit)
        is Err -> {
          when (it.error) {
            FileMonitor.ErrorType.TIMED_OUT -> SafeFuture.failedFuture(
              ErrorResponse(ProverErrorType.ResponseTimeout, "Monitoring timed out").asException()
            )
          }
        }
      }
    }
  }
}
