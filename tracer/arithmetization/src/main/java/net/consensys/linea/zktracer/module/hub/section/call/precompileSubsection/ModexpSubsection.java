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
import static net.consensys.linea.zktracer.module.blake2fmodexpdata.BlakeModexpDataOperation.MODEXP_COMPONENT_BYTE_SIZE;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractBase;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractBbs;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractEbs;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractExponent;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractMbs;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpExtractModulus;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpFullResultCopy;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpLoadLead;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall.forModexpPartialResultCopy;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp.ModexpXbsCase.OOB_INST_MODEXP_BBS;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp.ModexpXbsCase.OOB_INST_MODEXP_EBS;
import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp.ModexpXbsCase.OOB_INST_MODEXP_MBS;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileScenario.PRC_FAILURE_KNOWN_TO_RAM;

import java.math.BigInteger;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.exp.ExpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.exp.ModexpLogExpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp.ModexpCallDataSizeOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp.ModexpExtractOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp.ModexpLeadOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp.ModexpPricingOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp.ModexpXbsOobCall;
import net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata;
import net.consensys.linea.zktracer.module.hub.section.call.CallSection;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.apache.tuweni.bytes.Bytes;

public class ModexpSubsection extends PrecompileSubsection {

  public final ModexpMetadata modexpMetaData;
  private ModexpPricingOobCall sixthOobCall;
  private ImcFragment seventhImcFragment;
  public boolean transactionWillBePopped = false;

  public ModexpSubsection(final Hub hub, final CallSection callSection) {
    super(hub, callSection);

    modexpMetaData = new ModexpMetadata(getCallDataRange());
    if (modexpMetaData
                .bbs()
                .toUnsignedBigInteger()
                .compareTo(BigInteger.valueOf(MODEXP_COMPONENT_BYTE_SIZE))
            > 0
        || modexpMetaData
                .mbs()
                .toUnsignedBigInteger()
                .compareTo(BigInteger.valueOf(MODEXP_COMPONENT_BYTE_SIZE))
            > 0
        || modexpMetaData
                .ebs()
                .toUnsignedBigInteger()
                .compareTo(BigInteger.valueOf(MODEXP_COMPONENT_BYTE_SIZE))
            > 0) {
      hub.modexpEffectiveCall().updateTally(Integer.MAX_VALUE);
      hub.defers().unscheduleForContextReEntry(this, hub.currentFrame());
      transactionWillBePopped = true;
      return;
    }

    final ModexpCallDataSizeOobCall firstOobCall = new ModexpCallDataSizeOobCall(modexpMetaData);
    firstImcFragment.callOob(firstOobCall);

    final ImcFragment secondImcFragment = ImcFragment.empty(hub);
    fragments().add(secondImcFragment);
    if (modexpMetaData.extractBbs()) {
      final MmuCall mmuCall = forModexpExtractBbs(hub, this, modexpMetaData);
      secondImcFragment.callMmu(mmuCall);
    }

    final ModexpXbsOobCall secondOobCall =
        new ModexpXbsOobCall(modexpMetaData, OOB_INST_MODEXP_BBS);
    secondImcFragment.callOob(secondOobCall);

    final ImcFragment thirdImcFragment = ImcFragment.empty(hub);
    fragments().add(thirdImcFragment);
    if (modexpMetaData.extractEbs()) {
      final MmuCall mmuCall = forModexpExtractEbs(hub, this, modexpMetaData);
      thirdImcFragment.callMmu(mmuCall);
    }
    final ModexpXbsOobCall thirdOobCall = new ModexpXbsOobCall(modexpMetaData, OOB_INST_MODEXP_EBS);
    thirdImcFragment.callOob(thirdOobCall);

    final ImcFragment fourthImcFragment = ImcFragment.empty(hub);
    fragments().add(fourthImcFragment);
    if (modexpMetaData.extractMbs()) {
      final MmuCall mmuCall = forModexpExtractMbs(hub, this, modexpMetaData);
      fourthImcFragment.callMmu(mmuCall);
    }
    final ModexpXbsOobCall fourthOobCall =
        new ModexpXbsOobCall(modexpMetaData, OOB_INST_MODEXP_MBS);
    fourthImcFragment.callOob(fourthOobCall);

    final ImcFragment fifthImcFragment = ImcFragment.empty(hub);
    fragments().add(fifthImcFragment);
    final ModexpLeadOobCall fifthOobCall = new ModexpLeadOobCall(modexpMetaData);
    fifthImcFragment.callOob(fifthOobCall);
    if (modexpMetaData.loadRawLeadingWord()) {
      final MmuCall mmuCall = forModexpLoadLead(hub, this, modexpMetaData);
      fifthImcFragment.callMmu(mmuCall);
      final ExpCall modexpLogCallToExp = new ModexpLogExpCall(modexpMetaData);
      fifthImcFragment.callExp(modexpLogCallToExp);
    }

    final ImcFragment sixthImcFragment = ImcFragment.empty(hub);
    fragments().add(sixthImcFragment);
    final long calleeGas = callSection.stpCall.effectiveChildContextGasAllowance();
    sixthOobCall = new ModexpPricingOobCall(modexpMetaData, calleeGas);
    sixthImcFragment.callOob(sixthOobCall);

    // We need to trigger the OOB before CALL's execution
    if (sixthOobCall.isRamSuccess()) {
      seventhImcFragment = ImcFragment.empty(hub);
      final ModexpExtractOobCall seventhOobCall = new ModexpExtractOobCall(modexpMetaData);
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

    final Bytes returnData = extractReturnData();

    modexpMetaData.rawResult(returnData);
    hub.blakeModexpData().callModexp(modexpMetaData, exoModuleOperationId());

    fragments().add(seventhImcFragment);
    if (modexpMetaData.extractModulus()) {
      final MmuCall mmuCall = forModexpExtractBase(hub, this, modexpMetaData);
      seventhImcFragment.callMmu(mmuCall);
    }

    final ImcFragment eighthImcFragment = ImcFragment.empty(hub);
    fragments().add(eighthImcFragment);
    if (modexpMetaData.extractModulus()) {
      final MmuCall mmuCall = forModexpExtractExponent(hub, this, modexpMetaData);
      eighthImcFragment.callMmu(mmuCall);
    }

    final ImcFragment ninthImcFragment = ImcFragment.empty(hub);
    fragments().add(ninthImcFragment);
    if (modexpMetaData.extractModulus()) {
      final MmuCall mmuCall = forModexpExtractModulus(hub, this, modexpMetaData);
      ninthImcFragment.callMmu(mmuCall);
    }

    final ImcFragment tenthImcFragment = ImcFragment.empty(hub);
    fragments().add(tenthImcFragment);
    if (modexpMetaData.extractModulus()) {
      final MmuCall mmuCall = forModexpFullResultCopy(hub, this, modexpMetaData);
      tenthImcFragment.callMmu(mmuCall);
    }

    final ImcFragment eleventhImcFragment = ImcFragment.empty(hub);
    fragments().add(eleventhImcFragment);
    if (modexpMetaData.mbsNonZero() && !getReturnAtRange().isEmpty()) {
      final MmuCall mmuCall = forModexpPartialResultCopy(hub, this, modexpMetaData);
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
