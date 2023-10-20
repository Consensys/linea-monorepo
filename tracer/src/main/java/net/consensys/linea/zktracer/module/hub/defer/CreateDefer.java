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

package net.consensys.linea.zktracer.module.hub.defer;

import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.section.CreateSection;
import net.consensys.linea.zktracer.module.runtime.callstack.CallFrame;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;

public record CreateDefer(
    AccountSnapshot oldCreatorSnapshot,
    AccountSnapshot oldCreatedSnapshot,
    ContextFragment preContext,
    CallFrame callerFrame)
    implements PostExecDefer, NextContextDefer {
  // PostExec defer
  @Override
  public void runPostExec(Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    // The post-exec behaves identically to the new context defer; albeit with different global
    // state
    this.runNextContext(hub, frame);
  }

  // Context defer
  @Override
  public void runNextContext(Hub hub, MessageFrame frame) {
    Address creatorAddress = oldCreatorSnapshot.address();
    AccountSnapshot newCreatorSnapshot =
        AccountSnapshot.fromAccount(
            frame.getWorldUpdater().getAccount(creatorAddress),
            true,
            hub.conflation().deploymentInfo().number(creatorAddress),
            hub.conflation().deploymentInfo().isDeploying(creatorAddress));

    Address createdAddress = oldCreatedSnapshot.address();
    AccountSnapshot newCreatedSnapshot =
        AccountSnapshot.fromAccount(
            frame.getWorldUpdater().getAccount(createdAddress),
            true,
            hub.conflation().deploymentInfo().number(createdAddress),
            hub.conflation().deploymentInfo().isDeploying(createdAddress));

    hub.addTraceSection(
        new CreateSection(
            hub,
            callerFrame,
            preContext,
            // 3× own account
            new AccountFragment(oldCreatorSnapshot, newCreatorSnapshot, false, 0, false),
            new AccountFragment(oldCreatorSnapshot, newCreatorSnapshot, false, 0, false),
            new AccountFragment(oldCreatorSnapshot, newCreatorSnapshot, false, 0, false),
            // 2×created account
            new AccountFragment(oldCreatedSnapshot, newCreatedSnapshot, false, 0, true),
            new AccountFragment(oldCreatedSnapshot, newCreatedSnapshot, false, 0, false)));
  }
}
