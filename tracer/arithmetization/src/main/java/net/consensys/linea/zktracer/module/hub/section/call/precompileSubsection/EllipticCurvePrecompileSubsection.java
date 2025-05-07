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
package net.consensys.linea.zktracer.module.hub.section.call.precompileSubsection;

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.zktracer.Trace.WORD_SIZE;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileFlag.*;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileScenario.PRC_FAILURE_KNOWN_TO_HUB;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileScenario.PRC_FAILURE_KNOWN_TO_RAM;

import java.math.BigInteger;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.common.CommonPrecompileOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.common.EcPairingOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.common.ecAddMulRecover.EcAddOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.common.ecAddMulRecover.EcMulOobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.common.ecAddMulRecover.EcRecoverOobCall;
import net.consensys.linea.zktracer.module.hub.section.call.CallSection;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.apache.tuweni.bytes.Bytes;

public class EllipticCurvePrecompileSubsection extends PrecompileSubsection {
  final CommonPrecompileOobCall oobCall;

  public EllipticCurvePrecompileSubsection(Hub hub, CallSection callSection) {
    super(hub, callSection);

    final BigInteger calleeGas =
        BigInteger.valueOf(callSection.stpCall.effectiveChildContextGasAllowance());
    oobCall =
        switch (flag()) {
          case PRC_ECRECOVER -> new EcRecoverOobCall(calleeGas);
          case PRC_ECADD -> new EcAddOobCall(calleeGas);
          case PRC_ECMUL -> new EcMulOobCall(calleeGas);
          case PRC_ECPAIRING -> new EcPairingOobCall(calleeGas);
          default -> throw new IllegalArgumentException(
              String.format(
                  "Precompile address %s not supported by constructor", this.flag().toString()));
        };

    firstImcFragment.callOob(oobCall);

    // Recall that the default scenario is PRC_SUCCESS_WONT_REVERT
    if (!oobCall.isHubSuccess()) {
      this.setScenario(PRC_FAILURE_KNOWN_TO_HUB);
    }
  }

  @Override
  public void resolveAtContextReEntry(Hub hub, CallFrame callFrame) {
    super.resolveAtContextReEntry(hub, callFrame);

    final Bytes returnData = extractReturnData();

    // sanity checks
    switch (flag()) {
      case PRC_ECRECOVER -> {
        checkArgument(oobCall.isHubSuccess() == callSuccess);
        checkArgument(
            callSuccess
                ? (returnData == Bytes.EMPTY || returnData.size() == WORD_SIZE)
                : returnData == Bytes.EMPTY);
      }
      case PRC_ECPAIRING -> checkArgument(
          returnDataRange.extract().size() == (callSuccess ? WORD_SIZE : 0));
      case PRC_ECADD, PRC_ECMUL -> checkArgument(
          returnDataRange.extract().size() == (callSuccess ? 2 * WORD_SIZE : 0));
      default -> throw new IllegalArgumentException("Not an elliptic curve precompile");
    }

    if (!oobCall.isHubSuccess()) {
      return;
    }

    // NOTE: from here on out
    //    hubSuccess ≡ true

    // ECRECOVER can only be FAILURE_KNOWN_TO_HUB or some form of SUCCESS_XXXX_REVERT
    if (flag().isAnyOf(PRC_ECADD, PRC_ECMUL, PRC_ECPAIRING)) {
      if (oobCall.isHubSuccess() && !callSuccess) {
        precompileScenarioFragment.scenario(PRC_FAILURE_KNOWN_TO_RAM);
      }
    }

    final MmuCall firstMmuCall;
    final boolean nonemptyCallData = !getCallDataRange().isEmpty();

    final boolean successBitMmuCall = flag() == PRC_ECRECOVER ? !returnData.isEmpty() : callSuccess;

    if (nonemptyCallData) {
      switch (flag()) {
        case PRC_ECRECOVER -> firstMmuCall =
            MmuCall.callDataExtractionForEcrecover(hub, this, successBitMmuCall);
        case PRC_ECADD -> firstMmuCall =
            MmuCall.callDataExtractionForEcadd(hub, this, successBitMmuCall);
        case PRC_ECMUL -> firstMmuCall =
            MmuCall.callDataExtractionForEcmul(hub, this, successBitMmuCall);
        case PRC_ECPAIRING -> firstMmuCall =
            MmuCall.callDataExtractionForEcpairing(hub, this, successBitMmuCall);
        default -> throw new IllegalArgumentException("Not an elliptic curve precompile");
      }
      firstImcFragment.callMmu(firstMmuCall);

      hub.ecData.callEcData(exoModuleOperationId(), flag(), extractCallData(), returnData);
    }

    if (!callSuccess) return;

    final ImcFragment secondImcFragment = ImcFragment.empty(hub);
    this.fragments().add(secondImcFragment);

    final ImcFragment thirdImcFragment = ImcFragment.empty(hub);
    this.fragments().add(thirdImcFragment);

    MmuCall secondMmuCall = null;
    MmuCall thirdMmuCall = null;
    final boolean producesNonemptyReturnData = !returnData.isEmpty();
    final boolean callerMayReceiveReturnData = !getReturnAtRange().isEmpty();

    if (producesNonemptyReturnData) {
      switch (flag()) {
        case PRC_ECRECOVER -> {
          secondMmuCall = MmuCall.fullReturnDataTransferForEcrecover(hub, this, successBitMmuCall);
          if (callerMayReceiveReturnData) {
            thirdMmuCall = MmuCall.partialReturnDataCopyForEcrecover(hub, this);
          }
        }
        case PRC_ECADD -> {
          if (nonemptyCallData) {
            secondMmuCall = MmuCall.fullTransferOfReturnDataForEcadd(hub, this, successBitMmuCall);
          }
          if (callerMayReceiveReturnData) {
            thirdMmuCall = MmuCall.partialCopyOfReturnDataForEcadd(hub, this);
          }
        }
        case PRC_ECMUL -> {
          if (nonemptyCallData) {
            secondMmuCall = MmuCall.fullReturnDataTransferForEcmul(hub, this, successBitMmuCall);
          }
          if (callerMayReceiveReturnData) {
            thirdMmuCall = MmuCall.partialCopyOfReturnDataForEcmul(hub, this);
          }
        }
        case PRC_ECPAIRING -> {
          secondMmuCall = MmuCall.fullReturnDataTransferForEcpairing(hub, this, successBitMmuCall);
          if (callerMayReceiveReturnData) {
            thirdMmuCall = MmuCall.partialCopyOfReturnDataForEcpairing(hub, this);
          }
        }
        default -> throw new IllegalArgumentException("Not an elliptic curve precompile");
      }
      if (secondMmuCall != null) {
        secondImcFragment.callMmu(secondMmuCall);
      }
      if (thirdMmuCall != null) {
        thirdImcFragment.callMmu(thirdMmuCall);
      }
    }
  }
}
