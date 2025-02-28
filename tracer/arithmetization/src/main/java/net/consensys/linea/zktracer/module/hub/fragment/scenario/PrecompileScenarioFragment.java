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

package net.consensys.linea.zktracer.module.hub.fragment.scenario;

import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileFlag.*;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileScenario.*;

import java.util.Map;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.section.call.precompileSubsection.PrecompileSubsection;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;

@Getter
@AllArgsConstructor
@Accessors(fluent = true)
public class PrecompileScenarioFragment implements TraceFragment {

  public enum PrecompileScenario {
    PRC_FAILURE_KNOWN_TO_HUB,
    PRC_FAILURE_KNOWN_TO_RAM,
    PRC_SUCCESS_WILL_REVERT,
    PRC_SUCCESS_WONT_REVERT;

    public boolean isFailure() {
      return this == PRC_FAILURE_KNOWN_TO_HUB || this == PRC_FAILURE_KNOWN_TO_RAM;
    }

    public boolean isSuccess() {
      return this == PRC_SUCCESS_WILL_REVERT || this == PRC_SUCCESS_WONT_REVERT;
    }
  }

  public enum PrecompileFlag {
    PRC_ECRECOVER,
    PRC_SHA2_256,
    PRC_RIPEMD_160,
    PRC_IDENTITY,
    PRC_MODEXP,
    PRC_ECADD,
    PRC_ECMUL,
    PRC_ECPAIRING,
    PRC_BLAKE2F;

    private static final Map<Address, PrecompileFlag> ADDRESS_TO_FLAG_MAP =
        Map.of(
            Address.ECREC, PRC_ECRECOVER,
            Address.SHA256, PRC_SHA2_256,
            Address.RIPEMD160, PRC_RIPEMD_160,
            Address.ID, PRC_IDENTITY,
            Address.MODEXP, PRC_MODEXP,
            Address.ALTBN128_ADD, PRC_ECADD,
            Address.ALTBN128_MUL, PRC_ECMUL,
            Address.ALTBN128_PAIRING, PRC_ECPAIRING,
            Address.BLAKE2B_F_COMPRESSION, PRC_BLAKE2F);

    private static final Map<PrecompileFlag, Integer> DATA_PHASE_MAP =
        Map.of(
            PRC_ECRECOVER, Trace.PHASE_ECRECOVER_DATA,
            PRC_SHA2_256, Trace.PHASE_SHA2_DATA,
            PRC_RIPEMD_160, Trace.PHASE_RIPEMD_DATA,
            // IDENTITY not supported
            // MODEXP not supported
            PRC_ECADD, Trace.PHASE_ECADD_DATA,
            PRC_ECMUL, Trace.PHASE_ECMUL_DATA,
            PRC_ECPAIRING, Trace.PHASE_ECPAIRING_DATA
            // BLAKE2f not supported
            );

    private static final Map<PrecompileFlag, Integer> RESULT_PHASE_MAP =
        Map.of(
            PRC_ECRECOVER, Trace.PHASE_ECRECOVER_RESULT,
            PRC_SHA2_256, Trace.PHASE_SHA2_RESULT,
            PRC_RIPEMD_160, Trace.PHASE_RIPEMD_RESULT,
            // IDENTITY not supported
            PRC_MODEXP, Trace.PHASE_MODEXP_RESULT,
            PRC_ECADD, Trace.PHASE_ECADD_RESULT,
            PRC_ECMUL, Trace.PHASE_ECMUL_RESULT,
            PRC_ECPAIRING, Trace.PHASE_ECPAIRING_RESULT,
            PRC_BLAKE2F, Trace.PHASE_BLAKE_RESULT);

    public static PrecompileFlag addressToPrecompileFlag(Address precompileAddress) {
      if (!ADDRESS_TO_FLAG_MAP.containsKey(precompileAddress)) {
        throw new IllegalArgumentException("Not valid London precompile address");
      }
      return ADDRESS_TO_FLAG_MAP.get(precompileAddress);
    }

    public int dataPhase() {
      if (!DATA_PHASE_MAP.containsKey(this)) {
        throw new IllegalArgumentException("Precompile not supported by the DATA_PHASE_MAP");
      }
      return DATA_PHASE_MAP.get(this);
    }

    public int resultPhase() {
      if (!RESULT_PHASE_MAP.containsKey(this)) {
        throw new IllegalArgumentException("Precompile not supported by the RESULT_PHASE_MAP");
      }
      return RESULT_PHASE_MAP.get(this);
    }

    public boolean isEcdataPrecompile() {
      return this.isAnyOf(PRC_ECRECOVER, PRC_ECADD, PRC_ECMUL, PRC_ECPAIRING);
    }

    public boolean isAnyOf(PrecompileFlag... flags) {
      for (PrecompileFlag flag : flags) {
        if (this == flag) {
          return true;
        }
      }
      return false;
    }
  }

  final PrecompileSubsection precompileSubSection;
  @Setter public PrecompileScenario scenario;
  @Setter public PrecompileFlag flag;

  public PrecompileScenarioFragment(
      final PrecompileSubsection precompileSubsection,
      final PrecompileFlag flag,
      final PrecompileScenario scenario) {
    this.precompileSubSection = precompileSubsection;
    this.flag = flag;
    this.scenario = scenario;
  }

  public boolean isPrcFailure() {
    return scenario == PRC_FAILURE_KNOWN_TO_HUB || scenario == PRC_FAILURE_KNOWN_TO_RAM;
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    trace
        .peekAtScenario(true)
        // // Precompile scenarios
        ////////////////////
        .pScenarioPrcEcrecover(flag == PRC_ECRECOVER)
        .pScenarioPrcSha2256(flag == PRC_SHA2_256)
        .pScenarioPrcRipemd160(flag == PRC_RIPEMD_160)
        .pScenarioPrcIdentity(flag == PRC_IDENTITY)
        .pScenarioPrcModexp(flag == PRC_MODEXP)
        .pScenarioPrcEcadd(flag == PRC_ECADD)
        .pScenarioPrcEcmul(flag == PRC_ECMUL)
        .pScenarioPrcEcpairing(flag == PRC_ECPAIRING)
        .pScenarioPrcBlake2f(flag == PRC_BLAKE2F)
        .pScenarioPrcSuccessCallerWillRevert(scenario == PRC_SUCCESS_WILL_REVERT)
        .pScenarioPrcSuccessCallerWontRevert(scenario == PRC_SUCCESS_WONT_REVERT)
        .pScenarioPrcFailureKnownToHub(scenario == PRC_FAILURE_KNOWN_TO_HUB)
        .pScenarioPrcFailureKnownToRam(scenario == PRC_FAILURE_KNOWN_TO_RAM)
        .pScenarioPrcCallerGas(Bytes.ofUnsignedLong(precompileSubSection.callerGas()))
        .pScenarioPrcCalleeGas(Bytes.ofUnsignedLong(precompileSubSection.calleeGas()))
        .pScenarioPrcReturnGas(Bytes.ofUnsignedLong(precompileSubSection.returnGas()))
        .pScenarioPrcCdo(precompileSubSection.callDataOffset())
        .pScenarioPrcCds(precompileSubSection.callDataSize())
        .pScenarioPrcRao(precompileSubSection.returnAtOffset())
        .pScenarioPrcRac(precompileSubSection.returnAtCapacity());
    return trace;
  }
}
