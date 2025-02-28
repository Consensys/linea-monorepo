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

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.OOB_INST_DEPLOYMENT;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.types.EWord;

@Getter
@Setter
public class DeploymentOobCall extends OobCall {
  EWord size;
  boolean maxCodeSizeException;

  public DeploymentOobCall() {
    super(OOB_INST_DEPLOYMENT);
  }

  public BigInteger sizeHi() {
    return size.hiBigInt();
  }

  public BigInteger sizeLo() {
    return size.loBigInt();
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .data1(bigIntegerToBytes(sizeHi()))
        .data2(bigIntegerToBytes(sizeLo()))
        .data3(ZERO)
        .data4(ZERO)
        .data5(ZERO)
        .data6(ZERO)
        .data7(booleanToBytes(maxCodeSizeException))
        .data8(ZERO)
        .data9(ZERO);
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(sizeHi()))
        .pMiscOobData2(bigIntegerToBytes(sizeLo()))
        .pMiscOobData3(ZERO)
        .pMiscOobData4(ZERO)
        .pMiscOobData5(ZERO)
        .pMiscOobData6(ZERO)
        .pMiscOobData7(booleanToBytes(maxCodeSizeException))
        .pMiscOobData8(ZERO)
        .pMiscOobData9(ZERO);
  }
}
