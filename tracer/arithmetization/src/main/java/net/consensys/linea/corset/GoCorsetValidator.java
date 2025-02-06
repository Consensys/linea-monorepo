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

import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator.Result;

/**
 * Responsible for running the command-line <code>go-corset</code> tool to check that a given trace
 * is accepted by the zkevm constraints. The <code>corset</code> tool has three IR levels at which
 * it can do this (HIR/MIR/AIR). The lowest level (AIR) provides the most accurate results (i.e.
 * most comparable with prover).
 *
 * <p>The configuration can be set using the environment variable <code>GO_CORSET_FLAGS</code>. If
 * this is not set, then this validator will not be activated. </code>.
 */
@Slf4j
public class GoCorsetValidator extends AbstractExecutable {
  /** Indicates whether or not this validator is active (i.e. we located the go-corset binary). */
  @Getter private boolean active = false;

  public GoCorsetValidator() {
    configGoCorset();
  }

  /**
   * Attempt to validate a given trace against a given zkEvmBin file.
   *
   * @param traceFile Path to the trace file being validated.
   * @param zkEvmBin Path to the zkEvmBin file being validated.
   * @return A result which tells us whether or not the trace file was accepted, and provides
   *     additional information for debugging purposes.
   */
  public Result validate(final Path traceFile, final String zkEvmBin) {
    if (active) {
      Outcome outcome;
      try {
        List<String> commands = buildCommandLine(traceFile, zkEvmBin);
        log.debug("{}", commands);
        // Execute corset with a 5s timeout.
        outcome = super.exec(5, commands);
      } catch (Throwable e) {
        log.error("Corset validation has thrown an exception: %s".formatted(e.getMessage()));
        throw new RuntimeException(e);
      }
      // Check for success or failure
      if (outcome.exitcode() != 0) {
        log.error("Validation failed: %s".formatted(outcome.output()));
        return new Result(false, traceFile.toFile(), outcome.output());
      }
      // success!
      return new Result(true, traceFile.toFile(), outcome.output());
    }
    // Tool is not active
    log.debug("(inactive)");
    return null;
  }

  /**
   * Configure corset from the <code>CORSET_FLAGS</code> environment variable (if set). If the
   * environment variable is not set, the default configuration is retained. If the environment
   * variable is set, but its value is malformed then an exception is raised.
   */
  private void configGoCorset() {
    // If we can execute go-corset then use it!
    this.active = super.isExecutable("go-corset", "--help");
  }

  /**
   * Construct the command-line to use for running corset to check a given trace file is accepted
   * (or not) by a given bin file. Amongst other things, this will configure the command-line flags
   * as dictated by the CORSET_FLAGS environment variable.
   *
   * @return
   */
  private List<String> buildCommandLine(Path traceFile, String zkEvmBin) {
    ArrayList<String> options = new ArrayList<>();
    // Determine options to use (either default or override)
    String flags =
        System.getenv().getOrDefault("GOCORSET_FLAGS", "-w --report --report-context 2 --hir ");
    // Specify corset binary
    options.add("go-corset");
    // Specify corset "check" command.
    options.add("check");
    // Add all options
    options.addAll(Arrays.asList(flags.split(" ")));
    // Specify trace file to use
    options.add(traceFile.toAbsolutePath().toString());
    // Specify the zkevm.bin file.
    options.add(zkEvmBin);
    // Done
    return options;
  }
}
