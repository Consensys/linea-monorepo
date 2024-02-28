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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;

public record MxpCall(
    boolean mxpException,
    int opCode,
    boolean deploys,
    int memorySize,
    long gasMxp,
    EWord offset1,
    EWord offset2,
    EWord size1,
    EWord size2)
    implements TraceSubFragment {
  public static MxpCall build(Hub hub) {
    final OpCode opCode = hub.currentFrame().opCode();

    // TODO: call the MXP here
    // TODO: get from Mxp all the following
    // TODO: check hub mxpx == mxp mxpx
    long gasMxp = 0;
    EWord offset1 = EWord.ZERO;
    EWord offset2 = EWord.ZERO;
    EWord size1 = EWord.ZERO;
    EWord size2 = EWord.ZERO;

    return new MxpCall(
        hub.pch().exceptions().outOfMemoryExpansion(),
        hub.currentFrame().opCodeData().value(),
        opCode == OpCode.RETURN && hub.currentFrame().underDeployment(),
        hub.currentFrame().frame().memoryWordSize(),
        gasMxp,
        offset1,
        offset2,
        size1,
        size2);
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscMxpFlag(true)
        .pMiscMxpMxpx(this.mxpException)
        .pMiscMxpInst(Bytes.ofUnsignedInt(this.opCode))
        .pMiscMxpDeploys(this.deploys)
        .pMiscMxpWords(Bytes.ofUnsignedLong(this.memorySize))
        .pMiscMxpGasMxp(Bytes.ofUnsignedLong(this.gasMxp))
        .pMiscMxpOffset1Hi(this.offset1.hi())
        .pMiscMxpOffset1Lo(this.offset1.lo())
        .pMiscMxpOffset2Hi(this.offset2.hi())
        .pMiscMxpOffset2Lo(this.offset2.lo())
        .pMiscMxpSize1Hi(this.size1.hi())
        .pMiscMxpSize1Lo(this.size1.lo())
        .pMiscMxpSize2Hi(this.size2.hi())
        .pMiscMxpSize2Lo(this.size2.lo());
  }
}
