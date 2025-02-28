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

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.OOB_INST_CDL;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.types.EWord;

@Getter
@Setter
public class CallDataLoadOobCall extends OobCall {
  EWord offset;
  BigInteger cds;
  boolean cdlOutOfBounds;

  public CallDataLoadOobCall() {
    super(OOB_INST_CDL);
  }

  public BigInteger offsetHi() {
    return offset.hiBigInt();
  }

  public BigInteger offsetLo() {
    return offset.loBigInt();
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .data1(bigIntegerToBytes(offsetHi()))
        .data2(bigIntegerToBytes(offsetLo()))
        .data3(ZERO)
        .data4(ZERO)
        .data5(bigIntegerToBytes(cds))
        .data6(ZERO)
        .data7(booleanToBytes(cdlOutOfBounds))
        .data8(ZERO)
        .data9(ZERO);
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(offsetHi()))
        .pMiscOobData2(bigIntegerToBytes(offsetLo()))
        .pMiscOobData3(ZERO)
        .pMiscOobData4(ZERO)
        .pMiscOobData5(bigIntegerToBytes(cds))
        .pMiscOobData6(ZERO)
        .pMiscOobData7(booleanToBytes(cdlOutOfBounds))
        .pMiscOobData8(ZERO)
        .pMiscOobData9(ZERO);
  }
}
