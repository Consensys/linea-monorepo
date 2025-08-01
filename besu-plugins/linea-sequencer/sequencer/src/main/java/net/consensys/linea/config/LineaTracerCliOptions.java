/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.config;

import com.google.common.base.MoreObjects;
import net.consensys.linea.plugins.LineaCliOptions;
import net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator;
import picocli.CommandLine;

public class LineaTracerCliOptions implements LineaCliOptions {
  public static final String CONFIG_KEY = "tracer-sequencer-config";

  public static final String MODULE_LIMIT_FILE_PATH = "--plugin-linea-module-limit-file-path";
  public static final String DEFAULT_MODULE_LIMIT_FILE_PATH = "moduleLimitFile.toml";

  @CommandLine.Option(
      names = {MODULE_LIMIT_FILE_PATH},
      hidden = true,
      paramLabel = "<STRING>",
      description =
          "Path to the toml file containing the module limits (default: ${DEFAULT-VALUE})")
  private String moduleLimitFilePath = DEFAULT_MODULE_LIMIT_FILE_PATH;

  private LineaTracerCliOptions() {}

  /**
   * Create Linea cli options.
   *
   * @return the Linea cli options
   */
  public static LineaTracerCliOptions create() {
    return new LineaTracerCliOptions();
  }

  /**
   * Linea cli options from config.
   *
   * @param config the config
   * @return the Linea cli options
   */
  public static LineaTracerCliOptions fromConfig(final LineaTracerConfiguration config) {
    final LineaTracerCliOptions options = create();
    options.moduleLimitFilePath = config.moduleLimitsFilePath();
    return options;
  }

  /**
   * To domain object Linea factory configuration.
   *
   * @return the Linea factory configuration
   */
  @Override
  public LineaTracerLineLimitConfiguration toDomainObject() {
    return LineaTracerLineLimitConfiguration.builder()
        .moduleLimitsFilePath(moduleLimitFilePath)
        .moduleLimitsMap(ModuleLineCountValidator.createLimitModules(moduleLimitFilePath))
        .build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(MODULE_LIMIT_FILE_PATH, moduleLimitFilePath)
        .toString();
  }
}
