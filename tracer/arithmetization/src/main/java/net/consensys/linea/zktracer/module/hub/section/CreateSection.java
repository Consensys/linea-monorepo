/*
 * Copyright ConsenSys Inc.
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
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_ABORT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_EMPTY_INIT_CODE_WILL_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_EMPTY_INIT_CODE_WONT_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_EXCEPTION;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_FAILURE_CONDITION_WILL_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_FAILURE_CONDITION_WONT_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_NON_EMPTY_INIT_CODE_FAILURE_WILL_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_NON_EMPTY_INIT_CODE_FAILURE_WONT_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WILL_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WONT_REVERT;
import static net.consensys.linea.zktracer.opcode.OpCode.*;
import static net.consensys.linea.zktracer.types.AddressUtils.getDeploymentAddress;

import java.util.Optional;

import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.*;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.RlpAddrSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.MxpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.StpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.opcodes.CreateOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment;
import net.consensys.linea.zktracer.module.hub.signals.AbortingConditions;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.module.shakiradata.ShakiraDataOperation;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class CreateSection extends TraceSection
    implements PostOpcodeDefer,
        ImmediateContextEntryDefer,
        PostRollbackDefer,
        ContextReEntryDefer,
        PostTransactionDefer {

  // Just before create
  private AccountSnapshot preOpcodeCreatorSnapshot;
  private AccountSnapshot preOpcodeCreateeSnapshot;

  // Just at the entry child frame
  private AccountSnapshot childEntryCreatorSnapshot;
  private AccountSnapshot childEntryCreateeSnapshot;

  // Just at the entry child frame
  private AccountSnapshot reEntryCreatorSnapshot;
  private AccountSnapshot reEntryCreateeSnapshot;

  private RlpAddrSubFragment rlpAddrSubFragment;

  // row i+0
  final CreateScenarioFragment scenarioFragment;
  // row i+?
  private ContextFragment finalContextFragment;

  private boolean requiresRomLex;
  private Wei value;

  // TODO: according to our preliminary conclusion in issue #866
  //  CREATE's that raise a failure condition _do spawn a child context_.
  public CreateSection(Hub hub) {
    super(hub, maxNumberOfLines(hub.pch().exceptions(), hub.pch().abortingConditions()));
    final short exceptions = hub.pch().exceptions();

    this.addStack(hub);

    // row  i+ 0
    scenarioFragment = new CreateScenarioFragment();
    this.addFragment(scenarioFragment);

    // row i + 1
    final ContextFragment currentContextFragment = ContextFragment.readCurrentContextData(hub);
    this.addFragment(currentContextFragment);

    // row: i + 2
    final ImcFragment imcFragment = ImcFragment.empty(hub);
    this.addFragment(imcFragment);

    // STATICX case
    // Note: in the static case this imc fragment remains empty
    if (Exceptions.staticFault(exceptions)) {
      scenarioFragment.setScenario(CREATE_EXCEPTION);
      return;
    }

    final MxpCall mxpCall = new MxpCall(hub);
    imcFragment.callMxp(mxpCall);
    checkArgument(mxpCall.mxpx == Exceptions.memoryExpansionException(exceptions));

    // MXPX case
    if (mxpCall.mxpx) {
      scenarioFragment.setScenario(CREATE_EXCEPTION);
      return;
    }

    final StpCall stpCall = new StpCall(hub, mxpCall.getGasMxp());
    imcFragment.callStp(stpCall);

    checkArgument(stpCall.outOfGasException() == Exceptions.outOfGasException(exceptions));

    // OOGX case
    if (Exceptions.outOfGasException(exceptions)) {
      scenarioFragment.setScenario(CREATE_EXCEPTION);
      return;
    }

    // The CREATE(2) is now unexceptional
    checkArgument(Exceptions.none(exceptions));
    hub.currentFrame().childSpanningSection(this);

    final CreateOobCall oobCall = new CreateOobCall();
    imcFragment.callOob(oobCall);

    final AbortingConditions aborts = hub.pch().abortingConditions().snapshot();
    checkArgument(oobCall.isAbortingCondition() == aborts.any());

    final CallFrame callFrame = hub.currentFrame();
    final MessageFrame messageFrame = hub.messageFrame();

    final Address creatorAddress = callFrame.accountAddress();
    preOpcodeCreatorSnapshot = AccountSnapshot.canonical(hub, creatorAddress);

    final Address createeAddress = getDeploymentAddress(messageFrame);
    preOpcodeCreateeSnapshot = AccountSnapshot.canonical(hub, createeAddress);

    if (aborts.any()) {
      scenarioFragment.setScenario(CREATE_ABORT);
      this.finishAbort(hub);
      hub.defers().scheduleForPostExecution(this);
      return;
    }

    // The CREATE(2) is now unexceptional and unaborted
    checkArgument(aborts.none());
    hub.defers().scheduleForImmediateContextEntry(this); // when we add the two account fragments
    hub.defers().scheduleForPostRollback(this, hub.currentFrame()); // in case of Rollback
    hub.defers().scheduleForPostTransaction(this); // when we add the last context row

    rlpAddrSubFragment = RlpAddrSubFragment.makeFragment(hub, createeAddress);

    final Optional<Account> deploymentAccount =
        Optional.ofNullable(messageFrame.getWorldUpdater().get(createeAddress));
    final boolean createdAddressHasNonZeroNonce =
        deploymentAccount.map(a -> a.getNonce() != 0).orElse(false);
    final boolean createdAddressHasNonEmptyCode =
        deploymentAccount.map(AccountState::hasCode).orElse(false);

    final boolean failedCreate = createdAddressHasNonZeroNonce || createdAddressHasNonEmptyCode;
    final boolean emptyInitCode = hub.transients().op().initCodeSegment().isEmpty();

    final long offset = Words.clampedToLong(messageFrame.getStackItem(1));
    final long size = Words.clampedToLong(messageFrame.getStackItem(2));

    // Trigger MMU & SHAKIRA to hash the (non-empty) InitCode of CREATE2 - even for failed CREATE2
    if (hub.opCode() == CREATE2 && !emptyInitCode) {
      final Bytes create2InitCode = messageFrame.shadowReadMemory(offset, size);

      final MmuCall mmuCall = MmuCall.create2(hub, create2InitCode, failedCreate);
      imcFragment.callMmu(mmuCall);

      final ShakiraDataOperation shakiraDataOperation =
          new ShakiraDataOperation(hub.stamp(), create2InitCode);
      hub.shakiraData().call(shakiraDataOperation);

      writeHashInfoResult(shakiraDataOperation.result());
    }

    value = failedCreate ? Wei.ZERO : Wei.of(UInt256.fromBytes(messageFrame.getStackItem(0)));

    if (failedCreate) {
      finalContextFragment = ContextFragment.nonExecutionProvidesEmptyReturnData(hub);
      scenarioFragment.setScenario(CREATE_FAILURE_CONDITION_WONT_REVERT);
      hub.failureConditionForCreates = true;
      return;
    }

    if (emptyInitCode) {
      finalContextFragment = ContextFragment.nonExecutionProvidesEmptyReturnData(hub);
      scenarioFragment.setScenario(CREATE_EMPTY_INIT_CODE_WONT_REVERT);
      hub.transients().conflation().deploymentInfo().newDeploymentSansExecutionAt(createeAddress);
      return;
    }

    // Finally, non-exceptional, non-aborting, non-failing, non-emptyInitCode create
    ////////////////////////////////////////////////////////////////////////////////

    // we capture revert information about the child context: CCSR and CCRS
    hub.defers().scheduleForContextReEntry(imcFragment, hub.currentFrame());

    // The current execution context pays (63/64)ths of it current gas to the child context
    commonValues.payGasPaidOutOfPocket(hub);
    hub.defers()
        .scheduleForContextReEntry(this, callFrame); // To get the success bit of the CREATE(2)

    requiresRomLex = true;
    hub.romLex().callRomLex(messageFrame);
    hub.transients()
        .conflation()
        .deploymentInfo()
        .newDeploymentWithExecutionAt(createeAddress, messageFrame.shadowReadMemory(offset, size));

    // Note: the case CREATE2 has been set before, we need to do it even in the failure case
    if (hub.opCode() == CREATE) {
      final MmuCall mmuCall = MmuCall.create(hub);
      imcFragment.callMmu(mmuCall);
    }

    finalContextFragment = ContextFragment.initializeNewExecutionContext(hub);
  }

  @Override
  public void resolveUponContextEntry(Hub hub) {
    childEntryCreatorSnapshot =
        AccountSnapshot.canonical(hub, preOpcodeCreatorSnapshot.address())
            // .raiseNonceByOne() // the nonce was already raised
            .decrementBalanceBy(value);
    childEntryCreateeSnapshot =
        AccountSnapshot.canonical(hub, preOpcodeCreateeSnapshot.address())
            .raiseNonceByOne()
            .incrementBalanceBy(value);

    final AccountFragment.AccountFragmentFactory accountFragmentFactory =
        hub.factories().accountFragment();

    final AccountFragment creatorAccountFragment =
        accountFragmentFactory.make(
            preOpcodeCreatorSnapshot,
            childEntryCreatorSnapshot,
            DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));
    creatorAccountFragment.rlpAddrSubFragment(rlpAddrSubFragment);

    final AccountFragment createeAccountFragment =
        accountFragmentFactory.makeWithTrm(
            preOpcodeCreateeSnapshot,
            childEntryCreateeSnapshot,
            preOpcodeCreateeSnapshot.address().trimLeadingZeros(),
            DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

    createeAccountFragment.requiresRomlex(requiresRomLex);

    this.addFragments(creatorAccountFragment, createeAccountFragment);
  }

  @Override
  public void resolveAtContextReEntry(Hub hub, CallFrame frame) {
    final boolean deploymentSuccess = !frame.frame().getStackItem(0).isZero();

    if (deploymentSuccess) {
      scenarioFragment.setScenario(CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WONT_REVERT);
    } else {
      scenarioFragment.setScenario(CREATE_NON_EMPTY_INIT_CODE_FAILURE_WONT_REVERT);

      reEntryCreatorSnapshot = AccountSnapshot.canonical(hub, preOpcodeCreatorSnapshot.address());
      reEntryCreateeSnapshot = AccountSnapshot.canonical(hub, preOpcodeCreateeSnapshot.address());

      final AccountFragment.AccountFragmentFactory accountFragmentFactory =
          hub.factories().accountFragment();

      final int childRevertStamp = hub.getLastChildCallFrame(frame).revertStamp();

      final AccountFragment undoCreator =
          accountFragmentFactory.make(
              childEntryCreatorSnapshot,
              reEntryCreatorSnapshot,
              DomSubStampsSubFragment.revertsWithChildDomSubStamps(
                  this.hubStamp(), childRevertStamp, 2));

      final AccountFragment undoCreatee =
          accountFragmentFactory.make(
              childEntryCreateeSnapshot,
              reEntryCreateeSnapshot,
              DomSubStampsSubFragment.revertsWithChildDomSubStamps(
                  this.hubStamp(), childRevertStamp, 3));

      this.addFragments(undoCreator, undoCreatee);
    }
  }

  @Override
  public void resolveUponRollback(Hub hub, MessageFrame messageFrame, CallFrame callFrame) {
    scenarioFragment.setScenario(switchToRevert(scenarioFragment.getScenario()));

    final AccountFragment.AccountFragmentFactory accountFragmentFactory =
        hub.factories().accountFragment();

    final int revertStamp = callFrame.revertStamp();
    final boolean firstUndo =
        scenarioFragment.getScenario() != CREATE_NON_EMPTY_INIT_CODE_FAILURE_WILL_REVERT;

    final AccountFragment undoCreator =
        accountFragmentFactory.make(
            firstUndo ? childEntryCreatorSnapshot : reEntryCreatorSnapshot,
            preOpcodeCreatorSnapshot,
            DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                this.hubStamp(), revertStamp, firstUndo ? 2 : 4));

    final AccountFragment undoCreatee =
        accountFragmentFactory.make(
            firstUndo ? childEntryCreateeSnapshot : reEntryCreateeSnapshot,
            preOpcodeCreateeSnapshot,
            DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                this.hubStamp(), revertStamp, firstUndo ? 3 : 5));

    this.addFragments(undoCreator, undoCreatee);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    addFragment(finalContextFragment);
  }

  private static short maxNumberOfLines(final short exceptions, final AbortingConditions abort) {
    if (Exceptions.any(exceptions)) {
      return 6;
    }
    if (abort.any()) {
      return 7;
    }
    return 11; // Note: could be lower for unreverted successful CREATE(s)
  }

  private void finishAbort(final Hub hub) {
    final AccountFragment.AccountFragmentFactory accountFragmentFactory =
        hub.factories().accountFragment();
    final AccountFragment creatorAccountFragment =
        accountFragmentFactory.make(
            preOpcodeCreateeSnapshot,
            preOpcodeCreateeSnapshot,
            DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));

    final ContextFragment updatedCurrentContextFragment =
        ContextFragment.nonExecutionProvidesEmptyReturnData(hub);

    this.addFragments(creatorAccountFragment, updatedCurrentContextFragment);
  }

  private static CreateScenarioFragment.CreateScenario switchToRevert(
      final CreateScenarioFragment.CreateScenario previousScenario) {
    return switch (previousScenario) {
      case CREATE_FAILURE_CONDITION_WONT_REVERT -> CREATE_FAILURE_CONDITION_WILL_REVERT;
      case CREATE_EMPTY_INIT_CODE_WONT_REVERT -> CREATE_EMPTY_INIT_CODE_WILL_REVERT;
      case CREATE_NON_EMPTY_INIT_CODE_FAILURE_WONT_REVERT -> CREATE_NON_EMPTY_INIT_CODE_FAILURE_WILL_REVERT;
      case CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WONT_REVERT -> CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WILL_REVERT;
      default -> throw new IllegalArgumentException("unexpected Create scenario");
    };
  }

  public boolean isAbortedCreate() {
    return scenarioFragment.isAbortedCreate();
  }

  // we unlatched the stack after a CREATE if and only if we don't "contextEnter" the CREATE.
  // "failure condition CREATE's" do enter the CREATE context.
  @Override
  public void resolvePostExecution(
      Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    checkState(isAbortedCreate());
    hub.unlatchStack(frame, this);
  }
}
