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

import static net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobInstruction.OOB_INST_MODEXP_LEAD;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import org.apache.tuweni.bytes.Bytes;

@Getter
@Setter
public class ModexpLeadOobCall extends OobCall {

  BigInteger bbs;
  BigInteger cds;
  BigInteger ebs;

  boolean loadLead;
  int cdsCutoff;
  int ebsCutoff;
  int subEbs32;

  public ModexpLeadOobCall() {
    super(OOB_INST_MODEXP_LEAD);
  }

  @Override
  public net.consensys.linea.zktracer.module.oob.Trace trace(
      net.consensys.linea.zktracer.module.oob.Trace trace) {
    return trace
        .data1(bigIntegerToBytes(bbs))
        .data2(bigIntegerToBytes(cds))
        .data3(bigIntegerToBytes(ebs))
        .data4(booleanToBytes(loadLead))
        .data5(ZERO)
        .data6(Bytes.ofUnsignedInt(cdsCutoff))
        .data7(Bytes.ofUnsignedInt(ebsCutoff))
        .data8(Bytes.ofUnsignedInt(subEbs32))
        .data9(ZERO);
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(oobInstructionValue())
        .pMiscOobData1(bigIntegerToBytes(bbs))
        .pMiscOobData2(bigIntegerToBytes(cds))
        .pMiscOobData3(bigIntegerToBytes(ebs))
        .pMiscOobData4(booleanToBytes(loadLead))
        .pMiscOobData5(ZERO)
        .pMiscOobData6(Bytes.ofUnsignedInt(cdsCutoff))
        .pMiscOobData7(Bytes.ofUnsignedInt(ebsCutoff))
        .pMiscOobData8(Bytes.ofUnsignedInt(subEbs32))
        .pMiscOobData9(ZERO);
  }
}
