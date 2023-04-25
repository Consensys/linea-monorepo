/*
 * Copyright ConsenSys AG.
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
package net.consensys.linea.zktracer.corset;

import static java.nio.file.StandardOpenOption.WRITE;

import java.io.IOException;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.concurrent.TimeUnit;

import org.apache.commons.io.IOUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class CorsetValidator {
  private static final Logger LOG = LoggerFactory.getLogger(CorsetValidator.class);
  private static final String ZK_EVM_BIN_ENV = "ZK_EVM_BIN";

  private static String CORSET_BIN;
  private static String ZK_EVM_BIN;

  static {
    init();
  }

  private static void init() {
    ZK_EVM_BIN = System.getenv(ZK_EVM_BIN_ENV);
    if (ZK_EVM_BIN == null) {
      LOG.error("Environment variable " + ZK_EVM_BIN_ENV + " is not set");
      throw new RuntimeException("Environment variable " + ZK_EVM_BIN_ENV + " is not set");
    }

    final Process whichCorsetProcess;

    try {
      whichCorsetProcess = Runtime.getRuntime().exec(new String[] {"which", "corset"});
    } catch (IOException e) {
      LOG.error("Error while searching for corset" + e.getMessage());
      throw new RuntimeException(e);
    }

    final String whichCorsetProcessOutput;
    try {
      whichCorsetProcessOutput =
          IOUtils.toString(whichCorsetProcess.getInputStream(), Charset.defaultCharset());
    } catch (IOException e) {
      LOG.error("Error while catching output whichCorsetProcess: " + e.getMessage());
      throw new RuntimeException(e);
    }

    try {
      whichCorsetProcess.waitFor(5, TimeUnit.SECONDS);
    } catch (InterruptedException e) {
      LOG.error("Timeout while searching for corset: " + e.getMessage());
      throw new RuntimeException(e);
    }

    if (whichCorsetProcess.exitValue() != 0) {
      LOG.error("Can't run corset executable with path: " + whichCorsetProcessOutput);
      throw new RuntimeException();
    }

    CORSET_BIN = whichCorsetProcessOutput.trim();
  }

  public static boolean isValid(final String trace) {
    final Path traceFile;

    try {
      traceFile = Files.createTempFile("", ".tmp.json");
      LOG.info("Trace file: " + traceFile.toAbsolutePath());
    } catch (IOException e) {
      LOG.error("Can't create temporary trace file: " + e.getMessage());
      throw new RuntimeException(e);
    }

    try {
      Files.write(traceFile, trace.getBytes(StandardCharsets.UTF_8), WRITE);
    } catch (IOException e) {
      LOG.error("Can't write to temporary trace file: " + e.getMessage());
      throw new RuntimeException(e);
    }

    final Process corsetValidationProcess;
    try {
      corsetValidationProcess =
          Runtime.getRuntime()
              .exec(
                  new String[] {
                    CORSET_BIN,
                    "check",
                    "-T",
                    traceFile.toFile().getAbsolutePath(),
                    "-v",
                    ZK_EVM_BIN
                  });
    } catch (IOException e) {
      LOG.error("Corset validation has thrown  an exception: " + e.getMessage());
      throw new RuntimeException(e);
    }

    final String corsetStdOutput;
    final String corsetErrorOutput;
    try {
      corsetStdOutput =
          IOUtils.toString(corsetValidationProcess.getInputStream(), Charset.defaultCharset());
      corsetErrorOutput =
          IOUtils.toString(corsetValidationProcess.getErrorStream(), Charset.defaultCharset());
    } catch (IOException e) {
      LOG.error("Error while catching output corsetValidationProcess: " + e.getMessage());
      throw new RuntimeException(e);
    }

    try {
      corsetValidationProcess.waitFor(5, TimeUnit.SECONDS);
    } catch (InterruptedException e) {
      LOG.error("Timeout while validating trace file: " + e.getMessage());
      throw new RuntimeException(e);
    }

    if (corsetValidationProcess.exitValue() != 0) {
      LOG.error("Validation failed: " + corsetStdOutput);
      LOG.error(corsetErrorOutput);
      return false;
    }

    return true;
  }
}
