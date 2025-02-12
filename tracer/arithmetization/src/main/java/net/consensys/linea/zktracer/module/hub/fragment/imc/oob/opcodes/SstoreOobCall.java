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
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;

@Getter
@Setter
public class SstoreOobCall extends OobCall {
  BigInteger gas;
  boolean sstorex;

  public SstoreOobCall() {
    super(OOB_INST_SSTORE);
  }

  @Override
  public net.consensys.linea.zktracer.module.oob.Trace trace(
      net.consensys.linea.zktracer.module.oob.Trace trace) {
    return trace
        .data1(ZERO)
        .data2(ZERO)
        .data3(ZERO)
        .data4(ZERO)
        .data5(bigIntegerToBytes(gas))
        .data6(ZERO)
        .data7(booleanToBytes(sstorex))
        .data8(ZERO)
        .data9(ZERO);
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(ZERO)
        .pMiscOobData2(ZERO)
        .pMiscOobData3(ZERO)
        .pMiscOobData4(ZERO)
        .pMiscOobData5(bigIntegerToBytes(gas))
        .pMiscOobData6(ZERO)
        .pMiscOobData7(booleanToBytes(sstorex))
        .pMiscOobData8(ZERO)
        .pMiscOobData9(ZERO);
  }
}
