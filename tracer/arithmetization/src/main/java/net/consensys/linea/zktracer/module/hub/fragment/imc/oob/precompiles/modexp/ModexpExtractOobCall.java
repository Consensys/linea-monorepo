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

import static net.consensys.linea.zktracer.Trace.OOB_INST_MODEXP_EXTRACT;
import static net.consensys.linea.zktracer.Trace.Oob.CT_MAX_MODEXP_EXTRACT;
import static net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata.BASE_MIN_OFFSET;
import static net.consensys.linea.zktracer.module.oob.OobExoCall.callToIsZero;
import static net.consensys.linea.zktracer.module.oob.OobExoCall.callToLT;
import static net.consensys.linea.zktracer.runtime.callstack.CallFrame.getOpCode;
import static net.consensys.linea.zktracer.types.Conversions.*;

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
public class ModexpExtractOobCall extends OobCall {

  final ModexpMetadata metadata;
  EWord cds;

  boolean extractBase;
  boolean extractExponent;
  boolean extractModulus;

  public ModexpExtractOobCall(ModexpMetadata metadata) {
    super();
    this.metadata = metadata;
  }

  @Override
  public void setInputData(MessageFrame frame, Hub hub) {
    final OpCode opCode = getOpCode(frame);
    final int cdsIndex = opCode.callHasValueArgument() ? 4 : 3;
    setCds(EWord.of(frame.getStackItem(cdsIndex)));
  }

  @Override
  public void callExoModules(Add add, Mod mod, Wcp wcp) {
    // row i
    final OobExoCall bbsIsZeroCall = callToIsZero(wcp, metadata.bbs());
    exoCalls.add(bbsIsZeroCall);
    final boolean bbsIsZero = bytesToBoolean(bbsIsZeroCall.result());

    // row i + 1
    final OobExoCall ebsIsZeroCall = callToIsZero(wcp, metadata.ebs());
    exoCalls.add(ebsIsZeroCall);
    final boolean ebsIsZero = bytesToBoolean(ebsIsZeroCall.result());

    // row i + 2
    final OobExoCall mbsIsZeroCall = callToIsZero(wcp, metadata.mbs());
    exoCalls.add(mbsIsZeroCall);
    final boolean mbsIsZero = bytesToBoolean(mbsIsZeroCall.result());

    // row i + 3
    final OobExoCall callDataExtendsBeyondExponentCall =
        callToLT(
            wcp,
            Bytes.ofUnsignedLong(BASE_MIN_OFFSET + metadata.bbsInt() + metadata.ebsInt()),
            cds);
    exoCalls.add(callDataExtendsBeyondExponentCall);
    final boolean callDataExtendsBeyondExponent =
        bytesToBoolean(callDataExtendsBeyondExponentCall.result());

    setExtractModulus(callDataExtendsBeyondExponent && !mbsIsZero);
    setExtractBase(extractModulus && !bbsIsZero);
    setExtractExponent(extractModulus && !ebsIsZero);
  }

  @Override
  public int ctMax() {
    return CT_MAX_MODEXP_EXTRACT;
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .isModexpExtract(true)
        .oobInst(OOB_INST_MODEXP_EXTRACT)
        .data2(cds.trimLeadingZeros())
        .data3(metadata.bbs().trimLeadingZeros())
        .data4(metadata.ebs().trimLeadingZeros())
        .data5(metadata.mbs().trimLeadingZeros())
        .data6(booleanToBytes(extractBase))
        .data7(booleanToBytes(extractExponent))
        .data8(booleanToBytes(extractModulus));
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(OOB_INST_MODEXP_EXTRACT)
        .pMiscOobData2(cds.trimLeadingZeros())
        .pMiscOobData3(metadata.bbs().trimLeadingZeros())
        .pMiscOobData4(metadata.ebs().trimLeadingZeros())
        .pMiscOobData5(metadata.mbs().trimLeadingZeros())
        .pMiscOobData6(booleanToBytes(extractBase))
        .pMiscOobData7(booleanToBytes(extractExponent))
        .pMiscOobData8(booleanToBytes(extractModulus));
  }
}
