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

package net.consensys.linea.zktracer.module.hub.fragment.misc.subfragment.oob;

import java.math.BigInteger;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

public record JumpSubFragment(
    EWord targetPc,
    EWord jumpCondition,
    int codeSize,
    OpCode opCode,
    boolean event1,
    boolean event2)
    implements TraceSubFragment {

  public static JumpSubFragment build(Hub hub, MessageFrame frame) {
    final OpCode opCode = hub.currentFrame().opCode();
    final long targetPc = Words.clampedToLong(frame.getStackItem(0));
    final boolean invalidDestination = frame.getCode().isJumpDestInvalid((int) targetPc);

    long jumpCondition = 0;
    boolean event1;
    switch (opCode) {
      case JUMP -> event1 = invalidDestination;
      case JUMPI -> {
        jumpCondition = Words.clampedToLong(frame.getStackItem(1));
        event1 = (jumpCondition != 0) && invalidDestination;
      }
      default -> throw new IllegalArgumentException("Unexpected opcode");
    }

    return new JumpSubFragment(
        EWord.of(BigInteger.valueOf(targetPc)),
        EWord.of(BigInteger.valueOf(jumpCondition)),
        frame.getWorldUpdater().get(hub.currentFrame().codeAddress()).getCode().size(),
        opCode,
        event1,
        jumpCondition > 0);
  }

  @Override
  public Trace.TraceBuilder trace(Trace.TraceBuilder trace) {
    return trace
        .pMiscellaneousOobOutgoingData1(this.targetPc.hiBigInt())
        .pMiscellaneousOobOutgoingData2(this.targetPc.loBigInt())
        .pMiscellaneousOobOutgoingData3(this.jumpCondition.hiBigInt())
        .pMiscellaneousOobOutgoingData4(this.jumpCondition.loBigInt())
        .pMiscellaneousOobOutgoingData5(BigInteger.valueOf(this.codeSize))
        .pMiscellaneousOobOutgoingData6(BigInteger.ZERO)
        .pMiscellaneousOobInst(BigInteger.valueOf(this.opCode.byteValue()))
        .pMiscellaneousOobEvent1(this.event1)
        .pMiscellaneousOobEvent2(this.event2);
  }
}
