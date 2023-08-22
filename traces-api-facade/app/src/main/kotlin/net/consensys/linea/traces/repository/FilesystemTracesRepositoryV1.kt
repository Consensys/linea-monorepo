package net.consensys.linea.traces.repository

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.get
import com.github.michaelbull.result.getError
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.BlockTraces
import net.consensys.linea.ErrorType
import net.consensys.linea.TracesError
import net.consensys.linea.TracesRepositoryV1
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.metrics.monitoring.elapsedTimeInMillisSince
import net.consensys.linea.traces.TracesFileNameSupplier
import net.consensys.linea.traces.TracesFiles
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path

class FilesystemTracesRepositoryV1(
  private val vertx: Vertx,
  private val config: Config,
  private val fileNameSupplier: TracesFileNameSupplier = TracesFiles::rawTracesFileNameSupplierV1,
  private val tracesOnlyFilter: (content: JsonObject) -> JsonObject = ::tracesOnlyFromContent
) : TracesRepositoryV1 {
  data class Config(
    val tracesDirectory: Path,
    val tracesEngineVersion: String,
    val tracesFileExtension: String
  )

  private val log: Logger = LogManager.getLogger(this::class.java)
  private val fsHelper = FilesystemHelper(vertx, log = log)

  private fun findTracesFile(block: BlockNumberAndHash): Result<String, TracesError> {
    val tracesFileName = fileNameSupplier(
      block.number,
      block.hash,
      config.tracesEngineVersion,
      config.tracesFileExtension
    )
    val tracesFile = config.tracesDirectory.resolve(tracesFileName).toFile()
    return if (tracesFile.exists()) {
      Ok(tracesFile.absolutePath)
    } else {
      Err(
        TracesError(
          ErrorType.TRACES_UNAVAILABLE,
          "Traces not available for block ${block.number}. Target file: ${tracesFile.absolutePath}"
        )
      )
    }
  }

  private fun loadTracesFileContent(
    filePath: String,
    block: BlockNumberAndHash
  ): SafeFuture<BlockTraces> {
    val startTime = System.nanoTime()
    return fsHelper.readGzipedJsonFile(Path.of(filePath))
      .map { json -> BlockTraces(block.number, tracesOnlyFilter(json)) }
      .toSafeFuture()
      .whenComplete { _, _ ->
        log.debug(
          "load time=${elapsedTimeInMillisSince(startTime)}ms blockNumber=${block.number}"
        )
      }
  }

  override fun getTraces(block: BlockNumberAndHash): SafeFuture<Result<BlockTraces, TracesError>> {
    return when (val result = findTracesFile(block)) {
      is Ok<String> -> loadTracesFileContent(result.value, block).thenApply { Ok(it) }
      is Err<TracesError> -> SafeFuture.completedFuture(result)
    }
  }

  override fun getTraces(blocks: List<BlockNumberAndHash>): SafeFuture<Result<List<BlockTraces>, TracesError>> {
    val blocksFiles: List<Pair<BlockNumberAndHash, Result<String, TracesError>>> =
      blocks.map { it to findTracesFile(it) }

    val fileMissingError: TracesError? = blocksFiles.find { it.second is Err }?.second?.getError()
    if (fileMissingError != null) {
      return SafeFuture.completedFuture(Err(fileMissingError))
    }

    return SafeFuture.collectAll(
      blocksFiles.map { loadTracesFileContent(it.second.get()!!, it.first) }.stream()
    )
      .thenApply { listOfTraces: List<BlockTraces> -> Ok(listOfTraces) }
  }
}
