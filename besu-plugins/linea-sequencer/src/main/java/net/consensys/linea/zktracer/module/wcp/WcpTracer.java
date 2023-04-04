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
package net.consensys.linea.zktracer.module.wcp;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.hyperledger.besu.evm.frame.MessageFrame;

import java.util.List;

import org.apache.tuweni.bytes.Bytes32;

public class WcpTracer implements ModuleTracer {
  private static final int LIMB_SIZE = 16;
  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "wcp";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of(OpCode.LT, OpCode.GT, OpCode.SLT, OpCode.SGT, OpCode.EQ, OpCode.ISZERO);
  }

  @Override
  public Object trace(final MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

    final WcpData data = new WcpData(opCode, arg1, arg2);
    final WcpTrace.Trace.Builder builder = WcpTrace.Trace.Builder.newInstance();

    stamp++;
    for (int ct = 0; ct < maxCt(data.isOneLineInstruction()); ct++) {
      builder
          .appendWcpStamp(stamp)
          .appendOneLineInstruction(data.isOneLineInstruction())
          .appendCounter(ct)
          .appendInst(UnsignedByte.of(opCode.value))
          .appendArg1Hi(data.getArg1Hi().toUnsignedBigInteger())
          .appendArg1Lo(data.getArg1Lo().toUnsignedBigInteger())
          .appendArg2Hi(data.getArg2Hi().toUnsignedBigInteger())
          .appendArg2Lo(data.getArg2Lo().toUnsignedBigInteger())
          .appendResHi(data.getResHi())
          .appendResLo(data.getResLo())
          .appendBits(data.getBits().get(ct))
          .appendNeg1(data.getNeg1())
          .appendNeg2(data.getNeg2())
          .appendByte1(UnsignedByte.of(data.getArg1Hi().get(ct)))
          .appendByte2(UnsignedByte.of(data.getArg1Lo().get(ct)))
          .appendByte3(UnsignedByte.of(data.getArg2Hi().get(ct)))
          .appendByte4(UnsignedByte.of(data.getArg2Lo().get(ct)))
          .appendByte5(UnsignedByte.of(data.getAdjHi().get(ct)))
          .appendByte6(UnsignedByte.of(data.getAdjLo().get(ct)))
          .appendAcc1(data.getArg1Hi().slice(0, 1 + ct).toUnsignedBigInteger())
          .appendAcc2(data.getArg1Lo().slice(0, 1 + ct).toUnsignedBigInteger())
          .appendAcc3(data.getArg2Hi().slice(0, 1 + ct).toUnsignedBigInteger())
          .appendAcc4(data.getArg2Lo().slice(0, 1 + ct).toUnsignedBigInteger())
          .appendAcc5(data.getAdjHi().slice(0, 1 + ct).toUnsignedBigInteger())
          .appendAcc6(data.getAdjLo().slice(0, 1 + ct).toUnsignedBigInteger())
          .appendBit1(data.getBit1())
          .appendBit2(data.getBit2())
          .appendBit3(data.getBit3())
          .appendBit4(data.getBit4());
    }

    builder.setStamp(stamp);
    return builder.build();
  }

  private int maxCt(final boolean isOneLineInstruction) {
    return isOneLineInstruction ? 1 : LIMB_SIZE;
  }
}
