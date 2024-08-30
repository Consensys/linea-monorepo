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
      Hub hub, WorldView world, TransactionProcessingMetadata txMetadata, Transients transients) {
    super(hub, (short) 4);
    this.txMetadata = txMetadata;
    hub.defers().scheduleForPostTransaction(this);

    // From account information
    final Address senderAddress = txMetadata.getBesuTransaction().getSender();
    senderAccountSnapshotBefore =
        AccountSnapshot.fromAccount(
            world.get(senderAddress),
            isPrecompile(senderAddress),
            transients.conflation().deploymentInfo().deploymentNumber(senderAddress),
            false);

    // Recipiet account information
    final Address recipientAddress = txMetadata.getEffectiveRecipient();
    recipientAccountSnapshotBefore =
        AccountSnapshot.fromAccount(
            world.get(recipientAddress),
            isPrecompile(recipientAddress),
            transients.conflation().deploymentInfo().deploymentNumber(recipientAddress),
            false);

    // the updated deployment info appears in the "updated" account fragment
    if (txMetadata.isDeployment()) {
      transients.conflation().deploymentInfo().newDeploymentAtForTxSkip(recipientAddress);
    }

    // Coinbase account information
    final Address coinbaseAddress = txMetadata.getCoinbase();
    coinbaseAccountSnapshotBefore =
        AccountSnapshot.fromAccount(
            world.get(coinbaseAddress),
            isPrecompile(coinbaseAddress),
            transients
                .conflation()
                .deploymentInfo()
                .deploymentNumber(transients.block().coinbaseAddress()),
            false);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    final Address senderAddress = this.senderAccountSnapshotBefore.address();
    final Address recipientAddress = this.recipientAccountSnapshotBefore.address();
    final Address coinbaseAddress = this.txMetadata.getCoinbase();

    // TODO: @François: it seems to me this doesn't take into account the
    //  "sender = recipient, recipient = coinbase, coinbase = sender"
    //  shenanigans.
    final AccountSnapshot senderAccountSnapshotAfter =
        AccountSnapshot.fromAccount(
            state.get(senderAddress),
            this.senderAccountSnapshotBefore.isWarm(),
            hub.transients().conflation().deploymentInfo().deploymentNumber(senderAddress),
            false);

    final AccountSnapshot recipientAccountSnapshotAfter =
        AccountSnapshot.fromAccount(
            state.get(recipientAddress),
            this.recipientAccountSnapshotBefore.isWarm(),
            hub.transients().conflation().deploymentInfo().deploymentNumber(recipientAddress),
            false);

    final AccountSnapshot coinbaseAccountSnapshotAfter =
        AccountSnapshot.fromAccount(
            state.get(coinbaseAddress),
            this.coinbaseAccountSnapshotBefore.isWarm(),
            hub.transients().conflation().deploymentInfo().deploymentNumber(coinbaseAddress),
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
