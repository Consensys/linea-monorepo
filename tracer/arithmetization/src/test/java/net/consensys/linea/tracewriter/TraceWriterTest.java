/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package net.consensys.linea.tracewriter;

import static org.assertj.core.api.Assertions.assertThat;

import java.nio.file.Path;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

class TraceWriterTest {

  @TempDir Path tempDir;

  private TraceWriter traceWriter;
  private TraceWriter compressedTraceWriter;

  @BeforeEach
  void setUp() {
    traceWriter = new TraceWriter(tempDir, false);
    compressedTraceWriter = new TraceWriter(tempDir, true);
  }

  @Test
  void traceFilePathGeneratesCorrectPath() {
    Path path = traceWriter.traceFilePath(100L, 200L, "0.2.3", "24.1.0");

    assertThat(path.getFileName().toString()).isEqualTo("100-200.conflated.0.2.3.24.1.0.lt");
    assertThat(path.getParent()).isEqualTo(tempDir);
  }

  @Test
  void traceFilePathWithCompressionHasGzExtension() {
    Path path = compressedTraceWriter.traceFilePath(100L, 200L, "0.2.3", "24.1.0");

    assertThat(path.getFileName().toString()).isEqualTo("100-200.conflated.0.2.3.24.1.0.lt.gz");
  }

  @Test
  void outputDirectoryCreatedIfNotExists() {
    Path newDir = tempDir.resolve("traces");
    TraceWriter writer = new TraceWriter(newDir, false);

    // traceFilePath should work even if directory doesn't exist
    // (directory is created when writing)
    Path path = writer.traceFilePath(100L, 100L, "0.2.3", "24.1.0");

    assertThat(path).isNotNull();
    assertThat(path.getParent()).isEqualTo(newDir);
  }

  @Test
  void singleBlockConflationNaming() {
    // When start and end block are the same
    Path path = traceWriter.traceFilePath(100L, 100L, "0.2.3", "24.1.0");

    assertThat(path.getFileName().toString()).isEqualTo("100-100.conflated.0.2.3.24.1.0.lt");
  }

  @Test
  void virtualBlockTraceFilePathGeneratesCorrectPath() {
    Path path = traceWriter.virtualBlockTraceFilePath(100L, "0.2.3");

    assertThat(path.getFileName().toString()).isEqualTo("100-.conflated.0.2.3.lt");
    assertThat(path.getParent()).isEqualTo(tempDir);
  }

  @Test
  void virtualBlockTraceFilePathWithCompressionHasGzExtension() {
    Path path = compressedTraceWriter.virtualBlockTraceFilePath(100L, "0.2.3");

    assertThat(path.getFileName().toString()).isEqualTo("100-.conflated.0.2.3.lt.gz");
  }

  @Test
  void virtualBlockTraceFilePathWithDifferentBlockNumbers() {
    Path path1 = traceWriter.virtualBlockTraceFilePath(1L, "1.0.0");
    Path path2 = traceWriter.virtualBlockTraceFilePath(999999L, "2.0.0");

    assertThat(path1.getFileName().toString()).isEqualTo("1-.conflated.1.0.0.lt");
    assertThat(path2.getFileName().toString()).isEqualTo("999999-.conflated.2.0.0.lt");
  }
}
