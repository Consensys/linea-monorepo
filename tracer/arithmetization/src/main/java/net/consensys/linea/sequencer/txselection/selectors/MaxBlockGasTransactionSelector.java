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
package net.consensys.linea.sequencer.txselection.selectors;

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_GAS_EXCEEDS_USER_MAX_BLOCK_GAS;
import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_TOO_LARGE_FOR_REMAINING_USER_GAS;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;

@Slf4j
@RequiredArgsConstructor
public class MaxBlockGasTransactionSelector implements PluginTransactionSelector {

  private final long maxGasPerBlock;
  private long cumulativeBlockGasUsed;

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final PendingTransaction pendingTransaction,
      final TransactionProcessingResult processingResult) {

    final long gasUsedByTransaction = processingResult.getEstimateGasUsedByTransaction();

    if (gasUsedByTransaction > maxGasPerBlock) {
      log.trace(
          "Not selecting transaction, gas used {} greater than max user gas per block {}",
          gasUsedByTransaction,
          maxGasPerBlock);
      return TX_GAS_EXCEEDS_USER_MAX_BLOCK_GAS;
    }

    if (isTransactionExceedingMaxBlockGasLimit(gasUsedByTransaction)) {
      log.trace(
          "Not selecting transaction, cumulative gas used {} greater than max user gas per block {}",
          cumulativeBlockGasUsed,
          maxGasPerBlock);
      return TX_TOO_LARGE_FOR_REMAINING_USER_GAS;
    }
    return SELECTED;
  }

  private boolean isTransactionExceedingMaxBlockGasLimit(long transactionGasUsed) {
    try {
      return Math.addExact(cumulativeBlockGasUsed, transactionGasUsed) > maxGasPerBlock;
    } catch (final ArithmeticException e) {
      // Overflow won't occur as cumulativeBlockGasUsed won't exceed Long.MAX_VALUE
      return true;
    }
  }

  @Override
  public void onTransactionSelected(
      final PendingTransaction pendingTransaction,
      final TransactionProcessingResult processingResult) {
    cumulativeBlockGasUsed +=
        Math.addExact(cumulativeBlockGasUsed, processingResult.getEstimateGasUsedByTransaction());
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final PendingTransaction pendingTransaction) {
    // Evaluation done in post-processing, no action needed here.
    return SELECTED;
  }
}
