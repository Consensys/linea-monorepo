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
package net.consensys.linea.sequencer;

import com.google.common.base.MoreObjects;
import picocli.CommandLine;

/** The RocksDb cli options. */
public class LineaCLIOptions {

  public static final int DEFAULT_MAX_TX_CALLDATA_SIZE = 61440;
  public static final int DEFAULT_MAX_BLOCK_CALLDATA_SIZE = 71680;
  public static final String MAX_TX_CALLDATA_SIZE = "--plugin-linea-max-tx-calldata-size";
  public static final String MAX_BLOCK_CALLDATA_SIZE = "--plugin-linea-max-block-calldata-size";

  @CommandLine.Option(
      names = {MAX_TX_CALLDATA_SIZE},
      hidden = true,
      defaultValue = "61440",
      paramLabel = "<INTEGER>",
      description = "Maximum size for the calldata of a Transaction (default: ${DEFAULT-VALUE})")
  int maxTxCalldataSize;

  @CommandLine.Option(
      names = {MAX_BLOCK_CALLDATA_SIZE},
      hidden = true,
      defaultValue = "71680",
      paramLabel = "<INTEGER>",
      description = "Maximum size for the calldata of a Block (default: ${DEFAULT-VALUE})")
  int maxBlockCalldataSize;

  private LineaCLIOptions() {}

  /**
   * Create Linea cli options.
   *
   * @return the Linea cli options
   */
  public static LineaCLIOptions create() {
    return new LineaCLIOptions();
  }

  /**
   * Linea cli options from config.
   *
   * @param config the config
   * @return the Linea cli options
   */
  public static LineaCLIOptions fromConfig(final LineaConfiguration config) {
    final LineaCLIOptions options = create();
    options.maxTxCalldataSize = config.getMaxTxCalldataSize();
    options.maxBlockCalldataSize = config.getMaxBlockCalldataSize();
    return options;
  }

  /**
   * To domain object rocks db factory configuration.
   *
   * @return the rocks db factory configuration
   */
  public LineaConfiguration toDomainObject() {
    return new LineaConfiguration(maxTxCalldataSize, maxBlockCalldataSize);
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(MAX_TX_CALLDATA_SIZE, maxTxCalldataSize)
        .add(MAX_BLOCK_CALLDATA_SIZE, maxBlockCalldataSize)
        .toString();
  }
}
