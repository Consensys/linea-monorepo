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

package net.consensys.linea.zktracer.module.hub.fragment.imc;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.State;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;

@RequiredArgsConstructor
public class MxpCall implements TraceSubFragment {

  public final Hub hub;

  // filled in by MXP module
  @Getter @Setter public OpCodeData opCodeData;
  @Getter @Setter public boolean deploys;
  @Getter @Setter public EWord offset1 = EWord.ZERO;
  @Getter @Setter public EWord size1 = EWord.ZERO;
  @Getter @Setter public EWord offset2 = EWord.ZERO;
  @Getter @Setter public EWord size2 = EWord.ZERO;
  @Setter public boolean mayTriggerNontrivialMmuOperation;

  /** mxpx is short of Memory eXPansion eXception */
  @Getter @Setter public boolean mxpx;

  @Getter @Setter public long memorySizeInWords;
  @Getter @Setter public long gasMxp;

  public static MxpCall build(Hub hub) {
    return new MxpCall(hub);
  }

  static boolean getMemoryExpansionException(Hub hub) {
    return Exceptions.memoryExpansionException(hub.pch().exceptions());
  }

  public boolean getSize1NonZeroNoMxpx() {
    return !this.mxpx && !this.size1.isZero();
  }

  public boolean getSize2NonZeroNoMxpx() {
    return !this.mxpx && !this.size2.isZero();
  }

  public Trace trace(Trace trace, State hubState) {
    hubState.incrementMxpStamp();
    return trace
        .pMiscMxpFlag(true)
        .pMiscMxpInst(this.opCodeData.value())
        .pMiscMxpDeploys(this.deploys)
        .pMiscMxpOffset1Hi(this.offset1.hi())
        .pMiscMxpOffset1Lo(this.offset1.lo())
        .pMiscMxpSize1Hi(this.size1.hi())
        .pMiscMxpSize1Lo(this.size1.lo())
        .pMiscMxpOffset2Hi(this.offset2.hi())
        .pMiscMxpOffset2Lo(this.offset2.lo())
        .pMiscMxpSize2Hi(this.size2.hi())
        .pMiscMxpSize2Lo(this.size2.lo())
        .pMiscMxpMtntop(this.mayTriggerNontrivialMmuOperation)
        .pMiscMxpSize1NonzeroNoMxpx(this.getSize1NonZeroNoMxpx())
        .pMiscMxpSize2NonzeroNoMxpx(this.getSize2NonZeroNoMxpx())
        .pMiscMxpMxpx(this.mxpx)
        .pMiscMxpWords(Bytes.ofUnsignedLong(this.memorySizeInWords))
        .pMiscMxpGasMxp(Bytes.ofUnsignedLong(this.gasMxp));
  }
}
