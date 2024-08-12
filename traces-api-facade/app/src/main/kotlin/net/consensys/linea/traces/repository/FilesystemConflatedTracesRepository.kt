package net.consensys.linea.traces.repository

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import io.vertx.core.Vertx
import io.vertx.core.json.JsonObject
import net.consensys.linea.ConflatedTracesRepository
import net.consensys.linea.TracesConflation
import net.consensys.linea.async.toSafeFuture
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.io.FileOutputStream
import java.io.OutputStream
import java.nio.file.Path
import java.util.concurrent.Callable
import java.util.zip.GZIPOutputStream
import kotlin.system.measureTimeMillis

class FilesystemConflatedTracesRepository(
  private val vertx: Vertx,
  private val tracesDirectory: Path,
  private val gzipCompressionEnabled: Boolean = true,
  private val objectMapper: ObjectMapper = jacksonObjectMapper()
) : ConflatedTracesRepository {
  private val log: Logger = LogManager.getLogger(this.javaClass)

  private fun destinationFileName(
    startBlockNumber: ULong,
    endBlockNumber: ULong,
    tracesVersion: String
  ): String {
    val extension = if (gzipCompressionEnabled) "json.gz" else "json"
    return "$startBlockNumber-$endBlockNumber.conflated.v$tracesVersion.$extension"
  }

  private fun inProgressDestinationFileName(
    startBlockNumber: ULong,
    endBlockNumber: ULong,
    tracesVersion: String
  ): String {
    return "${destinationFileName(startBlockNumber, endBlockNumber, tracesVersion)}.inprogress"
  }

  override fun findConflatedTraces(
    startBlockNumber: ULong,
    endBlockNumber: ULong,
    tracesVersion: String
  ): SafeFuture<String?> {
    val fileName = destinationFileName(
      startBlockNumber,
      endBlockNumber,
      tracesVersion
    )

    tracesDirectory.resolve(fileName).toFile().let { file ->
      return if (file.exists()) {
        SafeFuture.completedFuture(file.name)
      } else {
        SafeFuture.completedFuture(null)
      }
    }
  }

  override fun saveConflatedTraces(conflation: TracesConflation): SafeFuture<String> {
    val inProgressFileName =
      inProgressDestinationFileName(
        conflation.startBlockNumber,
        conflation.endBlockNumber,
        conflation.traces.version
      )
    val inProgressFilePath = tracesDirectory.resolve(inProgressFileName)

    return vertx
      .executeBlocking(
        Callable {
          if (gzipCompressionEnabled) {
            saveConflatedTracesGzipCompressed(inProgressFilePath, conflation.traces.result)
          } else {
            saveConflatedTracesRawJson(inProgressFilePath, conflation.traces.result)
          }
        },
        false
      )
      .toSafeFuture()
      .thenApply {
        destinationFileName(
          conflation.startBlockNumber,
          conflation.endBlockNumber,
          conflation.traces.version
        ).also { destinationFileName ->
          tracesDirectory.resolve(destinationFileName).run {
            inProgressFilePath.toFile().renameTo(this.toFile())
          }
        }
      }
  }

  private fun saveConflatedTracesGzipCompressed(filePath: Path, traces: JsonObject) {
    var serializationTime: Long
    log.info("saving conflation to {}", filePath.fileName)
    val time = measureTimeMillis {
      FileOutputStream(filePath.toString()).use { outputStream: OutputStream ->
        GZIPOutputStream(outputStream).use { gzipOutputStream ->
          serializationTime = measureTimeMillis { objectMapper.writeValue(gzipOutputStream, traces) }
        }
      }
    }
    log.debug(
      "total_time={}ms (json_encode + gzip + fs_write={}) in {}",
      time,
      serializationTime,
      filePath.fileName
    )
  }

  private fun saveConflatedTracesRawJson(filePath: Path, traces: JsonObject) {
    FileOutputStream(filePath.toString()).use { outputStream: OutputStream ->
      outputStream.write(traces.toString().toByteArray())
    }
  }
}

// Keeping for quick debugging in the future and fast iteration
// fun main() {
//   val vertx = Vertx.vertx()
//   val filePath = Path.of("tmp/")
//   val repository = FilesystemConflatedTracesRepository(
//     vertx,
//     filePath,
//     true
//   )
//
//   val jsonObject = JsonObject.of(
//     "key", "value",
//     "key2", "value2",
//     "key3", "value3"
//   )
//   repository.saveConflatedTraces(TracesConflation(1u, 4u, VersionedResult("", jsonObject)))
//     .get()
//   vertx.close()
// }
