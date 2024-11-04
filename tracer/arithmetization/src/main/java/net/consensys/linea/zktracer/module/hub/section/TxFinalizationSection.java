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

import static com.google.common.base.Preconditions.*;

import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TransactionFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.transients.DeploymentInfo;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class TxFinalizationSection extends TraceSection implements PostTransactionDefer {
  private final TransactionProcessingMetadata txMetadata;
  private final AccountSnapshot senderSnapshotBeforeFinalization;
  private final AccountSnapshot recipientSnapshotBeforeFinalization;
  private final AccountSnapshot coinbaseSnapshotBeforeTxFinalization;
  private @Setter AccountSnapshot senderSnapshotAfterTxFinalization;
  private @Setter AccountSnapshot recipientSnapshotAfterTxFinalization;
  private @Setter AccountSnapshot coinbaseSnapshotAfterFinalization;

  public TxFinalizationSection(Hub hub, WorldView world) {
    super(hub, (short) 4);
    txMetadata = hub.txStack().current();

    final Address senderAddress = txMetadata.getSender();
    final Address recipientAddress = txMetadata.getEffectiveRecipient();
    final Address coinbaseAddress = txMetadata.getCoinbase();

    // recipient
    senderSnapshotBeforeFinalization = AccountSnapshot.canonical(hub, senderAddress);
    recipientSnapshotBeforeFinalization = AccountSnapshot.canonical(hub, recipientAddress);
    coinbaseSnapshotBeforeTxFinalization = AccountSnapshot.canonical(hub, coinbaseAddress);

    //  TODO: re-enable checks
    // checkArgument(
    //         senderSnapshotBeforeFinalization.isWarm(),
    //         "The sender account ought to be warm during TX_FINL");
    // checkArgument(
    //         recipientSnapshotBeforeFinalization.isWarm(),
    //         "The recipient account ought to be warm during TX_FINL");
    checkArgument(
        txMetadata.isCoinbaseWarmAtTransactionEnd()
            == coinbaseSnapshotBeforeTxFinalization.isWarm(),
        "isCoinbaseWarmAtTransactiondEnd prediction is wrong");

    hub.defers().scheduleForPostTransaction(this);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView world, Transaction tx, boolean isSuccessful) {

    final boolean coinbaseWarmth = txMetadata.isCoinbaseWarmAtTransactionEnd();

    final Address senderAddress = senderSnapshotBeforeFinalization.address();
    senderSnapshotAfterTxFinalization = AccountSnapshot.canonical(hub, senderAddress);
    senderSnapshotAfterTxFinalization.turnOnWarmth(); // purely constraints based

    final Address recipientAddress = recipientSnapshotBeforeFinalization.address();
    recipientSnapshotAfterTxFinalization = AccountSnapshot.canonical(hub, recipientAddress);
    recipientSnapshotAfterTxFinalization.turnOnWarmth(); // purely constraints based

    final Address coinbaseAddress = coinbaseSnapshotBeforeTxFinalization.address();
    coinbaseSnapshotAfterFinalization = AccountSnapshot.canonical(hub, coinbaseAddress);
    coinbaseSnapshotAfterFinalization.setWarmthTo(coinbaseWarmth); // purely constraints based

    DeploymentInfo deploymentInfo = hub.transients().conflation().deploymentInfo();
    checkArgument(isSuccessful == txMetadata.statusCode());

    // TODO: do we switch off the deployment status at the end of a deployment ?
    checkArgument(
        !deploymentInfo.getDeploymentStatus(senderAddress),
        "The sender may not be under deployment");
    checkArgument(
        !deploymentInfo.getDeploymentStatus(recipientAddress),
        "The recipient may not be under deployment");
    checkArgument(
        !deploymentInfo.getDeploymentStatus(coinbaseAddress),
        "The coinbase may not be under deployment");

    if (isSuccessful) {
      successfulFinalization(hub);
    } else {
      unsuccessfulFinalization(hub);
    }
  }

  private void successfulFinalization(Hub hub) {

    if (!senderIsCoinbase()) {

      AccountFragment senderAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  senderSnapshotBeforeFinalization,
                  senderSnapshotAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

      AccountFragment coinbaseAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  coinbaseSnapshotBeforeTxFinalization,
                  coinbaseSnapshotAfterFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

      TransactionFragment currentTransactionFragment =
          TransactionFragment.prepare(hub.txStack().current());

      this.addFragments(senderAccountFragment, coinbaseAccountFragment, currentTransactionFragment);
      return;
    }

    // TODO: verify it works
    AccountFragment senderAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                senderSnapshotBeforeFinalization,
                senderSnapshotBeforeFinalization
                    .deepCopy()
                    .incrementBalanceBy(txMetadata.getGasRefundInWei()),
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

    AccountFragment coinbaseAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                senderSnapshotBeforeFinalization,
                coinbaseSnapshotAfterFinalization,
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

    TransactionFragment currentTransactionFragment =
        TransactionFragment.prepare(hub.txStack().current());

    this.addFragments(senderAccountFragment, coinbaseAccountFragment, currentTransactionFragment);
  }

  private void unsuccessfulFinalization(Hub hub) {
    if (noAddressCollisions()) {

      AccountFragment senderAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  senderSnapshotBeforeFinalization,
                  senderSnapshotAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

      AccountFragment recipientAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  recipientSnapshotBeforeFinalization,
                  recipientSnapshotAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

      AccountFragment coinbaseAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  coinbaseSnapshotBeforeTxFinalization,
                  coinbaseSnapshotAfterFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2));

      TransactionFragment currentTransactionFragment =
          TransactionFragment.prepare(hub.txStack().current());

      this.addFragments(
          senderAccountFragment,
          recipientAccountFragment,
          coinbaseAccountFragment,
          currentTransactionFragment);

    } else {

      Wei transactionValue = (Wei) txMetadata.getBesuTransaction().getValue();

      // FIRST ROW
      ////////////
      AccountSnapshot senderSnapshotAfterValueAndGasRefunds =
          senderSnapshotBeforeFinalization
              .deepCopy()
              .incrementBalanceBy(transactionValue)
              .incrementBalanceBy(txMetadata.getGasRefundInWei());

      AccountFragment senderAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  senderSnapshotBeforeFinalization,
                  senderSnapshotAfterValueAndGasRefunds,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

      // SECOND ROW
      /////////////
      AccountSnapshot recipientSnapshotBeforeSecondRow =
          (senderIsRecipient())
              ? senderSnapshotAfterValueAndGasRefunds
              : recipientSnapshotBeforeFinalization;

      AccountSnapshot recipientSnapshotAfterSecondRow =
          recipientSnapshotBeforeSecondRow.deepCopy().decrementBalanceBy(transactionValue);

      AccountFragment recipientAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  recipientSnapshotBeforeSecondRow,
                  recipientSnapshotAfterSecondRow,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

      // THIRD ROW
      ////////////
      AccountSnapshot coinbaseSnapshotBefore =
          coinbaseSnapshotAfterFinalization
              .deepCopy()
              .decrementBalanceBy(txMetadata.getCoinbaseReward());

      AccountFragment coinbaseAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  coinbaseSnapshotBefore,
                  coinbaseSnapshotAfterFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2));

      TransactionFragment currentTransactionFragment =
          TransactionFragment.prepare(hub.txStack().current());

      this.addFragments(
          senderAccountFragment,
          recipientAccountFragment,
          coinbaseAccountFragment,
          currentTransactionFragment);
    }
  }

  private boolean senderIsCoinbase() {
    return this.senderSnapshotBeforeFinalization
        .address()
        .equals(coinbaseSnapshotBeforeTxFinalization.address());
  }

  private boolean senderIsRecipient() {
    return this.senderSnapshotBeforeFinalization
        .address()
        .equals(recipientSnapshotBeforeFinalization.address());
  }

  private boolean recipientIsCoinbase() {
    return this.recipientSnapshotBeforeFinalization
        .address()
        .equals(coinbaseSnapshotBeforeTxFinalization.address());
  }

  private boolean addressCollision() {
    return senderIsCoinbase() || senderIsRecipient() || recipientIsCoinbase();
  }

  private boolean noAddressCollisions() {
    return !addressCollision();
  }
}
