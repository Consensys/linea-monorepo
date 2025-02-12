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

import static com.google.common.base.Preconditions.*;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction;

@Getter
@Setter
public class PrecompileCommonOobCall extends OobCall {

  BigInteger calleeGas;
  BigInteger cds;
  BigInteger returnAtCapacity;
  boolean hubSuccess;
  BigInteger returnGas;
  boolean returnAtCapacityNonZero;
  boolean cdsIsZero; // Necessary to compute extractCallData and emptyCallData

  public PrecompileCommonOobCall(OobInstruction oobInstruction, long calleeGas) {
    super(oobInstruction);
    this.calleeGas = BigInteger.valueOf(calleeGas);
    checkArgument(oobInstruction.isCommonPrecompile());
  }

  public boolean getExtractCallData() {
    return hubSuccess && !cdsIsZero;
  }

  public boolean getCallDataIsEmpty() {
    return hubSuccess && cdsIsZero;
  }

  @Override
  public net.consensys.linea.zktracer.module.oob.Trace trace(
      net.consensys.linea.zktracer.module.oob.Trace trace) {
    return trace
        .data1(bigIntegerToBytes(calleeGas))
        .data2(bigIntegerToBytes(cds))
        .data3(bigIntegerToBytes(returnAtCapacity))
        .data4(booleanToBytes(hubSuccess)) // Set after the constructor
        .data5(bigIntegerToBytes(returnGas)) // Set after the constructor
        .data6(booleanToBytes(getExtractCallData())) // Derived from other parameters
        .data7(booleanToBytes(getCallDataIsEmpty())) // Derived from other parameters
        .data8(booleanToBytes(returnAtCapacityNonZero)) // Set after the constructor
        .data9(ZERO);
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(calleeGas))
        .pMiscOobData2(bigIntegerToBytes(cds))
        .pMiscOobData3(bigIntegerToBytes(returnAtCapacity))
        .pMiscOobData4(booleanToBytes(hubSuccess)) // Set after the constructor
        .pMiscOobData5(bigIntegerToBytes(returnGas)) // Set after the constructor
        .pMiscOobData6(booleanToBytes(getExtractCallData())) // Derived from other parameters
        .pMiscOobData7(booleanToBytes(getCallDataIsEmpty())) // Derived from other parameters
        .pMiscOobData8(booleanToBytes(returnAtCapacityNonZero)) // Set after the constructor
        .pMiscOobData9(ZERO);
  }
}
