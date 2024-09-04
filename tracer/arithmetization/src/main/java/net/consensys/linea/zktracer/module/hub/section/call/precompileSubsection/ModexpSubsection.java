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
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractBase;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractBbs;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractEbs;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractExponent;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractMbs;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractModulus;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpFullResultCopy;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpLoadLead;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpPartialResultCopy;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.ModexpXbsCase.OOB_INST_MODEXP_BBS;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.ModexpXbsCase.OOB_INST_MODEXP_EBS;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.ModexpXbsCase.OOB_INST_MODEXP_MBS;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileScenario.PRC_FAILURE_KNOWN_TO_RAM;
import static net.consensys.linea.zktracer.module.limits.precompiles.ModexpEffectiveCall.PROVER_MAX_INPUT_BYTE_SIZE;

import java.math.BigInteger;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.exp.ExpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.exp.ModexpLogExpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.ModexpCallDataSizeOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.ModexpExtractOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.ModexpLeadOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.ModexpPricingOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.ModexpXbsOobCall;
import net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata;
import net.consensys.linea.zktracer.module.hub.section.call.CallSection;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;

public class ModexpSubsection extends PrecompileSubsection {

  private final ModexpMetadata modExpMetadata;
  private ModexpPricingOobCall sixthOobCall;
  private ImcFragment seventhImcFragment;

  public ModexpSubsection(final Hub hub, final CallSection callSection) {
    super(hub, callSection);

    modExpMetadata = new ModexpMetadata(hub, callData);
    if (modExpMetadata
                .bbs()
                .toUnsignedBigInteger()
                .compareTo(BigInteger.valueOf(PROVER_MAX_INPUT_BYTE_SIZE))
            >= 0
        || modExpMetadata
                .mbs()
                .toUnsignedBigInteger()
                .compareTo(BigInteger.valueOf(PROVER_MAX_INPUT_BYTE_SIZE))
            >= 0
        || modExpMetadata
                .ebs()
                .toUnsignedBigInteger()
                .compareTo(BigInteger.valueOf(PROVER_MAX_INPUT_BYTE_SIZE))
            >= 0) {
      hub.modexpEffectiveCall().addPrecompileLimit(Integer.MAX_VALUE);
      return;
    }

    final ModexpCallDataSizeOobCall firstOobCall = new ModexpCallDataSizeOobCall();
    firstImcFragment.callOob(firstOobCall);

    final ImcFragment secondImcFragment = ImcFragment.empty(hub);
    fragments().add(secondImcFragment);
    if (modExpMetadata.extractBbs()) {
      final MmuCall mmuCall = forModexpExtractBbs(hub, this, modExpMetadata);
      secondImcFragment.callMmu(mmuCall);
    }
    final ModexpXbsOobCall secondOobCall = new ModexpXbsOobCall(OOB_INST_MODEXP_BBS);
    secondImcFragment.callOob(secondOobCall);

    final ImcFragment thirdImcFragment = ImcFragment.empty(hub);
    fragments().add(thirdImcFragment);
    if (modExpMetadata.extractEbs()) {
      final MmuCall mmuCall = forModexpExtractEbs(hub, this, modExpMetadata);
      thirdImcFragment.callMmu(mmuCall);
    }
    final ModexpXbsOobCall thirdOobCall = new ModexpXbsOobCall(OOB_INST_MODEXP_EBS);
    thirdImcFragment.callOob(thirdOobCall);

    final ImcFragment fourthImcFragment = ImcFragment.empty(hub);
    fragments().add(fourthImcFragment);
    if (modExpMetadata.extractMbs()) {
      final MmuCall mmuCall = forModexpExtractMbs(hub, this, modExpMetadata);
      fourthImcFragment.callMmu(mmuCall);
    }
    final ModexpXbsOobCall fourthOobCall = new ModexpXbsOobCall(OOB_INST_MODEXP_MBS);
    fourthImcFragment.callOob(fourthOobCall);

    final ImcFragment fifthImcFragment = ImcFragment.empty(hub);
    fragments().add(fifthImcFragment);
    final ModexpLeadOobCall fifthOobCall = new ModexpLeadOobCall();
    fifthImcFragment.callOob(fifthOobCall);
    if (modExpMetadata.loadRawLeadingWord()) {
      final MmuCall mmuCall = forModexpLoadLead(hub, this, modExpMetadata);
      fifthImcFragment.callMmu(mmuCall);
      final ExpCall modexpLogCallToExp = new ModexpLogExpCall(modExpMetadata);
      fifthImcFragment.callExp(modexpLogCallToExp);
    }

    final ImcFragment sixthImcFragment = ImcFragment.empty(hub);
    fragments().add(sixthImcFragment);
    sixthOobCall = new ModexpPricingOobCall();
    sixthImcFragment.callOob(sixthOobCall);

    // We need to trigger the OOB before CALL's execution
    if (sixthOobCall.isRamSuccess()) {
      seventhImcFragment = ImcFragment.empty(hub);
      final ModexpExtractOobCall seventhOobCall = new ModexpExtractOobCall();
      seventhImcFragment.callOob(seventhOobCall);
    }
  }

  @Override
  public void resolveAtContextReEntry(Hub hub, CallFrame callFrame) {
    super.resolveAtContextReEntry(hub, callFrame);

    // sanity check
    checkArgument(callSuccess == sixthOobCall.isRamSuccess());

    if (!callSuccess) {
      precompileScenarioFragment.scenario(PRC_FAILURE_KNOWN_TO_RAM);
      return;
    }

    modExpMetadata.rawResult(returnData);
    hub.blakeModexpData().callModexp(modExpMetadata, this.exoModuleOperationId());

    fragments().add(seventhImcFragment);
    if (modExpMetadata.extractModulus()) {
      final MmuCall mmuCall = forModexpExtractBase(hub, this, modExpMetadata);
      seventhImcFragment.callMmu(mmuCall);
    }

    final ImcFragment eighthImcFragment = ImcFragment.empty(hub);
    fragments().add(eighthImcFragment);
    if (modExpMetadata.extractModulus()) {
      final MmuCall mmuCall = forModexpExtractExponent(hub, this, modExpMetadata);
      eighthImcFragment.callMmu(mmuCall);
    }

    final ImcFragment ninthImcFragment = ImcFragment.empty(hub);
    fragments().add(ninthImcFragment);
    if (modExpMetadata.extractModulus()) {
      final MmuCall mmuCall = forModexpExtractModulus(hub, this, modExpMetadata);
      ninthImcFragment.callMmu(mmuCall);
    }

    final ImcFragment tenthImcFragment = ImcFragment.empty(hub);
    fragments().add(tenthImcFragment);
    if (modExpMetadata.extractModulus()) {
      final MmuCall mmuCall = forModexpFullResultCopy(hub, this, modExpMetadata);
      tenthImcFragment.callMmu(mmuCall);
    }

    final ImcFragment eleventhImcFragment = ImcFragment.empty(hub);
    fragments().add(eleventhImcFragment);
    if (modExpMetadata.mbsNonZero() && !parentReturnDataTarget.isEmpty()) {
      final MmuCall mmuCall = forModexpPartialResultCopy(hub, this, modExpMetadata);
      eleventhImcFragment.callMmu(mmuCall);
    }
  }

  // 13 = 1 + 12 (scenario row + up to 12 miscellaneous fragments)
  @Override
  protected short maxNumberOfLines() {
    return 13;
    // Note: we don't have the successBit available at the moment
    // and can't provide the "real" value (8 in case of failure.)
  }
}
