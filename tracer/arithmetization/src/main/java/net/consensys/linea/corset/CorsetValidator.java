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
import java.nio.file.Path;

import lombok.extern.slf4j.Slf4j;

/**
 * Responsible for running the command-line <code>corset</code> tool to check that a given trace is
 * accepted by the zkevm constraints. The <code>corset</code> tool has variables levels of
 * "expansion" which it can apply before performing the check. Greater levels of expansion imply
 * more accurate checks (i.e. more realistic compared to the prover). Furthermore, <code>corset
 * </code> can be configured to use field arithmetic or simply big integers (with the latter option
 * intended to offer faster but much less precise checking). The default configuration is set to
 * give good accuracy, but without significant overhead. A greater level can be configured by
 * enabling the <code>autoConstraints</code>.
 *
 * <p>The configuration can be set using the environment variable <code>CORSET_FLAGS</code>. Example
 * values for this environment variable include <code>fields,expand</code> (sets <code>
 * fieldArithmetic=true</code> and <code>expansion=0</code>) and <code>fields,expand,expand,
 * auto</code> (enables <code>fieldArithmetic</code> and <code>autoConstraints</code> and sets
 * <code>expansion=2</code>). Note, it doesn't make sense to have <code>expand</code> without <code>
 * fields</code>. Likewise, it doesn't make sense to have <code>auto</code> without <code>expand
 * </code>.
 */
@Slf4j
public class CorsetValidator {
  public record Result(boolean isValid, File traceFile, String corsetOutput) {}

  /** */
  private static final String ZK_EVM_RELATIVE_PATH = "/linea-constraints/";

  private static final String ZK_EVM_BIN = "zkevm.bin";

  /** Specifies the default zkEVM.bin file to use (including its path). */
  private String defaultZkEvm = null;

  /** Interface to Go corset tool. */
  private final GoCorsetValidator goCorset;

  public CorsetValidator() {
    // Construct and initialise Go corset.
    this.goCorset = new GoCorsetValidator();
    // Configure default path to the zkevm.bin file.
    initDefaultZkEvm();
  }

  /**
   * Attempt to validate a given tracefile against a given set of zkEVM constraints. A default
   * location for the zkevm.bin file is used.
   *
   * @param traceFile The tracefile being validated.
   * @return A result which tells us whether or not the trace file was accepted, and provides
   *     additional information for debugging purposes.
   */
  public Result validate(final Path traceFile) throws RuntimeException {
    if (defaultZkEvm == null) {
      throw new IllegalArgumentException("Default zkevm.bin not set.");
    }
    return validate(traceFile, defaultZkEvm);
  }

  /**
   * Attempt to validate a given tracefile against a given set of zkEVM constraints.
   *
   * @param traceFile The tracefile being validated.
   * @param zkEvmBin The zkEVM constraints file (compiled using corset).
   * @return A result which tells us whether or not the trace file was accepted, and provides
   *     additional information for debugging purposes.
   */
  public Result validate(final Path traceFile, final String zkEvmBin) {
    // Generate result from Go corset tool.
    Result rg = goCorset.validate(traceFile, zkEvmBin);
    // Sanity check at least one validator is active
    if (!goCorset.isActive()) {
      throw new RuntimeException("go-corset not available");
    } else {
      return rg;
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

    final String zkEvmBinInCurrentDir = currentDir + ZK_EVM_RELATIVE_PATH + binName();
    if (new File(zkEvmBinInCurrentDir).exists()) {
      defaultZkEvm = zkEvmBinInCurrentDir;
      return;
    }

    final String zkEvmBinInDirAbove = currentDir + "/.." + ZK_EVM_RELATIVE_PATH + binName();
    if (new File(zkEvmBinInDirAbove).exists()) {
      defaultZkEvm = zkEvmBinInDirAbove;
      return;
    }

    log.warn("Could not find default path for {}", binName());
  }

  private String binName() {
    return System.getenv("ZKEVM_BIN") != null ? System.getenv("ZKEVM_BIN") : ZK_EVM_BIN;
  }
}
