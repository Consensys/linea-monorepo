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

import static net.consensys.linea.zktracer.module.oob.Trace.OOB_INST_modexp_lead;
import static net.consensys.linea.zktracer.types.Conversions.booleanToBytes;

import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.OobCall;
import net.consensys.linea.zktracer.module.oob.OobDataChannel;
import org.apache.tuweni.bytes.Bytes;

public record ModexpLead(int bbsLo, long callDataSize, int ebsLo) implements OobCall {

  @Override
  public Bytes data(OobDataChannel i) {
    return switch (i) {
      case DATA_1 -> Bytes.ofUnsignedLong(bbsLo);
      case DATA_2 -> Bytes.ofUnsignedLong(callDataSize);
      case DATA_3 -> Bytes.ofUnsignedLong(ebsLo);
      case DATA_4 -> booleanToBytes(callDataSize > 96 + bbsLo && ebsLo > 0);
      case DATA_6 -> Bytes.ofUnsignedInt(
          callDataSize > 96 + bbsLo ? Math.min(callDataSize - 96 - bbsLo, 32) : 0);
      case DATA_7 -> Bytes.ofUnsignedLong(Math.min(ebsLo, 32));
      case DATA_8 -> Bytes.ofUnsignedLong(ebsLo < 32 ? 0 : ebsLo - 32);
      default -> Bytes.EMPTY;
    };
  }

  @Override
  public int oobInstruction() {
    return OOB_INST_modexp_lead;
  }
}
