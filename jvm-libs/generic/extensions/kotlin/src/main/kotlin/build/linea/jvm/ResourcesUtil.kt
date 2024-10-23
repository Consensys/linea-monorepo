package build.linea.jvm

import java.io.File
import java.nio.file.Files
import java.nio.file.Path

object ResourcesUtil {
  @JvmStatic
  fun copyResourceToTmpDir(
    resourcePath: String,
    tmpDirPrefix: String = "linea-resources-"
  ): Path {
    val fileDestination = File(
      Files.createTempDirectory(tmpDirPrefix)
        .resolve(Path.of(resourcePath).fileName)
        .toString()
    )
    Files.copy(
      Thread.currentThread().getContextClassLoader().getResourceAsStream(resourcePath),
      fileDestination.toPath()
    )
    return fileDestination.toPath()
  }
}
