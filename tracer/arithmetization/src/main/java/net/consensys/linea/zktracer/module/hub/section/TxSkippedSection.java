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
import org.hyperledger.besu.evm.worldstate.WorldView;

/**
 * SkippedTransaction latches data at the pre-execution of the transaction data that will be used
 * later, through a {@link PostTransactionDefer}, to generate the trace chunks required for the
 * proving of a pure transaction.
 */
public class TxSkippedSection extends TraceSection implements PostTransactionDefer {

  final TransactionProcessingMetadata txMetadata;
  final AccountSnapshot senderAccountSnapshotBefore;
  final AccountSnapshot recipientAccountSnapshotBefore;
  final AccountSnapshot coinbaseAccountSnapshotBefore;

  public TxSkippedSection(
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
        AccountSnapshot.canonical(hub, world, senderAddress, isPrecompile(senderAddress));
    recipientAccountSnapshotBefore =
        AccountSnapshot.canonical(hub, world, recipientAddress, isPrecompile(recipientAddress));
    coinbaseAccountSnapshotBefore =
        AccountSnapshot.canonical(hub, world, coinbaseAddress, isPrecompile(coinbaseAddress));

    // arithmetization restriction
    checkArgument(
        !isPrecompile(recipientAddress),
        "Arithmetization restriction: recipient address is a precompile.");

    // sanity check + EIP-3607
    checkArgument(world.get(senderAddress) != null);
    checkArgument(!world.get(senderAddress).hasCode());

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
  /**
   * This doesn't take into account the "sender = recipient, recipient = coinbase, coinbase =
   * sender" shenanigans.
   *
   * <p>TODO: issue #1018, {@link https://github.com/Consensys/linea-tracer/issues/1018}
   */
  public void resolvePostTransaction(Hub hub, WorldView state, Transaction tx, boolean statusCode) {
    final Address senderAddress = txMetadata.getBesuTransaction().getSender();
    final Address recipientAddress = txMetadata.getEffectiveRecipient();
    final Address coinbaseAddress = txMetadata.getCoinbase();

    checkArgument(statusCode, "TX_SKIP transactions should be successful");
    checkArgument(
        statusCode == txMetadata.statusCode(), "meta data suggests an unsuccessful TX_SKIP");

    final AccountSnapshot senderAccountSnapshotAfter =
        AccountSnapshot.fromAccount(
            state.get(senderAddress),
            this.senderAccountSnapshotBefore.isWarm(),
            hub.deploymentNumberOf(senderAddress),
            false);

    final AccountSnapshot recipientAccountSnapshotAfter =
        AccountSnapshot.fromAccount(
            state.get(recipientAddress),
            this.recipientAccountSnapshotBefore.isWarm(),
            hub.deploymentNumberOf(recipientAddress),
            false);

    final AccountSnapshot coinbaseAccountSnapshotAfter =
        AccountSnapshot.fromAccount(
            state.get(coinbaseAddress),
            this.coinbaseAccountSnapshotBefore.isWarm(),
            hub.deploymentNumberOf(coinbaseAddress),
            false);

    // sender account fragment
    final AccountFragment senderAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                senderAccountSnapshotBefore,
                senderAccountSnapshotAfter,
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));

    // recipient account fragment
    final AccountFragment recipientAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                recipientAccountSnapshotBefore,
                recipientAccountSnapshotAfter,
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

    // coinbase account fragment
    final AccountFragment coinbaseAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                coinbaseAccountSnapshotBefore,
                coinbaseAccountSnapshotAfter,
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2));

    // transaction fragment
    final TransactionFragment transactionFragment =
        TransactionFragment.prepare(hub.txStack().current());

    this.addFragments(
        senderAccountFragment,
        recipientAccountFragment,
        coinbaseAccountFragment,
        transactionFragment);
  }
}
