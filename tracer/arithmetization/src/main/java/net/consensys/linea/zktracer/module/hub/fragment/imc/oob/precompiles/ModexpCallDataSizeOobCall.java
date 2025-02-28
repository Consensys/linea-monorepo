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

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.OOB_INST_MODEXP_CDS;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;

@Getter
@Setter
public class ModexpCallDataSizeOobCall extends OobCall {

  BigInteger cds;
  boolean extractBbs;
  boolean extractEbs;
  boolean extractMbs;

  public ModexpCallDataSizeOobCall() {
    super(OOB_INST_MODEXP_CDS);
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .data1(ZERO)
        .data2(bigIntegerToBytes(cds))
        .data3(booleanToBytes(extractBbs))
        .data4(booleanToBytes(extractEbs))
        .data5(booleanToBytes(extractMbs))
        .data6(ZERO)
        .data7(ZERO)
        .data8(ZERO)
        .data9(ZERO);
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(ZERO)
        .pMiscOobData2(bigIntegerToBytes(cds))
        .pMiscOobData3(booleanToBytes(extractBbs))
        .pMiscOobData4(booleanToBytes(extractEbs))
        .pMiscOobData5(booleanToBytes(extractMbs))
        .pMiscOobData6(ZERO)
        .pMiscOobData7(ZERO)
        .pMiscOobData8(ZERO)
        .pMiscOobData9(ZERO);
  }
}
