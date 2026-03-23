package linea.blob.process

import linea.blob.GoNativeBlobCompressor
import org.apache.logging.log4j.LogManager
import java.io.BufferedReader
import java.io.Closeable
import java.io.IOException
import java.io.PrintWriter
import java.util.HexFormat
import java.util.concurrent.TimeUnit

/**
 * Implements [GoNativeBlobCompressor] by communicating with a child JVM process
 * over newline-delimited JSON on stdin/stdout.
 *
 * The child process ([GoNativeBlobCompressorSubprocessServer]) loads the same `.so` via
 * JNA but runs in its own OS address space, giving full isolation of the Go
 * runtime (GC, goroutine scheduler, global state).
 *
 * Use [GoNativeBlobCompressorSubprocessFactory.create] to obtain an instance.
 * Call [close] when the compressor is no longer needed.
 */
class GoNativeBlobCompressorSubprocessClient internal constructor(private val process: Process) :
  GoNativeBlobCompressor, Closeable {

  private val log = LogManager.getLogger(GoNativeBlobCompressorSubprocessClient::class.java)
  private val stdin: PrintWriter = PrintWriter(process.outputStream.bufferedWriter())
  private val stdout: BufferedReader = process.inputStream.bufferedReader()

  // ── IPC ────────────────────────────────────────────────────────────────────

  @Synchronized
  private fun rpc(json: String): String {
    check(process.isAlive) {
      "Compressor subprocess has terminated unexpectedly (exit=${process.exitValue()})"
    }
    log.trace("→ {}", json)
    stdin.println(json)
    stdin.flush()
    val response = stdout.readLine()
      ?: throw IOException("Compressor subprocess closed stdout unexpectedly")
    log.trace("← {}", response)
    if (extractField(response, "ok") == "false") {
      val msg = extractField(response, "error") ?: "unknown error"
      throw IOException("Compressor subprocess error: $msg  (request=$json)")
    }
    return response
  }

  // ── GoNativeBlobCompressor ─────────────────────────────────────────────────

  override fun Init(dataLimit: Int, dictPath: String): Boolean {
    val resp = rpc("""{"method":"Init","dataLimit":$dataLimit,"dictPath":${jsonStr(dictPath)}}""")
    return extractBool(resp) ?: false
  }

  override fun Reset() {
    rpc("""{"method":"Reset"}""")
  }

  override fun StartNewBatch() {
    rpc("""{"method":"StartNewBatch"}""")
  }

  override fun Write(data: ByteArray, data_len: Int): Boolean {
    val resp = rpc("""{"method":"Write","data":"${toHex(data, data_len)}"}""")
    return extractBool(resp) ?: false
  }

  override fun CanWrite(data: ByteArray, data_len: Int): Boolean {
    val resp = rpc("""{"method":"CanWrite","data":"${toHex(data, data_len)}"}""")
    return extractBool(resp) ?: false
  }

  override fun Error(): String? = extractField(rpc("""{"method":"Error"}"""), "str")

  override fun Len(): Int = extractInt(rpc("""{"method":"Len"}""")) ?: 0

  override fun Bytes(out: ByteArray) {
    val resp = rpc("""{"method":"Bytes"}""")
    val hex = extractField(resp, "hex") ?: return
    fromHex(hex).copyInto(out)
  }

  override fun WorstCompressedBlockSize(data: ByteArray, data_len: Int): Int {
    val resp = rpc("""{"method":"WorstCompressedBlockSize","data":"${toHex(data, data_len)}"}""")
    return extractInt(resp) ?: -1
  }

  override fun WorstCompressedTxSize(data: ByteArray, data_len: Int): Int {
    val resp = rpc("""{"method":"WorstCompressedTxSize","data":"${toHex(data, data_len)}"}""")
    return extractInt(resp) ?: -1
  }

  override fun RawCompressedSize(data: ByteArray, data_len: Int): Int {
    val resp = rpc("""{"method":"RawCompressedSize","data":"${toHex(data, data_len)}"}""")
    return extractInt(resp) ?: -1
  }

  // ── Closeable ──────────────────────────────────────────────────────────────

  override fun close() {
    stdin.close()
    if (!process.waitFor(5, TimeUnit.SECONDS)) {
      log.warn("Compressor subprocess did not exit within 5 s – destroying forcibly")
      process.destroyForcibly()
    }
  }

  // ── JSON helpers (client side – parses responses) ─────────────────────────

  private fun toHex(data: ByteArray, len: Int): String = HexFormat.of().formatHex(data, 0, len)
  private fun fromHex(hex: String): ByteArray = HexFormat.of().parseHex(hex)
  private fun extractBool(json: String) = extractField(json, "bool")?.toBooleanStrictOrNull()
  private fun extractInt(json: String) = extractField(json, "int")?.toIntOrNull()
}

// ─── shared JSON helpers (package-private) ───────────────────────────────────

/** Wraps [s] as a JSON string, escaping backslashes, double-quotes and newlines. */
internal fun jsonStr(s: String): String =
  "\"${s.replace("\\", "\\\\").replace("\"", "\\\"").replace("\n", "\\n")}\""

/**
 * Minimal single-pass extractor for the compact JSON produced by this protocol.
 * Handles JSON strings (with basic escape sequences) and scalar values
 * (booleans, integers, null).  Fields must not contain nested objects or arrays.
 */
internal fun extractField(json: String, field: String): String? {
  val marker = "\"$field\":"
  val idx = json.indexOf(marker)
  if (idx == -1) return null
  val start = idx + marker.length
  return if (json[start] == '"') {
    var i = start + 1
    val sb = StringBuilder()
    while (i < json.length && json[i] != '"') {
      if (json[i] == '\\' && i + 1 < json.length) {
        sb.append(json[i + 1])
        i += 2
      } else {
        sb.append(json[i++])
      }
    }
    sb.toString()
  } else {
    val end = json.indexOf(',', start).takeIf { it != -1 } ?: json.indexOf('}', start)
    val raw = json.substring(start, end)
    if (raw == "null") null else raw
  }
}
