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

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.section.CallSection;
import net.consensys.linea.zktracer.module.runtime.callstack.CallFrame;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;

public record CallDefer(
    AccountSnapshot oldCallerSnapshot,
    AccountSnapshot oldCalledSnapshot,
    ContextFragment preContext,
    CallFrame parentFrame)
    implements PostExecDefer, NextContextDefer {
  // PostExec defer
  @Override
  public void run(Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    // The post-exec behaves identically to the new context defer; albeit with different global
    // state
    this.run(hub, frame);
  }

  // Context defer
  @Override
  public void run(Hub hub, MessageFrame frame) {
    Address callerAddress = oldCallerSnapshot.address();
    AccountSnapshot newCallerSnapshot =
        AccountSnapshot.fromAccount(
            frame.getWorldUpdater().getAccount(callerAddress),
            true,
            hub.conflation().deploymentInfo().number(callerAddress),
            hub.conflation().deploymentInfo().isDeploying(callerAddress));

    Address calledAddress = oldCalledSnapshot.address();
    AccountSnapshot newCalledSnapshot =
        AccountSnapshot.fromAccount(
            frame.getWorldUpdater().getAccount(calledAddress),
            true,
            hub.conflation().deploymentInfo().number(calledAddress),
            hub.conflation().deploymentInfo().isDeploying(calledAddress));

    hub.addTraceSection(
        new CallSection(
            hub,
            parentFrame,
            // 2× own account
            new AccountFragment(oldCallerSnapshot, newCallerSnapshot, false, 0, false),
            new AccountFragment(oldCallerSnapshot, newCallerSnapshot, false, 0, false),
            // 2× target code account
            new AccountFragment(oldCalledSnapshot, newCalledSnapshot, false, 0, false),
            new AccountFragment(oldCalledSnapshot, newCalledSnapshot, false, 0, false),
            // context -- only if not precompile
            new ContextFragment(hub.callStack(), hub.currentFrame(), false)));
  }
}
