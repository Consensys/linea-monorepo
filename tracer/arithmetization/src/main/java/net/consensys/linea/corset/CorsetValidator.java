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

package net.consensys.linea.corset;

import static java.nio.file.StandardOpenOption.WRITE;

import java.io.IOException;
import java.nio.charset.Charset;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.concurrent.TimeUnit;

import lombok.extern.slf4j.Slf4j;
import org.apache.commons.io.IOUtils;

@Slf4j
public class CorsetValidator {
  private static String CORSET_BIN;
  private static final String ZK_EVM_BIN = "../zkevm-constraints/zkevm.bin";

  static {
    init();
  }

  private static void init() {
    final Process whichCorsetProcess;

    try {
      whichCorsetProcess = Runtime.getRuntime().exec(new String[] {"which", "corset"});
    } catch (IOException e) {
      log.error("Error while searching for corset: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }

    final String whichCorsetProcessOutput;
    try {
      whichCorsetProcessOutput =
          IOUtils.toString(whichCorsetProcess.getInputStream(), Charset.defaultCharset());
    } catch (IOException e) {
      log.error("Error while catching output whichCorsetProcess: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }

    try {
      whichCorsetProcess.waitFor(5, TimeUnit.SECONDS);
    } catch (InterruptedException e) {
      log.error("Timeout while searching for corset: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }

    if (whichCorsetProcess.exitValue() != 0) {
      log.error("Cannot run corset executable with path: %s".formatted(whichCorsetProcessOutput));
      throw new RuntimeException();
    }

    CORSET_BIN = whichCorsetProcessOutput.trim();
  }

  public static boolean isValid(final String trace) {
    final Path traceFile;

    try {
      traceFile = Files.createTempFile("", ".tmp.json");
      log.info("Trace file: %s".formatted(traceFile.toAbsolutePath()));
    } catch (IOException e) {
      log.error("Can't create temporary trace file: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }

    try {
      Files.writeString(traceFile, trace, WRITE);
    } catch (IOException e) {
      log.error("Cannot write to temporary trace file: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }

    final Process corsetValidationProcess;
    try {
      corsetValidationProcess =
          new ProcessBuilder(
                  CORSET_BIN,
                  "check",
                  "-T",
                  traceFile.toFile().getAbsolutePath(),
                  "-q",
                  "-r",
                  "-d",
                  "-s",
                  "-t",
                  "2",
                  ZK_EVM_BIN)
              .redirectInput(ProcessBuilder.Redirect.INHERIT)
              .redirectErrorStream(true)
              .start();
    } catch (IOException e) {
      log.error("Corset validation has thrown an exception: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }

    final String corsetOutput;
    try {
      corsetOutput =
          IOUtils.toString(corsetValidationProcess.getInputStream(), Charset.defaultCharset());
    } catch (IOException e) {
      log.error(
          "Error while catching output corsetValidationProcess: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }

    try {
      corsetValidationProcess.waitFor(5, TimeUnit.SECONDS);
    } catch (InterruptedException e) {
      log.error("Timeout while validating trace file: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }

    if (corsetValidationProcess.exitValue() != 0) {
      log.error("Validation failed: %s".formatted(corsetOutput));
      return false;
    }

    return true;
  }
}
