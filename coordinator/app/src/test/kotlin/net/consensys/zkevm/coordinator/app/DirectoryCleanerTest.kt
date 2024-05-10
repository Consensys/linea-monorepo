package net.consensys.zkevm.coordinator.app

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import net.consensys.linea.async.get
import net.consensys.linea.async.toSafeFuture
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import java.io.File
import java.io.FileFilter
import java.nio.file.Files
import kotlin.io.path.absolutePathString
import kotlin.io.path.nameWithoutExtension

@ExtendWith(VertxExtension::class)
class DirectoryCleanerTest {

  @Test
  fun test_tmp_directory_cleanup(vertx: Vertx) {
    val tmpDirectory = Files.createTempDirectory(
      DirectoryCleanerTest::class.simpleName + "-test_tmp_directory_cleanup"
    )
    for (i in 1..10) {
      val fileExtension = if (i % 2 == 0) { ".json" } else { ".json.in-progress" }
      Files.createTempFile(
        tmpDirectory,
        "directory_cleaner_test-request-$i",
        fileExtension
      )
    }
    assertThat(vertx.fileSystem().readDir(tmpDirectory.absolutePathString()).get().size).isEqualTo(10)
    val directoryCleaner = DirectoryCleaner(vertx, listOf(tmpDirectory), DirectoryCleaner.JSON_FILE_FILTER)
    directoryCleaner.cleanup().get()
    var remainingFiles = 0
    vertx.fileSystem().readDir(tmpDirectory.absolutePathString()).toSafeFuture().thenApply { filePaths ->
      filePaths.forEach { filePath ->
        val file = File(filePath)
        assertThat(file.extension).isEqualTo("in-progress")
        remainingFiles += 1
      }
    }.get()
    assertThat(remainingFiles).isEqualTo(5)
  }

  @Test
  fun test_deletion_of_absent_file_does_not_throw_exception(vertx: Vertx) {
    val tmpDirectory = Files.createTempDirectory(
      DirectoryCleanerTest::class.simpleName + "-test_deletion_of_absent_file_throws_exception"
    )
    val tmpFile = Files.createTempFile(
      tmpDirectory,
      "absent_file",
      ".json"
    )
    val tmpFile2 = Files.createTempFile(
      tmpDirectory,
      "directory_cleaner_test-request-1",
      ".json"
    )

    val tmpFile3 = Files.createTempFile(
      tmpDirectory,
      "directory_cleaner_test-request-2",
      ".json.in-progress"
    )

    val mockFileFilter = mock<FileFilter> {}
    whenever(mockFileFilter.accept(any())).thenAnswer {
      val file = it.getArgument<File>(0)
      if (file.nameWithoutExtension.contains(tmpFile.nameWithoutExtension)) {
        vertx.fileSystem().deleteBlocking(it.getArgument<File>(0).absolutePath)
      }
      if (file.extension.equals("json")) true else false
    }
    assertThat(vertx.fileSystem().readDir(tmpDirectory.absolutePathString()).get().size).isEqualTo(3)

    val directoryCleaner = DirectoryCleaner(vertx, listOf(tmpDirectory), mockFileFilter)
    directoryCleaner.cleanup().get()

    val remainingFiles = vertx.fileSystem().readDir(tmpDirectory.absolutePathString()).get()
    assertThat(remainingFiles.size).isEqualTo(1)
    assertThat(remainingFiles.first()).contains("in-progress")
  }

  @Test
  fun test_json_file_filter() {
    assertThat(DirectoryCleaner.JSON_FILE_FILTER.accept(File("11-27-getZkAggregatedProof.json"))).isTrue()
    assertThat(DirectoryCleaner.JSON_FILE_FILTER.accept(File("11-27-getZkAggregatedProof.Json"))).isTrue()
    assertThat(DirectoryCleaner.JSON_FILE_FILTER.accept(File("11-27-getZkAggregatedProof.JSON"))).isTrue()
    assertThat(
      DirectoryCleaner.JSON_FILE_FILTER.accept(File("11-27-getZkAggregatedProof.json.in-progress"))
    ).isFalse()
  }
}
