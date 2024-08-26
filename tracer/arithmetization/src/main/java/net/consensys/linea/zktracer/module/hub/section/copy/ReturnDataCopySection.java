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

package net.consensys.linea.zktracer.module.hub.section.copy;

import static net.consensys.linea.zktracer.module.hub.signals.Exceptions.OUT_OF_GAS_EXCEPTION;

import com.google.common.base.Preconditions;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.MxpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.opcodes.ReturnDataCopyOobCall;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;

public class ReturnDataCopySection extends TraceSection {

  public ReturnDataCopySection(Hub hub) {
    super(hub, maxNumberOfRows(hub));

    final ContextFragment currentContext = ContextFragment.readCurrentContextData(hub);
    final ImcFragment imcFragment = ImcFragment.empty(hub);
    final ReturnDataCopyOobCall oobCall = new ReturnDataCopyOobCall();
    imcFragment.callOob(oobCall);

    this.addStack(hub);
    this.addFragment(currentContext);
    this.addFragment(imcFragment);

    final short exceptions = hub.pch().exceptions();
    final boolean returnDataCopyException = oobCall.isRdcx();
    Preconditions.checkArgument(
        returnDataCopyException == Exceptions.returnDataCopyFault(exceptions));

    // returnDataCopyException case
    if (returnDataCopyException) {
      return;
    }

    final MxpCall mxpCall = new MxpCall(hub);
    imcFragment.callMxp(mxpCall);

    Preconditions.checkArgument(mxpCall.mxpx == Exceptions.memoryExpansionException(exceptions));

    // memoryExpansionException case
    if (mxpCall.mxpx) {
      return;
    }

    // outOfGasException case
    if (Exceptions.any(exceptions)) {
      Preconditions.checkArgument(exceptions == OUT_OF_GAS_EXCEPTION);
      return;
    }

    // beyond this point unexceptional
    final boolean triggerMmu = mxpCall.mayTriggerNontrivialMmuOperation;
    if (triggerMmu) {
      final MmuCall mmuCall = MmuCall.returnDataCopy(hub);
      imcFragment.callMmu(mmuCall);
    }
  }

  private static short maxNumberOfRows(Hub hub) {
    return (short) (hub.opCode().numberOfStackRows() + 3);
  }
}
