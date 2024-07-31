package net.consensys.linea.traces.repository

import com.fasterxml.jackson.databind.ObjectMapper
import io.vertx.core.Future
import io.vertx.core.json.JsonObject
import net.consensys.linea.metrics.micrometer.elapsedTimeInMillisSince
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.nio.file.Path
import java.util.zip.GZIPInputStream

class FilesystemHelper(
  val objectMapper: ObjectMapper = ObjectMapper(),
  val log: Logger = LogManager.getLogger(FilesystemHelper::class.java)
) {
  fun readGzipedJsonFile(filePath: Path): Future<JsonObject> {
    val startTime = System.nanoTime()
    return java.io.FileInputStream(filePath.toFile()).use { fileIs ->
      val filesystemLoadTime = elapsedTimeInMillisSince(startTime)
      val startUngzipTime = System.nanoTime()
      GZIPInputStream(fileIs).use { gzipInputStream ->
        @Suppress("UNCHECKED_CAST")
        val jsonm = objectMapper.readValue(gzipInputStream, Map::class.java) as Map<String, Any>
        val json = JsonObject(jsonm)
        val jsonTime = elapsedTimeInMillisSince(startUngzipTime)
        val totalTime = elapsedTimeInMillisSince(startTime)
        log.debug(
          "total_time={}ms (file_load={} unzip+json_parse={}) in {}",
          totalTime,
          filesystemLoadTime,
          jsonTime,
          filePath.fileName
        )
        Future.succeededFuture(json)
      }
    }
  }

  fun readGzipedJsonFileAsString(filePath: Path): Future<String> {
    val startTime = System.nanoTime()
    return java.io.FileInputStream(filePath.toFile()).use { fileIs ->
      val filesystemLoadTime = elapsedTimeInMillisSince(startTime)
      GZIPInputStream(fileIs).use { gzipInputStream ->
        val unzipTime = elapsedTimeInMillisSince(startTime)
        try {
          val allBytes = gzipInputStream.readAllBytes()
          val result = String(allBytes)
          val toStringTime = elapsedTimeInMillisSince(startTime)
          log.debug(
            "total_time={}ms (file_load={} unzip={} toString={}) in {}",
            elapsedTimeInMillisSince(startTime),
            unzipTime - filesystemLoadTime,
            toStringTime - unzipTime - filesystemLoadTime,
            filePath.fileName
          )
          Future.succeededFuture(result)
        } catch (e: OutOfMemoryError) {
          log.warn("File is too large to read into String: ${filePath.fileName} error={}", e.message)
          Future.failedFuture(e)
        }
      }
    }
  }
}
