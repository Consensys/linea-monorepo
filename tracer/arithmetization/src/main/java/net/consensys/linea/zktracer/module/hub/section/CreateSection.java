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
import static net.consensys.linea.zktracer.module.shakiradata.HashType.KECCAK;
import static net.consensys.linea.zktracer.opcode.OpCode.CREATE;
import static net.consensys.linea.zktracer.opcode.OpCode.CREATE2;
import static net.consensys.linea.zktracer.types.AddressUtils.getDeploymentAddress;
import static org.hyperledger.besu.crypto.Hash.keccak256;

import java.util.Optional;

import com.google.common.base.Preconditions;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.ContextReEntryDefer;
import net.consensys.linea.zktracer.module.hub.defer.ImmediateContextEntryDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostRollbackDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
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
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class CreateSection extends TraceSection
    implements ImmediateContextEntryDefer,
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
    Preconditions.checkArgument(mxpCall.mxpx == Exceptions.memoryExpansionException(exceptions));

    // MXPX case
    if (mxpCall.mxpx) {
      scenarioFragment.setScenario(CREATE_EXCEPTION);
      return;
    }

    final StpCall stpCall = new StpCall(hub, mxpCall.getGasMxp());
    imcFragment.callStp(stpCall);

    Preconditions.checkArgument(
        stpCall.outOfGasException() == Exceptions.outOfGasException(exceptions));

    // OOGX case
    if (Exceptions.outOfGasException(exceptions)) {
      scenarioFragment.setScenario(CREATE_EXCEPTION);
      return;
    }

    // The CREATE(2) is now unexceptional
    Preconditions.checkArgument(Exceptions.none(exceptions));
    hub.currentFrame().childSpanningSection(this);

    final CreateOobCall oobCall = new CreateOobCall();
    imcFragment.callOob(oobCall);

    final AbortingConditions aborts = hub.pch().abortingConditions().snapshot();
    Preconditions.checkArgument(oobCall.isAbortingCondition() == aborts.any());

    final CallFrame callFrame = hub.currentFrame();
    final MessageFrame frame = callFrame.frame();

    final Address creatorAddress = callFrame.accountAddress();
    preOpcodeCreatorSnapshot = AccountSnapshot.canonical(hub, creatorAddress);

    final Address createeAddress = getDeploymentAddress(frame);
    preOpcodeCreateeSnapshot = AccountSnapshot.canonical(hub, createeAddress);

    if (aborts.any()) {
      scenarioFragment.setScenario(CREATE_ABORT);
      this.finishAbortCreate(hub);
      return;
    }

    // The CREATE(2) is now unexceptional and unaborted
    Preconditions.checkArgument(aborts.none());
    hub.defers().scheduleForImmediateContextEntry(this); // when we add the two account fragments
    hub.defers().scheduleForPostRollback(this, hub.currentFrame()); // in case of Rollback
    hub.defers().scheduleForPostTransaction(this); // when we add the last context row

    rlpAddrSubFragment = RlpAddrSubFragment.makeFragment(hub, createeAddress);

    final Optional<Account> deploymentAccount =
        Optional.ofNullable(frame.getWorldUpdater().get(createeAddress));
    final boolean createdAddressHasNonZeroNonce =
        deploymentAccount.map(a -> a.getNonce() != 0).orElse(false);
    final boolean createdAddressHasNonEmptyCode =
        deploymentAccount.map(AccountState::hasCode).orElse(false);

    final boolean failedCreate = createdAddressHasNonZeroNonce || createdAddressHasNonEmptyCode;
    final boolean emptyInitCode = hub.transients().op().initCodeSegment().isEmpty();

    // Trigger MMU & SHAKIRA to hash the (non-empty) InitCode of CREATE2 - even for failed CREATE2
    if (hub.opCode() == CREATE2 && !emptyInitCode) {
      final MmuCall mmuCall = MmuCall.create2(hub, failedCreate);
      imcFragment.callMmu(mmuCall);

      final long offset = Words.clampedToLong(hub.messageFrame().getStackItem(1));
      final long size = Words.clampedToLong(hub.messageFrame().getStackItem(2));
      final Bytes create2InitCode = frame.shadowReadMemory(offset, size);
      final Bytes32 hash = keccak256(create2InitCode);
      final ShakiraDataOperation shakiraDataOperation =
          new ShakiraDataOperation(hub.stamp(), KECCAK, create2InitCode, hash);
      hub.shakiraData().call(shakiraDataOperation);
    }

    if (failedCreate || emptyInitCode) {
      finalContextFragment = ContextFragment.nonExecutionProvidesEmptyReturnData(hub);

      if (failedCreate) {
        scenarioFragment.setScenario(CREATE_FAILURE_CONDITION_WONT_REVERT);
      }
      if (emptyInitCode) {
        scenarioFragment.setScenario(CREATE_EMPTY_INIT_CODE_WONT_REVERT);
      }

      return;
    }

    // Finally, non-exceptional, non-aborting, non-failing, non-emptyInitCode create
    hub.defers()
        .scheduleForContextReEntry(
            this, hub.currentFrame()); // To get the success bit of the CREATE(2)

    hub.romLex().callRomLex(frame);

    // Note: the case CREATE2 has been set before, we need to do it even in the failure case
    if (hub.opCode() == CREATE) {
      final MmuCall mmuCall = MmuCall.create(hub);
      imcFragment.callMmu(mmuCall);
    }

    this.finalContextFragment = ContextFragment.initializeNewExecutionContext(hub);
  }

  @Override
  public void resolveUponImmediateContextEntry(Hub hub) {
    childEntryCreatorSnapshot = AccountSnapshot.canonical(hub, preOpcodeCreatorSnapshot.address());
    childEntryCreateeSnapshot = AccountSnapshot.canonical(hub, preOpcodeCreateeSnapshot.address());

    final AccountFragment.AccountFragmentFactory accountFragmentFactory =
        hub.factories().accountFragment();

    final AccountFragment creatorAccountFragment =
        accountFragmentFactory.make(
            preOpcodeCreatorSnapshot,
            childEntryCreatorSnapshot,
            DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));
    creatorAccountFragment.rlpAddrSubFragment(rlpAddrSubFragment);

    final AccountFragment createeAccountFragment =
        accountFragmentFactory.make(
            preOpcodeCreateeSnapshot,
            childEntryCreateeSnapshot,
            DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

    this.addFragments(creatorAccountFragment, createeAccountFragment);
  }

  @Override
  public void resolveAtContextReEntry(Hub hub, CallFrame frame) {
    final boolean deploymentSuccess = !frame.frame().getStackItem(0).isZero();

    if (!deploymentSuccess) {
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
  public void resolvePostRollback(Hub hub, MessageFrame messageFrame, CallFrame callFrame) {
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

  private void finishAbortCreate(final Hub hub) {
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
}
