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

package net.consensys.linea.zktracer.module.hub.section.call.precompileSubsection;

import static com.google.common.base.Preconditions.*;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.OOB_INST_RIPEMD;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.OOB_INST_SHA2;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileFlag.PRC_SHA2_256;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileScenario.PRC_FAILURE_KNOWN_TO_HUB;
import static net.consensys.linea.zktracer.module.shakiradata.HashFunction.RIPEMD;
import static net.consensys.linea.zktracer.module.shakiradata.HashFunction.SHA256;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.PrecompileCommonOobCall;
import net.consensys.linea.zktracer.module.hub.section.call.CallSection;
import net.consensys.linea.zktracer.module.shakiradata.ShakiraDataOperation;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;

public class ShaTwoOrRipemdSubSection extends PrecompileSubsection {
  final PrecompileCommonOobCall oobCall;

  public ShaTwoOrRipemdSubSection(Hub hub, CallSection callSection) {
    super(hub, callSection);

    oobCall =
        switch (flag()) {
          case PRC_SHA2_256 -> new PrecompileCommonOobCall(OOB_INST_SHA2);
          case PRC_RIPEMD_160 -> new PrecompileCommonOobCall(OOB_INST_RIPEMD);
          default -> throw new IllegalArgumentException(
              String.format(
                  "Precompile address %s not supported by constructor", this.flag().toString()));
        };
    firstImcFragment.callOob(oobCall);

    if (!oobCall.isHubSuccess()) {
      this.setScenario(PRC_FAILURE_KNOWN_TO_HUB);
    }
  }

  @Override
  public void resolveAtContextReEntry(Hub hub, CallFrame callFrame) {
    super.resolveAtContextReEntry(hub, callFrame);

    // sanity check
    checkArgument(callSuccess == oobCall.isHubSuccess());

    if (!callSuccess) {
      return;
    }

    // NOTE: we trigger the SHAKIRA module for nonempty call data only
    if (!callData.isEmpty()) {
      final ShakiraDataOperation shakiraCall =
          new ShakiraDataOperation(
              callSection.hubStamp(),
              this.flag() == PRC_SHA2_256 ? SHA256 : RIPEMD,
              callData(),
              callFrame.frame().getReturnData());
      hub.shakiraData().call(shakiraCall);

      final MmuCall mmuCall = MmuCall.callDataExtractionForShaTwoAndRipemd(hub, this);
      firstImcFragment.callMmu(mmuCall);
    }

    // the full result transfer happens in all cases
    final ImcFragment secondImcFragment = ImcFragment.empty(hub);
    this.fragments().add(secondImcFragment);

    final MmuCall fullOutputDataTransfer = MmuCall.fullResultTransferForShaTwoAndRipemd(hub, this);
    secondImcFragment.callMmu(fullOutputDataTransfer);

    final ImcFragment thirdImcFragment = ImcFragment.empty(hub);
    this.fragments().add(thirdImcFragment);

    // the partial copy of return data happens only if the caller context
    // provided a nonempty return data target
    if (!parentReturnDataTarget.isEmpty()) {
      final MmuCall partialReturnDataCopy =
          MmuCall.partialReturnDataCopyForShaTwoAndRipemd(hub, this);
      thirdImcFragment.callMmu(partialReturnDataCopy);
    }
  }

  // 4 = 1 + 3 (scenario row + up to 3 miscellaneous fragments)
  @Override
  protected short maxNumberOfLines() {
    return 4;
    // Note: we don't have the callSuccess available at the moment
    // and can't provide the "real" value (2 in case of FKTH.)
  }
}
