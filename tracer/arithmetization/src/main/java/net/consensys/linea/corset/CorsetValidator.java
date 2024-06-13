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
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.TimeUnit;

import lombok.Getter;
import lombok.Setter;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.io.IOUtils;

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

  private static final String ZK_EVM_RELATIVE_PATH = "/zkevm-constraints/zkevm.bin";

  /** Specifies the default zkEVM.bin file to use (including its path). */
  private String defaultZkEvm = null;

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

  public CorsetValidator() {
    initCorset();
    configCorset();
    initDefaultZkEvm();
  }

  public Result validate(final Path filename) throws RuntimeException {
    if (defaultZkEvm == null) {
      throw new IllegalArgumentException("Default zkevm.bin not set.");
    }
    return validate(filename, defaultZkEvm);
  }

  public Result validate(final Path filename, final String zkEvmBin) throws RuntimeException {
    final Process corsetValidationProcess;
    try {
      List<String> options = buildOptions(filename, zkEvmBin);
      corsetValidationProcess =
          new ProcessBuilder(options)
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
      // Check for default case (empty string)
      if (!flags.isEmpty()) {
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
              // Error
              throw new RuntimeException("Unknown Corset configuration flag: %s".formatted(flag));
          }
        }
      }
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

  /**
   * Construct the list of options to be used when running Corset.
   *
   * @return
   */
  private List<String> buildOptions(Path filename, String zkEvmBin) {
    ArrayList<String> options = new ArrayList<>();
    // Specify corset binary
    options.add(corsetBin);
    // Specify corset "check" command.
    options.add("check");
    // Specify corset trace file to use
    options.add("-T");
    options.add(filename.toAbsolutePath().toString());
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
    // Specify number of threads to use.
    options.add("-t");
    options.add(determineNumberOfThreads());
    // Specify the zkevm.bin file.
    options.add(zkEvmBin);
    log.info("Corset options: " + options);
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
    return Optional.ofNullable(System.getenv("CORSET_THREADS")).orElse("2");
  }
}
