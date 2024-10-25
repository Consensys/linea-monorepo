package build.linea.jvm

import build.linea.jvm.ResourcesUtil.copyResourceToTmpDir
import org.assertj.core.api.AssertionsForClassTypes.assertThat
import org.junit.jupiter.api.Test
import java.nio.file.Files

class ResourcesUtilTest {
  @Test
  fun `copyResourceToTmpDir should copy when in root directory`() {
    val file = copyResourceToTmpDir(resourcePath = "root-resource.txt", tmpDirPrefix = "testing-resources-")
    assertThat(Files.readString(file).trim()).isEqualTo("root resource 1")
  }

  @Test
  fun `copyResourceToTmpDir should copy when not in root directory`() {
    val file1Path =
      copyResourceToTmpDir(resourcePath = "test/folder/nested-resource.txt", tmpDirPrefix = "testing-resources-")
    val file2Path =
      copyResourceToTmpDir(resourcePath = "test/folder2/nested-resource.txt", tmpDirPrefix = "testing-resources-")

    // make sure files with same name in different directories are not overwritten inside the same tmp directory
    assertThat(Files.readString(file1Path).trim()).isEqualTo("nested resource 1")
    assertThat(Files.readString(file2Path).trim()).isEqualTo("nested resource 2")
  }
}
