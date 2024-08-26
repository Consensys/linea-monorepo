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
  final AccountSnapshot oldFromAccount;
  final AccountSnapshot oldToAccount;
  final AccountSnapshot oldMinerAccount;

  public TxSkippedSection(
      Hub hub, WorldView world, TransactionProcessingMetadata txMetadata, Transients transients) {
    super(hub, (short) 4);
    this.txMetadata = txMetadata;
    hub.defers().scheduleForPostTransaction(this);

    // From account information
    final Address fromAddress = txMetadata.getBesuTransaction().getSender();
    oldFromAccount =
        AccountSnapshot.fromAccount(
            world.get(fromAddress),
            isPrecompile(fromAddress),
            transients.conflation().deploymentInfo().number(fromAddress),
            false);

    // To account information
    final Address toAddress = txMetadata.getEffectiveTo();
    if (txMetadata.isDeployment()) {
      transients.conflation().deploymentInfo().deploy(toAddress);
    }
    oldToAccount =
        AccountSnapshot.fromAccount(
            world.get(toAddress),
            isPrecompile(toAddress),
            transients.conflation().deploymentInfo().number(toAddress),
            false);

    // Miner account information
    final Address minerAddress = txMetadata.getCoinbase();
    oldMinerAccount =
        AccountSnapshot.fromAccount(
            world.get(minerAddress),
            isPrecompile(minerAddress),
            transients.conflation().deploymentInfo().number(transients.block().minerAddress()),
            false);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    final Address fromAddress = this.oldFromAccount.address();
    final Address toAddress = this.oldToAccount.address();
    final Address minerAddress = this.txMetadata.getCoinbase();
    hub.transients().conflation().deploymentInfo().unmarkDeploying(toAddress);

    final AccountSnapshot newFromAccount =
        AccountSnapshot.fromAccount(
            state.get(fromAddress),
            this.oldFromAccount.isWarm(),
            hub.transients().conflation().deploymentInfo().number(fromAddress),
            false);

    final AccountSnapshot newToAccount =
        AccountSnapshot.fromAccount(
            state.get(toAddress),
            this.oldToAccount.isWarm(),
            hub.transients().conflation().deploymentInfo().number(toAddress),
            false);

    final AccountSnapshot newMinerAccount =
        AccountSnapshot.fromAccount(
            state.get(minerAddress),
            this.oldMinerAccount.isWarm(),
            hub.transients().conflation().deploymentInfo().number(minerAddress),
            false);

    // From
    this.addFragment(
        hub.factories()
            .accountFragment()
            .make(
                oldFromAccount,
                newFromAccount,
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0)));

    // To
    this.addFragment(
        hub.factories()
            .accountFragment()
            .make(
                oldToAccount,
                newToAccount,
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1)));

    // Miner
    this.addFragment(
        hub.factories()
            .accountFragment()
            .make(
                oldMinerAccount,
                newMinerAccount,
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 2)));

    // Transaction data
    this.addFragment(TransactionFragment.prepare(hub.txStack().current()));
  }
}
