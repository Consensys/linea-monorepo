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

import static net.consensys.linea.zktracer.Trace.OOB_INST_MODEXP_CDS;
import static net.consensys.linea.zktracer.Trace.Oob.CT_MAX_MODEXP_CDS;
import static net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata.*;
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
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Getter
@Setter
public class ModexpCallDataSizeOobCall extends OobCall {

  final ModexpMetadata modexpMetadata;
  EWord cds;

  public ModexpCallDataSizeOobCall(final ModexpMetadata modexpMetadata) {
    super();
    this.modexpMetadata = modexpMetadata;
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
    exoCalls.add(callToLT(wcp, Bytes.of(BBS_MIN_OFFSET), cds));

    // row i + 1
    exoCalls.add(callToLT(wcp, Bytes.of(EBS_MIN_OFFSET), cds));

    // row i + 2
    exoCalls.add(callToLT(wcp, Bytes.of(MBS_MIN_OFFSET), cds));
  }

  @Override
  public int ctMax() {
    return CT_MAX_MODEXP_CDS;
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .isModexpCds(true)
        .oobInst(OOB_INST_MODEXP_CDS)
        .data2(cds.trimLeadingZeros())
        .data3(booleanToBytes(modexpMetadata.extractBbs()))
        .data4(booleanToBytes(modexpMetadata.extractEbs()))
        .data5(booleanToBytes(modexpMetadata.extractMbs()));
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(OOB_INST_MODEXP_CDS)
        .pMiscOobData2(cds.trimLeadingZeros())
        .pMiscOobData3(booleanToBytes(modexpMetadata.extractBbs()))
        .pMiscOobData4(booleanToBytes(modexpMetadata.extractEbs()))
        .pMiscOobData5(booleanToBytes(modexpMetadata.extractMbs()));
  }
}
