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
import net.consensys.linea.zktracer.module.hub.defer.NextContextDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostExecDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.ScenarioFragment;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class SmartContractCallSection extends TraceSection
    implements PostTransactionDefer, PostExecDefer, NextContextDefer {
  private final Bytes rawCalledAddress;
  private final CallFrame callerCallFrame;
  private final int calledCallFrameId;
  private final AccountSnapshot preCallCallerAccountSnapshot;
  private final AccountSnapshot preCallCalledAccountSnapshot;

  private AccountSnapshot inCallCallerAccountSnapshot;
  private AccountSnapshot inCallCalledAccountSnapshot;

  private AccountSnapshot postCallCallerAccountSnapshot;
  private AccountSnapshot postCallCalledAccountSnapshot;

  private final ScenarioFragment scenarioFragment;
  private final ImcFragment imcFragment;

  public SmartContractCallSection(
      Hub hub,
      AccountSnapshot preCallCallerAccountSnapshot,
      AccountSnapshot preCallCalledAccountSnapshot,
      Bytes rawCalledAddress,
      ImcFragment imcFragment) {
    this.rawCalledAddress = rawCalledAddress;
    this.callerCallFrame = hub.currentFrame();
    this.calledCallFrameId = hub.callStack().futureId();
    this.preCallCallerAccountSnapshot = preCallCallerAccountSnapshot;
    this.preCallCalledAccountSnapshot = preCallCalledAccountSnapshot;
    this.imcFragment = imcFragment;
    this.scenarioFragment =
        ScenarioFragment.forSmartContractCallSection(
            hub, calledCallFrameId, this.callerCallFrame.id());

    this.addStack(hub);

    hub.defers().postExec(this);
    hub.defers().nextContext(this, hub.currentFrame().id());
    hub.defers().postTx(this);
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
            hub.transients().conflation().deploymentInfo().number(callerAddress),
            hub.transients().conflation().deploymentInfo().isDeploying(callerAddress));
    this.postCallCalledAccountSnapshot =
        AccountSnapshot.fromAccount(
            calledAccount,
            frame.isAddressWarm(calledAddress),
            hub.transients().conflation().deploymentInfo().number(calledAddress),
            hub.transients().conflation().deploymentInfo().isDeploying(calledAddress));
  }

  @Override
  public void runNextContext(Hub hub, MessageFrame frame) {
    final Address callerAddress = preCallCallerAccountSnapshot.address();
    final Account callerAccount = frame.getWorldUpdater().get(callerAddress);
    final Address calledAddress = preCallCalledAccountSnapshot.address();
    final Account calledAccount = frame.getWorldUpdater().get(calledAddress);

    this.inCallCallerAccountSnapshot =
        AccountSnapshot.fromAccount(
            callerAccount,
            frame.isAddressWarm(callerAddress),
            hub.transients().conflation().deploymentInfo().number(callerAddress),
            hub.transients().conflation().deploymentInfo().isDeploying(callerAddress));
    this.inCallCalledAccountSnapshot =
        AccountSnapshot.fromAccount(
            calledAccount,
            frame.isAddressWarm(calledAddress),
            hub.transients().conflation().deploymentInfo().number(calledAddress),
            hub.transients().conflation().deploymentInfo().isDeploying(calledAddress));
  }

  @Override
  public void runPostTx(Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    final AccountFragment.AccountFragmentFactory accountFragmentFactory =
        hub.factories().accountFragment();
    final CallFrame calledCallFrame = hub.callStack().getById(this.calledCallFrameId);
    this.scenarioFragment.runPostTx(hub, state, tx, isSuccessful);

    this.addFragmentsWithoutStack(
        hub,
        callerCallFrame,
        this.scenarioFragment,
        ContextFragment.readContextData(hub.callStack()),
        this.imcFragment,
        accountFragmentFactory.make(
            this.preCallCallerAccountSnapshot, this.inCallCallerAccountSnapshot),
        accountFragmentFactory.makeWithTrm(
            this.preCallCalledAccountSnapshot,
            this.inCallCalledAccountSnapshot,
            this.rawCalledAddress));

    if (callerCallFrame.hasReverted()) {
      if (calledCallFrame.hasReverted()) {
        this.addFragmentsWithoutStack(
            hub,
            callerCallFrame,
            accountFragmentFactory.make(
                this.inCallCallerAccountSnapshot, this.preCallCallerAccountSnapshot),
            accountFragmentFactory.make(
                this.inCallCalledAccountSnapshot, this.postCallCalledAccountSnapshot),
            accountFragmentFactory.make(
                this.postCallCalledAccountSnapshot, this.preCallCalledAccountSnapshot));
      } else {
        this.addFragmentsWithoutStack(
            hub,
            callerCallFrame,
            accountFragmentFactory.make(
                this.inCallCallerAccountSnapshot, this.preCallCallerAccountSnapshot),
            accountFragmentFactory.make(
                this.inCallCalledAccountSnapshot, this.preCallCalledAccountSnapshot));
      }
    } else {
      if (calledCallFrame.hasReverted()) {
        this.addFragmentsWithoutStack(
            hub,
            callerCallFrame,
            accountFragmentFactory.make(
                this.inCallCallerAccountSnapshot, this.postCallCallerAccountSnapshot),
            accountFragmentFactory.make(
                this.inCallCalledAccountSnapshot, this.postCallCalledAccountSnapshot));
      }
    }

    this.addFragmentsWithoutStack(
        hub, callerCallFrame, ContextFragment.enterContext(hub.callStack(), calledCallFrame));
  }
}
