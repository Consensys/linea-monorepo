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

import java.math.BigDecimal;

import com.google.common.base.MoreObjects;
import jakarta.validation.constraints.Positive;
import picocli.CommandLine;

/** The Linea CLI options. */
public class LineaTransactionSelectorCliOptions {
  public static final int DEFAULT_MAX_BLOCK_CALLDATA_SIZE = 70_000;
  private static final String DEFAULT_MODULE_LIMIT_FILE_PATH = "moduleLimitFile.toml";
  public static final long DEFAULT_MAX_GAS_PER_BLOCK = 30_000_000L;
  public static final int DEFAULT_VERIFICATION_GAS_COST = 1_200_000;
  public static final int DEFAULT_VERIFICATION_CAPACITY = 90_000;
  public static final int DEFAULT_GAS_PRICE_RATIO = 15;
  public static final BigDecimal DEFAULT_MIN_MARGIN = BigDecimal.ONE;
  private static final String MAX_BLOCK_CALLDATA_SIZE = "--plugin-linea-max-block-calldata-size";
  private static final String MODULE_LIMIT_FILE_PATH = "--plugin-linea-module-limit-file-path";
  private static final String MAX_GAS_PER_BLOCK = "--plugin-linea-max-block-gas";
  private static final String VERIFICATION_GAS_COST = "--plugin-linea-verification-gas-cost";
  private static final String VERIFICATION_CAPACITY = "--plugin-linea-verification-capacity";
  private static final String GAS_PRICE_RATIO = "--plugin-linea-gas-price-ratio";
  private static final String MIN_MARGIN = "--plugin-linea-min-margin";

  @Positive
  @CommandLine.Option(
      names = {MAX_BLOCK_CALLDATA_SIZE},
      hidden = true,
      paramLabel = "<INTEGER>",
      description = "Maximum size for the calldata of a block (default: ${DEFAULT-VALUE})")
  private int maxBlockCallDataSize = DEFAULT_MAX_BLOCK_CALLDATA_SIZE;

  @CommandLine.Option(
      names = {MODULE_LIMIT_FILE_PATH},
      hidden = true,
      paramLabel = "<STRING>",
      description =
          "Path to the toml file containing the module limits (default: ${DEFAULT-VALUE})")
  private String moduleLimitFilePath = DEFAULT_MODULE_LIMIT_FILE_PATH;

  @Positive
  @CommandLine.Option(
      names = {MAX_GAS_PER_BLOCK},
      hidden = true,
      paramLabel = "<LONG>",
      description = "Sets max gas per block  (default: ${DEFAULT-VALUE})")
  private Long maxGasPerBlock = DEFAULT_MAX_GAS_PER_BLOCK;

  @Positive
  @CommandLine.Option(
      names = {VERIFICATION_GAS_COST},
      hidden = true,
      paramLabel = "<INTEGER>",
      description = "L1 verification gas cost (default: ${DEFAULT-VALUE})")
  private int verificationGasCost = DEFAULT_VERIFICATION_GAS_COST;

  @Positive
  @CommandLine.Option(
      names = {VERIFICATION_CAPACITY},
      hidden = true,
      paramLabel = "<INTEGER>",
      description = "L1 verification capacity (default: ${DEFAULT-VALUE})")
  private int verificationCapacity = DEFAULT_VERIFICATION_CAPACITY;

  @Positive
  @CommandLine.Option(
      names = {GAS_PRICE_RATIO},
      hidden = true,
      paramLabel = "<INTEGER>",
      description = "L1/L2 gas price ratio (default: ${DEFAULT-VALUE})")
  private int gasPriceRatio = DEFAULT_GAS_PRICE_RATIO;

  @Positive
  @CommandLine.Option(
      names = {MIN_MARGIN},
      hidden = true,
      paramLabel = "<FLOAT>",
      description = "Minimum margin of a transaction to be selected (default: ${DEFAULT-VALUE})")
  private BigDecimal minMargin = DEFAULT_MIN_MARGIN;

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
    options.maxGasPerBlock = config.maxGasPerBlock();
    options.verificationGasCost = config.getVerificationGasCost();
    options.verificationCapacity = config.getVerificationCapacity();
    options.gasPriceRatio = config.getGasPriceRatio();
    options.minMargin = BigDecimal.valueOf(config.getMinMargin());
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
        .maxGasPerBlock(maxGasPerBlock)
        .verificationGasCost(verificationGasCost)
        .verificationCapacity(verificationCapacity)
        .gasPriceRatio(gasPriceRatio)
        .minMargin(minMargin.doubleValue())
        .build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(MAX_BLOCK_CALLDATA_SIZE, maxBlockCallDataSize)
        .add(MODULE_LIMIT_FILE_PATH, moduleLimitFilePath)
        .add(MAX_GAS_PER_BLOCK, maxGasPerBlock)
        .add(VERIFICATION_GAS_COST, verificationGasCost)
        .add(VERIFICATION_CAPACITY, verificationCapacity)
        .add(GAS_PRICE_RATIO, gasPriceRatio)
        .add(MIN_MARGIN, minMargin)
        .toString();
  }
}
