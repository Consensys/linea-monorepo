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

package net.consensys.linea.zktracer.module.txn_data;

import java.util.Optional;

import lombok.Getter;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

/**
 * This class gathers the block-related information required to trace the {@link TxnData} module.
 */
public class BlockSnapshot {
  /** Sequential ID of this block within a conflation */
  @Getter int id;

  /** A list of {@link TransactionSnapshot} contained in this block */
  @Getter private final StackedList<TransactionSnapshot> txs = new StackedList<>();

  /** The base fee of this block */
  @Getter private final Optional<Wei> baseFee;

  /** The coinbase of this block */
  @Getter private final Address coinbaseAddress;

  BlockSnapshot(int id, ProcessableBlockHeader header) {
    this.id = id;
    this.baseFee = header.getBaseFee().map(x -> (Wei) x);
    this.coinbaseAddress = header.getCoinbase();
  }

  /**
   * Returns the latest transaction snapshotted in this block.
   *
   * @return the latest {@link TransactionSnapshot}
   */
  TransactionSnapshot currentTx() {
    return this.txs.getLast();
  }

  /**
   * Start capturing a transaction in this block.
   *
   * @param worldView a view on the state
   * @param tx the {@link Transaction}
   */
  void captureTx(int codeIdBeforeLex, WorldView worldView, Transaction tx) {
    final TransactionSnapshot snapshot =
        TransactionSnapshot.fromTransaction(codeIdBeforeLex, tx, worldView, this.baseFee);
    this.txs.add(snapshot);
  }

  /**
   * Finishes capturing a transaction in this block.
   *
   * @param cumulativeGasUsed
   * @param leftoverGas
   * @param refundCounter
   * @param status true if the transaction was successful
   */
  void endTx(long cumulativeGasUsed, long leftoverGas, long refundCounter, boolean status) {
    final long effectiveGasRefund = this.getEffectiveGasRefund(leftoverGas, refundCounter);
    this.currentTx().status(status);
    this.currentTx().refundCounter(refundCounter);
    this.currentTx().leftoverGas(leftoverGas);
    this.currentTx().effectiveGasRefund(effectiveGasRefund);
    this.currentTx().cumulativeGasConsumption(cumulativeGasUsed);
  }

  /**
   * Computes the amount of gas being effectively refunded during transaction finalization.
   *
   * @param leftoverGas
   * @param refundCounter
   * @return
   */
  long getEffectiveGasRefund(long leftoverGas, long refundCounter) {
    long gasLimitMinusLeftoverGasOverTwo = 0;
    if (!this.txs.isEmpty()) {
      gasLimitMinusLeftoverGasOverTwo = (this.currentTx().gasLimit() - leftoverGas) >> 1;
    }

    return leftoverGas + Long.min(refundCounter, gasLimitMinusLeftoverGasOverTwo);
  }
}
