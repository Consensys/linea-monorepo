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

package net.consensys.linea.zktracer.module.hub.fragment.misc.subfragment;

import java.math.BigInteger;

import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.opcode.OpCode;

public record MxpSubFragment(
    boolean mxpException,
    byte opCode,
    boolean deploys,
    int memorySize,
    long gasMxp,
    EWord offset1,
    EWord offset2,
    EWord size1,
    EWord size2)
    implements TraceSubFragment {
  public static MxpSubFragment build(Hub hub) {
    final OpCode opCode = hub.currentFrame().opCode();

    // TODO: call the MXP here
    // TODO: get from Mxp all the following
    // TODO: check hub mxpx == mxp mxpx
    long gasMxp = 0;
    EWord offset1 = EWord.ZERO;
    EWord offset2 = EWord.ZERO;
    EWord size1 = EWord.ZERO;
    EWord size2 = EWord.ZERO;

    return new MxpSubFragment(
        hub.exceptions().outOfMemoryExpansion(),
        opCode.byteValue(),
        opCode == OpCode.RETURN && hub.currentFrame().codeDeploymentStatus(),
        hub.currentFrame().frame().memoryWordSize(),
        gasMxp,
        offset1,
        offset2,
        size1,
        size2);
  }

  @Override
  public Trace.TraceBuilder trace(Trace.TraceBuilder trace) {
    return trace
        .pMiscellaneousMxpMxpx(this.mxpException)
        .pMiscellaneousMxpInst(BigInteger.valueOf(this.opCode))
        .pMiscellaneousMxpDeploys(this.deploys)
        .pMiscellaneousMxpWords(BigInteger.valueOf(this.memorySize))
        .pMiscellaneousMxpGasMxp(BigInteger.valueOf(this.gasMxp))
        .pMiscellaneousMxpOffset1Hi(this.offset1.hiBigInt())
        .pMiscellaneousMxpOffset1Lo(this.offset1.loBigInt())
        .pMiscellaneousMxpOffset2Hi(this.offset2.hiBigInt())
        .pMiscellaneousMxpOffset2Lo(this.offset2.loBigInt())
        .pMiscellaneousMxpSize1Hi(this.size1.hiBigInt())
        .pMiscellaneousMxpSize1Lo(this.size1.loBigInt())
        .pMiscellaneousMxpSize2Hi(this.size2.hiBigInt())
        .pMiscellaneousMxpSize2Lo(this.size2.loBigInt());
  }
}
