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

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.nio.file.attribute.FileAttribute;
import java.nio.file.attribute.PosixFilePermission;
import java.nio.file.attribute.PosixFilePermissions;
import java.util.Set;

import lombok.RequiredArgsConstructor;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ZkTracer;

@Slf4j
@RequiredArgsConstructor
public class TraceWriter {
  private static final String TRACE_FILE_EXTENSION = ".lt";
  private static final String TRACE_TEMP_FILE_EXTENSION = ".lt.tmp";

  private final ZkTracer tracer;

  @SneakyThrows(IOException.class)
  public Path writeTraceToFile(
      final Path tracesOutputDirPath,
      final long startBlockNumber,
      final long endBlockNumber,
      final String expectedTracesEngineVersion) {
    // Generate the original and final trace file name.
    final String origTraceFileName =
        generateOutputFileName(startBlockNumber, endBlockNumber, expectedTracesEngineVersion);
    // Generate and resolve the original and final trace file path.
    final Path origTraceFilePath =
        generateOutputFilePath(tracesOutputDirPath, origTraceFileName + TRACE_FILE_EXTENSION);

    // Write the trace at the original and final trace file path, but with the suffix .tmp at the
    // end of the file.
    final Path tmpTraceFilePath =
        writeToTmpFile(
            tracesOutputDirPath,
            origTraceFileName + ".",
            TRACE_TEMP_FILE_EXTENSION,
            startBlockNumber,
            endBlockNumber);
    // After trace writing is complete, rename the file by removing the .tmp prefix, indicating
    // the file is complete and should not be corrupted due to trace writing issues.
    final Path finalizedTraceFilePath =
        Files.move(tmpTraceFilePath, origTraceFilePath, StandardCopyOption.ATOMIC_MOVE);

    return finalizedTraceFilePath.toAbsolutePath();
  }

  public Path writeToTmpFile(
      final Path rootDir,
      final String prefix,
      final String suffix,
      final long startBlockNumber,
      final long endBlockNumber) {
    Path traceFile;
    try {
      FileAttribute<Set<PosixFilePermission>> perms =
          PosixFilePermissions.asFileAttribute(PosixFilePermissions.fromString("rw-r--r--"));
      traceFile = Files.createTempFile(rootDir, prefix, suffix, perms);
    } catch (IOException e) {
      log.error(
          "Error while creating tmp file {} {} {}. Trying without setting the permissions",
          rootDir,
          prefix,
          suffix);
      try {
        traceFile = Files.createTempFile(rootDir, prefix, suffix);
      } catch (IOException f) {
        log.error("Still Failing while creating tmp file {} {} {}", rootDir, prefix, suffix);
        throw new RuntimeException(e);
      }
    }

    tracer.writeToFile(traceFile, startBlockNumber, endBlockNumber);

    return traceFile;
  }

  private Path generateOutputFilePath(final Path tracesOutputDirPath, final String traceFileName) {
    if (!Files.isDirectory(tracesOutputDirPath) && !tracesOutputDirPath.toFile().mkdirs()) {
      throw new RuntimeException(
          String.format(
              "Trace directory '%s' does not exist and could not be made.",
              tracesOutputDirPath.toAbsolutePath()));
    }

    return tracesOutputDirPath.resolve(Paths.get(traceFileName));
  }

  private String generateOutputFileName(
      final long startBlockNumber,
      final long endBlockNumber,
      final String expectedTracesEngineVersion) {
    return "%s-%s.conflated.%s"
        .formatted(startBlockNumber, endBlockNumber, expectedTracesEngineVersion);
  }
}
