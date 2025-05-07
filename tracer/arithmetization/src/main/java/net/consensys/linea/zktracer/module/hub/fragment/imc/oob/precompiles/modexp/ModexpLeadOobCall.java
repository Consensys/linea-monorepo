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

package net.consensys.linea.zktracer.module.hub.fragment.imc.oob.precompiles.modexp;

import static net.consensys.linea.zktracer.Trace.OOB_INST_MODEXP_LEAD;
import static net.consensys.linea.zktracer.Trace.Oob.CT_MAX_MODEXP_LEAD;
import static net.consensys.linea.zktracer.Trace.WORD_SIZE;
import static net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata.BASE_MIN_OFFSET;
import static net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata.EBS_MIN_OFFSET;
import static net.consensys.linea.zktracer.module.oob.OobExoCall.*;
import static net.consensys.linea.zktracer.runtime.callstack.CallFrame.getOpCode;
import static net.consensys.linea.zktracer.types.Conversions.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.oob.OobCall;
import net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.oob.OobExoCall;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Getter
@Setter
public class ModexpLeadOobCall extends OobCall {

  final ModexpMetadata metadata;
  EWord cds;

  boolean loadLead;
  int cdsCutoff;
  int ebsCutoff;
  int subEbs32;

  public ModexpLeadOobCall(ModexpMetadata metadata) {
    super();
    this.metadata = metadata;
  }

  @Override
  public void setInputData(MessageFrame frame, Hub hub) {
    final OpCode opCode = getOpCode(frame);
    final int cdsIndex = opCode.callHasValueArgument() ? 4 : 3;
    cds = EWord.of(frame.getStackItem(cdsIndex));
  }

  @Override
  public void callExoModules(Add add, Mod mod, Wcp wcp) {
    // row i
    final OobExoCall ebsIsZeroCall = callToIsZero(wcp, metadata.ebs());
    exoCalls.add(ebsIsZeroCall);
    final boolean ebsIsZero = bytesToBoolean(ebsIsZeroCall.result());

    // row i + 1
    final OobExoCall ebsLessThan32Call =
        callToLT(wcp, metadata.ebs(), Bytes.ofUnsignedInt(EBS_MIN_OFFSET));
    exoCalls.add(ebsLessThan32Call);
    final boolean ebsLessThan32 = bytesToBoolean(ebsLessThan32Call.result());

    // row i + 2
    final OobExoCall callDataContainsExponentBytesCall =
        callToLT(wcp, metadata.bbs().add(BASE_MIN_OFFSET), cds);
    exoCalls.add(callDataContainsExponentBytesCall);
    final boolean callDataContainsExponentBytes =
        bytesToBoolean(callDataContainsExponentBytesCall.result());

    // row i + 3
    final OobExoCall compCall =
        callDataContainsExponentBytes
            ? callToLT(
                wcp,
                cds.subtract(BASE_MIN_OFFSET).subtract(metadata.bbs()),
                Bytes.ofUnsignedInt(WORD_SIZE))
            : noCall();
    exoCalls.add(compCall);
    final boolean comp = bytesToBoolean(compCall.result());

    final boolean loadLead = callDataContainsExponentBytes && !ebsIsZero;
    setLoadLead(loadLead);

    // Set cdsCutoff
    if (!callDataContainsExponentBytes) {
      setCdsCutoff(0);
    } else {
      setCdsCutoff(
          comp
              ? (cds.toUnsignedBigInteger()
                  .subtract(BigInteger.valueOf(96).add(metadata.bbs().toUnsignedBigInteger()))
                  .intValue())
              : 32);
    }
    // Set ebsCutoff
    setEbsCutoff(ebsLessThan32 ? metadata.ebs().intValue() : 32);

    // Set subEbs32
    setSubEbs32(ebsLessThan32 ? 0 : metadata.ebs().intValue() - 32);
  }

  @Override
  public int ctMax() {
    return CT_MAX_MODEXP_LEAD;
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .isModexpLead(true)
        .oobInst(OOB_INST_MODEXP_LEAD)
        .data1(metadata.bbs().trimLeadingZeros())
        .data2(cds.trimLeadingZeros())
        .data3(metadata.ebs().trimLeadingZeros())
        .data4(booleanToBytes(loadLead))
        .data6(Bytes.ofUnsignedInt(cdsCutoff))
        .data7(Bytes.ofUnsignedInt(ebsCutoff))
        .data8(Bytes.ofUnsignedInt(subEbs32));
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(OOB_INST_MODEXP_LEAD)
        .pMiscOobData1(metadata.bbs().trimLeadingZeros())
        .pMiscOobData2(cds.trimLeadingZeros())
        .pMiscOobData3(metadata.ebs().trimLeadingZeros())
        .pMiscOobData4(booleanToBytes(loadLead))
        .pMiscOobData6(Bytes.ofUnsignedInt(cdsCutoff))
        .pMiscOobData7(Bytes.ofUnsignedInt(ebsCutoff))
        .pMiscOobData8(Bytes.ofUnsignedInt(subEbs32));
  }
}
