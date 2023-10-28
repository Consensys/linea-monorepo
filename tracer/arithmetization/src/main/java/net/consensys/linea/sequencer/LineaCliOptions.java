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

package net.consensys.linea.sequencer;

import com.google.common.base.MoreObjects;
import picocli.CommandLine;

/** The Linea CLI options. */
public class LineaCliOptions {
  public static final int DEFAULT_MAX_TX_CALLDATA_SIZE = 60000;
  public static final int DEFAULT_MAX_BLOCK_CALLDATA_SIZE = 70000;

  private static final String MAX_TX_CALLDATA_SIZE = "--plugin-linea-max-tx-calldata-size";
  private static final String MAX_BLOCK_CALLDATA_SIZE = "--plugin-linea-max-block-calldata-size";

  @CommandLine.Option(
      names = {MAX_TX_CALLDATA_SIZE},
      hidden = true,
      paramLabel = "<INTEGER>",
      description =
          "Maximum size for the calldata of a Transaction (default: "
              + DEFAULT_MAX_TX_CALLDATA_SIZE
              + ")")
  private int maxTxCallDataSize = DEFAULT_MAX_TX_CALLDATA_SIZE;

  @CommandLine.Option(
      names = {MAX_BLOCK_CALLDATA_SIZE},
      hidden = true,
      paramLabel = "<INTEGER>",
      description =
          "Maximum size for the calldata of a Block (default: "
              + DEFAULT_MAX_BLOCK_CALLDATA_SIZE
              + ")")
  private int maxBlockCallDataSize = DEFAULT_MAX_BLOCK_CALLDATA_SIZE;

  private LineaCliOptions() {}

  /**
   * Create Linea cli options.
   *
   * @return the Linea cli options
   */
  public static LineaCliOptions create() {
    return new LineaCliOptions();
  }

  /**
   * Linea cli options from config.
   *
   * @param config the config
   * @return the Linea cli options
   */
  public static LineaCliOptions fromConfig(final LineaConfiguration config) {
    final LineaCliOptions options = create();
    options.maxTxCallDataSize = config.maxTxCallDataSize();
    options.maxBlockCallDataSize = config.maxBlockCallDataSize();
    return options;
  }

  /**
   * To domain object Linea factory configuration.
   *
   * @return the Linea factory configuration
   */
  public LineaConfiguration toDomainObject() {
    return new LineaConfiguration.Builder()
        .maxTxCallDataSize(maxTxCallDataSize)
        .maxBlockCallDataSize(maxBlockCallDataSize)
        .build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(MAX_TX_CALLDATA_SIZE, maxTxCallDataSize)
        .add(MAX_BLOCK_CALLDATA_SIZE, maxBlockCallDataSize)
        .toString();
  }
}
