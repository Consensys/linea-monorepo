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

package net.consensys.linea.zktracer.module.hub.fragment.imc.oob.opcodes;

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.*;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.types.EWord;

@Getter
@Setter
public class JumpOobCall extends OobCall {
  EWord pcNew;
  BigInteger codeSize;
  boolean jumpGuaranteedException;
  boolean jumpMustBeAttempted;

  public JumpOobCall() {
    super(OOB_INST_JUMP);
  }

  public BigInteger pcNewHi() {
    return pcNew.hiBigInt();
  }

  public BigInteger pcNewLo() {
    return pcNew.loBigInt();
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .data1(bigIntegerToBytes(pcNewHi()))
        .data2(bigIntegerToBytes(pcNewLo()))
        .data3(ZERO)
        .data4(ZERO)
        .data5(bigIntegerToBytes(codeSize))
        .data6(ZERO)
        .data7(booleanToBytes(jumpGuaranteedException))
        .data8(booleanToBytes(jumpMustBeAttempted))
        .data9(ZERO);
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(pcNewHi()))
        .pMiscOobData2(bigIntegerToBytes(pcNewLo()))
        .pMiscOobData3(ZERO)
        .pMiscOobData4(ZERO)
        .pMiscOobData5(bigIntegerToBytes(codeSize))
        .pMiscOobData6(ZERO)
        .pMiscOobData7(booleanToBytes(jumpGuaranteedException))
        .pMiscOobData8(booleanToBytes(jumpMustBeAttempted))
        .pMiscOobData9(ZERO);
  }
}
