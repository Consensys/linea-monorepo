package net.consensys.linea.traces.repository

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.map
import io.vertx.core.Future
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject
import net.consensys.linea.BlockTraces
import net.consensys.linea.ErrorType
import net.consensys.linea.TracesError
import net.consensys.linea.TracesRepository
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.metrics.monitoring.elapsedTimeInMillisSince
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path

internal fun tracesOnlyFromContent(content: JsonObject): JsonObject {
  // TODO: filter out from the file objects that are not traces
  return content
}

class FilesystemTracesRepository(
  private val vertx: Vertx,
  private val tracesDirectory: Path,
  private val tracesFileExtension: String = "gz",
  private val tracesOnlyFilter: (content: JsonObject) -> JsonObject = ::tracesOnlyFromContent
) : TracesRepository {
  private val log: Logger = LogManager.getLogger(FilesystemTracesRepository::class.java)
  private val fsHelper = FilesystemHelper(vertx, log = log)
  private val tracesDirectoryPathStr = tracesDirectory.toString()
  private fun findTracesFile(blockNumber: UInt): Future<Result<String, TracesError>> {
    return vertx
      .fileSystem()
      .readDir(tracesDirectoryPathStr, "$blockNumber-.*.$tracesFileExtension")
      .map { listOfFiles ->
        when {
          listOfFiles.isEmpty() ->
            Err(
              TracesError(
                ErrorType.TRACES_UNAVAILABLE,
                "Traces not available for block $blockNumber."
              )
            )
          listOfFiles.size > 1 ->
            Err(
              TracesError(
                ErrorType.TRACES_AMBIGUITY,
                "Found multiple traces for the same block $blockNumber: [${listOfFiles.joinToString(",")}]"
              )
            )
          else -> Ok(listOfFiles.first())
        }
      }
  }

  private fun loadTracesFileContent(
    blockNumber: UInt
  ): Future<Result<Pair<String, JsonObject>, TracesError>> {
    return findTracesFile(blockNumber).flatMap { fileFindResult: Result<String, TracesError> ->
      when (fileFindResult) {
        is Err -> Future.succeededFuture(fileFindResult)
        is Ok -> {
          fsHelper.readGzipedJsonFile(Path.of(fileFindResult.value)).map { json ->
            Ok(Pair(fileFindResult.value, tracesOnlyFilter(json)))
          }
        }
      }
    }
  }

  override fun getTraces(blockNumber: UInt): SafeFuture<Result<BlockTraces, TracesError>> {
    val startTime = System.nanoTime()
    return loadTracesFileContent(blockNumber)
      .map { result ->
        result.map { (_: String, jsonContent: JsonObject) ->
          BlockTraces(blockNumber.toULong(), tracesOnlyFromContent(jsonContent))
        }
      }
      .toSafeFuture()
      .whenComplete { _, _ ->
        log.debug(
          "scanning folder + load time=${elapsedTimeInMillisSince(startTime)}ms blockNumber=$blockNumber"
        )
      }
  }
}
