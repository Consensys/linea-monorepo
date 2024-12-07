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
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CreateScenarioFragment.CreateScenario.*;
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
import net.consensys.linea.zktracer.opcode.OpCode;
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
        ContextEntryDefer,
        PostRollbackDefer,
        ContextReEntryDefer,
        PostTransactionDefer {

  private final Address creatorAddress;
  private final Address createeAddress;

  private final AccountFragment.AccountFragmentFactory accountFragmentFactory;

  // Just before create
  private AccountSnapshot preOpcodeCreatorSnapshot;
  private AccountSnapshot preOpcodeCreateeSnapshot;

  // Just at the entry child frame
  private AccountSnapshot childContextEntryCreatorSnapshot;
  private AccountSnapshot childContextEntryCreateeSnapshot;

  // Just at the entry child frame
  private AccountSnapshot reEntryCreatorSnapshot;
  private AccountSnapshot reEntryCreateeSnapshot;

  private RlpAddrSubFragment rlpAddrSubFragment;

  final CreateScenarioFragment scenarioFragment; // row i + 0
  final ContextFragment currentContextFragment; // row i + 1
  final ImcFragment imcFragment; // row i + 2
  private ContextFragment finalContextFragment; // row i+?

  private boolean requiresRomLex;
  private Wei value;
  private boolean success = false;

  // TODO: according to our preliminary conclusion in issue #866
  //  CREATE's that raise a failure condition _do spawn a child context_.
  public CreateSection(Hub hub, MessageFrame frame) {
    super(hub, maxNumberOfLines(hub.pch().exceptions(), hub.pch().abortingConditions()));
    accountFragmentFactory = hub.factories().accountFragment();

    creatorAddress = frame.getRecipientAddress();
    createeAddress = getDeploymentAddress(frame);
    value = Wei.of(UInt256.fromBytes(frame.getStackItem(0)));

    scenarioFragment = new CreateScenarioFragment();
    currentContextFragment = ContextFragment.readCurrentContextData(hub);
    imcFragment = ImcFragment.empty(hub);

    this.addStack(hub);
    this.addFragment(scenarioFragment);
    this.addFragment(currentContextFragment);
    this.addFragment(imcFragment);

    refineCreateScenario(hub, frame);
    scheduleSection(hub);

    final short exceptions = hub.pch().exceptions();

    // STATICX case
    if (Exceptions.staticFault(exceptions)) {
      return;
    }

    // MXPX case
    final MxpCall mxpCall = new MxpCall(hub);
    imcFragment.callMxp(mxpCall);
    checkArgument(mxpCall.mxpx == Exceptions.memoryExpansionException(exceptions));
    if (mxpCall.mxpx) {
      return;
    }

    // OOGX case
    final StpCall stpCall = new StpCall(hub, frame, mxpCall.getGasMxp());
    imcFragment.callStp(stpCall);
    checkArgument(stpCall.outOfGasException() == Exceptions.outOfGasException(exceptions));
    if (Exceptions.outOfGasException(exceptions)) {
      return;
    }

    // The CREATE(2) is now unexceptional
    /////////////////////////////////////

    checkArgument(Exceptions.none(exceptions));
    hub.currentFrame().childSpanningSection(this);

    final CreateOobCall oobCall = new CreateOobCall();
    imcFragment.callOob(oobCall);

    preOpcodeCreatorSnapshot =
        AccountSnapshot.canonical(hub, frame.getWorldUpdater(), creatorAddress);
    preOpcodeCreateeSnapshot =
        AccountSnapshot.canonical(hub, frame.getWorldUpdater(), createeAddress);

    final boolean aborts = scenarioFragment.getScenario() == CREATE_ABORT;
    final boolean failedCreate =
        scenarioFragment.getScenario() == CREATE_FAILURE_CONDITION_WONT_REVERT;
    final boolean emptyInitCode =
        scenarioFragment.getScenario() == CREATE_EMPTY_INIT_CODE_WONT_REVERT;

    checkArgument(oobCall.isAbortingCondition() == aborts);
    if (aborts) {
      this.traceAbort(hub);
      return;
    }

    // The CREATE(2) is now unexceptional, unaborted
    ////////////////////////////////////////////////

    rlpAddrSubFragment = RlpAddrSubFragment.makeFragment(hub, createeAddress);

    final long offset = Words.clampedToLong(frame.getStackItem(1));
    final long size = Words.clampedToLong(frame.getStackItem(2));

    // Trigger MMU & SHAKIRA to hash the (non-empty) InitCode of CREATE2 - even for failed CREATE2
    if (nontrivialCreate2(hub.opCode(), size)) {
      final Bytes create2InitCode = frame.shadowReadMemory(offset, size);

      final MmuCall mmuCall = MmuCall.create2(hub, create2InitCode, failedCreate);
      imcFragment.callMmu(mmuCall);

      final ShakiraDataOperation shakiraDataOperation =
          new ShakiraDataOperation(hub.stamp(), create2InitCode);
      hub.shakiraData().call(shakiraDataOperation);

      writeHashInfoResult(shakiraDataOperation.result());
    }

    if (failedCreate) {
      finalContextFragmentSquashesReturnData(hub);
      commonValues.payGasPaidOutOfPocket(hub);
      hub.failureConditionForCreates = true;
      return;
    }

    if (emptyInitCode) {
      success = true;
      finalContextFragmentSquashesReturnData(hub);
      hub.transients().conflation().deploymentInfo().newDeploymentSansExecutionAt(createeAddress);
      return;
    }

    // unexceptional, unaborted, non-failing, non-emptyInitCode CREATE(2)
    /////////////////////////////////////////////////////////////////////

    // we charge for the gas paid out of pocket
    commonValues.payGasPaidOutOfPocket(hub);

    requiresRomLex = true;
    hub.romLex().callRomLex(frame);
    hub.transients()
        .conflation()
        .deploymentInfo()
        .newDeploymentWithExecutionAt(createeAddress, frame.shadowReadMemory(offset, size));

    // Note: the case CREATE2 has been set before, we need to do it even in the failure case
    if (hub.opCode() == CREATE) {
      final MmuCall mmuCall = MmuCall.create(hub);
      imcFragment.callMmu(mmuCall);
    }

    finalContextFragment = ContextFragment.initializeExecutionContext(hub);
  }

  @Override
  public void resolveUponContextEntry(Hub hub, MessageFrame frame) {
    childContextEntryCreatorSnapshot =
        AccountSnapshot.canonical(hub, frame.getWorldUpdater(), preOpcodeCreatorSnapshot.address())
            // .raiseNonceByOne() // the nonce was already raised
            .decrementBalanceBy(value);
    childContextEntryCreateeSnapshot =
        AccountSnapshot.canonical(hub, frame.getWorldUpdater(), preOpcodeCreateeSnapshot.address())
            .raiseNonceByOne()
            .incrementBalanceBy(value);

    final AccountFragment creatorAccountFragment =
        accountFragmentFactory.make(
            preOpcodeCreatorSnapshot,
            childContextEntryCreatorSnapshot,
            DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));
    creatorAccountFragment.rlpAddrSubFragment(rlpAddrSubFragment);

    final AccountFragment createeAccountFragment =
        accountFragmentFactory.makeWithTrm(
            preOpcodeCreateeSnapshot,
            childContextEntryCreateeSnapshot,
            createeAddress.trimLeadingZeros(),
            DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

    createeAccountFragment.requiresRomlex(requiresRomLex);

    this.addFragments(creatorAccountFragment, createeAccountFragment);
  }

  @Override
  public void resolveAtContextReEntry(Hub hub, CallFrame frame) {
    success = !frame.frame().getStackItem(0).isZero();

    CreateScenarioFragment.CreateScenario scenario = scenarioFragment.getScenario();

    switch (scenario) {
      case CREATE_FAILURE_CONDITION_WONT_REVERT -> {
        checkState(!success);
        reEntryCreatorSnapshot = preOpcodeCreatorSnapshot.deepCopy().raiseNonceByOne();
        reEntryCreateeSnapshot = preOpcodeCreateeSnapshot.deepCopy().turnOnWarmth();
        final AccountFragment firstCreatorFragment =
            accountFragmentFactory.make(
                preOpcodeCreatorSnapshot,
                reEntryCreatorSnapshot,
                DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));
        firstCreatorFragment.rlpAddrSubFragment(rlpAddrSubFragment);

        final AccountFragment firstCreateeFragment =
            accountFragmentFactory.makeWithTrm(
                preOpcodeCreateeSnapshot,
                reEntryCreateeSnapshot,
                createeAddress.trimLeadingZeros(),
                DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

        this.addFragments(firstCreatorFragment, firstCreateeFragment);
        return;
      }
      case CREATE_EMPTY_INIT_CODE_WONT_REVERT -> {
        checkState(success);
        reEntryCreatorSnapshot =
            AccountSnapshot.canonical(hub, frame.frame().getWorldUpdater(), creatorAddress);
        reEntryCreateeSnapshot =
            AccountSnapshot.canonical(hub, frame.frame().getWorldUpdater(), createeAddress);
        final AccountFragment firstCreatorFragment =
            accountFragmentFactory.make(
                preOpcodeCreatorSnapshot,
                reEntryCreatorSnapshot,
                DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));
        firstCreatorFragment.rlpAddrSubFragment(rlpAddrSubFragment);

        final AccountFragment firstCreateeFragment =
            accountFragmentFactory.makeWithTrm(
                preOpcodeCreateeSnapshot,
                reEntryCreateeSnapshot,
                createeAddress.trimLeadingZeros(),
                DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

        this.addFragments(firstCreatorFragment, firstCreateeFragment);
        return;
      }
      default -> {}
    }

    if (success) {
      scenarioFragment.setScenario(CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WONT_REVERT);
      return;
    }

    reEntryCreatorSnapshot = childContextEntryCreatorSnapshot.deepCopy().incrementBalanceBy(value);
    reEntryCreateeSnapshot =
        childContextEntryCreateeSnapshot
            .deepCopy()
            .decrementBalanceBy(value)
            .deploymentStatus(false)
            .deploymentNumber(hub.deploymentNumberOf(createeAddress))
            .nonce(0);

    final int childRevertStamp = hub.getLastChildCallFrame(frame).revertStamp();

    final AccountFragment undoCreator =
        accountFragmentFactory.make(
            childContextEntryCreatorSnapshot,
            reEntryCreatorSnapshot,
            DomSubStampsSubFragment.revertsWithChildDomSubStamps(
                this.hubStamp(), childRevertStamp, 0));

    final AccountFragment undoCreatee =
        accountFragmentFactory.make(
            childContextEntryCreateeSnapshot,
            reEntryCreateeSnapshot,
            DomSubStampsSubFragment.revertsWithChildDomSubStamps(
                this.hubStamp(), childRevertStamp, 1));

    this.addFragments(undoCreator, undoCreatee);
  }

  @Override
  public void resolveUponRollback(Hub hub, MessageFrame messageFrame, CallFrame callFrame) {

    final CreateScenarioFragment.CreateScenario scenario = scenarioFragment.getScenario();
    checkState(
        scenario.isAnyOf(
            CREATE_FAILURE_CONDITION_WONT_REVERT,
            CREATE_EMPTY_INIT_CODE_WONT_REVERT,
            CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WONT_REVERT,
            CREATE_NON_EMPTY_INIT_CODE_FAILURE_WONT_REVERT));
    scenarioFragment.setScenario(switchToRevertingScenario(scenario));

    final int revertStamp = callFrame.revertStamp();

    switch (scenario) {
      case CREATE_FAILURE_CONDITION_WONT_REVERT, CREATE_EMPTY_INIT_CODE_WONT_REVERT -> {
        final AccountFragment undoCreator =
            accountFragmentFactory.make(
                reEntryCreatorSnapshot.deepCopy().setDeploymentInfo(hub),
                preOpcodeCreatorSnapshot.deepCopy().setDeploymentInfo(hub),
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    this.hubStamp(), revertStamp, 0));

        final AccountFragment undoCreatee =
            accountFragmentFactory.make(
                reEntryCreateeSnapshot.deepCopy().setDeploymentInfo(hub),
                preOpcodeCreateeSnapshot.deepCopy().setDeploymentInfo(hub),
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    this.hubStamp(), revertStamp, 1));
        this.addFragments(undoCreator, undoCreatee);
        return;
      }
        // case CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WONT_REVERT ->
        // revertNonEmptyInitCodeDeploymentSuccessCase();
        // case CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WILL_REVERT ->
        // revertNonEmptyInitCodeDeploymentFailureCase();
      default -> {}
    }

    final boolean firstUndo =
        scenarioFragment.getScenario() != CREATE_NON_EMPTY_INIT_CODE_FAILURE_WILL_REVERT;

    final AccountFragment undoCreator =
        accountFragmentFactory.make(
            firstUndo ? childContextEntryCreatorSnapshot : reEntryCreatorSnapshot,
            preOpcodeCreatorSnapshot.deepCopy().setDeploymentInfo(hub),
            DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                this.hubStamp(), revertStamp, firstUndo ? 0 : 2));

    final AccountFragment undoCreatee =
        accountFragmentFactory.make(
            firstUndo ? childContextEntryCreateeSnapshot : reEntryCreateeSnapshot,
            preOpcodeCreateeSnapshot.deepCopy().setDeploymentInfo(hub),
            DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                this.hubStamp(), revertStamp, firstUndo ? 1 : 3));

    this.addFragments(undoCreator, undoCreatee);
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

  private void traceAbort(final Hub hub) {
    final AccountFragment creatorAccountFragment =
        accountFragmentFactory.make(
            preOpcodeCreatorSnapshot,
            preOpcodeCreatorSnapshot,
            DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));

    finalContextFragmentSquashesReturnData(hub);

    this.addFragments(creatorAccountFragment, finalContextFragment);
  }

  private void finalContextFragmentSquashesReturnData(Hub hub) {
    finalContextFragment = ContextFragment.nonExecutionProvidesEmptyReturnData(hub);
  }

  private static CreateScenarioFragment.CreateScenario switchToRevertingScenario(
      final CreateScenarioFragment.CreateScenario previousScenario) {
    return switch (previousScenario) {
      case CREATE_FAILURE_CONDITION_WONT_REVERT -> CREATE_FAILURE_CONDITION_WILL_REVERT;
      case CREATE_EMPTY_INIT_CODE_WONT_REVERT -> CREATE_EMPTY_INIT_CODE_WILL_REVERT;
      case CREATE_NON_EMPTY_INIT_CODE_FAILURE_WONT_REVERT -> CREATE_NON_EMPTY_INIT_CODE_FAILURE_WILL_REVERT;
      case CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WONT_REVERT -> CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WILL_REVERT;
      default -> throw new IllegalArgumentException("unexpected Create scenario");
    };
  }

  private boolean hasEmptyInitCode(Hub hub) {
    return hub.transients().op().initCodeSegment().isEmpty();
  }

  private void refineCreateScenario(Hub hub, MessageFrame frame) {
    if (hub.isExceptional()) {
      scenarioFragment.setScenario(CREATE_EXCEPTION);
      return;
    }

    if (hub.pch().abortingConditions().any()) {
      scenarioFragment.setScenario(CREATE_ABORT);
      return;
    }

    if (raisesFailureCondition(frame)) {
      scenarioFragment.setScenario(CREATE_FAILURE_CONDITION_WONT_REVERT);
      return;
    }

    scenarioFragment.setScenario(
        hasEmptyInitCode(hub)
            ? CREATE_EMPTY_INIT_CODE_WONT_REVERT
            : CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WONT_REVERT);
  }

  private void scheduleSection(Hub hub) {
    CreateScenarioFragment.CreateScenario scenario = scenarioFragment.getScenario();
    final CallFrame currentFrame = hub.currentFrame();
    switch (scenario) {
      case CREATE_EXCEPTION -> {}
      case CREATE_ABORT -> hub.defers().scheduleForPostExecution(this); // unlatch the stack
      case CREATE_FAILURE_CONDITION_WONT_REVERT, CREATE_EMPTY_INIT_CODE_WONT_REVERT -> {
        hub.defers().scheduleForContextReEntry(this, currentFrame);
        hub.defers().scheduleForPostRollback(this, currentFrame);
        hub.defers().scheduleForPostTransaction(this);
      }
      case CREATE_NON_EMPTY_INIT_CODE_SUCCESS_WONT_REVERT -> {
        // The current execution context pays (63/64)ths of it current gas to the child context
        // To get the success bit of the CREATE(2) operation
        hub.defers().scheduleForContextEntry(this);
        hub.defers().scheduleForContextReEntry(this, currentFrame);
        hub.defers().scheduleForPostRollback(this, currentFrame);
        hub.defers().scheduleForPostTransaction(this);

        // we capture revert information about the child context: CCSR and CCRS
        hub.defers().scheduleForContextReEntry(imcFragment, hub.currentFrame());
      }
      default -> throw new IllegalStateException(
          scenario.name() + " not allowed when defining the schedule");
    }
  }

  private boolean raisesFailureCondition(MessageFrame frame) {

    final Optional<Account> deploymentAccount =
        Optional.ofNullable(frame.getWorldUpdater().get(createeAddress));
    final boolean createdAddressHasNonZeroNonce =
        deploymentAccount.map(a -> a.getNonce() != 0).orElse(false);
    final boolean createdAddressHasNonEmptyCode =
        deploymentAccount.map(AccountState::hasCode).orElse(false);

    return createdAddressHasNonZeroNonce || createdAddressHasNonEmptyCode;
  }

  // we unlatched the stack after a CREATE if and only if we don't "contextEnter" the CREATE.
  // "failure condition CREATE's" do enter the CREATE context.
  @Override
  public void resolvePostExecution(
      Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    checkState(scenarioFragment.isAbortedCreate());
    hub.unlatchStack(frame, this);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    addFragment(finalContextFragment);
  }

  private boolean nontrivialCreate2(OpCode opCode, long size) {
    return (opCode == CREATE2 && size != 0);
  }
}
