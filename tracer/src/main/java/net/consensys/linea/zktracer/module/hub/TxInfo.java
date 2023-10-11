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

package net.consensys.linea.zktracer.module.hub;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.container.StackedContainer;
import org.hyperledger.besu.datatypes.Quantity;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.worldstate.WorldView;

/** Stores transaction-specific information. */
@Accessors(fluent = true)
@Getter
public class TxInfo implements StackedContainer {
  private static final GasCalculator gc = ZkTracer.gasCalculator;

  private int preExecNumber = 0;

  private int number = 0;
  private Transaction transaction;
  @Setter private TxState state;
  @Setter private Boolean status;
  private long initialGas;
  StorageInfo storage;

  /**
   * Returns the transaction result, or throws an exception if it is being accessed outside of its
   * specified lifetime -- between the conclusion of a transaction and the start of a new one.
   *
   * @return the transaction final status
   */
  public boolean status() {
    if (this.status == null) {
      throw new RuntimeException("TX state can not be queried for now.");
    }

    return this.status;
  }

  boolean shouldSkip(WorldView world) {
    return (this.transaction.getTo().isPresent()
            && world.get(this.transaction.getTo().get()).getCode().isEmpty()) // pure transaction
        || (this.transaction.getTo().isEmpty()
            && this.transaction.getInit().isEmpty()); // contract creation without init code
  }

  public Wei gasPrice() {
    return Wei.of(
        this.transaction.getGasPrice().map(Quantity::getAsBigInteger).orElse(BigInteger.ZERO));
  }

  public static BigInteger computeInitGas(Transaction tx) {
    boolean isDeployment = tx.getTo().isEmpty();
    return BigInteger.valueOf(
        tx.getGasLimit()
            - gc.transactionIntrinsicGasCost(tx.getPayload(), isDeployment)
            - tx.getAccessList().map(gc::accessListGasCost).orElse(0L));
  }

  /**
   * Update transaction-specific information on new transaction start.
   *
   * @param tx the new transaction
   */
  void update(final Transaction tx) {
    if (tx.getTo().isPresent() && Hub.isPrecompile(tx.getTo().get())) {
      throw new RuntimeException("Call to precompile forbidden");
    } else {
      this.number++;
    }
    this.storage = new StorageInfo();
    this.status = null;
    this.transaction = tx;
    this.initialGas = tx.getGasLimit();
  }

  @Override
  public void enter() {
    this.preExecNumber = this.number;
  }

  @Override
  public void pop() {
    this.number = this.preExecNumber;
  }
}
