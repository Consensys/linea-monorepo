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

package net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles;

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.*;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.booleanToBytes;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;

@Getter
@Setter
public class ModexpPricingOobCall extends OobCall {

  final BigInteger callGas;
  BigInteger returnAtCapacity;
  boolean ramSuccess;
  BigInteger exponentLog;
  int maxMbsBbs;

  BigInteger returnGas;
  boolean returnAtCapacityNonZero;

  public ModexpPricingOobCall(long calleeGas) {
    super(OOB_INST_MODEXP_PRICING);
    this.callGas = BigInteger.valueOf(calleeGas);
  }

  @Override
  public net.consensys.linea.zktracer.module.oob.Trace trace(
      net.consensys.linea.zktracer.module.oob.Trace trace) {
    return trace
        .data1(bigIntegerToBytes(callGas))
        .data2(ZERO)
        .data3(bigIntegerToBytes(returnAtCapacity))
        .data4(booleanToBytes(ramSuccess))
        .data5(bigIntegerToBytes(returnGas))
        .data6(bigIntegerToBytes(exponentLog))
        .data7(bigIntegerToBytes(BigInteger.valueOf(maxMbsBbs)))
        .data8(booleanToBytes(returnAtCapacityNonZero))
        .data9(ZERO);
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(callGas))
        .pMiscOobData2(ZERO)
        .pMiscOobData3(bigIntegerToBytes(returnAtCapacity))
        .pMiscOobData4(booleanToBytes(ramSuccess))
        .pMiscOobData5(bigIntegerToBytes(returnGas))
        .pMiscOobData6(bigIntegerToBytes(exponentLog))
        .pMiscOobData7(bigIntegerToBytes(BigInteger.valueOf(maxMbsBbs)))
        .pMiscOobData8(booleanToBytes(returnAtCapacityNonZero))
        .pMiscOobData9(ZERO);
  }
}
