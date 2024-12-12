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

import static net.consensys.linea.zktracer.opcode.OpCode.*;

import com.google.common.base.Preconditions;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostRollbackDefer;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

public class AccountSection extends TraceSection implements PostRollbackDefer {

  Bytes rawTargetAddress;
  Address targetAddress;
  AccountSnapshot firstAccountSnapshot;
  AccountSnapshot firstAccountSnapshotNew;
  AccountSnapshot secondAccountSnapshot;
  AccountSnapshot secondAccountSnapshotNew;
  int hubStamp;

  public AccountSection(Hub hub) {
    super(hub, maxNumberOfRows(hub));
    hubStamp = hub.stamp();
    this.addStack(hub);

    final short exceptions = hub.pch().exceptions();

    if (hub.opCode().isAnyOf(OpCode.SELFBALANCE, OpCode.CODESIZE)) {
      if (Exceptions.any(exceptions)) {
        // the "squash parent return data" context row is all there is
        // The following is true since we do not enter here in case of a STACK_OVERFLOW_EXCEPTION
        Preconditions.checkArgument(Exceptions.outOfGasException(exceptions));
        return;
      }

      final ContextFragment currentContext = ContextFragment.readCurrentContextData(hub);
      this.addFragment(currentContext);
    }

    final MessageFrame frame = hub.messageFrame();

    targetAddress =
        switch (hub.opCode()) {
          case BALANCE, EXTCODESIZE, EXTCODEHASH -> {
            hub.defers().scheduleForPostRollback(this, hub.currentFrame());
            rawTargetAddress = frame.getStackItem(0);
            yield Words.toAddress(this.rawTargetAddress);
          }
          case SELFBALANCE -> frame.getRecipientAddress();
          case CODESIZE -> frame.getContractAddress();
          default -> throw new RuntimeException("Not an ACCOUNT instruction");
        };

    firstAccountSnapshot = AccountSnapshot.canonical(hub, targetAddress);
    firstAccountSnapshotNew = firstAccountSnapshot.deepCopy();

    if (Exceptions.none(exceptions)) {
      firstAccountSnapshotNew.turnOnWarmth();
    }

    final DomSubStampsSubFragment doingDomSubStamps =
        DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0);

    final AccountFragment doingAccountFragment =
        switch (hub.opCode()) {
          case BALANCE, EXTCODESIZE, EXTCODEHASH -> hub.factories()
              .accountFragment()
              .makeWithTrm(
                  firstAccountSnapshot,
                  firstAccountSnapshotNew,
                  rawTargetAddress,
                  doingDomSubStamps);
          case SELFBALANCE, CODESIZE -> hub.factories()
              .accountFragment()
              .make(firstAccountSnapshot, firstAccountSnapshotNew, doingDomSubStamps);
          default -> throw new IllegalStateException("Not an ACCOUNT instruction");
        };
    this.addFragment(doingAccountFragment);
  }

  public void resolveUponRollback(Hub hub, MessageFrame messageFrame, CallFrame callFrame) {

    secondAccountSnapshot = firstAccountSnapshotNew.deepCopy().setDeploymentNumber(hub);
    secondAccountSnapshotNew = firstAccountSnapshot.deepCopy().setDeploymentNumber(hub);
    final DomSubStampsSubFragment undoingDomSubStamps =
        DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
            hubStamp, hub.currentFrame().revertStamp(), 1);

    this.addFragment(
        hub.factories()
            .accountFragment()
            .make(secondAccountSnapshot, secondAccountSnapshotNew, undoingDomSubStamps));
  }

  private static short maxNumberOfRows(Hub hub) {
    final OpCode opCode = hub.opCode();

    if (opCode.isAnyOf(BALANCE, EXTCODESIZE, EXTCODEHASH)) {
      return (short) (opCode.numberOfStackRows() + 3);
    }

    return (short) (opCode.numberOfStackRows() + (Exceptions.any(hub.pch().exceptions()) ? 1 : 2));
  }
}
