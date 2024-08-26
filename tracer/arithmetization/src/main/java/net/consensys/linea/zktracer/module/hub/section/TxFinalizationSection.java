/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.hub.section;

import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TransactionFragment;
import net.consensys.linea.zktracer.module.hub.transients.DeploymentInfo;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class TxFinalizationSection extends TraceSection implements PostTransactionDefer {
  private final TransactionProcessingMetadata txMetadata;
  private final AccountSnapshot fromAccountBeforeTxFinalization;
  private final AccountSnapshot toAccountBeforeTxFinalization;
  private final AccountSnapshot minerAccountBeforeTxFinalization;
  private @Setter AccountSnapshot fromAccountAfterTxFinalization;
  private @Setter AccountSnapshot toAccountAfterTxFinalization;
  private @Setter AccountSnapshot minerAccountAfterTxFinalization;

  public TxFinalizationSection(Hub hub, WorldView world) {
    super(hub, (short) 4);
    this.txMetadata = hub.txStack().current();

    final DeploymentInfo depInfo = hub.transients().conflation().deploymentInfo();

    // old Sender account snapshot
    final Address senderAddress = txMetadata.getSender();
    this.fromAccountBeforeTxFinalization =
        AccountSnapshot.fromAccount(
            world.get(senderAddress),
            true,
            depInfo.number(senderAddress),
            depInfo.isDeploying(senderAddress));

    // old Recipient account snapshot
    final Address toAddress = txMetadata.getEffectiveTo();
    this.toAccountBeforeTxFinalization =
        AccountSnapshot.fromAccount(
            world.get(toAddress), true, depInfo.number(toAddress), depInfo.isDeploying(toAddress));

    // old Miner Account snapshot
    final Address minerAddress = txMetadata.getCoinbase();
    this.minerAccountBeforeTxFinalization =
        AccountSnapshot.fromAccount(
            world.get(minerAddress),
            txMetadata.isMinerWarmAtEndTx(),
            depInfo.number(minerAddress),
            depInfo.isDeploying(minerAddress));

    hub.defers().scheduleForPostTransaction(this);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView world, Transaction tx, boolean isSuccessful) {
    final DeploymentInfo depInfo = hub.transients().conflation().deploymentInfo();

    final Address fromAddress = fromAccountBeforeTxFinalization.address();
    this.fromAccountAfterTxFinalization =
        AccountSnapshot.fromAccount(
            world.get(fromAddress),
            true,
            depInfo.number(fromAddress),
            depInfo.isDeploying(fromAddress));

    final Address toAddress = toAccountBeforeTxFinalization.address();
    this.toAccountAfterTxFinalization =
        AccountSnapshot.fromAccount(
            world.get(toAddress), true, depInfo.number(toAddress), depInfo.isDeploying(toAddress));

    final Address minerAddress = minerAccountBeforeTxFinalization.address();
    this.minerAccountAfterTxFinalization =
        AccountSnapshot.fromAccount(
            world.get(minerAddress),
            txMetadata.isMinerWarmAtEndTx(),
            depInfo.number(minerAddress),
            depInfo.isDeploying(minerAddress));

    if (txMetadata.statusCode()) {
      successfulFinalization(hub);
    } else {
      unsuccessfulFinalization(hub);
    }
  }

  private void successfulFinalization(Hub hub) {
    if (!senderIsMiner()) {
      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  this.fromAccountBeforeTxFinalization,
                  this.fromAccountAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0)));

      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  this.minerAccountBeforeTxFinalization,
                  this.minerAccountAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1)));
      this.addFragment(TransactionFragment.prepare(hub.txStack().current()));
    } else {
      // TODO: verify it works
      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  this.fromAccountBeforeTxFinalization,
                  this.fromAccountBeforeTxFinalization.incrementBalance(
                      txMetadata.getGasRefundInWei()),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0)));

      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  this.fromAccountBeforeTxFinalization,
                  this.minerAccountAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1)));
      this.addFragment(TransactionFragment.prepare(hub.txStack().current()));
    }
  }

  private void unsuccessfulFinalization(Hub hub) {
    if (noAccountCollision()) {
      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  this.fromAccountBeforeTxFinalization,
                  this.fromAccountAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0)));

      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  this.toAccountBeforeTxFinalization,
                  this.toAccountAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1)));

      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  this.minerAccountBeforeTxFinalization,
                  this.minerAccountAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2)));
      this.addFragment(TransactionFragment.prepare(hub.txStack().current()));
    } else {
      // TODO: test this
      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  this.fromAccountBeforeTxFinalization,
                  this.fromAccountBeforeTxFinalization
                      .incrementBalance((Wei) txMetadata.getBesuTransaction().getValue())
                      .incrementBalance(txMetadata.getGasRefundInWei()),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0)));
      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  senderIsTo()
                      ? this.fromAccountBeforeTxFinalization
                      : this.toAccountBeforeTxFinalization,
                  senderIsTo()
                      ? this.fromAccountBeforeTxFinalization.decrementBalance(
                          (Wei) txMetadata.getBesuTransaction().getValue())
                      : this.toAccountBeforeTxFinalization.decrementBalance(
                          (Wei) txMetadata.getBesuTransaction().getValue()),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1)));

      this.addFragment(
          hub.factories()
              .accountFragment()
              .make(
                  this.minerAccountAfterTxFinalization.decrementBalance(
                      txMetadata.getMinerReward()),
                  this.minerAccountAfterTxFinalization.incrementBalance(
                      txMetadata.getMinerReward()),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2)));
      this.addFragment(TransactionFragment.prepare(hub.txStack().current()));
    }
  }

  private boolean senderIsMiner() {
    return this.fromAccountBeforeTxFinalization
        .address()
        .equals(this.minerAccountBeforeTxFinalization.address());
  }

  private boolean senderIsTo() {
    return this.fromAccountBeforeTxFinalization
        .address()
        .equals(this.toAccountBeforeTxFinalization.address());
  }

  private boolean toIsMiner() {
    return this.toAccountBeforeTxFinalization
        .address()
        .equals(this.minerAccountBeforeTxFinalization.address());
  }

  private boolean noAccountCollision() {
    return !senderIsMiner() && !senderIsTo() && !toIsMiner();
  }
}
