/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.modulelimit;

import com.google.common.io.Resources;
import java.io.File;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.Map;
import java.util.function.Function;
import java.util.stream.Collectors;
import lombok.extern.slf4j.Slf4j;
import org.apache.tuweni.toml.Toml;
import org.apache.tuweni.toml.TomlParseResult;
import org.apache.tuweni.toml.TomlTable;

/**
 * Verifies line counts for modules based on provided limits. It supports verifying whether current
 * transaction exceed these limits.
 */
@Slf4j
public class ModuleLineCountValidator {
  private final Map<String, Integer> moduleLineCountLimits;

  /**
   * Constructs a new accumulator with specified module line count limits.
   *
   * @param moduleLineCountLimits A map of module names to their respective line count limits.
   */
  public ModuleLineCountValidator(Map<String, Integer> moduleLineCountLimits) {
    this.moduleLineCountLimits = Map.copyOf(moduleLineCountLimits);
  }

  /**
   * Verifies if the current accumulated line counts for modules exceed the predefined limits.
   *
   * @param currentAccumulatedLineCounts A map of module names to their current accumulated line
   *     counts.
   * @return A {@link ModuleLimitsValidationResult} indicating the outcome of the verification.
   */
  public ModuleLimitsValidationResult validate(
      final Map<String, Integer> currentAccumulatedLineCounts) {
    return validate(currentAccumulatedLineCounts, initialLineCountLimits());
  }

  /**
   * Verifies whether the current accumulated line counts, against previous accumulation line
   * counts, for modules exceed the predefined limits.
   *
   * @param currentAccumulatedLineCounts A map of module names to their current accumulated line
   *     counts.
   * @param prevAccumulatedLineCounts A map with previous accumulated line counts.
   * @return A {@link ModuleLimitsValidationResult} indicating the outcome of the verification.
   */
  public ModuleLimitsValidationResult validate(
      final Map<String, Integer> currentAccumulatedLineCounts,
      final Map<String, Integer> prevAccumulatedLineCounts) {
    for (Map.Entry<String, Integer> moduleEntry : currentAccumulatedLineCounts.entrySet()) {
      final String moduleName = moduleEntry.getKey();
      final int currentTotalLineCountForModule = moduleEntry.getValue();
      if (currentTotalLineCountForModule < 0) {
        log.error(
            "Negative line count {} returned for module '{}'.",
            currentAccumulatedLineCounts,
            moduleName);
        return ModuleLimitsValidationResult.invalidLineCount(
            moduleName, currentTotalLineCountForModule);
      }
      final Integer lineCountLimitForModule = moduleLineCountLimits.get(moduleName);

      if (lineCountLimitForModule == null) {
        log.error("Module '{}' is not defined in limits config.", moduleName);
        return ModuleLimitsValidationResult.moduleNotDefined(moduleName);
      }

      final int lineCountAddedByCurrentTx =
          currentTotalLineCountForModule - prevAccumulatedLineCounts.get(moduleName);

      if (lineCountAddedByCurrentTx > lineCountLimitForModule) {
        return ModuleLimitsValidationResult.txModuleLineCountOverflow(
            moduleName,
            lineCountAddedByCurrentTx,
            lineCountLimitForModule,
            currentTotalLineCountForModule,
            lineCountLimitForModule);
      }

      if (currentTotalLineCountForModule > lineCountLimitForModule) {
        return ModuleLimitsValidationResult.blockModuleLineCountFull(
            moduleName,
            lineCountAddedByCurrentTx,
            lineCountLimitForModule,
            currentTotalLineCountForModule,
            lineCountLimitForModule);
      }
    }
    return ModuleLimitsValidationResult.VALID;
  }

  private Map<String, Integer> initialLineCountLimits() {
    return moduleLineCountLimits.keySet().stream()
        .collect(Collectors.toMap(Function.identity(), unused -> 0));
  }

  /** Enumerates possible outcomes of verifying module line counts against their limits. */
  public enum ModuleLineCountResult {
    VALID,
    TX_MODULE_LINE_COUNT_OVERFLOW,
    BLOCK_MODULE_LINE_COUNT_FULL,
    MODULE_NOT_DEFINED,
    INVALID_LINE_COUNT
  }

  public static Map<String, Integer> createLimitModules(String moduleLimitsFilePath) {
    try {
      URL url = new File(moduleLimitsFilePath).toURI().toURL();
      final String tomlString = Resources.toString(url, StandardCharsets.UTF_8);
      TomlParseResult result = Toml.parse(tomlString);
      final TomlTable table = result.getTable("traces-limits");
      final Map<String, Integer> limitsMap =
          table.toMap().entrySet().stream()
              .collect(
                  Collectors.toUnmodifiableMap(
                      Map.Entry::getKey, e -> Math.toIntExact((Long) e.getValue())));

      return limitsMap;
    } catch (final Exception e) {
      final String errorMsg =
          "Problem reading the toml file containing the limits for the modules: "
              + moduleLimitsFilePath;
      log.error(errorMsg);
      throw new RuntimeException(errorMsg, e);
    }
  }
}
