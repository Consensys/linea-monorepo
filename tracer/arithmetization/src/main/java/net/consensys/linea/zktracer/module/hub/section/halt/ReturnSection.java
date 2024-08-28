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
package net.consensys.linea.zktracer.module.hub.section.halt;

import static net.consensys.linea.zktracer.module.hub.fragment.scenario.ReturnScenarioFragment.ReturnScenario.*;
import static net.consensys.linea.zktracer.module.hub.signals.Exceptions.OUT_OF_GAS_EXCEPTION;
import static net.consensys.linea.zktracer.module.hub.signals.Exceptions.memoryExpansionException;
import static net.consensys.linea.zktracer.types.Conversions.bytesToBoolean;

import com.google.common.base.Preconditions;
import lombok.Getter;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.ContextReEntryDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostRollbackDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.MxpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.opcodes.DeploymentOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.opcodes.XCallOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.ReturnScenarioFragment;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

@Getter
public class ReturnSection extends TraceSection
    implements ContextReEntryDefer, PostRollbackDefer, PostTransactionDefer {

  final boolean returnFromMessageCall;
  final boolean returnFromDeployment;
  boolean nonemptyByteCode;
  final ReturnScenarioFragment returnScenarioFragment;
  AccountFragment deploymentFragment;

  AccountSnapshot preDeploymentAccountSnapshot;
  AccountSnapshot postDeploymentAccountSnapshot;
  AccountSnapshot undoingDeploymentAccountSnapshot;
  ContextFragment squashParentContextReturnData;
  Address deploymentAddress;

  boolean successfulMessageCallExpected; // for sanity check
  boolean successfulDeploymentExpected; // for sanity check

  // TODO: trigger SHAKIRA

  public ReturnSection(Hub hub) {
    super(hub, maxNumberOfRows(hub));

    final CallFrame currentFrame = hub.currentFrame();
    final MessageFrame frame = hub.messageFrame();

    returnFromMessageCall = currentFrame.isMessageCall();
    returnFromDeployment = currentFrame.isDeployment();

    Preconditions.checkArgument(
        returnFromDeployment
            == hub.transients()
                .conflation()
                .deploymentInfo()
                .isDeploying(frame.getContractAddress()));

    returnScenarioFragment = new ReturnScenarioFragment();
    final ContextFragment currentContextFragment = ContextFragment.readCurrentContextData(hub);
    final ImcFragment firstImcFragment = ImcFragment.empty(hub);
    final MxpCall mxpCall = new MxpCall(hub);
    firstImcFragment.callMxp(mxpCall);

    this.addStack(hub);
    this.addFragment(returnScenarioFragment);
    this.addFragment(currentContextFragment);
    this.addFragment(firstImcFragment);

    final short exceptions = hub.pch().exceptions();

    if (Exceptions.any(exceptions)) {
      returnScenarioFragment.setScenario(RETURN_EXCEPTION);
    }

    Preconditions.checkArgument(mxpCall.mxpx == memoryExpansionException(exceptions));

    if (mxpCall.mxpx) {
      return;
    }

    if (Exceptions.outOfGasException(exceptions) && returnFromMessageCall) {
      Preconditions.checkArgument(exceptions == OUT_OF_GAS_EXCEPTION);
      return;
    }

    if (Exceptions.any(exceptions)) {
      // exceptional message calls are dealt with;
      // if exceptions remain they must be related
      // to deployments:
      Preconditions.checkArgument(returnFromDeployment);
    }

    // maxCodeSizeException case
    boolean triggerOobForMaxCodeSizeException = Exceptions.codeSizeOverflow(exceptions);
    if (triggerOobForMaxCodeSizeException) {
      OobCall oobCall = new XCallOobCall();
      firstImcFragment.callOob(oobCall);
      return;
    }

    // invalidCodePrefixException case
    final boolean nontrivialMmuOperation = mxpCall.mayTriggerNontrivialMmuOperation;
    final boolean triggerMmuForInvalidCodePrefix = Exceptions.invalidCodePrefix(exceptions);
    if (triggerMmuForInvalidCodePrefix) {
      Preconditions.checkArgument(returnFromDeployment && nontrivialMmuOperation);

      final MmuCall actuallyInvalidCodePrefixMmuCall = MmuCall.invalidCodePrefix(hub);
      firstImcFragment.callMmu(actuallyInvalidCodePrefixMmuCall);

      Preconditions.checkArgument(!actuallyInvalidCodePrefixMmuCall.successBit());
      return;
    }

    // Unexceptional RETURN's
    // (we have exceptions ≡ ∅ by the checkArgument)
    ////////////////////////////////////////////////

    Preconditions.checkArgument(Exceptions.none(exceptions));

    // RETURN_FROM_MESSAGE_CALL cases
    if (returnFromMessageCall) {
      successfulMessageCallExpected = true;
      final boolean messageCallReturnTouchesRam =
          !currentFrame.isRoot()
              && nontrivialMmuOperation // [size ≠ 0] ∧ ¬MXPX
              && !currentFrame.returnDataTargetInCaller().isEmpty(); // [r@c ≠ 0]

      returnScenarioFragment.setScenario(
          messageCallReturnTouchesRam
              ? RETURN_FROM_MESSAGE_CALL_WILL_TOUCH_RAM
              : RETURN_FROM_MESSAGE_CALL_WONT_TOUCH_RAM);

      if (messageCallReturnTouchesRam) {
        final MmuCall returnFromMessageCall = MmuCall.returnFromMessageCall(hub);
        firstImcFragment.callMmu(returnFromMessageCall);
      }

      final ContextFragment updateCallerReturnData =
          ContextFragment.executionProvidesReturnData(
              hub,
              hub.callStack().getById(currentFrame.parentFrameId()).contextNumber(),
              currentFrame.contextNumber());
      this.addFragment(updateCallerReturnData);

      return;
    }

    // RETURN_FROM_DEPLOYMENT cases
    if (returnFromDeployment) {
      successfulDeploymentExpected = true;

      // TODO: @Olivier and @François: what happens when "re-entering" the root's parent context ?
      //  we may need to improve the triggering of the resolution to also kick in at transaction
      //  end for stuff that happens after the root returns ...
      hub.defers()
          .scheduleForContextReEntry(
              this, hub.callStack().parent()); // post deployment account snapshot
      hub.defers().scheduleForPostRollback(this, currentFrame); // undo deployment
      hub.defers().scheduleForPostTransaction(this); // inserting the final context row;

      squashParentContextReturnData = ContextFragment.executionProvidesEmptyReturnData(hub);
      deploymentAddress = frame.getRecipientAddress();
      nonemptyByteCode = mxpCall.mayTriggerNontrivialMmuOperation;
      preDeploymentAccountSnapshot = AccountSnapshot.canonical(hub, deploymentAddress);
      returnScenarioFragment.setScenario(
          nonemptyByteCode
              ? RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WONT_REVERT
              : RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WONT_REVERT);

      final Bytes byteCodeSize = frame.getStackItem(1);
      Preconditions.checkArgument(nonemptyByteCode == (!byteCodeSize.isZero()));

      // Empty deployments
      if (!nonemptyByteCode) {
        return;
      }

      hub.romLex().callRomLex(frame);

      final MmuCall invalidCodePrefixCheckMmuCall = MmuCall.invalidCodePrefix(hub);
      firstImcFragment.callMmu(invalidCodePrefixCheckMmuCall);

      final DeploymentOobCall maxCodeSizeOobCall = new DeploymentOobCall();
      firstImcFragment.callOob(maxCodeSizeOobCall);

      // sanity checks
      Preconditions.checkArgument(invalidCodePrefixCheckMmuCall.successBit());
      Preconditions.checkArgument(!maxCodeSizeOobCall.isMaxCodeSizeException());

      final ImcFragment secondImcFragment = ImcFragment.empty(hub);
      this.addFragment(secondImcFragment);

      final MmuCall nonemptyDeploymentMmuCall = MmuCall.returnFromDeployment(hub);
      secondImcFragment.callMmu(nonemptyDeploymentMmuCall);
    }
  }

  @Override
  public void resolveAtContextReEntry(Hub hub, CallFrame frame) {

    // TODO: optional sanity check that may be removed
    if (returnFromMessageCall) {
      Bytes topOfTheStack = hub.messageFrame().getStackItem(0);
      boolean messageCallWasSuccessful = bytesToBoolean(topOfTheStack);
      Preconditions.checkArgument(messageCallWasSuccessful == successfulMessageCallExpected);
    }

    // TODO: optional sanity check that may be removed
    if (returnFromDeployment) {
      Bytes topOfTheStack = hub.messageFrame().getStackItem(0);
      boolean deploymentWasSuccess = !topOfTheStack.isZero();
      Preconditions.checkArgument(deploymentWasSuccess == successfulDeploymentExpected);
    }

    postDeploymentAccountSnapshot = AccountSnapshot.canonical(hub, deploymentAddress);
    final AccountFragment deploymentAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                preDeploymentAccountSnapshot,
                postDeploymentAccountSnapshot,
                DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));

    if (nonemptyByteCode) {
      // TODO: we require the
      //  - triggerHashInfo stuff on the first stack row (automatic AFAICT)
      //  - triggerROMLEX on the deploymentAccountFragment row (see below)
      deploymentAccountFragment.requiresRomlex(true);
    }

    this.addFragment(deploymentAccountFragment);
  }

  @Override
  public void resolvePostRollback(Hub hub, MessageFrame messageFrame, CallFrame callFrame) {

    Preconditions.checkArgument(returnFromDeployment);
    returnScenarioFragment.setScenario(
        nonemptyByteCode
            ? RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WILL_REVERT
            : RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WILL_REVERT);

    undoingDeploymentAccountSnapshot = AccountSnapshot.canonical(hub, deploymentAddress);

    // TODO: does this account for updates to
    //  - deploymentNumber and status ?
    //  - MARKED_FOR_SELF_DESTRUCT(_NEW) ?
    final AccountFragment undoingDeploymentAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                postDeploymentAccountSnapshot,
                undoingDeploymentAccountSnapshot,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    this.hubStamp(), hub.callStack().current().revertStamp(), 1));

    this.addFragment(undoingDeploymentAccountFragment);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {

    Preconditions.checkArgument(returnFromDeployment);
    this.addFragment(squashParentContextReturnData);
  }

  private static short maxNumberOfRows(Hub hub) {
    return (short)
        (hub.opCode().numberOfStackRows() + (Exceptions.any(hub.pch().exceptions()) ? 4 : 7));
  }
}
