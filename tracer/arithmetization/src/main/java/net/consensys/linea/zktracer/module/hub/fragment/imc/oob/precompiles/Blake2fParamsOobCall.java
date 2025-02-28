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

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.OOB_INST_BLAKE_PARAMS;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;

@Getter
@Setter
public class Blake2fParamsOobCall extends OobCall {

  BigInteger calleeGas;
  BigInteger blakeR;
  BigInteger blakeF;

  boolean ramSuccess;
  BigInteger returnGas;

  public Blake2fParamsOobCall(long calleeGas) {
    super(OOB_INST_BLAKE_PARAMS);
    this.calleeGas = BigInteger.valueOf(calleeGas);
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .data1(bigIntegerToBytes(calleeGas))
        .data2(ZERO)
        .data3(ZERO)
        .data4(booleanToBytes(ramSuccess)) // Set after the constructor
        .data5(bigIntegerToBytes(returnGas)) // Set after the constructor
        .data6(bigIntegerToBytes(blakeR))
        .data7(bigIntegerToBytes(blakeF))
        .data8(ZERO)
        .data9(ZERO);
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(calleeGas))
        .pMiscOobData2(ZERO)
        .pMiscOobData3(ZERO)
        .pMiscOobData4(booleanToBytes(ramSuccess)) // Set after the constructor
        .pMiscOobData5(bigIntegerToBytes(returnGas)) // Set after the constructor
        .pMiscOobData6(bigIntegerToBytes(blakeR))
        .pMiscOobData7(bigIntegerToBytes(blakeF))
        .pMiscOobData8(ZERO)
        .pMiscOobData9(ZERO);
  }
}
