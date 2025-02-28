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

package net.consensys.linea.zktracer.module.ecdata;

import java.util.List;
import java.util.Set;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.module.OperationListModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedList;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment;
import net.consensys.linea.zktracer.module.limits.precompiles.EcAddEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcMulEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcPairingFinalExponentiations;
import net.consensys.linea.zktracer.module.limits.precompiles.EcPairingG2MembershipCalls;
import net.consensys.linea.zktracer.module.limits.precompiles.EcPairingMillerLoops;
import net.consensys.linea.zktracer.module.limits.precompiles.EcRecoverEffectiveCall;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;

@RequiredArgsConstructor
@Getter
@Accessors(fluent = true)
public class EcData implements OperationListModule<EcDataOperation> {
  public static final Set<Address> EC_PRECOMPILES =
      Set.of(Address.ECREC, Address.ALTBN128_ADD, Address.ALTBN128_MUL, Address.ALTBN128_PAIRING);

  private final ModuleOperationStackedList<EcDataOperation> operations =
      new ModuleOperationStackedList<>();

  private final Wcp wcp;
  private final Ext ext;

  private final EcAddEffectiveCall ecAddEffectiveCall;
  private final EcMulEffectiveCall ecMulEffectiveCall;
  private final EcRecoverEffectiveCall ecRecoverEffectiveCall;

  private final EcPairingG2MembershipCalls ecPairingG2MembershipCalls;
  private final EcPairingMillerLoops ecPairingMillerLoops;
  private final EcPairingFinalExponentiations ecPairingFinalExponentiations;

  @Getter private EcDataOperation ecDataOperation;

  @Override
  public String moduleKey() {
    return "EC_DATA";
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders() {
    return Trace.Ecdata.headers(this.lineCount());
  }

  @Override
  public void commit(Trace trace) {
    int stamp = 0;
    long previousId = 0;
    for (EcDataOperation op : operations.getAll()) {
      op.trace(trace.ecdata, ++stamp, previousId);
      previousId = op.id();
    }
  }

  public void callEcData(
      final int id,
      final PrecompileScenarioFragment.PrecompileFlag precompileFlag,
      final Bytes callData,
      final Bytes returnData) {
    ecDataOperation =
        EcDataOperation.of(this.wcp, this.ext, id, precompileFlag, callData, returnData);
    operations.add(ecDataOperation);

    switch (ecDataOperation.precompileFlag()) {
      case PRC_ECADD -> ecAddEffectiveCall.addPrecompileLimit(
          ecDataOperation.internalChecksPassed() ? 1 : 0);
      case PRC_ECMUL -> ecMulEffectiveCall.addPrecompileLimit(
          ecDataOperation.internalChecksPassed() ? 1 : 0);
      case PRC_ECRECOVER -> ecRecoverEffectiveCall.addPrecompileLimit(
          ecDataOperation.internalChecksPassed() ? 1 : 0);
      case PRC_ECPAIRING -> {
        // ecPairingG2MembershipCalls case
        // NOTE: the other precompile limits are managed below
        // NOTE: see EC_DATA specs Figure 3.5 for a graphical representation of this case analysis
        if (!ecDataOperation.internalChecksPassed()) {
          ecPairingG2MembershipCalls.addPrecompileLimit(0);
          // The circuit is never invoked in the case of internal checks failing
        }
        // NOTE: the && of the conditions may seem not necessary since in the specs
        // !internalChecksPassed => !notOnG2AccMax
        // however, in EcDataOperation implementation the notOnG2AccMax takes into consideration
        // only large points G2 membership
        // , and it has to be && with internalChecksPassed to compute the actual
        // NOT_ON_G2_ACC_MAX to trace
        if (ecDataOperation.internalChecksPassed() && ecDataOperation.notOnG2AccMax()) {
          ecPairingG2MembershipCalls.addPrecompileLimit(1);
          // The circuit is invoked only once if there is at least one point predicted to be not on
          // G2
        }
        if (ecDataOperation.internalChecksPassed()
            && !ecDataOperation.notOnG2AccMax()
            && ecDataOperation.overallTrivialPairing().getLast()) {
          ecPairingG2MembershipCalls.addPrecompileLimit(0);
          // The circuit is never invoked in the case of a trivial pairing
        }
        if (ecDataOperation.internalChecksPassed()
            && !ecDataOperation.notOnG2AccMax()
            && !ecDataOperation.overallTrivialPairing().getLast()) {
          ecPairingG2MembershipCalls.addPrecompileLimit(
              ecDataOperation.circuitSelectorG2MembershipCounter());
          // The circuit is invoked as many times as there are points predicted to be on G2
        }

        // NOTE: a similar case analysis to the one above may be done for the other
        // precompile limits. However, circuitSelectorEcPairingCounter already takes
        // it into consideration and what follows is enough

        // ecPairingMillerLoops case
        // NOTE: the pairings that require Miller Loops are the valid ones where
        // the small point is on C_1, the large point is on G_2, and they are not
        // points at infinity (valid trivial pairings and valid pairings with the
        // small point at infinity are excluded from this counting)
        ecPairingMillerLoops.addPrecompileLimit(ecDataOperation.circuitSelectorEcPairingCounter());

        // ecPairingFinalExponentiation case
        // NOTE: if at least one Miller Loop is computed, the final exponentiation is 1
        ecPairingFinalExponentiations.addPrecompileLimit(
            ecDataOperation.circuitSelectorEcPairingCounter() > 0
                ? 1
                : 0); // See https://eprint.iacr.org/2008/490.pdf
      }
      default -> throw new IllegalArgumentException("Operation not supported by EcData");
    }
  }
}
