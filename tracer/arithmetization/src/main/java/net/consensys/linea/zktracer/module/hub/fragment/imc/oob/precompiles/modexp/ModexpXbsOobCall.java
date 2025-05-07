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

import static net.consensys.linea.zktracer.Trace.OOB_INST_MODEXP_XBS;
import static net.consensys.linea.zktracer.Trace.Oob.CT_MAX_MODEXP_XBS;
import static net.consensys.linea.zktracer.module.oob.OobExoCall.callToIsZero;
import static net.consensys.linea.zktracer.module.oob.OobExoCall.callToLT;
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
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Getter
@Setter
public class ModexpXbsOobCall extends OobCall {

  final ModexpMetadata modexpMetadata;
  final ModexpXbsCase modexpXbsCase;

  Bytes maxXbsYbs;
  boolean xbsNonZero;

  public ModexpXbsOobCall(ModexpMetadata modexpMetaData, ModexpXbsCase modexpXbsCase) {
    super();
    this.modexpMetadata = modexpMetaData;
    this.modexpXbsCase = modexpXbsCase;
  }

  @Override
  public void setInputData(MessageFrame frame, Hub hub) {}

  @Override
  public void callExoModules(Add add, Mod mod, Wcp wcp) {
    // row i
    exoCalls.add(callToLT(wcp, xbs(), Bytes.ofUnsignedInt(513)));

    // row i + 1
    final OobExoCall compareXbsYbsCall = callToLT(wcp, xbs().lo(), ybsLo());
    exoCalls.add(compareXbsYbsCall);
    final boolean comp = bytesToBoolean(compareXbsYbsCall.result());
    setMaxXbsYbs(computeMax() ? (comp ? ybsLo() : xbs().lo()) : Bytes.EMPTY);

    // row i + 2
    final OobExoCall xbsNonZerCall = callToIsZero(wcp, xbs().lo());
    exoCalls.add(xbsNonZerCall);
    setXbsNonZero(computeMax() ? !bytesToBoolean(xbsNonZerCall.result()) : false);
  }

  @Override
  public int ctMax() {
    return CT_MAX_MODEXP_XBS;
  }

  private EWord xbs() {
    return switch (modexpXbsCase) {
      case OOB_INST_MODEXP_BBS -> modexpMetadata.bbs();
      case OOB_INST_MODEXP_EBS -> modexpMetadata.ebs();
      case OOB_INST_MODEXP_MBS -> modexpMetadata.mbs();
    };
  }

  private Bytes ybsLo() {
    return switch (modexpXbsCase) {
      case OOB_INST_MODEXP_BBS, OOB_INST_MODEXP_EBS -> Bytes.EMPTY;
      case OOB_INST_MODEXP_MBS -> modexpMetadata.bbs().lo();
    };
  }

  private boolean computeMax() {
    return switch (modexpXbsCase) {
      case OOB_INST_MODEXP_BBS, OOB_INST_MODEXP_EBS -> false;
      case OOB_INST_MODEXP_MBS -> true;
    };
  }

  @Override
  public Trace.Oob trace(Trace.Oob trace) {
    return trace
        .isModexpXbs(true)
        .oobInst(OOB_INST_MODEXP_XBS)
        .data1(xbs().hi())
        .data2(xbs().lo())
        .data3(ybsLo())
        .data4(booleanToBytes(computeMax()))
        .data7(maxXbsYbs)
        .data8(booleanToBytes(xbsNonZero));
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    return trace
        .pMiscOobFlag(true)
        .pMiscOobInst(OOB_INST_MODEXP_XBS)
        .pMiscOobData1(xbs().hi())
        .pMiscOobData2(xbs().lo())
        .pMiscOobData3(ybsLo())
        .pMiscOobData4(booleanToBytes(computeMax()))
        .pMiscOobData7(maxXbsYbs)
        .pMiscOobData8(booleanToBytes(xbsNonZero));
  }
}
