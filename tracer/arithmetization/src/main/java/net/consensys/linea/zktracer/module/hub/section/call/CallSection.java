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

package net.consensys.linea.zktracer.module.hub.section.call;

import static net.consensys.linea.zktracer.module.hub.AccountSnapshot.canonical;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.*;
import static net.consensys.linea.zktracer.types.AddressUtils.isPrecompile;
import static net.consensys.linea.zktracer.types.Conversions.bytesToBoolean;
import static org.hyperledger.besu.datatypes.Address.*;

import java.util.Map;
import java.util.Optional;
import java.util.function.BiFunction;

import com.google.common.base.Preconditions;
import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Factories;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.ContextExitDefer;
import net.consensys.linea.zktracer.module.hub.defer.ContextReEntryDefer;
import net.consensys.linea.zktracer.module.hub.defer.ImmediateContextEntryDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostOpcodeDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostRollbackDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.MxpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.StpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.opcodes.CallOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.opcodes.XCallOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.module.hub.section.call.precompileSubsection.*;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class CallSection extends TraceSection
    implements PostOpcodeDefer,
        ImmediateContextEntryDefer,
        ContextExitDefer,
        ContextReEntryDefer,
        PostRollbackDefer,
        PostTransactionDefer {

  private static final Map<Address, BiFunction<Hub, CallSection, PrecompileSubsection>>
      ADDRESS_TO_PRECOMPILE =
          Map.of(
              ECREC, EllipticCurvePrecompileSubsection::new,
              SHA256, ShaTwoOrRipemdSubSection::new,
              RIPEMD160, ShaTwoOrRipemdSubSection::new,
              ID, IdentitySubsection::new,
              MODEXP, ModexpSubsection::new,
              ALTBN128_ADD, EllipticCurvePrecompileSubsection::new,
              ALTBN128_MUL, EllipticCurvePrecompileSubsection::new,
              ALTBN128_PAIRING, EllipticCurvePrecompileSubsection::new,
              BLAKE2B_F_COMPRESSION, BlakeSubsection::new);

  public Optional<Address> precompileAddress;

  // row i+0
  private final CallScenarioFragment scenarioFragment = new CallScenarioFragment();

  // last row
  @Setter private ContextFragment finalContextFragment;

  private Bytes rawCalleeAddress;

  // Just before the CALL Opcode
  private AccountSnapshot preOpcodeCallerSnapshot;
  private AccountSnapshot preOpcodeCalleeSnapshot;

  // Just after the CALL Opcode
  private AccountSnapshot postOpcodeCallerSnapshot;
  private AccountSnapshot postOpcodeCalleeSnapshot;

  // Just before re-entry
  private AccountSnapshot childContextExitCallerSnapshot;
  private AccountSnapshot childContextExitCalleeSnapshot;

  // Just after re-entry
  private AccountSnapshot reEntryCallerSnapshot;
  private AccountSnapshot reEntryCalleeSnapshot;

  private boolean selfCallWithNonzeroValueTransfer;

  private Wei value;

  private AccountSnapshot postRollbackCalleeSnapshot;
  private AccountSnapshot postRollbackCallerSnapshot;

  //
  private PrecompileSubsection precompileSubsection;

  public CallSection(Hub hub) {
    super(hub, maxNumberOfLines(hub));

    final short exceptions = hub.pch().exceptions();

    // row i + 1
    final ContextFragment currentContextFragment = ContextFragment.readCurrentContextData(hub);
    // row i + 2
    final ImcFragment firstImcFragment = ImcFragment.empty(hub);

    this.addStackAndFragments(hub, scenarioFragment, currentContextFragment, firstImcFragment);

    if (Exceptions.any(exceptions)) {
      scenarioFragment.setScenario(CALL_EXCEPTION);
      final XCallOobCall oobCall = new XCallOobCall();
      firstImcFragment.callOob(oobCall);
    }

    // STATICX cases
    if (Exceptions.staticFault(exceptions)) {
      return;
    }

    final MxpCall mxpCall = new MxpCall(hub);
    firstImcFragment.callMxp(mxpCall);
    Preconditions.checkArgument(mxpCall.mxpx == Exceptions.memoryExpansionException(exceptions));

    // MXPX case
    if (Exceptions.memoryExpansionException(exceptions)) {
      return;
    }

    final StpCall stpCall = new StpCall(hub, mxpCall.gasMxp);
    firstImcFragment.callStp(stpCall);
    Preconditions.checkArgument(
        stpCall.outOfGasException() == Exceptions.outOfGasException(exceptions));

    final Address callerAddress = hub.currentFrame().callerAddress();
    preOpcodeCallerSnapshot = canonical(hub, callerAddress);

    rawCalleeAddress = hub.currentFrame().frame().getStackItem(1);
    final Address calleeAddress = Address.extract(EWord.of(rawCalleeAddress)); // TODO check this
    preOpcodeCalleeSnapshot = canonical(hub, calleeAddress);

    // OOGX case
    if (Exceptions.outOfGasException(exceptions)) {
      this.oogXCall(hub);
      return;
    }

    // The CALL is now unexceptional
    Preconditions.checkArgument(Exceptions.none(exceptions));
    hub.currentFrame().childSpanningSection(this);

    final CallOobCall oobCall = new CallOobCall();
    firstImcFragment.callOob(oobCall);

    final boolean aborts = hub.pch().abortingConditions().any();
    Preconditions.checkArgument(oobCall.isAbortingCondition() == aborts);

    hub.defers().scheduleForPostRollback(this, hub.currentFrame());
    hub.defers().scheduleForPostTransaction(this);

    if (aborts) {
      this.abortingCall(hub);
      hub.defers().scheduleForPostExecution(this);
      return;
    }

    // The CALL is now unexceptional and un-aborted
    hub.defers().scheduleForImmediateContextEntry(this);
    hub.defers().scheduleForContextReEntry(this, hub.currentFrame());
    final WorldUpdater world = hub.messageFrame().getWorldUpdater();

    if (isPrecompile(calleeAddress)) {
      precompileAddress = Optional.of(calleeAddress);
      scenarioFragment.setScenario(CALL_PRC_UNDEFINED);
      // Account rows for precompile are traced at contextReEntry

      precompileSubsection =
          ADDRESS_TO_PRECOMPILE.get(preOpcodeCalleeSnapshot.address()).apply(hub, this);
    } else {
      Optional.ofNullable(world.get(calleeAddress))
          .ifPresentOrElse(
              account -> {
                scenarioFragment.setScenario(
                    account.hasCode() ? CALL_SMC_UNDEFINED : CALL_EOA_SUCCESS_WONT_REVERT);
              },
              () -> {
                scenarioFragment.setScenario(CALL_EOA_SUCCESS_WONT_REVERT);
              });

      // TODO is world == worldUpdater & what happen if get doesn't work ?

      ;

      // TODO is world == worldUpdater & what happen if get
      //  doesn't work ?
      // TODO: write a test where the recipient of the call does not exist in the state
    }

    if (scenarioFragment.getScenario() == CALL_SMC_UNDEFINED) {
      finalContextFragment = ContextFragment.initializeNewExecutionContext(hub);
      final boolean callCanTransferValue = hub.currentFrame().opCode().callCanTransferValue();
      final boolean isSelfCall = callerAddress.equals(calleeAddress);
      value = Wei.of(hub.messageFrame().getStackItem(2).toUnsignedBigInteger());
      selfCallWithNonzeroValueTransfer = isSelfCall && callCanTransferValue && !value.isZero();
      hub.romLex().callRomLex(hub.currentFrame().frame());
      hub.defers().scheduleForContextExit(this, hub.callStack().futureId());
    }

    if (scenarioFragment.getScenario() == CALL_EOA_SUCCESS_WONT_REVERT) {
      finalContextFragment = ContextFragment.nonExecutionProvidesEmptyReturnData(hub);
    }
  }

  private static short maxNumberOfLines(final Hub hub) {
    // 99 % of the time this number of rows will be sufficient
    if (Exceptions.any(hub.pch().exceptions())) {
      return 8;
    }
    if (hub.pch().abortingConditions().any()) {
      return 9;
    }
    return 12; // 12 = 2 (stack) + 5 (CALL prequel) + 5 (successful PRC, except BLAKE and MODEXP)
  }

  private void oogXCall(Hub hub) {

    final Factories factories = hub.factories();
    final AccountFragment callerAccountFragment =
        factories
            .accountFragment()
            .make(
                preOpcodeCallerSnapshot,
                preOpcodeCallerSnapshot,
                DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));

    final AccountFragment calleeAccountFragment =
        factories
            .accountFragment()
            .makeWithTrm(
                preOpcodeCalleeSnapshot,
                preOpcodeCalleeSnapshot,
                rawCalleeAddress,
                DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

    this.addFragments(callerAccountFragment, calleeAccountFragment);
  }

  private void abortingCall(Hub hub) {
    scenarioFragment.setScenario(CALL_ABORT_WONT_REVERT);
    finalContextFragment = ContextFragment.nonExecutionProvidesEmptyReturnData(hub);
  }

  @Override
  public void resolvePostExecution(
      Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    Preconditions.checkArgument(scenarioFragment.getScenario() == CALL_ABORT_WONT_REVERT);
    postOpcodeCallerSnapshot = canonical(hub, preOpcodeCallerSnapshot.address());
    postOpcodeCalleeSnapshot = canonical(hub, preOpcodeCalleeSnapshot.address());
  }

  @Override
  public void resolveUponImmediateContextEntry(Hub hub) {
    postOpcodeCallerSnapshot = canonical(hub, preOpcodeCallerSnapshot.address());
    postOpcodeCalleeSnapshot = canonical(hub, preOpcodeCalleeSnapshot.address());

    switch (scenarioFragment.getScenario()) {
      case CALL_SMC_UNDEFINED -> {
        if (selfCallWithNonzeroValueTransfer) {
          // In case of a self-call that transfers value, the balance of the caller
          // is decremented by the value transferred. This becomes the initial state
          // of the callee, which is then credited by that value. This can happen
          // only for the SMC case.
          postOpcodeCallerSnapshot.decrementBalanceBy(value);
          preOpcodeCalleeSnapshot.decrementBalanceBy(value);
        }

        final Factories factories = hub.factories();
        final AccountFragment firstCallerAccountFragment =
            factories
                .accountFragment()
                .make(
                    preOpcodeCallerSnapshot,
                    postOpcodeCallerSnapshot,
                    DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));

        final AccountFragment firstCalleeAccountFragment =
            factories
                .accountFragment()
                .makeWithTrm(
                    preOpcodeCalleeSnapshot,
                    postOpcodeCalleeSnapshot,
                    rawCalleeAddress,
                    DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

        firstCalleeAccountFragment.requiresRomlex(true);

        this.addFragments(firstCallerAccountFragment, firstCalleeAccountFragment);
      }

      case CALL_PRC_UNDEFINED -> {}

      case CALL_EOA_SUCCESS_WONT_REVERT -> {
        // Account rows for EOA calls are traced at contextReEntry
        return;
      }

      default -> throw new IllegalArgumentException("Should be in one of the three scenario above");
    }
  }

  /** Resolution happens as the child context is about to terminate. */
  @Override
  public void resolveUponExitingContext(Hub hub, CallFrame frame) {
    Preconditions.checkArgument(scenarioFragment.getScenario() == CALL_SMC_UNDEFINED);

    childContextExitCallerSnapshot = canonical(hub, preOpcodeCallerSnapshot.address());
    childContextExitCalleeSnapshot = canonical(hub, preOpcodeCalleeSnapshot.address());

    // TODO: what follows assumes that the caller's stack has been updated
    //  to contain the success bit of the call at traceContextReEntry.
    //  See issue #872.
    // TODO: when does the callFrame update its output data?
    // TODO: when does the callFrame update to the parent callFrame ?
    finalContextFragment.returnDataContextNumber(hub.currentFrame().contextNumber());
    finalContextFragment.returnDataSegment(hub.currentFrame().outputDataSpan());
  }

  @Override
  public void resolveAtContextReEntry(Hub hub, CallFrame frame) {
    // TODO: what follows assumes that the caller's stack has been updated
    //  to contain the success bit of the call at traceContextReEntry.
    //  See issue #872.
    // The callSuccess will only be set
    // if the call is acted upon i.e. if the call is un-exceptional and un-aborted
    final boolean successBit = bytesToBoolean(hub.messageFrame().getStackItem(0));

    reEntryCallerSnapshot = canonical(hub, preOpcodeCallerSnapshot.address());
    reEntryCalleeSnapshot = canonical(hub, preOpcodeCalleeSnapshot.address());

    switch (scenarioFragment.getScenario()) {
      case CALL_EOA_SUCCESS_WONT_REVERT -> {
        emptyCodeFirstCoupleOfAccountFragments(hub);
      }

      case CALL_PRC_UNDEFINED -> {
        if (successBit) {
          scenarioFragment.setScenario(CALL_PRC_SUCCESS_WONT_REVERT);
        } else {
          scenarioFragment.setScenario(CALL_PRC_FAILURE);
        }
        emptyCodeFirstCoupleOfAccountFragments(hub);
      }

      case CALL_SMC_UNDEFINED -> {
        if (successBit) {
          scenarioFragment.setScenario(CALL_SMC_SUCCESS_WONT_REVERT);
          return;
        }

        scenarioFragment.setScenario(CALL_SMC_FAILURE_WONT_REVERT);

        if (selfCallWithNonzeroValueTransfer) {
          childContextExitCallerSnapshot.decrementBalanceBy(value);
          reEntryCalleeSnapshot.decrementBalanceBy(value);
        }

        final AccountFragment postReEntryCallerAccountFragment =
            hub.factories()
                .accountFragment()
                .make(
                    childContextExitCallerSnapshot,
                    reEntryCallerSnapshot,
                    DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                        this.hubStamp(), this.revertStamp(), 2));

        final AccountFragment postReEntryCalleeAccountFragment =
            hub.factories()
                .accountFragment()
                .make(
                    childContextExitCalleeSnapshot,
                    reEntryCalleeSnapshot,
                    DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                        this.hubStamp(), this.revertStamp(), 3));

        this.addFragments(postReEntryCallerAccountFragment, postReEntryCalleeAccountFragment);
      }

      default -> throw new IllegalArgumentException("Illegal CALL scenario");
    }
  }

  @Override
  public void resolvePostRollback(Hub hub, MessageFrame messageFrame, CallFrame callFrame) {
    final Factories factory = hub.factories();
    postRollbackCalleeSnapshot = canonical(hub, preOpcodeCalleeSnapshot.address());
    postRollbackCallerSnapshot = canonical(hub, preOpcodeCallerSnapshot.address());

    final CallScenarioFragment.CallScenario callScenario = scenarioFragment.getScenario();
    switch (callScenario) {
      case CALL_ABORT_WONT_REVERT -> completeAbortWillRevert(factory);
      case CALL_EOA_SUCCESS_WONT_REVERT -> completeEoaSuccessWillRevert(factory);
      case CALL_SMC_FAILURE_WONT_REVERT -> completeSmcFailureWillRevert(factory);
      case CALL_SMC_SUCCESS_WONT_REVERT,
          CALL_PRC_SUCCESS_WONT_REVERT -> completeSmcSuccessWillRevertOrPrcSuccessWillRevert(
          factory);
      case CALL_PRC_FAILURE -> {
        // Note: no undoing required
        //  - account snapshots were taken with value transfers undone
        //  - precompiles are warm by definition so no warmth undoing required
        return;
      }
      default -> throw new IllegalArgumentException("Illegal CALL scenario");
    }
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {

    final CallScenarioFragment.CallScenario scenario = scenarioFragment.getScenario();

    Preconditions.checkArgument(scenario.noLongerUndefined());

    if (scenario.isPrecompileScenario()) {
      this.addFragments(precompileSubsection.fragments());
    }

    this.addFragment(finalContextFragment);
  }

  private void completeAbortWillRevert(Factories factory) {
    scenarioFragment.setScenario(CALL_ABORT_WILL_REVERT);
    final AccountFragment undoingCalleeAccountFragment =
        factory
            .accountFragment()
            .make(
                postOpcodeCalleeSnapshot,
                postRollbackCalleeSnapshot,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    this.hubStamp(), this.revertStamp(), 0));
    this.addFragment(undoingCalleeAccountFragment);
  }

  private void completeEoaSuccessWillRevert(Factories factory) {
    scenarioFragment.setScenario(CALL_EOA_SUCCESS_WILL_REVERT);

    final AccountFragment undoingCallerAccountFragment =
        factory
            .accountFragment()
            .make(
                postOpcodeCallerSnapshot,
                postRollbackCallerSnapshot,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    this.hubStamp(), this.revertStamp(), 2));

    final AccountFragment undoingCalleeAccountFragment =
        factory
            .accountFragment()
            .make(
                postOpcodeCalleeSnapshot,
                postRollbackCalleeSnapshot,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    this.hubStamp(), this.revertStamp(), 3));

    this.addFragments(undoingCallerAccountFragment, undoingCalleeAccountFragment);
  }

  private void completeSmcFailureWillRevert(Factories factory) {
    scenarioFragment.setScenario(CALL_SMC_FAILURE_WILL_REVERT);

    // this (should) work for both self calls and foreign address calls
    final AccountFragment undoingCalleeWarmthAccountFragment =
        factory
            .accountFragment()
            .make(
                reEntryCalleeSnapshot,
                postRollbackCalleeSnapshot,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    this.hubStamp(), this.revertStamp(), 4));

    this.addFragment(undoingCalleeWarmthAccountFragment);
  }

  private void completeSmcSuccessWillRevertOrPrcSuccessWillRevert(Factories factory) {

    final CallScenarioFragment.CallScenario callScenario = scenarioFragment.getScenario();
    if (callScenario == CALL_SMC_SUCCESS_WONT_REVERT) {
      scenarioFragment.setScenario(CALL_SMC_SUCCESS_WILL_REVERT);
    } else {
      scenarioFragment.setScenario(CALL_PRC_SUCCESS_WILL_REVERT);
    }

    if (selfCallWithNonzeroValueTransfer) {
      reEntryCallerSnapshot.decrementBalanceBy(value);
      postRollbackCalleeSnapshot.decrementBalanceBy(value);
    }

    final AccountFragment undoingCallerAccountFragment =
        factory
            .accountFragment()
            .make(
                reEntryCallerSnapshot,
                postRollbackCallerSnapshot,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    this.hubStamp(), this.revertStamp(), 2));
    final AccountFragment undoingCalleeAccountFragment =
        factory
            .accountFragment()
            .make(
                reEntryCalleeSnapshot,
                postRollbackCalleeSnapshot,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    this.hubStamp(), this.revertStamp(), 3));

    this.addFragments(undoingCallerAccountFragment, undoingCalleeAccountFragment);
  }

  private void emptyCodeFirstCoupleOfAccountFragments(final Hub hub) {
    final Factories factories = hub.factories();
    final AccountFragment firstCallerAccountFragment =
        factories
            .accountFragment()
            .make(
                preOpcodeCallerSnapshot,
                reEntryCallerSnapshot,
                DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));

    final AccountFragment firstCalleeAccountFragment =
        factories
            .accountFragment()
            .makeWithTrm(
                preOpcodeCalleeSnapshot,
                reEntryCalleeSnapshot,
                rawCalleeAddress,
                DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

    this.addFragments(firstCallerAccountFragment, firstCalleeAccountFragment);
  }
}
