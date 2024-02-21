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
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.slf4j.spi.LoggingEventBuilder;

@Slf4j
public class TransactionProfitabilityCalculator {

  private final LineaTransactionSelectorConfiguration conf;
  private final double preComputedValue;

  public TransactionProfitabilityCalculator(final LineaTransactionSelectorConfiguration conf) {
    this.conf = conf;
    this.preComputedValue =
        conf.estimateGasMinMargin() * conf.gasPriceRatio() * conf.verificationGasCost();
  }

  public Wei profitablePriorityFeePerGas(
      final Transaction transaction, final Wei minGasPrice, final long gas) {
    final double compressedTxSize = getCompressedTxSize(transaction);

    final var profitAt =
        preComputedValue
            * compressedTxSize
            * minGasPrice.getAsBigInteger().doubleValue()
            / (gas * conf.verificationCapacity());

    final var profitAtWei = Wei.ofNumber(BigDecimal.valueOf(profitAt).toBigInteger());

    log.atDebug()
        .setMessage(
            "Estimated profitable priorityFeePerGas: {}; estimateGasMinMargin={}, verificationCapacity={}, "
                + "verificationGasCost={}, gasPriceRatio={}, gas={}, minGasPrice={}, "
                + "l1GasPrice={}, txSize={}, compressedTxSize={}")
        .addArgument(profitAtWei::toHumanReadableString)
        .addArgument(conf.estimateGasMinMargin())
        .addArgument(conf.verificationCapacity())
        .addArgument(conf.verificationGasCost())
        .addArgument(conf.gasPriceRatio())
        .addArgument(gas)
        .addArgument(minGasPrice::toHumanReadableString)
        .addArgument(() -> minGasPrice.multiply(conf.gasPriceRatio()).toHumanReadableString())
        .addArgument(transaction::getSize)
        .addArgument(compressedTxSize)
        .log();

    return profitAtWei;
  }

  public boolean isProfitable(
      final String step,
      final Transaction transaction,
      final double minGasPrice,
      final double effectiveGasPrice,
      final long gas) {
    final double revenue = effectiveGasPrice * gas;

    final double l1GasPrice = minGasPrice * conf.gasPriceRatio();
    final double compressedTxSize = getCompressedTxSize(transaction);
    final double verificationGasCostSlice =
        (compressedTxSize / conf.verificationCapacity()) * conf.verificationGasCost();
    final double cost = l1GasPrice * verificationGasCostSlice;

    final double margin = revenue / cost;

    if (margin < conf.minMargin()) {
      log(
          log.atDebug(),
          step,
          transaction,
          margin,
          effectiveGasPrice,
          gas,
          minGasPrice,
          l1GasPrice,
          compressedTxSize);
      return false;
    } else {
      log(
          log.atTrace(),
          step,
          transaction,
          margin,
          effectiveGasPrice,
          gas,
          minGasPrice,
          l1GasPrice,
          compressedTxSize);
      return true;
    }
  }

  private double getCompressedTxSize(final Transaction transaction) {
    final byte[] bytes = transaction.encoded().toArrayUnsafe();
    return LibCompress.CompressedSize(bytes, bytes.length);
  }

  private void log(
      final LoggingEventBuilder leb,
      final String context,
      final Transaction transaction,
      final double margin,
      final double effectiveGasPrice,
      final long gasUsed,
      final double minGasPrice,
      final double l1GasPrice,
      final double compressedTxSize) {
    leb.setMessage(
            "Context {}. Transaction {} has a margin of {}, minMargin={}, verificationCapacity={}, "
                + "verificationGasCost={}, gasPriceRatio={}, effectiveGasPrice={}, gasUsed={}, minGasPrice={}, "
                + "l1GasPrice={}, txSize={}, compressedTxSize={}")
        .addArgument(context)
        .addArgument(transaction::getHash)
        .addArgument(margin)
        .addArgument(conf.minMargin())
        .addArgument(conf.verificationCapacity())
        .addArgument(conf.verificationGasCost())
        .addArgument(conf.gasPriceRatio())
        .addArgument(effectiveGasPrice)
        .addArgument(gasUsed)
        .addArgument(minGasPrice)
        .addArgument(l1GasPrice)
        .addArgument(transaction::getSize)
        .addArgument(compressedTxSize)
        .log();
  }
}
