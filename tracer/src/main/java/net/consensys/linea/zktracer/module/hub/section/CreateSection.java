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

package net.consensys.linea.zktracer.module.hub.section;

import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.NextContextDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostExecDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class CreateSection extends TraceSection
    implements PostExecDefer, NextContextDefer, PostTransactionDefer {
  private final AccountSnapshot oldCreatorSnapshot;
  private final AccountSnapshot oldCreatedSnapshot;
  private AccountSnapshot newCreatorSnapshot;
  private AccountSnapshot newCreatedSnapshot;

  public CreateSection(
      Hub hub, AccountSnapshot oldCreatorSnapshot, AccountSnapshot oldCreatedSnapshot) {
    this.oldCreatorSnapshot = oldCreatorSnapshot;
    this.oldCreatedSnapshot = oldCreatedSnapshot;

    this.addStack(hub);
  }

  @Override
  public void runPostExec(Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    // The post-exec behaves identically to the new context defer; albeit with different global
    // state
    this.runNextContext(hub, frame);
  }

  @Override
  public void runNextContext(Hub hub, MessageFrame frame) {
    Address creatorAddress = oldCreatorSnapshot.address();
    this.newCreatorSnapshot =
        AccountSnapshot.fromAccount(
            frame.getWorldUpdater().getAccount(creatorAddress),
            true,
            hub.conflation().deploymentInfo().number(creatorAddress),
            hub.conflation().deploymentInfo().isDeploying(creatorAddress));

    Address createdAddress = oldCreatedSnapshot.address();
    this.newCreatedSnapshot =
        AccountSnapshot.fromAccount(
            frame.getWorldUpdater().getAccount(createdAddress),
            true,
            hub.conflation().deploymentInfo().number(createdAddress),
            hub.conflation().deploymentInfo().isDeploying(createdAddress));
  }

  @Override
  public void seal(Hub hub) {}

  @Override
  public void runPostTx(Hub hub, WorldView state, Transaction tx) {
    final boolean updateReturnData = false; // TODO:

    this.addChunksWithoutStack(
        hub,
        new ContextFragment(hub.callStack(), hub.currentFrame(), updateReturnData),
        new AccountFragment(oldCreatorSnapshot, newCreatorSnapshot, false, 0, false),
        new AccountFragment(oldCreatorSnapshot, newCreatorSnapshot, false, 0, false),
        new AccountFragment(oldCreatorSnapshot, newCreatorSnapshot, false, 0, false),
        // 2×created account
        new AccountFragment(oldCreatedSnapshot, newCreatedSnapshot, false, 0, true),
        new AccountFragment(oldCreatedSnapshot, newCreatedSnapshot, false, 0, false));
  }
}
