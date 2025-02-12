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

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.OOB_INST_BLAKE_CDS;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;

@Getter
@Setter
public class Blake2fCallDataSizeOobCall extends OobCall {
  BigInteger cds;
  BigInteger returnAtCapacity;
  boolean hubSuccess;
  boolean returnAtCapacityNonZero;

  public Blake2fCallDataSizeOobCall() {
    super(OOB_INST_BLAKE_CDS);
  }

  @Override
  public net.consensys.linea.zktracer.module.oob.Trace trace(
      net.consensys.linea.zktracer.module.oob.Trace trace) {
    return trace
        .data1(ZERO)
        .data2(bigIntegerToBytes(cds))
        .data3(bigIntegerToBytes(returnAtCapacity))
        .data4(booleanToBytes(hubSuccess)) // Set after the constructor
        .data5(ZERO)
        .data6(ZERO)
        .data7(ZERO)
        .data8(booleanToBytes(returnAtCapacityNonZero)) // Set after the constructor
        .data9(ZERO);
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(ZERO)
        .pMiscOobData2(bigIntegerToBytes(cds))
        .pMiscOobData3(bigIntegerToBytes(returnAtCapacity))
        .pMiscOobData4(booleanToBytes(hubSuccess)) // Set after the constructor
        .pMiscOobData5(ZERO)
        .pMiscOobData6(ZERO)
        .pMiscOobData7(ZERO)
        .pMiscOobData8(booleanToBytes(returnAtCapacityNonZero)) // Set after the constructor
        .pMiscOobData9(ZERO);
  }
}
