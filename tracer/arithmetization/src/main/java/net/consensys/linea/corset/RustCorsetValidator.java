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
import java.util.List;
import java.util.Optional;

import lombok.Getter;
import lombok.Setter;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator.Result;

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
public class RustCorsetValidator extends AbstractExecutable {
  /** Specifies the Corset binary to use (including its path). */
  private String corsetBin;

  /**
   * Specifies whether field arithmetic should be used. To best reflect the prover, this should be
   * enabled.
   */
  @Getter @Setter private boolean fieldArithmetic = true;

  /**
   * Specifies how much expansion should be applied to constraints (max is 4). For example,
   * normalisation expressions are expanded into columns containing the multiplicative inverse, etc.
   * To best reflect the prover, this should be enabled. Also, the effect of this option is limited
   * unless field arithmetic is also enabled.
   */
  @Getter @Setter private int expansion = 4;

  /**
   * Specifies whether expansion of "auto constraints" should be performed. This is the final step
   * of expansion taking us to the lowest level. To best reflect the prover, this should be enabled.
   * However, it is not enabled by default because this imposes a high performance overhead.
   */
  @Getter @Setter private boolean autoConstraints = false;

  /**
   * Specifies the number of rows to show either side for a failing constraint. This can faciliate
   * debugging, since the greater the width the more information can be seen. At the same time,
   * however, too much information can make the report very hard to read.
   */
  @Getter @Setter private int reportWidth = 8;

  /** Indicates whether or not this validator is active (i.e. we located the corset binary). */
  @Getter private boolean active = false;

  public RustCorsetValidator() {
    // initCorset();
    configCorset();
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
   * Attempt to locate the corset executable. Ideally, this is on the PATH and works out-of-the-box.
   * However, this will fall back to a default location (based on where Cargo normally places
   * binaries) if that is not the case.
   */
  private void initCorset() {
    // Try default
    if (isExecutable("corset", "--version")) {
      corsetBin = "corset";
      active = true;
    } else {
      // Try fall back
      final String homePath = System.getenv("HOME");
      corsetBin = homePath + "/.cargo/bin/corset";
      log.warn("Trying to use default corset path: %s".formatted(corsetBin));
      // Sanity check we can execute it
      active = isExecutable(corsetBin, "--version");
    }
  }

  /**
   * Configure corset from the <code>CORSET_FLAGS</code> environment variable (if set). If the
   * environment variable is not set, the default configuration is retained. If the environment
   * variable is set, but its value is malformed then an exception is raised.
   */
  private void configCorset() {
    String flags = System.getenv().get("CORSET_FLAGS");
    if (flags != null) {
      log.info(
          "Configuring corset from CORSET_FLAGS environment variable: \"%s\"".formatted(flags));
      // Reset default configuration
      this.fieldArithmetic = false;
      this.expansion = 0;
      this.autoConstraints = false;
      //
      if (flags.equals("disable")) {
        // Special case used to disable corset even when it is available.
        active = false;
      } else if (!flags.isEmpty()) {
        // split flags by separator
        String[] splitFlags = flags.split(",");
        // Build configuration based on flags
        for (String flag : splitFlags) {
          switch (flag) {
            case "fields":
              this.fieldArithmetic = true;
              break;
            case "expand":
              if (expansion >= 4) {
                throw new RuntimeException(
                    ("Malformed Corset configuration flags (expansion "
                            + "beyond four meaningless): %s")
                        .formatted(flags));
              }
              this.expansion++;
              break;
            case "auto":
              this.autoConstraints = true;
              break;
            default:
              if (flag.startsWith("trace-span=")) {
                this.reportWidth = Integer.parseInt(flag.substring(11));
              } else {
                // Error
                throw new RuntimeException("Unknown Corset configuration flag: %s".formatted(flag));
              }
          }
        }
      }
    }
  }

  /**
   * Construct the command-line to use for running corset to check a given trace file is accepted
   * (or not) by a given bin file. Amongst other things, this will configure the command-line flags
   * as dictated by the CORSET_FLAGS environment variable.
   *
   * @return
   */
  private List<String> buildCommandLine(Path traceFile, String zkEvmBin) {
    List<String> options = new ArrayList<>();
    // Specify corset binary
    options.add(corsetBin);
    // Specify corset "check" command.
    options.add("check");
    // Specify corset trace file to use
    options.add("-T");
    options.add(traceFile.toAbsolutePath().toString());
    // Specify reporting options where:
    //
    // -q Decrease logging verbosity
    // -r detail failing constraint
    // -d dim unimportant expressions for failing constraints
    // -s display original source along with compiled form
    options.add("-qrds");
    // Enable field arithmetic (if applicable)
    if (fieldArithmetic) {
      options.add("-N");
    }
    // Enable expansion (if applicable)
    if (expansion != 0) {
      String es = "e".repeat(expansion);
      options.add("-" + es);
    }
    // Enable auto constraints (if applicable)
    if (autoConstraints) {
      options.add("--auto-constraints");
      options.add("nhood,sorts");
    }
    // Specify span width to use
    options.add("--trace-span");
    options.add(Integer.toString(this.reportWidth));
    // Specify number of threads to use.
    options.add("-t");
    options.add(determineNumberOfThreads());
    // Specify the zkevm.bin file.
    options.add(zkEvmBin);
    // Done
    return options;
  }

  /**
   * Determine the number of threads to use when checking constraints. The default is "2", but this
   * can be overriden using an environment variable <code>CORSET_THREADS</code>.
   *
   * @return
   */
  private String determineNumberOfThreads() {
    int ncpus = Runtime.getRuntime().availableProcessors();
    return Optional.ofNullable(System.getenv("CORSET_THREADS")).orElse(Integer.toString(ncpus));
  }
}
