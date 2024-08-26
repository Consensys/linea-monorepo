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

import com.google.common.base.Preconditions;
import lombok.Getter;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.State;
import net.consensys.linea.zktracer.module.hub.defer.PostRollbackDefer;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.opcodes.SstoreOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.storage.StorageFragment;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

@Getter
public class SstoreSection extends TraceSection implements PostRollbackDefer {

  final WorldView world;
  final Address address;
  final int deploymentNumber;
  final Bytes32 storageKey;
  final boolean incomingWarmth;
  final EWord valueOriginal;
  final EWord valueCurrent;
  final EWord valueNext;

  public SstoreSection(Hub hub, WorldView world) {
    super(
        hub,
        (short)
            (hub.opCode().numberOfStackRows() + (Exceptions.any(hub.pch().exceptions()) ? 5 : 4)));

    this.world = world;
    this.address = hub.messageFrame().getRecipientAddress();
    this.deploymentNumber = hub.currentFrame().accountDeploymentNumber();
    this.storageKey = Bytes32.leftPad(hub.messageFrame().getStackItem(0));
    this.incomingWarmth = hub.messageFrame().getWarmedUpStorage().contains(address, storageKey);
    this.valueOriginal =
        EWord.of(world.get(address).getOriginalStorageValue(UInt256.fromBytes(storageKey)));
    this.valueCurrent = EWord.of(world.get(address).getStorageValue(UInt256.fromBytes(storageKey)));
    this.valueNext = EWord.of(hub.messageFrame().getStackItem(1));
    final short exceptions = hub.pch().exceptions();

    final boolean staticContextException = Exceptions.staticFault(exceptions);
    final boolean sstoreException = Exceptions.outOfSStore(exceptions);

    // CONTEXT fragment
    final ContextFragment readCurrentContext = ContextFragment.readCurrentContextData(hub);
    this.addStackAndFragments(hub, readCurrentContext);

    if (staticContextException) {
      return;
    }

    // MISC fragment
    final ImcFragment miscForSstore = ImcFragment.empty(hub);
    this.addFragment(miscForSstore);

    miscForSstore.callOob(new SstoreOobCall());

    if (sstoreException) {
      return;
    }

    Preconditions.checkArgument(
        Exceptions.outOfGasException(exceptions) || Exceptions.none(exceptions));

    hub.defers().scheduleForPostRollback(this, hub.currentFrame());

    // STORAGE fragment (for doing)
    final StorageFragment doingSstore = this.doingSstore(hub);
    this.addFragment(doingSstore);

    // Update the First Last time seen map of storage keys
    final State.StorageSlotIdentifier storageSlotIdentifier =
        new State.StorageSlotIdentifier(
            address,
            hub.transients().conflation().deploymentInfo().number(address),
            EWord.of(storageKey));
    hub.state.updateOrInsertStorageSlotOccurrence(storageSlotIdentifier, doingSstore);

    // set the refundDelta
    this.commonValues.refundDelta(
        Hub.GAS_PROJECTOR
            .of(hub.currentFrame().frame(), hub.opCode())
            .refund()); // TODO should use Besu's value
  }

  private StorageFragment doingSstore(Hub hub) {

    return new StorageFragment(
        hub.state,
        new State.StorageSlotIdentifier(
            this.address, this.deploymentNumber, EWord.of(this.storageKey)),
        this.valueOriginal,
        this.valueCurrent,
        this.valueNext,
        this.incomingWarmth,
        true,
        DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0),
        hub.state.firstAndLastStorageSlotOccurrences.size());
  }

  @Override
  public void resolvePostRollback(Hub hub, MessageFrame messageFrame, CallFrame callFrame) {
    final DomSubStampsSubFragment undoingDomSubStamps =
        DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
            this.hubStamp(), hub.callStack().current().revertStamp(), 0);

    final StorageFragment undoingSstoreStorageFragment =
        new StorageFragment(
            hub.state,
            new State.StorageSlotIdentifier(
                this.address, this.deploymentNumber, EWord.of(this.storageKey)),
            this.valueOriginal,
            this.valueNext,
            this.valueCurrent,
            true,
            this.incomingWarmth,
            undoingDomSubStamps,
            hub.state.firstAndLastStorageSlotOccurrences.size());

    this.addFragment(undoingSstoreStorageFragment);

    // undo the refund
    this.commonValues.refundDelta(0);
  }
}
