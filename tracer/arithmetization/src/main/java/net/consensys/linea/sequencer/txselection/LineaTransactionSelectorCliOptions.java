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

package net.consensys.linea.sequencer.txselection;

import com.google.common.base.MoreObjects;
import picocli.CommandLine;

/** The Linea CLI options. */
public class LineaTransactionSelectorCliOptions {
  public static final int DEFAULT_MAX_BLOCK_CALLDATA_SIZE = 70000;
  private static final String DEFAULT_MODULE_LIMIT_FILE_PATH = "moduleLimitFile.json";

  private static final String MAX_BLOCK_CALLDATA_SIZE = "--plugin-linea-max-block-calldata-size";
  private static final String MODULE_LIMIT_FILE_PATH = "--plugin-linea-module-limit-file-path";

  @CommandLine.Option(
      names = {MAX_BLOCK_CALLDATA_SIZE},
      hidden = true,
      paramLabel = "<INTEGER>",
      description =
          "Maximum size for the calldata of a Block (default: "
              + DEFAULT_MAX_BLOCK_CALLDATA_SIZE
              + ")")
  private int maxBlockCallDataSize = DEFAULT_MAX_BLOCK_CALLDATA_SIZE;

  @CommandLine.Option(
      names = {MODULE_LIMIT_FILE_PATH},
      hidden = true,
      paramLabel = "<STRING>",
      description =
          "Path to the json file containing the module limits (default: "
              + DEFAULT_MODULE_LIMIT_FILE_PATH
              + ")")
  private String moduleLimitFilePath = DEFAULT_MODULE_LIMIT_FILE_PATH;

  private LineaTransactionSelectorCliOptions() {}

  /**
   * Create Linea cli options.
   *
   * @return the Linea cli options
   */
  public static LineaTransactionSelectorCliOptions create() {
    return new LineaTransactionSelectorCliOptions();
  }

  /**
   * Linea cli options from config.
   *
   * @param config the config
   * @return the Linea cli options
   */
  public static LineaTransactionSelectorCliOptions fromConfig(
      final LineaTransactionSelectorConfiguration config) {
    final LineaTransactionSelectorCliOptions options = create();
    options.maxBlockCallDataSize = config.maxBlockCallDataSize();
    options.moduleLimitFilePath = config.moduleLimitsFilePath();
    return options;
  }

  /**
   * To domain object Linea factory configuration.
   *
   * @return the Linea factory configuration
   */
  public LineaTransactionSelectorConfiguration toDomainObject() {
    return new LineaTransactionSelectorConfiguration.Builder()
        .maxBlockCallDataSize(maxBlockCallDataSize)
        .moduleLimits(moduleLimitFilePath)
        .build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(MAX_BLOCK_CALLDATA_SIZE, maxBlockCallDataSize)
        .add(MODULE_LIMIT_FILE_PATH, moduleLimitFilePath)
        .toString();
  }
}
