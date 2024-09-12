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

package net.consensys.linea.config;

import java.net.URI;

import com.google.common.base.MoreObjects;
import jakarta.validation.constraints.Positive;
import net.consensys.linea.plugins.LineaCliOptions;
import picocli.CommandLine;

/** The Linea Transaction Selector CLI options. */
public class LineaTransactionSelectorCliOptions implements LineaCliOptions {
  public static final String CONFIG_KEY = "transaction-selector-config";

  public static final String MAX_BLOCK_CALLDATA_SIZE = "--plugin-linea-max-block-calldata-size";
  public static final int DEFAULT_MAX_BLOCK_CALLDATA_SIZE = 70_000;
  public static final String OVER_LINE_COUNT_LIMIT_CACHE_SIZE =
      "--plugin-linea-over-line-count-limit-cache-size";
  public static final int DEFAULT_OVER_LINE_COUNT_LIMIT_CACHE_SIZE = 10_000;

  public static final String MAX_GAS_PER_BLOCK = "--plugin-linea-max-block-gas";
  public static final long DEFAULT_MAX_GAS_PER_BLOCK = 30_000_000L;

  public static final String UNPROFITABLE_CACHE_SIZE = "--plugin-linea-unprofitable-cache-size";
  public static final int DEFAULT_UNPROFITABLE_CACHE_SIZE = 100_000;

  public static final String UNPROFITABLE_RETRY_LIMIT = "--plugin-linea-unprofitable-retry-limit";
  public static final int DEFAULT_UNPROFITABLE_RETRY_LIMIT = 10;

  public static final String REJECTED_TX_ENDPOINT = "--plugin-linea-rejected-tx-endpoint";

  @Positive
  @CommandLine.Option(
      names = {MAX_BLOCK_CALLDATA_SIZE},
      hidden = true,
      paramLabel = "<INTEGER>",
      description = "Maximum size for the calldata of a block (default: ${DEFAULT-VALUE})")
  private int maxBlockCallDataSize = DEFAULT_MAX_BLOCK_CALLDATA_SIZE;

  @Positive
  @CommandLine.Option(
      names = {OVER_LINE_COUNT_LIMIT_CACHE_SIZE},
      hidden = true,
      paramLabel = "<INTEGER>",
      description =
          "Max number of transactions that go over the line count limit we keep track of (default: ${DEFAULT-VALUE})")
  private int overLineCountLimitCacheSize = DEFAULT_OVER_LINE_COUNT_LIMIT_CACHE_SIZE;

  @Positive
  @CommandLine.Option(
      names = {MAX_GAS_PER_BLOCK},
      hidden = true,
      paramLabel = "<LONG>",
      description = "Sets max gas per block  (default: ${DEFAULT-VALUE})")
  private Long maxGasPerBlock = DEFAULT_MAX_GAS_PER_BLOCK;

  @Positive
  @CommandLine.Option(
      names = {UNPROFITABLE_CACHE_SIZE},
      hidden = true,
      paramLabel = "<INTEGER>",
      description =
          "Max number of unprofitable transactions we keep track of (default: ${DEFAULT-VALUE})")
  private int unprofitableCacheSize = DEFAULT_UNPROFITABLE_CACHE_SIZE;

  @Positive
  @CommandLine.Option(
      names = {UNPROFITABLE_RETRY_LIMIT},
      hidden = true,
      paramLabel = "<INTEGER>",
      description =
          "Max number of unprofitable transactions we retry on each block creation (default: ${DEFAULT-VALUE})")
  private int unprofitableRetryLimit = DEFAULT_UNPROFITABLE_RETRY_LIMIT;

  @CommandLine.Option(
      names = {REJECTED_TX_ENDPOINT},
      hidden = true,
      paramLabel = "<URI>",
      description = "Endpoint URI for reporting rejected transactions  (default: ${DEFAULT-VALUE})")
  private URI rejectedTxEndpoint = null;

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
    options.overLineCountLimitCacheSize = config.overLinesLimitCacheSize();
    options.maxGasPerBlock = config.maxGasPerBlock();
    options.unprofitableCacheSize = config.unprofitableCacheSize();
    options.unprofitableRetryLimit = config.unprofitableRetryLimit();
    options.rejectedTxEndpoint = config.rejectedTxEndpoint();
    return options;
  }

  /**
   * To domain object Linea factory configuration.
   *
   * @return the Linea factory configuration
   */
  @Override
  public LineaTransactionSelectorConfiguration toDomainObject() {
    return LineaTransactionSelectorConfiguration.builder()
        .maxBlockCallDataSize(maxBlockCallDataSize)
        .overLinesLimitCacheSize(overLineCountLimitCacheSize)
        .maxGasPerBlock(maxGasPerBlock)
        .unprofitableCacheSize(unprofitableCacheSize)
        .unprofitableRetryLimit(unprofitableRetryLimit)
        .rejectedTxEndpoint(rejectedTxEndpoint)
        .build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(MAX_BLOCK_CALLDATA_SIZE, maxBlockCallDataSize)
        .add(OVER_LINE_COUNT_LIMIT_CACHE_SIZE, overLineCountLimitCacheSize)
        .add(MAX_GAS_PER_BLOCK, maxGasPerBlock)
        .add(UNPROFITABLE_CACHE_SIZE, unprofitableCacheSize)
        .add(UNPROFITABLE_RETRY_LIMIT, unprofitableRetryLimit)
        .add(REJECTED_TX_ENDPOINT, rejectedTxEndpoint)
        .toString();
  }
}
