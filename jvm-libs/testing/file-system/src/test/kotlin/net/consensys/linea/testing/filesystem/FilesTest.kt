package net.consensys.linea.testing.filesystem

import net.consensys.encodeHex
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import kotlin.random.Random

class FilesTest {
  @Test
  fun `when file exists shall return path to it`() {
    val file = findPathFileOrDir("build.gradle")
    assertThat(file).isNotNull
    assertThat(file.toString()).contains("jvm-libs/testing/file-system")

    assertThat(findPathFileOrDir("settings.gradle", lookupParentDir = false)).isNull()
    assertThat(findPathFileOrDir("settings.gradle", lookupParentDir = true)).exists()
  }

  @Test
  fun `when file does not exists shall return null`() {
    val file = findPathFileOrDir("non-existing-file-${Random.nextBytes(10).encodeHex()}")
    assertThat(file).isNull()
  }
}
