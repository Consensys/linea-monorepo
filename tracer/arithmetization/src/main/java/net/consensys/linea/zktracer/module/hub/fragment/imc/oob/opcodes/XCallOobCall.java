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
public class XCallOobCall extends OobCall {
  EWord value;
  boolean valueIsNonzero;
  boolean valueIsZero;

  public XCallOobCall() {
    super(OOB_INST_XCALL);
  }

  public BigInteger valueHi() {
    return value.hiBigInt();
  }

  public BigInteger valueLo() {
    return value.loBigInt();
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .data1(bigIntegerToBytes(valueHi()))
        .data2(bigIntegerToBytes(valueLo()))
        .data3(ZERO)
        .data4(ZERO)
        .data5(ZERO)
        .data6(ZERO)
        .data7(booleanToBytes(valueIsNonzero))
        .data8(booleanToBytes(valueIsZero))
        .data9(ZERO);
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(valueHi()))
        .pMiscOobData2(bigIntegerToBytes(valueLo()))
        .pMiscOobData3(ZERO)
        .pMiscOobData4(ZERO)
        .pMiscOobData5(ZERO)
        .pMiscOobData6(ZERO)
        .pMiscOobData7(booleanToBytes(valueIsNonzero))
        .pMiscOobData8(booleanToBytes(valueIsZero))
        .pMiscOobData9(ZERO);
  }
}
