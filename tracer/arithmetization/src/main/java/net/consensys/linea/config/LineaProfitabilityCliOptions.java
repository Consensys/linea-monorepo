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

import java.math.BigDecimal;

import com.google.common.base.MoreObjects;
import jakarta.validation.constraints.Positive;
import net.consensys.linea.config.converters.WeiConverter;
import org.hyperledger.besu.datatypes.Wei;
import picocli.CommandLine;

/** The Linea profitability calculator CLI options. */
public class LineaProfitabilityCliOptions {
  public static final String VERIFICATION_GAS_COST = "--plugin-linea-verification-gas-cost";
  public static final int DEFAULT_VERIFICATION_GAS_COST = 1_200_000;

  public static final String VERIFICATION_CAPACITY = "--plugin-linea-verification-capacity";
  public static final int DEFAULT_VERIFICATION_CAPACITY = 90_000;

  public static final String GAS_PRICE_RATIO = "--plugin-linea-gas-price-ratio";
  public static final int DEFAULT_GAS_PRICE_RATIO = 15;

  public static final String GAS_PRICE_ADJUSTMENT = "--plugin-linea-gas-price-adjustment";
  public static final Wei DEFAULT_GAS_PRICE_ADJUSTMENT = Wei.ZERO;

  public static final String MIN_MARGIN = "--plugin-linea-min-margin";
  public static final BigDecimal DEFAULT_MIN_MARGIN = BigDecimal.ONE;

  public static final String ESTIMATE_GAS_MIN_MARGIN = "--plugin-linea-estimate-gas-min-margin";
  public static final BigDecimal DEFAULT_ESTIMATE_GAS_MIN_MARGIN = BigDecimal.ONE;

  public static final String TX_POOL_MIN_MARGIN = "--plugin-linea-tx-pool-min-margin";
  public static final BigDecimal DEFAULT_TX_POOL_MIN_MARGIN = BigDecimal.valueOf(0.5);

  public static final String TX_POOL_ENABLE_CHECK_API =
      "--plugin-linea-tx-pool-profitability-check-api-enabled";
  public static final boolean DEFAULT_TX_POOL_ENABLE_CHECK_API = true;

  public static final String TX_POOL_ENABLE_CHECK_P2P =
      "--plugin-linea-tx-pool-profitability-check-p2p-enabled";
  public static final boolean DEFAULT_TX_POOL_ENABLE_CHECK_P2P = false;

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

  @CommandLine.Option(
      names = {GAS_PRICE_ADJUSTMENT},
      hidden = true,
      converter = WeiConverter.class,
      paramLabel = "<WEI>",
      description =
          "Amount to add to the calculated profitable gas price (default: ${DEFAULT-VALUE})")
  private Wei gasPriceAdjustment = DEFAULT_GAS_PRICE_ADJUSTMENT;

  @Positive
  @CommandLine.Option(
      names = {MIN_MARGIN},
      hidden = true,
      paramLabel = "<FLOAT>",
      description = "Minimum margin of a transaction to be selected (default: ${DEFAULT-VALUE})")
  private BigDecimal minMargin = DEFAULT_MIN_MARGIN;

  @Positive
  @CommandLine.Option(
      names = {ESTIMATE_GAS_MIN_MARGIN},
      hidden = true,
      paramLabel = "<FLOAT>",
      description =
          "Recommend a specific gas price when using linea_estimateGas (default: ${DEFAULT-VALUE})")
  private BigDecimal estimageGasMinMargin = DEFAULT_ESTIMATE_GAS_MIN_MARGIN;

  @Positive
  @CommandLine.Option(
      names = {TX_POOL_MIN_MARGIN},
      hidden = true,
      paramLabel = "<FLOAT>",
      description =
          "The min margin an incoming tx must have to be accepted in the txpool (default: ${DEFAULT-VALUE})")
  private BigDecimal txPoolMinMargin = DEFAULT_TX_POOL_MIN_MARGIN;

  @CommandLine.Option(
      names = {TX_POOL_ENABLE_CHECK_API},
      arity = "0..1",
      hidden = true,
      paramLabel = "<BOOLEAN>",
      description =
          "Enable the profitability check for txs received via API? (default: ${DEFAULT-VALUE})")
  private boolean txPoolCheckApiEnabled = DEFAULT_TX_POOL_ENABLE_CHECK_API;

  @CommandLine.Option(
      names = {TX_POOL_ENABLE_CHECK_P2P},
      arity = "0..1",
      hidden = true,
      paramLabel = "<BOOLEAN>",
      description =
          "Enable the profitability check for txs received via p2p? (default: ${DEFAULT-VALUE})")
  private boolean txPoolCheckP2pEnabled = DEFAULT_TX_POOL_ENABLE_CHECK_P2P;

  private LineaProfitabilityCliOptions() {}

  /**
   * Create Linea cli options.
   *
   * @return the Linea cli options
   */
  public static LineaProfitabilityCliOptions create() {
    return new LineaProfitabilityCliOptions();
  }

  /**
   * Linea cli options from config.
   *
   * @param config the config
   * @return the Linea cli options
   */
  public static LineaProfitabilityCliOptions fromConfig(
      final LineaProfitabilityConfiguration config) {
    final LineaProfitabilityCliOptions options = create();
    options.verificationGasCost = config.verificationGasCost();
    options.verificationCapacity = config.verificationCapacity();
    options.gasPriceRatio = config.gasPriceRatio();
    options.gasPriceAdjustment = config.gasPriceAdjustment();
    options.minMargin = BigDecimal.valueOf(config.minMargin());
    options.estimageGasMinMargin = BigDecimal.valueOf(config.estimateGasMinMargin());
    options.txPoolMinMargin = BigDecimal.valueOf(config.txPoolMinMargin());
    return options;
  }

  /**
   * To domain object Linea factory configuration.
   *
   * @return the Linea factory configuration
   */
  public LineaProfitabilityConfiguration toDomainObject() {
    return LineaProfitabilityConfiguration.builder()
        .verificationGasCost(verificationGasCost)
        .verificationCapacity(verificationCapacity)
        .gasPriceRatio(gasPriceRatio)
        .gasPriceAdjustment(gasPriceAdjustment)
        .minMargin(minMargin.doubleValue())
        .estimateGasMinMargin(estimageGasMinMargin.doubleValue())
        .txPoolMinMargin(txPoolMinMargin.doubleValue())
        .build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(VERIFICATION_GAS_COST, verificationGasCost)
        .add(VERIFICATION_CAPACITY, verificationCapacity)
        .add(GAS_PRICE_RATIO, gasPriceRatio)
        .add(GAS_PRICE_ADJUSTMENT, gasPriceAdjustment)
        .add(MIN_MARGIN, minMargin)
        .add(ESTIMATE_GAS_MIN_MARGIN, estimageGasMinMargin)
        .add(TX_POOL_MIN_MARGIN, txPoolMinMargin)
        .toString();
  }
}
