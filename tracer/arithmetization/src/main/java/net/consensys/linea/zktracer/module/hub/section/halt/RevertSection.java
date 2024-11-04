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

import static com.google.common.base.Preconditions.*;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.MxpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;

public class RevertSection extends TraceSection {

  final ImcFragment imcFragment;
  MmuCall mmuCall;

  public RevertSection(Hub hub) {
    // up to 4 = 1 + 3 rows
    super(hub, (short) 4);

    short exceptions = hub.pch().exceptions();

    imcFragment = ImcFragment.empty(hub);
    this.addStackAndFragments(hub, imcFragment);

    // triggerExp = false
    // triggerOob = false
    // triggerStp = false
    // triggerMxp = true
    MxpCall mxpCall = new MxpCall(hub);
    imcFragment.callMxp(mxpCall);
    checkArgument(mxpCall.mxpx == Exceptions.memoryExpansionException(exceptions));

    if (Exceptions.memoryExpansionException(exceptions)) {
      return;
    }

    if (Exceptions.outOfGasException(exceptions)) {
      return;
    }

    // The XAHOY = 0 case
    /////////////////////
    checkArgument(Exceptions.none(exceptions));

    final boolean triggerMmu =
        (Exceptions.none(exceptions))
            && !hub.currentFrame().isRoot()
            && mxpCall.mayTriggerNontrivialMmuOperation // i.e. size ≠ 0 ∧ ¬MXPX
            && !hub.currentFrame().returnDataTargetInCaller().isEmpty();

    if (triggerMmu) {
      mmuCall = MmuCall.revert(hub);
      imcFragment.callMmu(mmuCall);
    }

    final CallFrame callFrame = hub.currentFrame();
    final ContextFragment currentContext = ContextFragment.readCurrentContextData(hub);
    final ContextFragment updateCallerReturnData =
        ContextFragment.executionProvidesReturnData(
            hub,
            hub.callStack().getById(callFrame.callerId()).contextNumber(),
            callFrame.contextNumber());

    this.addFragments(currentContext, updateCallerReturnData);
  }
}
