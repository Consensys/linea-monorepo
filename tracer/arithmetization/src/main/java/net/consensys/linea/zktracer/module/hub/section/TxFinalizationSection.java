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

  public TxFinalizationSection(Hub hub, WorldView world, boolean exceptionOrRevert) {
    super(hub, (short) 4);

    txMetadata = hub.txStack().current();

    final Address senderAddress = txMetadata.getSender();
    final Address recipientAddress = txMetadata.getEffectiveRecipient();
    final Address coinbaseAddress = txMetadata.getCoinbase();

    // recipient
    senderSnapshotBeforeFinalization =
        exceptionOrRevert
            ? hub.txStack().getInitializationSection().getSenderAfterPayingForTransaction()
            : AccountSnapshot.canonical(hub, world, senderAddress);
    recipientSnapshotBeforeFinalization =
        exceptionOrRevert
            ? hub.txStack().getInitializationSection().getRecipientAfterValueTransfer()
            : AccountSnapshot.canonical(hub, world, recipientAddress);
    coinbaseSnapshotBeforeTxFinalization = AccountSnapshot.canonical(hub, world, coinbaseAddress);

    hub.defers().scheduleForPostTransaction(this);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView world, Transaction tx, boolean isSuccessful) {

    final boolean coinbaseWarmth = txMetadata.isCoinbaseWarmAtTransactionEnd();

    final Address senderAddress = senderSnapshotBeforeFinalization.address();
    senderSnapshotAfterTxFinalization = AccountSnapshot.canonical(hub, world, senderAddress);
    senderSnapshotAfterTxFinalization.turnOnWarmth(); // purely constraints based

    final Address recipientAddress = recipientSnapshotBeforeFinalization.address();
    recipientSnapshotAfterTxFinalization = AccountSnapshot.canonical(hub, world, recipientAddress);
    recipientSnapshotAfterTxFinalization.turnOnWarmth(); // purely constraints based

    final Address coinbaseAddress = coinbaseSnapshotBeforeTxFinalization.address();
    coinbaseSnapshotAfterFinalization = AccountSnapshot.canonical(hub, world, coinbaseAddress);
    coinbaseSnapshotAfterFinalization.setWarmthTo(coinbaseWarmth); // purely constraints based

    DeploymentInfo deploymentInfo = hub.transients().conflation().deploymentInfo();
    checkArgument(isSuccessful == txMetadata.statusCode());

    // TODO: do we switch off the deployment status at the end of a deployment ?
    // checkArgument(
    //     !deploymentInfo.getDeploymentStatus(senderAddress),
    //     "The sender may not be under deployment");
    // checkArgument(
    //     !deploymentInfo.getDeploymentStatus(recipientAddress),
    //     "The recipient may not be under deployment");
    checkArgument(
        !deploymentInfo.getDeploymentStatus(coinbaseAddress),
        "The coinbase may not be under deployment");

    if (isSuccessful) {
      successFinalization(hub);
    } else {
      failureFinalization(hub);
    }
  }

  private void successFinalization(Hub hub) {

    if (!txMetadata.senderIsCoinbase()) {

      final AccountFragment senderAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  senderSnapshotBeforeFinalization,
                  senderSnapshotAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

      final AccountFragment coinbaseAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  coinbaseSnapshotBeforeTxFinalization,
                  coinbaseSnapshotAfterFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

      this.addFragments(senderAccountFragment, coinbaseAccountFragment);
    } else {
      // TODO: verify it works
      final AccountFragment senderAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  senderSnapshotBeforeFinalization,
                  senderSnapshotBeforeFinalization
                      .deepCopy()
                      .incrementBalanceBy(txMetadata.getGasRefundInWei()),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

      final AccountFragment coinbaseAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  senderSnapshotBeforeFinalization,
                  coinbaseSnapshotAfterFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

      this.addFragments(senderAccountFragment, coinbaseAccountFragment);
    }

    final TransactionFragment currentTransactionFragment =
        TransactionFragment.prepare(hub.txStack().current());
    this.addFragment(currentTransactionFragment);
  }

  private void failureFinalization(Hub hub) {
    if (txMetadata.noAddressCollisions()) {

      final AccountFragment senderAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  senderSnapshotBeforeFinalization,
                  senderSnapshotAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

      final AccountFragment recipientAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  recipientSnapshotBeforeFinalization,
                  recipientSnapshotAfterTxFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

      final AccountFragment coinbaseAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  coinbaseSnapshotBeforeTxFinalization,
                  coinbaseSnapshotAfterFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2));

      this.addFragments(senderAccountFragment, recipientAccountFragment, coinbaseAccountFragment);

    } else {

      final Wei transactionValue = (Wei) txMetadata.getBesuTransaction().getValue();

      // FIRST ROW
      final AccountSnapshot senderSnapshotAfterValueAndGasRefunds =
          senderSnapshotBeforeFinalization
              .deepCopy()
              .incrementBalanceBy(transactionValue)
              .incrementBalanceBy(txMetadata.getGasRefundInWei());

      final AccountFragment senderAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  senderSnapshotBeforeFinalization,
                  senderSnapshotAfterValueAndGasRefunds,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

      // SECOND ROW
      final AccountSnapshot recipientSnapshotBeforeSecondRow =
          (txMetadata.senderIsRecipient())
              ? senderSnapshotAfterValueAndGasRefunds
              : recipientSnapshotBeforeFinalization;

      final AccountSnapshot recipientSnapshotAfterSecondRow =
          recipientSnapshotBeforeSecondRow.deepCopy().decrementBalanceBy(transactionValue);

      final AccountFragment recipientAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  recipientSnapshotBeforeSecondRow,
                  recipientSnapshotAfterSecondRow,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

      // THIRD ROW
      final AccountSnapshot coinbaseSnapshotBefore =
          coinbaseSnapshotAfterFinalization
              .deepCopy()
              .decrementBalanceBy(txMetadata.getCoinbaseReward());

      final AccountFragment coinbaseAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  coinbaseSnapshotBefore,
                  coinbaseSnapshotAfterFinalization,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2));

      this.addFragments(senderAccountFragment, recipientAccountFragment, coinbaseAccountFragment);
    }
    final TransactionFragment currentTransactionFragment =
        TransactionFragment.prepare(hub.txStack().current());

    this.addFragment(currentTransactionFragment);
  }
}
