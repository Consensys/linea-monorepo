package linea.blob.process

import linea.blob.BlobCompressorVersion

/**
 * Creates [GoNativeBlobCompressorSubprocessClient] instances, each backed by an independent
 * child JVM process.
 *
 * The child process is started with the same classpath as the current JVM so
 * that it can load the JNA `.so` from the classpath resources.
 */
object GoNativeBlobCompressorSubprocessFactory {
  /**
   * Spawns a fresh child JVM and returns a [GoNativeBlobCompressorSubprocessClient]
   * connected to it.  Each call produces a fully isolated compressor instance.
   * The caller must [GoNativeBlobCompressorSubprocessClient.close] it when finished.
   */
  @JvmStatic
  fun create(version: BlobCompressorVersion): GoNativeBlobCompressorSubprocessClient {
    val javaExe = "${System.getProperty("java.home")}/bin/java"
    val classpath = System.getProperty("java.class.path")
    val process = ProcessBuilder(
      javaExe,
      "-cp",
      classpath,
      GoNativeBlobCompressorSubprocessServer::class.java.name,
      version.version,
    )
      .redirectError(ProcessBuilder.Redirect.INHERIT)
      .start()
    return GoNativeBlobCompressorSubprocessClient(process)
  }
}
