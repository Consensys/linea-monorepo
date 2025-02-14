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

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.BUNDLE_GAS_EXCEEDS_MAX_BUNDLE_BLOCK_GAS;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.rpc.methods.LineaSendBundle;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.AbstractStatefulPluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.SelectorsStateManager;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

@Slf4j
public class MaxBundleGasPerBlockTransactionSelector
    extends AbstractStatefulPluginTransactionSelector<Long> {
  private final long maxBundleGasPerBlock;

  public MaxBundleGasPerBlockTransactionSelector(
      final SelectorsStateManager selectorsStateManager, final long maxBundleGasPerBlock) {
    super(selectorsStateManager, 0L, SelectorsStateManager.StateDuplicator::duplicateLong);
    this.maxBundleGasPerBlock = maxBundleGasPerBlock;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext txContext) {

    // short circuit if we are not a PendingBundleTx
    if (!(txContext.getPendingTransaction() instanceof LineaSendBundle.PendingBundleTx)) {
      return SELECTED;
    }

    final long cumulativeBundleGas = getWorkingState();

    if (cumulativeBundleGas + txContext.getPendingTransaction().getTransaction().getGasLimit()
        > maxBundleGasPerBlock) {
      return BUNDLE_GAS_EXCEEDS_MAX_BUNDLE_BLOCK_GAS;
    }

    return SELECTED;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext txContext,
      final TransactionProcessingResult transactionProcessingResult) {

    // short circuit if we are not a PendingBundleTx
    if (!(txContext.getPendingTransaction() instanceof LineaSendBundle.PendingBundleTx)) {
      return SELECTED;
    }

    final long newCumulativeBundleGas =
        getWorkingState() + txContext.getPendingTransaction().getTransaction().getGasLimit();
    setWorkingState(newCumulativeBundleGas);

    return SELECTED;
  }
}
