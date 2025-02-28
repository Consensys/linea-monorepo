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

package net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles;

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.*;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;

@Getter
@Setter
public class ModexpXbsOobCall extends OobCall {

  final ModexpXbsCase modexpXbsCase;
  BigInteger xbsHi;
  BigInteger xbsLo;
  BigInteger ybsLo;
  boolean computeMax;

  BigInteger maxXbsYbs;
  boolean xbsNonZero;

  public ModexpXbsOobCall(ModexpXbsCase modexpXbsCase) {
    super(OOB_INST_MODEXP_XBS);
    this.modexpXbsCase = modexpXbsCase;
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .data1(bigIntegerToBytes(xbsHi))
        .data2(bigIntegerToBytes(xbsLo))
        .data3(bigIntegerToBytes(ybsLo))
        .data4(booleanToBytes(computeMax))
        .data5(ZERO)
        .data6(ZERO)
        .data7(bigIntegerToBytes(maxXbsYbs))
        .data8(booleanToBytes(xbsNonZero))
        .data9(ZERO);
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(xbsHi))
        .pMiscOobData2(bigIntegerToBytes(xbsLo))
        .pMiscOobData3(bigIntegerToBytes(ybsLo))
        .pMiscOobData4(booleanToBytes(computeMax))
        .pMiscOobData5(ZERO)
        .pMiscOobData6(ZERO)
        .pMiscOobData7(bigIntegerToBytes(maxXbsYbs))
        .pMiscOobData8(booleanToBytes(xbsNonZero))
        .pMiscOobData9(ZERO);
  }
}
