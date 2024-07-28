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

import java.util.List;
import java.util.Optional;

import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostExecDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.defer.ReEnterContextDefer;
import net.consensys.linea.zktracer.module.hub.fragment.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.ScenarioFragment;
import net.consensys.linea.zktracer.module.hub.precompiles.PrecompileInvocation;
import net.consensys.linea.zktracer.module.hub.precompiles.PrecompileLinesGenerator;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class NoCodeCallSection extends TraceSection
    implements PostTransactionDefer, PostExecDefer, ReEnterContextDefer {
  private final Bytes rawCalledAddress;
  private final Optional<PrecompileInvocation> precompileInvocation;
  private final CallFrame callerCallFrame;
  private final int calledCallFrameId;
  private boolean callSuccessful = false;
  private final AccountSnapshot preCallCallerAccountSnapshot;
  private final AccountSnapshot preCallCalledAccountSnapshot;

  private AccountSnapshot postCallCallerAccountSnapshot;
  private AccountSnapshot postCallCalledAccountSnapshot;
  private final ImcFragment imcFragment;
  private final ScenarioFragment scenarioFragment;

  private Optional<List<TraceFragment>> maybePrecompileLines = Optional.empty();

  public NoCodeCallSection(
      Hub hub,
      Optional<PrecompileInvocation> targetPrecompile,
      AccountSnapshot preCallCallerAccountSnapshot,
      AccountSnapshot preCallCalledAccountSnapshot,
      Bytes rawCalledAddress,
      ImcFragment imcFragment) {
    this.rawCalledAddress = rawCalledAddress;
    this.precompileInvocation = targetPrecompile;
    this.preCallCallerAccountSnapshot = preCallCallerAccountSnapshot;
    this.preCallCalledAccountSnapshot = preCallCalledAccountSnapshot;
    this.callerCallFrame = hub.currentFrame();
    this.calledCallFrameId = hub.callStack().futureId();
    this.imcFragment = imcFragment;
    this.scenarioFragment =
        ScenarioFragment.forNoCodeCallSection(
            hub, precompileInvocation, this.callerCallFrame.id(), this.calledCallFrameId);
    this.addStack(hub);

    hub.defers().postExec(this);
    hub.defers().reEntry(this);
    hub.defers().postTx(this);
  }

  @Override
  public void runAtReEnter(Hub hub, MessageFrame frame) {
    // The precompile lines will read the return data, so they need to be added after re-entry.
    this.maybePrecompileLines =
        this.precompileInvocation.map(p -> PrecompileLinesGenerator.generateFor(hub, p));
  }

  @Override
  public void runPostExec(Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    this.callSuccessful = !frame.getStackItem(0).isZero();
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
  public void runPostTx(Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    final AccountFragment.AccountFragmentFactory accountFragmentFactory =
        hub.factories().accountFragment();
    this.scenarioFragment.runPostTx(hub, state, tx, isSuccessful);

    this.addFragmentsWithoutStack(
        hub,
        callerCallFrame,
        this.scenarioFragment,
        this.imcFragment,
        ContextFragment.readContextData(hub.callStack()),
        accountFragmentFactory.make(
            this.preCallCallerAccountSnapshot, this.postCallCallerAccountSnapshot),
        accountFragmentFactory.makeWithTrm(
            this.preCallCalledAccountSnapshot,
            this.postCallCalledAccountSnapshot,
            this.rawCalledAddress));

    if (precompileInvocation.isPresent()) {
      if (this.callSuccessful && callerCallFrame.hasReverted()) {
        this.addFragmentsWithoutStack(
            hub,
            callerCallFrame,
            accountFragmentFactory.make(
                this.postCallCallerAccountSnapshot, this.preCallCallerAccountSnapshot),
            accountFragmentFactory.make(
                this.postCallCalledAccountSnapshot, this.preCallCalledAccountSnapshot));
      }
      this.addFragmentsWithoutStack(
          hub,
          ScenarioFragment.forPrecompileEpilogue(
              hub, precompileInvocation.get(), callerCallFrame.id(), calledCallFrameId));

      // for (TraceFragment f :
      //    this.maybePrecompileLines.orElseThrow(
      //        () ->
      //            new IllegalStateException(
      //                "missing precompile lines for %s".formatted(callerCallFrame.opCode())))) {
      //  this.addFragment(hub, callerCallFrame, f);
      // }

      // TODO: temporary hack, we should have only successful EcData Operations, will be fixed in PR
      // #748
      // This class is going to die in this PR
      if (this.maybePrecompileLines.isPresent()) {
        for (TraceFragment f : this.maybePrecompileLines.get()) {
          this.addFragment(hub, callerCallFrame, f);
        }
      }
    } else {
      if (callerCallFrame.hasReverted()) {
        this.addFragmentsWithoutStack(
            hub,
            callerCallFrame,
            accountFragmentFactory.make(
                this.postCallCallerAccountSnapshot, this.preCallCallerAccountSnapshot),
            accountFragmentFactory.make(
                this.postCallCalledAccountSnapshot, this.preCallCalledAccountSnapshot));
      }
      this.addFragmentsWithoutStack(
          hub, callerCallFrame, ContextFragment.nonExecutionEmptyReturnData(hub.callStack()));
    }
  }
}
