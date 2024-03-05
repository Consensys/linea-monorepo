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
package net.consensys.linea.bl;

import java.math.BigDecimal;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.compress.LibCompress;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.slf4j.spi.LoggingEventBuilder;

@Slf4j
public class TransactionProfitabilityCalculator {

  private final LineaProfitabilityConfiguration profitabilityConf;
  private final double preComputedValue;

  public TransactionProfitabilityCalculator(
      final LineaProfitabilityConfiguration profitabilityConf) {
    this.profitabilityConf = profitabilityConf;
    this.preComputedValue =
        profitabilityConf.gasPriceRatio() * profitabilityConf.verificationGasCost();
  }

  public Wei profitablePriorityFeePerGas(
      final Transaction transaction,
      final double minMargin,
      final Wei minGasPrice,
      final long gas) {
    final double compressedTxSize = getCompressedTxSize(transaction);

    final var profitAt =
        preComputedValue
            * minMargin
            * compressedTxSize
            * minGasPrice.getAsBigInteger().doubleValue()
            / (gas * profitabilityConf.verificationCapacity());

    final var adjustedProfit =
        Wei.ofNumber(BigDecimal.valueOf(profitAt).toBigInteger())
            .add(profitabilityConf.gasPriceAdjustment());

    log.atDebug()
        .setMessage(
            "Estimated profitable priorityFeePerGas: {}; estimateGasMinMargin={}, verificationCapacity={}, "
                + "verificationGasCost={}, gasPriceRatio={}, gasPriceAdjustment={}, gas={}, minGasPrice={}, "
                + "l1GasPrice={}, txSize={}, compressedTxSize={}")
        .addArgument(adjustedProfit::toHumanReadableString)
        .addArgument(profitabilityConf.estimateGasMinMargin())
        .addArgument(profitabilityConf.verificationCapacity())
        .addArgument(profitabilityConf.verificationGasCost())
        .addArgument(profitabilityConf.gasPriceRatio())
        .addArgument(profitabilityConf.gasPriceAdjustment()::toHumanReadableString)
        .addArgument(gas)
        .addArgument(minGasPrice::toHumanReadableString)
        .addArgument(
            () -> minGasPrice.multiply(profitabilityConf.gasPriceRatio()).toHumanReadableString())
        .addArgument(transaction::getSize)
        .addArgument(compressedTxSize)
        .log();

    return adjustedProfit;
  }

  public boolean isProfitable(
      final String context,
      final Transaction transaction,
      final double minMargin,
      final Wei minGasPrice,
      final Wei effectiveGasPrice,
      final long gas) {

    final Wei profitablePriorityFee =
        profitablePriorityFeePerGas(transaction, minMargin, minGasPrice, gas);

    if (effectiveGasPrice.lessThan(profitablePriorityFee)) {
      log(
          log.atDebug(),
          context,
          transaction,
          minMargin,
          effectiveGasPrice,
          profitablePriorityFee,
          gas,
          minGasPrice);
      return false;
    }

    log(
        log.atTrace(),
        context,
        transaction,
        minMargin,
        effectiveGasPrice,
        profitablePriorityFee,
        gas,
        minGasPrice);
    return true;
  }

  private double getCompressedTxSize(final Transaction transaction) {
    final byte[] bytes = transaction.encoded().toArrayUnsafe();
    return LibCompress.CompressedSize(bytes, bytes.length);
  }

  private void log(
      final LoggingEventBuilder leb,
      final String context,
      final Transaction transaction,
      final double minMargin,
      final Wei effectiveGasPrice,
      final Wei profitableGasPrice,
      final long gasUsed,
      final Wei minGasPrice) {
    leb.setMessage(
            "Context {}. Transaction {} has a margin of {}, minMargin={}, effectiveGasPrice={},"
                + " profitableGasPrice={}, verificationCapacity={}, verificationGasCost={}, gasPriceRatio={},, gasPriceAdjustment={}"
                + " gasUsed={}, minGasPrice={}")
        .addArgument(context)
        .addArgument(transaction::getHash)
        .addArgument(
            () ->
                effectiveGasPrice.toBigInteger().doubleValue()
                    / profitableGasPrice.toBigInteger().doubleValue())
        .addArgument(minMargin)
        .addArgument(effectiveGasPrice::toHumanReadableString)
        .addArgument(profitableGasPrice::toHumanReadableString)
        .addArgument(profitabilityConf.verificationCapacity())
        .addArgument(profitabilityConf.verificationGasCost())
        .addArgument(profitabilityConf.gasPriceRatio())
        .addArgument(profitabilityConf.gasPriceAdjustment()::toHumanReadableString)
        .addArgument(gasUsed)
        .addArgument(minGasPrice::toHumanReadableString)
        .log();
  }
}
