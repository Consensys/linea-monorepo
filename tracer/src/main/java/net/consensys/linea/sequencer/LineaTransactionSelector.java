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

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.TransactionSelector;

/**
 * Represents an implementation of a transaction selector which decides if a transaction is to be
 * added to a block.
 */
@Slf4j
@RequiredArgsConstructor
public class LineaTransactionSelector implements TransactionSelector {
  private final int maxTxCalldataSize;
  private final int maxBlockCalldataSize;
  private int blockCalldataSum;

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final PendingTransaction pendingTransaction) {
    final Transaction transaction = pendingTransaction.getTransaction();

    final int txCalldataSize = transaction.getPayload().size();

    if (txCalldataSize > maxTxCalldataSize) {
      log.warn(
          "Not adding transaction {} because calldata size {} is too big",
          transaction,
          txCalldataSize);
      return TransactionSelectionResult.invalid("Calldata too big");
    }
    try {
      blockCalldataSum = Math.addExact(blockCalldataSum, txCalldataSize);
    } catch (final ArithmeticException e) {
      // this should never happen
      log.warn(
          "Not adding transaction {} otherwise block calldata size {} overflows",
          transaction,
          blockCalldataSum);
      return TransactionSelectionResult.BLOCK_FULL;
    }
    if (blockCalldataSum > maxBlockCalldataSize) {
      return TransactionSelectionResult.BLOCK_FULL;
    }

    return TransactionSelectionResult.SELECTED;
  }
}
