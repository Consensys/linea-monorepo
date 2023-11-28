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

import java.io.File;
import java.io.IOException;
import java.nio.charset.Charset;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Optional;
import java.util.concurrent.TimeUnit;

import lombok.extern.slf4j.Slf4j;
import org.apache.commons.io.IOUtils;

@Slf4j
public class CorsetValidator {
  public record Result(boolean isValid, File traceFile, String corsetOutput) {}

  private static final String ZK_EVM_RELATIVE_PATH = "/zkevm-constraints/zkevm.bin";

  private String defaultZkEvm = null;
  private String corsetBin;

  public CorsetValidator() {
    initCorset();
    initDefaultZkEvm();
  }

  public Result validate(final Path filename) throws RuntimeException {
    return validate(filename, defaultZkEvm);
  }

  public Result validate(final Path filename, final String zkEvmBin) throws RuntimeException {
    final Process corsetValidationProcess;
    try {
      corsetValidationProcess =
          new ProcessBuilder(
                  corsetBin,
                  "check",
                  "-T",
                  filename.toAbsolutePath().toString(),
                  "-q",
                  "-r",
                  "-d",
                  "-s",
                  "-t",
                  Optional.ofNullable(System.getenv("CORSET_THREADS")).orElse("2"),
                  zkEvmBin)
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
      return new Result(false, filename.toFile(), corsetOutput);
    }

    return new Result(true, filename.toFile(), corsetOutput);
  }

  private void initCorset() {
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

    if (whichCorsetProcess.exitValue() == 0) {
      corsetBin = whichCorsetProcessOutput.trim();
      return;
    }

    log.warn("Could not find corset executable: %s".formatted(whichCorsetProcessOutput));

    final String homePath = System.getenv("HOME");
    corsetBin = homePath + "/.cargo/bin/corset";
    log.warn("Trying to use default corset path: %s".formatted(corsetBin));

    if (!Files.isExecutable(Path.of(corsetBin))) {
      throw new RuntimeException("Corset is not executable: %s".formatted(corsetBin));
    }
  }

  private void initDefaultZkEvm() {
    final String currentDir;

    try {
      currentDir = Path.of(".").toRealPath().toString();
    } catch (final IOException e) {
      log.error("Error while getting current directory: %s".formatted(e.getMessage()));
      throw new RuntimeException(e);
    }

    final String zkEvmBinInCurrentDir = currentDir + ZK_EVM_RELATIVE_PATH;
    if (new File(zkEvmBinInCurrentDir).exists()) {
      defaultZkEvm = zkEvmBinInCurrentDir;
      return;
    }

    final String zkEvmBinInDirAbove = currentDir + "/.." + ZK_EVM_RELATIVE_PATH;
    if (new File(zkEvmBinInDirAbove).exists()) {
      defaultZkEvm = zkEvmBinInDirAbove;
      return;
    }

    log.warn("Could not find default path for zkevm.bin");
  }
}
