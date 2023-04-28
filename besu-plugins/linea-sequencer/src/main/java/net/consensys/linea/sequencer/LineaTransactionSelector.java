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

import org.hyperledger.besu.plugin.data.Transaction;
import org.hyperledger.besu.plugin.data.TransactionReceipt;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.TransactionSelector;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class LineaTransactionSelector implements TransactionSelector {

  private static final Logger LOG = LoggerFactory.getLogger(LineaTransactionSelector.class);

  private final int maxTxCalldataSize;
  private final int maxBlockCalldataSize;
  private int blockCalldataSum;

  public LineaTransactionSelector(final int maxTxCalldataSize, final int maxBlockCalldataSize) {
    this.maxTxCalldataSize = maxTxCalldataSize;
    this.maxBlockCalldataSize = maxBlockCalldataSize;
  }

  @Override
  public TransactionSelectionResult selectTransaction(
      final Transaction transaction, final TransactionReceipt receipt) {

    LOG.trace(
        "Transaction: {}\nReceipt: {}\nCalldata sum: {}", transaction, receipt, blockCalldataSum);

    final int txCalldataSize = transaction.getPayload().size();
    if (txCalldataSize > maxTxCalldataSize) {
      return TransactionSelectionResult.DELETE_TRANSACTION_AND_CONTINUE;
    }
    try {
      blockCalldataSum = Math.addExact(blockCalldataSum, txCalldataSize);
    } catch (final ArithmeticException e) {
      LOG.debug(
          "Not adding transaction {} otherwise block calldata size {} overflows",
          transaction,
          blockCalldataSum);
      return TransactionSelectionResult.COMPLETE_OPERATION;
    }
    if (blockCalldataSum > maxBlockCalldataSize) {
      return TransactionSelectionResult.COMPLETE_OPERATION;
    }

    return TransactionSelectionResult.CONTINUE;
  }
}
