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

import java.util.Map;
import java.util.function.Supplier;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.tracer.BlockAwareOperationTracer;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;

/**
 * This class implements TransactionSelector and provides a specific implementation for evaluating
 * transactions based on the number of trace lines per module created by a transaction. It checks if
 * adding a transaction to the block pushes the trace lines for a module over the limit.
 */
@Slf4j
public class TraceLineLimitTransactionSelector implements PluginTransactionSelector {

  private final Supplier<Map<String, Integer>> moduleLimitsProvider;
  private final ZkTracer zkTracer;
  private final String limitFilePath;
  private Map<String, Integer> prevLineCount = Map.of();
  private Map<String, Integer> currLineCount;

  public TraceLineLimitTransactionSelector(
      final Supplier<Map<String, Integer>> moduleLimitsProvider, final String limitFilePath) {
    this.moduleLimitsProvider = moduleLimitsProvider;
    zkTracer = new ZkTracer();
    zkTracer.traceStartConflation(1L);
    this.limitFilePath = limitFilePath;
  }

  /**
   * No checking is done pre-processing.
   *
   * @param pendingTransaction The transaction to evaluate.
   * @return TransactionSelectionResult.SELECTED
   */
  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final PendingTransaction pendingTransaction) {
    return TransactionSelectionResult.SELECTED;
  }

  @Override
  public void onTransactionNotSelected(
      final PendingTransaction pendingTransaction,
      final TransactionSelectionResult transactionSelectionResult) {
    zkTracer.popTransaction(pendingTransaction);
  }

  @Override
  public void onTransactionSelected(
      final PendingTransaction pendingTransaction,
      final TransactionProcessingResult processingResult) {
    prevLineCount = currLineCount;
  }

  /**
   * Checking the created trace lines is performed post-processing.
   *
   * @param pendingTransaction The processed transaction.
   * @param processingResult The result of the transaction processing.
   * @return BLOCK_FULL if the trace lines for a module are over the limit, otherwise SELECTED.
   */
  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final PendingTransaction pendingTransaction,
      final TransactionProcessingResult processingResult) {
    final Map<String, Integer> moduleLimits = moduleLimitsProvider.get();
    // check that we are not exceed line number for any module
    currLineCount = zkTracer.getModulesLineCount();
    for (var e : currLineCount.entrySet()) {
      final String module = e.getKey();
      if (!moduleLimits.containsKey(module)) {
        final String errorMsg =
            "Module " + module + " does not exist in the limits file: " + limitFilePath;
        log.error(errorMsg);
        throw new RuntimeException(errorMsg);
      }

      final int currModuleLineCount = currLineCount.get(module);
      final int moduleLineCountLimit = moduleLimits.get(module);

      final int txModuleLineCount = currModuleLineCount - prevLineCount.getOrDefault(module, 0);
      if (txModuleLineCount > moduleLineCountLimit) {
        log.warn(
            "Tx {} line count for module {}={} is above the limit {}, removing from the txpool",
            pendingTransaction.getTransaction().getHash(),
            module,
            txModuleLineCount,
            moduleLineCountLimit);
        return TransactionSelectionResult.invalid("TX_MODULE_LINE_COUNT_OVERFLOW");
      }

      if (currModuleLineCount > moduleLineCountLimit) {
        return TransactionSelectionResult.BLOCK_FULL;
      }
    }
    return TransactionSelectionResult.SELECTED;
  }

  @Override
  public BlockAwareOperationTracer getOperationTracer() {
    return zkTracer;
  }
}
