/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.precompiles;

import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_ecadd;
import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_ecmul;
import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_ecpairing;
import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_ecrecover;
import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_identity;
import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_ripemd;
import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_sha2;
import static net.consensys.linea.zktracer.types.Conversions.booleanToBytes;

import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.OobCall;
import net.consensys.linea.zktracer.module.hub.precompiles.PrecompileInvocation;
import net.consensys.linea.zktracer.module.oob.OobDataChannel;
import net.consensys.linea.zktracer.types.Precompile;
import org.apache.tuweni.bytes.Bytes;

public record SimplePrecompileCall(
    PrecompileInvocation scenario, long callGas, long callDataSize, long returnDataRequestedSize)
    implements OobCall {

  @Override
  public Bytes data(OobDataChannel i) {
    return switch (i) {
      case DATA_1 -> Bytes.ofUnsignedLong(callGas);
      case DATA_2 -> Bytes.ofUnsignedLong(callDataSize);
      case DATA_3 -> Bytes.ofUnsignedLong(returnDataRequestedSize);
      case DATA_4 -> scenario.hubFailure() ? Bytes.EMPTY : Bytes.of(1);
      case DATA_5 -> scenario.hubSuccess()
          ? Bytes.ofUnsignedLong(callGas - scenario.precompilePrice())
          : Bytes.EMPTY;
      case DATA_6 -> scenario.precompile().equals(Precompile.EC_PAIRING)
          ? booleanToBytes(scenario.hubSuccess() && callDataSize > 0 && callDataSize % 192 == 0)
          : booleanToBytes(scenario.hubSuccess() && callDataSize > 0);
      case DATA_7 -> booleanToBytes(scenario.hubSuccess() && callDataSize == 0);
      case DATA_8 -> booleanToBytes(returnDataRequestedSize > 0);
    };
  }

  @Override
  public int oobInstruction() {
    return switch (scenario.precompile()) {
      case EC_RECOVER -> OOB_INST_ecrecover;
      case SHA2_256 -> OOB_INST_sha2;
      case RIPEMD_160 -> OOB_INST_ripemd;
      case IDENTITY -> OOB_INST_identity;
      case EC_ADD -> OOB_INST_ecadd;
      case EC_MUL -> OOB_INST_ecmul;
      case EC_PAIRING -> OOB_INST_ecpairing;
      default -> throw new IllegalArgumentException("unexpected complex precompile");
    };
  }
}
