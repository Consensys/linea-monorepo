package net.consensys.linea.nativecompressor

import com.sun.jna.Library
import com.sun.jna.Native
import java.io.File
import java.nio.file.Files
import java.nio.file.Path

object NativeLibUtil {
  fun copyResourceToTmpDir(resourceName: String): Path {
    val fileDestination = File(
      Files.createTempDirectory("linea-blob-compressor")
        .resolve(resourceName)
        .toString()
    )
    Files.copy(NativeLibUtil::class.java.getResourceAsStream(resourceName)!!, fileDestination.toPath())
    return fileDestination.toPath()
  }

  @JvmStatic
  fun <T : Library> loadJnaLibFromFileSystem(libFilePath: Path, clazz: Class<T>): T {
    return Native.load(libFilePath.toString(), clazz) as T
  }

  @JvmStatic
  fun <T : Library> loadJnaLibFromResource(libName: String, clazz: Class<T>): T {
    return loadJnaLibFromFileSystem(copyResourceToTmpDir(libName), clazz)
  }
}
