package net.consensys.zkevm.coordinator.blockcreation

import io.vertx.core.Future
import io.vertx.core.Vertx
import net.consensys.linea.async.AsyncRetryer
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.traces.TracesFileNameSupplier
import net.consensys.linea.traces.TracesFiles
import net.consensys.zkevm.coordinator.clients.TracesWatcher
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.io.FileNotFoundException
import java.nio.file.Files
import java.nio.file.Path
import kotlin.time.Duration

class TracesFilesManager(
  private val vertx: Vertx,
  private val config: Config,
  private val tracesFileNameSupplier: TracesFileNameSupplier = TracesFiles::rawTracesFileNameSupplierV1
) : TracesWatcher {
  data class Config(
    val tracesFolder: Path,
    val nonCanonicalTracesDir: Path,
    val pollingInterval: Duration,
    val tracesGenerationTimeout: Duration,
    val tracesEngineVersion: String,
    val tracesFileExtension: String,
    val createNonCanonicalTracesDirIfDoesNotExist: Boolean
  )

  private val log: Logger = LogManager.getLogger(this::class.java)
  private val retries: Int = run {
    config.tracesGenerationTimeout.inWholeMilliseconds /
      config.pollingInterval.inWholeMilliseconds.coerceAtLeast(1L)
  }.toInt()

  init {
    if (!Files.exists(config.nonCanonicalTracesDir)) {
      if (config.createNonCanonicalTracesDirIfDoesNotExist) {
        Files.createDirectories(config.nonCanonicalTracesDir)
      } else {
        throw FileNotFoundException("${config.nonCanonicalTracesDir} directory not found!")
      }
    }
  }

  override fun waitRawTracesGenerationOf(
    blockNumber: UInt64,
    blockHash: Bytes32
  ): SafeFuture<String> {
    val fileName =
      tracesFileNameSupplier(
        blockNumber.longValue().toULong(),
        blockHash,
        config.tracesEngineVersion,
        config.tracesFileExtension
      )

    val targetFile = config.tracesFolder.resolve(fileName).toFile()
    return AsyncRetryer.retry(
      vertx = vertx,
      maxRetries = retries,
      backoffDelay = config.pollingInterval
    ) {
      log.trace("Waiting for traces file: ${targetFile.absolutePath}")
      if (targetFile.exists()) {
        log.trace("Found for traces file: ${targetFile.absolutePath}")
        SafeFuture.completedFuture(targetFile.absolutePath)
      } else {
        val errorMessage = "File matching '$fileName' not found after ${config.tracesGenerationTimeout}."
        SafeFuture.failedFuture(FileNotFoundException(errorMessage))
      }
    }
  }

  internal fun cleanNonCanonicalSiblingsByHeight(
    blockNumber: UInt64,
    canonicalBlockHashToKeep: Bytes32
  ): SafeFuture<List<String>> {
    return vertx
      .fileSystem()
      .readDir(config.tracesFolder.toString())
      .flatMap { listOfFiles ->
        val filesToMove =
          listOfFiles.filter { fileAbsolutePath ->
            val fileName = Path.of(fileAbsolutePath).fileName.toString().lowercase()
            fileName.startsWith("$blockNumber-") &&
              fileName.endsWith(config.tracesFileExtension.lowercase()) &&
              !fileName.contains(canonicalBlockHashToKeep.toHexString().lowercase())
          }

        Future.all(
          filesToMove.map { fileAbsolutePath ->
            val destination =
              config.nonCanonicalTracesDir
                .resolve(Path.of(fileAbsolutePath).fileName)
                .toString()
            log.info("Moving non-canonical traces file $fileAbsolutePath --> $destination")
            vertx.fileSystem().move(fileAbsolutePath, destination)
          }
        )
          .map { filesToMove }
      }
      .toSafeFuture()
      .whenException { th -> log.error("Failed to move traces files: errorMessage={}", th.message, th) }
  }
}
