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

import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostExecDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.ScenarioFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.misc.MiscFragment;
import net.consensys.linea.zktracer.module.hub.subsection.PrecompileScenarioTraceSubsection;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class NoCodeCallSection extends TraceSection implements PostTransactionDefer, PostExecDefer {
  private final boolean targetIsPrecompile;
  private final CallFrame callerCallFrame;
  private final int calledCallFrameId;
  private final AccountSnapshot preCallCallerAccountSnapshot;
  private final AccountSnapshot preCallCalledAccountSnapshot;

  private AccountSnapshot postCallCallerAccountSnapshot;
  private AccountSnapshot postCallCalledAccountSnapshot;
  private final MiscFragment miscFragment;

  public NoCodeCallSection(
      Hub hub,
      boolean targetIsPrecompile,
      AccountSnapshot preCallCallerAccountSnapshot,
      AccountSnapshot preCallCalledAccountSnapshot,
      MiscFragment miscFragment) {
    this.targetIsPrecompile = targetIsPrecompile;
    this.preCallCallerAccountSnapshot = preCallCallerAccountSnapshot;
    this.preCallCalledAccountSnapshot = preCallCalledAccountSnapshot;
    this.callerCallFrame = hub.currentFrame();
    this.calledCallFrameId = hub.callStack().futureId();
    this.miscFragment = miscFragment;
    for (var stackChunk : hub.makeStackChunks(hub.currentFrame())) {
      this.addChunk(hub, hub.currentFrame(), stackChunk);
    }
  }

  @Override
  public void runPostExec(Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    final Address callerAddress = preCallCallerAccountSnapshot.address();
    final Account callerAccount = frame.getWorldUpdater().get(callerAddress);
    final Address calledAddress = preCallCalledAccountSnapshot.address();
    final Account calledAccount = frame.getWorldUpdater().get(calledAddress);

    this.postCallCallerAccountSnapshot =
        AccountSnapshot.fromAccount(
            callerAccount,
            frame.isAddressWarm(callerAddress),
            hub.conflation().deploymentInfo().number(callerAddress),
            hub.conflation().deploymentInfo().isDeploying(callerAddress));
    this.postCallCalledAccountSnapshot =
        AccountSnapshot.fromAccount(
            calledAccount,
            frame.isAddressWarm(calledAddress),
            hub.conflation().deploymentInfo().number(calledAddress),
            hub.conflation().deploymentInfo().isDeploying(calledAddress));
  }

  @Override
  public void runPostTx(Hub hub, WorldView state, Transaction tx) {
    this.addChunksWithoutStack(
        hub,
        callerCallFrame,
        new ScenarioFragment(
            targetIsPrecompile, false, false, this.callerCallFrame.id(), this.calledCallFrameId),
        new ContextFragment(hub.callStack(), hub.currentFrame(), false),
        this.miscFragment,
        new AccountFragment(this.preCallCallerAccountSnapshot, this.postCallCallerAccountSnapshot),
        new AccountFragment(this.preCallCalledAccountSnapshot, this.postCallCalledAccountSnapshot));

    if (callerCallFrame.hasReverted()) {
      if (targetIsPrecompile) {
        this.addChunksWithoutStack(
            hub,
            callerCallFrame,
            new AccountFragment(
                this.postCallCallerAccountSnapshot, this.preCallCallerAccountSnapshot),
            new AccountFragment(
                this.postCallCalledAccountSnapshot, this.preCallCalledAccountSnapshot));
        for (TraceFragment fragment : new PrecompileScenarioTraceSubsection().generate()) {
          this.addChunk(hub, callerCallFrame, fragment);
        }
      } else {
        this.addChunk(
            hub, callerCallFrame, new ContextFragment(hub.callStack(), this.callerCallFrame, true));
      }
    } else {
      if (targetIsPrecompile) {
        for (TraceFragment fragment : new PrecompileScenarioTraceSubsection().generate()) {
          this.addChunk(hub, callerCallFrame, fragment);
        }
      } else {
        this.addChunk(
            hub, callerCallFrame, new ContextFragment(hub.callStack(), this.callerCallFrame, true));
      }
    }
  }
}
