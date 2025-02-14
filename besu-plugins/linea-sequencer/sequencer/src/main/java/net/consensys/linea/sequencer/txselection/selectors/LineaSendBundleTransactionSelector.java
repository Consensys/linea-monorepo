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

import static java.lang.Boolean.TRUE;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

import net.consensys.linea.rpc.methods.LineaSendBundle;
import net.consensys.linea.rpc.services.BundlePoolService;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

public class LineaSendBundleTransactionSelector implements PluginTransactionSelector {

  final BundlePoolService bundlePool;

  public LineaSendBundleTransactionSelector(BundlePoolService bundlePool) {
    this.bundlePool = bundlePool;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext txContext) {

    // short circuit if we are not a PendingBundleTx
    if (!(txContext.getPendingTransaction() instanceof LineaSendBundle.PendingBundleTx)) {
      return TransactionSelectionResult.SELECTED;
    }

    final var blockHeader = txContext.getPendingBlockHeader();

    // map the pending transaction to its origin bundle:
    var bundle =
        bundlePool.getBundleByPendingTransaction(
            blockHeader.getNumber(), txContext.getPendingTransaction());

    if (bundle.isPresent()) {
      var satisfiesCriteria =
          bundle.stream()
              // filter if we have a min timestamp that has not been satisfied
              .filter(
                  b ->
                      b.minTimestamp()
                          .map(minTime -> minTime < Instant.now().getEpochSecond())
                          .orElse(TRUE))
              // filter if we have a max timestamp that has not been satisfied
              .filter(
                  b ->
                      b.maxTimestamp()
                          .map(maxTime -> maxTime > Instant.now().getEpochSecond())
                          .orElse(TRUE))
              .findAny();
      if (satisfiesCriteria.isPresent()) {
        return TransactionSelectionResult.SELECTED;
      } else {
        return TransactionSelectionResult.invalid("Failed Bundled Transaction Criteria");
      }
    }
    // if the bundle was not found return SELECTED so we do not block a non-bundle transaction
    return TransactionSelectionResult.SELECTED;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext txContext,
      final TransactionProcessingResult transactionProcessingResult) {

    // short circuit if we are not a PendingBundleTx
    if (!(txContext.getPendingTransaction() instanceof LineaSendBundle.PendingBundleTx)) {
      return TransactionSelectionResult.SELECTED;
    }

    if (transactionProcessingResult.isFailed()) {

      final var blockHeader = txContext.getPendingBlockHeader();
      // map the pending transaction to its origin bundle:
      Optional<List<Hash>> revertableList =
          bundlePool
              .getBundleByPendingTransaction(
                  blockHeader.getNumber(), txContext.getPendingTransaction())
              .flatMap(bundle -> bundle.revertingTxHashes());

      // if a bundle tx failed, but was not in a revertable list, we unselect and fail the bundle
      if (revertableList.isEmpty()
          || !revertableList
              .get()
              .contains(txContext.getPendingTransaction().getTransaction().getHash())) {
        return TransactionSelectionResult.invalid("Failed non revertable transaction in bundle");
      }
    }
    // if the bundle was not found return SELECTED so we do not block a non-bundle transaction
    return TransactionSelectionResult.SELECTED;
  }
}
