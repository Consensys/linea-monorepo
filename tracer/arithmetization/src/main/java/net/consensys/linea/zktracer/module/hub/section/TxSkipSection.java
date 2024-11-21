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

package net.consensys.linea.zktracer.module.hub.section;

import static com.google.common.base.Preconditions.*;
import static net.consensys.linea.zktracer.types.AddressUtils.isPrecompile;

import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TransactionFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.transients.Transients;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.worldstate.WorldView;

/**
 * SkippedTransaction latches data at the pre-execution of the transaction data that will be used
 * later, through a {@link PostTransactionDefer}, to generate the trace chunks required for the
 * proving of a pure transaction.
 */
public class TxSkipSection extends TraceSection implements PostTransactionDefer {

  final TransactionProcessingMetadata txMetadata;
  final AccountSnapshot senderAccountSnapshotBefore;
  final AccountSnapshot recipientAccountSnapshotBefore;
  final AccountSnapshot coinbaseAccountSnapshotBefore;

  public TxSkipSection(
      Hub hub,
      WorldView world,
      TransactionProcessingMetadata transactionProcessingMetadata,
      Transients transients) {
    super(hub, (short) 4);
    hub.defers().scheduleForPostTransaction(this);

    txMetadata = transactionProcessingMetadata;
    final Address senderAddress = txMetadata.getBesuTransaction().getSender();
    final Address recipientAddress = txMetadata.getEffectiveRecipient();
    final Address coinbaseAddress = txMetadata.getCoinbase();

    senderAccountSnapshotBefore =
        AccountSnapshot.canonical(hub, world, senderAddress, isPrecompile(senderAddress))
            .turnOnWarmth();
    recipientAccountSnapshotBefore =
        AccountSnapshot.canonical(hub, world, recipientAddress, isPrecompile(recipientAddress))
            .turnOnWarmth();
    coinbaseAccountSnapshotBefore =
        AccountSnapshot.canonical(hub, world, coinbaseAddress, isPrecompile(coinbaseAddress));

    // arithmetization restriction
    checkArgument(
        !isPrecompile(recipientAddress),
        "Arithmetization restriction: recipient address is a precompile.");

    // sanity check + EIP-3607
    checkArgument(world.get(senderAddress) != null, "Sender account must exists");
    checkArgument(!world.get(senderAddress).hasCode(), "Sender account must not have code");

    // deployments are local to a transaction, every address should have deploymentStatus == false
    // at the start of every transaction
    checkArgument(!hub.deploymentStatusOf(senderAddress));
    checkArgument(!hub.deploymentStatusOf(recipientAddress));
    checkArgument(!hub.deploymentStatusOf(coinbaseAddress));

    // the updated deployment info appears in the "updated" account fragment
    if (txMetadata.isDeployment()) {
      transients.conflation().deploymentInfo().newDeploymentSansExecutionAt(recipientAddress);
    }
  }

  @Override
  public void resolvePostTransaction(Hub hub, WorldView world, Transaction tx, boolean statusCode) {
    checkArgument(statusCode, "TX_SKIP transactions should be successful");
    checkArgument(txMetadata.statusCode(), "meta data suggests an unsuccessful TX_SKIP");

    if (txMetadata.noAddressCollisions()) {

      final AccountSnapshot senderAccountSnapshotAfter =
          AccountSnapshot.canonical(hub, world, senderAccountSnapshotBefore.address());

      final AccountSnapshot recipientAccountSnapshotAfter =
          AccountSnapshot.canonical(hub, world, recipientAccountSnapshotBefore.address());

      final AccountSnapshot coinbaseAccountSnapshotAfter =
          AccountSnapshot.canonical(hub, world, coinbaseAccountSnapshotBefore.address());

      // sender account fragment
      final AccountFragment senderAccountFragment =
          hub.factories()
              .accountFragment()
              .makeWithTrm(
                  senderAccountSnapshotBefore,
                  senderAccountSnapshotAfter,
                  senderAccountSnapshotBefore.address(),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

      // recipient account fragment
      final AccountFragment recipientAccountFragment =
          hub.factories()
              .accountFragment()
              .makeWithTrm(
                  recipientAccountSnapshotBefore,
                  recipientAccountSnapshotAfter,
                  recipientAccountSnapshotBefore.address(),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

      // coinbase account fragment
      final AccountFragment coinbaseAccountFragment =
          hub.factories()
              .accountFragment()
              .makeWithTrm(
                  coinbaseAccountSnapshotBefore,
                  coinbaseAccountSnapshotAfter,
                  coinbaseAccountSnapshotBefore.address(),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2));

      this.addFragments(senderAccountFragment, recipientAccountFragment, coinbaseAccountFragment);
    } else {
      final Wei value = (Wei) txMetadata.getBesuTransaction().getValue();

      final AccountSnapshot firstAccountFragmentSnapshotAfter =
          txMetadata.senderAddressCollision()
              ? senderAccountSnapshotBefore
                  .deepCopy()
                  .decrementBalanceBy(
                      value.add(txMetadata.getGasUsed() * txMetadata.getEffectiveGasPrice()))
                  .raiseNonceByOne()
              : AccountSnapshot.canonical(hub, world, senderAccountSnapshotBefore.address());

      final AccountSnapshot secondAccountFragmentSnapshotBefore =
          txMetadata.senderIsRecipient()
              ? firstAccountFragmentSnapshotAfter.deepCopy()
              : recipientAccountSnapshotBefore;

      final AccountSnapshot secondAccountFragmentSnapshotAfter =
          secondAccountFragmentSnapshotBefore.deepCopy().incrementBalanceBy(value);

      final AccountSnapshot coinbaseAccountSnapshotAfter =
          AccountSnapshot.canonical(hub, world, coinbaseAccountSnapshotBefore.address());

      final AccountSnapshot thirdAccountFragmentSnapshotBefore =
          txMetadata.coinbaseAddressCollision()
              ? coinbaseAccountSnapshotAfter
                  .deepCopy()
                  .decrementBalanceBy(txMetadata.getCoinbaseReward())
              : coinbaseAccountSnapshotBefore;

      // "sender" account fragment
      final AccountFragment firstAccountFragment =
          hub.factories()
              .accountFragment()
              .makeWithTrm(
                  senderAccountSnapshotBefore,
                  firstAccountFragmentSnapshotAfter,
                  senderAccountSnapshotBefore.address(),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

      // "recipient" account fragment
      final AccountFragment secondAccountFragment =
          hub.factories()
              .accountFragment()
              .makeWithTrm(
                  secondAccountFragmentSnapshotBefore,
                  secondAccountFragmentSnapshotAfter,
                  recipientAccountSnapshotBefore.address(),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

      // "coinbase" account fragment
      final AccountFragment thirdAccountFragment =
          hub.factories()
              .accountFragment()
              .makeWithTrm(
                  thirdAccountFragmentSnapshotBefore,
                  coinbaseAccountSnapshotAfter,
                  coinbaseAccountSnapshotBefore.address(),
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2));

      this.addFragments(firstAccountFragment, secondAccountFragment, thirdAccountFragment);
    }

    // transaction fragment
    final TransactionFragment transactionFragment =
        TransactionFragment.prepare(hub.txStack().current());

    this.addFragment(transactionFragment);
  }
}
