package linea.blob.process

import linea.blob.BlobCompressorVersion
import linea.blob.GoNativeBlobCompressor
import linea.blob.GoNativeBlobCompressorFactory
import java.io.PrintWriter
import java.util.HexFormat

/**
 * Entry point for the child JVM process spawned by [GoNativeBlobCompressorSubprocessFactory].
 *
 * Reads newline-delimited JSON commands from stdin and writes JSON responses to
 * stdout using the same protocol as [GoNativeBlobCompressorSubprocessClient].  The
 * compressor is the ordinary JNA-backed implementation; running it in a separate
 * process is what provides address-space isolation.
 *
 * Do not call this directly – it is invoked by [GoNativeBlobCompressorSubprocessFactory.create].
 */
object GoNativeBlobCompressorSubprocessServer {

  @JvmStatic
  fun main(args: Array<String>) {
    require(args.size == 1) { "Expected exactly one argument: <version>" }
    val version = BlobCompressorVersion.entries.first { it.version == args[0] }
    val compressor: GoNativeBlobCompressor = GoNativeBlobCompressorFactory.getInstance(version)

    val reader = System.`in`.bufferedReader()
    val writer = PrintWriter(System.out.bufferedWriter())

    fun respond(json: String) {
      writer.println(json)
      writer.flush()
    }

    var line = reader.readLine()
    while (line != null) {
      val method = extractField(line, "method")
      if (method == null) {
        respond(errResp("missing 'method' field"))
        line = reader.readLine()
        continue
      }

      val response = runCatching {
        when (method) {
          "Init" -> {
            val ok = compressor.Init(
              dataLimit = extractField(line, "dataLimit")?.toIntOrNull() ?: 0,
              dictPath = extractField(line, "dictPath") ?: "",
            )
            boolResp(ok)
          }
          "Reset" -> {
            compressor.Reset()
            voidResp()
          }
          "StartNewBatch" -> {
            compressor.StartNewBatch()
            voidResp()
          }
          "Write" -> boolResp(compressor.Write(data(line), data(line).size))
          "CanWrite" -> boolResp(compressor.CanWrite(data(line), data(line).size))
          "Error" -> strResp(compressor.Error())
          "Len" -> intResp(compressor.Len())
          "Bytes" -> {
            val buf = ByteArray(compressor.Len())
            compressor.Bytes(buf)
            hexResp(buf)
          }
          "WorstCompressedBlockSize" -> intResp(compressor.WorstCompressedBlockSize(data(line), data(line).size))
          "WorstCompressedTxSize" -> intResp(compressor.WorstCompressedTxSize(data(line), data(line).size))
          "RawCompressedSize" -> intResp(compressor.RawCompressedSize(data(line), data(line).size))
          else -> errResp("unknown method: $method")
        }
      }.getOrElse { errResp(it.message ?: it.javaClass.simpleName) }

      respond(response)
      line = reader.readLine()
    }
  }

  // ── response builders ────────────────────────────────────────────────────

  private fun voidResp() = """{"ok":true}"""
  private fun boolResp(b: Boolean) = """{"ok":true,"bool":$b}"""
  private fun intResp(n: Int) = """{"ok":true,"int":$n}"""
  private fun strResp(s: String?) = if (s != null) """{"ok":true,"str":${jsonStr(s)}}""" else """{"ok":true}"""
  private fun hexResp(b: ByteArray) = """{"ok":true,"hex":"${java.util.HexFormat.of().formatHex(b)}"}"""
  private fun errResp(msg: String) = """{"ok":false,"error":${jsonStr(msg)}}"""

  // ── helpers ───────────────────────────────────────────────────────────────

  /** Decodes the hex-encoded "data" field from [json]. */
  private fun data(json: String): ByteArray {
    val hex = extractField(json, "data") ?: return ByteArray(0)
    return if (hex.isEmpty()) ByteArray(0) else HexFormat.of().parseHex(hex)
  }
}
