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

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.OOB_INST_CALL;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.types.EWord;

@Getter
@Setter
public class CallOobCall extends OobCall {
  public EWord value;
  BigInteger balance;
  BigInteger callStackDepth;
  boolean abortingCondition;

  public CallOobCall() {
    super(OOB_INST_CALL);
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
        .data3(bigIntegerToBytes(balance))
        .data4(ZERO)
        .data5(ZERO)
        .data6(bigIntegerToBytes(callStackDepth))
        .data7(booleanToBytes(!value.isZero()))
        .data8(booleanToBytes(abortingCondition))
        .data9(ZERO);
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(valueHi()))
        .pMiscOobData2(bigIntegerToBytes(valueLo()))
        .pMiscOobData3(bigIntegerToBytes(balance))
        .pMiscOobData4(ZERO)
        .pMiscOobData5(ZERO)
        .pMiscOobData6(bigIntegerToBytes(callStackDepth))
        .pMiscOobData7(booleanToBytes(!value.isZero()))
        .pMiscOobData8(booleanToBytes(abortingCondition))
        .pMiscOobData9(ZERO);
  }
}
