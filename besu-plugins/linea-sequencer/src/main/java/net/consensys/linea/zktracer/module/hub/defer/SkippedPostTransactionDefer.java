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

package net.consensys.linea.zktracer.module.hub.defer;

import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TransactionFragment;
import net.consensys.linea.zktracer.module.hub.section.TxSkippedSection;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.worldstate.WorldView;

/**
 * SkippedTransaction latches data at the pre-execution of the transaction data that will be used
 * later, through a {@link PostTransactionDefer}, to generate the trace chunks required for the
 * proving of a pure transaction.
 *
 * @param oldFromAccount
 * @param oldToAccount
 * @param oldMinerAccount
 */
public record SkippedPostTransactionDefer(
    AccountSnapshot oldFromAccount,
    AccountSnapshot oldToAccount,
    AccountSnapshot oldMinerAccount,
    Wei gasPrice,
    Wei baseFee)
    implements PostTransactionDefer {
  @Override
  public void runPostTx(Hub hub, WorldView state, Transaction tx) {
    Address fromAddress = this.oldFromAccount.address();
    Address toAddress = this.oldToAccount.address();
    Address minerAddress = this.oldMinerAccount.address();
    hub.conflation().deploymentInfo().unmarkDeploying(toAddress);

    AccountSnapshot newFromAccount =
        AccountSnapshot.fromAccount(
            state.get(fromAddress),
            true,
            hub.conflation().deploymentInfo().number(fromAddress),
            false);

    AccountSnapshot newToAccount =
        AccountSnapshot.fromAccount(
            state.get(fromAddress),
            true,
            hub.conflation().deploymentInfo().number(toAddress),
            false);

    AccountSnapshot newMinerAccount =
        AccountSnapshot.fromAccount(
            state.get(minerAddress),
            true,
            hub.conflation().deploymentInfo().number(minerAddress),
            false);

    // Append the final chunk to the hub chunks
    hub.addTraceSection(
        new TxSkippedSection(
            hub,
            // 3 lines -- account changes
            // From
            new AccountFragment(oldFromAccount, newFromAccount, false, 0, false),
            // To
            new AccountFragment(oldToAccount, newToAccount, false, 0, false),
            // Miner
            new AccountFragment(oldMinerAccount, newMinerAccount, false, 0, false),

            // 1 line -- transaction data
            TransactionFragment.prepare(
                hub.conflation().number(),
                minerAddress,
                tx,
                false,
                gasPrice,
                baseFee,
                hub.tx().initialGas())));
  }
}
