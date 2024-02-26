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

package net.consensys.linea.zktracer.module.hub;

import static net.consensys.linea.zktracer.types.AddressUtils.isPrecompile;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import lombok.Builder;
import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.container.StackedContainer;
import net.consensys.linea.zktracer.module.hub.transients.StorageInitialValues;
import net.consensys.linea.zktracer.types.TxState;
import org.hyperledger.besu.datatypes.Quantity;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class TransactionStack implements StackedContainer {
  private final List<MetaTransaction> txs = new ArrayList<>(100);
  private int currentAbsNumber;

  public MetaTransaction current() {
    return this.txs.get(this.txs.size() - 1);
  }

  public MetaTransaction getById(int id) {
    return this.txs.get(id);
  }

  public MetaTransaction getByAbsNumber(int id) {
    for (MetaTransaction tx : this.txs) {
      if (tx.absNumber == id) {
        return tx;
      }
    }

    throw new IndexOutOfBoundsException("unknown tx");
  }

  @Override
  public void enter() {
    this.currentAbsNumber += 1;
  }

  @Override
  public void pop() {
    this.currentAbsNumber -= 1;
  }

  public void enterTransaction(final Transaction tx) {
    this.enter();
    if (tx.getTo().isPresent() && isPrecompile(tx.getTo().get())) {
      throw new RuntimeException("Call to precompile forbidden");
    } else {
      //      this.number++;
    }

    final MetaTransaction newTx =
        MetaTransaction.builder()
            .besuTx(tx)
            .absNumber(currentAbsNumber)
            .status(null)
            .initialGas(tx.getGasLimit())
            .build();
    this.txs.add(newTx);
  }

  public void exitTransaction(final Hub hub, boolean isSuccessful) {
    if (this.current().state() != TxState.TX_SKIP) {
      this.current().state(TxState.TX_FINAL);
    }
    this.current().status(!isSuccessful).endStamp(hub.stamp());
  }

  public static long computeInitGas(Transaction tx) {
    boolean isDeployment = tx.getTo().isEmpty();
    return tx.getGasLimit()
        - ZkTracer.gasCalculator.transactionIntrinsicGasCost(tx.getPayload(), isDeployment)
        - tx.getAccessList().map(ZkTracer.gasCalculator::accessListGasCost).orElse(0L);
  }

  @Builder
  @Accessors(fluent = true)
  public static class MetaTransaction {
    @Getter private int id;
    @Getter private Transaction besuTx;
    @Getter private int absNumber;
    @Getter @Setter private TxState state;
    @Setter @Builder.Default private Boolean status = null;
    @Getter private long initialGas;
    @Getter private final StorageInitialValues storage = new StorageInitialValues();
    @Getter @Setter @Builder.Default int endStamp = -1;

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

    public boolean shouldSkip(WorldView world) {
      return (this.besuTx.getTo().isPresent()
              && Optional.ofNullable(world.get(this.besuTx.getTo().get()))
                  .map(a -> a.getCode().isEmpty())
                  .orElse(true)) // pure transaction
          || (this.besuTx.getTo().isEmpty()
              && this.besuTx.getInit().isEmpty()); // contract creation without init code
    }

    public Wei gasPrice() {
      return Wei.of(
          this.besuTx.getGasPrice().map(Quantity::getAsBigInteger).orElse(BigInteger.ZERO));
    }
  }
}
