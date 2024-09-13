package net.consensys.linea.traces.repository

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import kotlin.io.path.Path

class FilesystemHelperTest {
  private val tracesDirectory = Path("../../testdata/traces/raw")
  lateinit var filesystemHelper: FilesystemHelper

  @BeforeEach
  fun beforeEach() {
    filesystemHelper = FilesystemHelper()
  }

  @Test
  fun readGzipedJsonFileAsString() {
    val traces = filesystemHelper.readGzippedJsonFileAsString(
      tracesDirectory.resolve("1-0xab538e7ab831af9442aab00443ee9803907654359dfcdfe1755f1a98fb87eafd.v0.0.1.json.gz")
    )
    assertThat(traces).isNotNull()
  }

  // @Test
  // disabled: meant for local prototyping only with production files
  fun readGzipedJsonFileAsString_canReadVeryLargeFile() {
    val filePath = Path("../../tmp/local/traces/raw")
      .resolve("2480859-0xeed3fd5ffcf442e9e7906d1d078ef6e607c1fc1aa015ebdb6cf2fff938d837a3.v0.2.0.json.gz")
    val traces = filesystemHelper.readGzippedJsonFileAsString(filePath)
    assertThat(traces).isNotNull()
  }
}
