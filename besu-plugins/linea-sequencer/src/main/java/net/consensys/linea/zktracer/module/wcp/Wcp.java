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

import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Module;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class Wcp implements Module {
  private static final int LIMB_SIZE = 16;

  final Trace.TraceBuilder builder = Trace.builder();
  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "wcp";
  }

  @Override
  public final List<OpCode> supportedOpCodes() {
    return List.of(OpCode.LT, OpCode.GT, OpCode.SLT, OpCode.SGT, OpCode.EQ, OpCode.ISZERO);
  }

  @Override
  public void trace(final MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));
    final WcpData data = new WcpData(opCode, arg1, arg2);

    stamp++;

    for (int ct = 0; ct < maxCt(data.isOneLineInstruction()); ct++) {
      builder
          .wordComparisonStampArg(stamp)
          .oneLineInstructionArg(data.isOneLineInstruction())
          .counterArg(ct)
          .instArg(UnsignedByte.of(opCode.value))
          .argument1HiArg(data.getArg1Hi().toUnsignedBigInteger())
          .argument1LoArg(data.getArg1Lo().toUnsignedBigInteger())
          .argument2HiArg(data.getArg2Hi().toUnsignedBigInteger())
          .argument2LoArg(data.getArg2Lo().toUnsignedBigInteger())
          .resultHiArg(data.getResHi())
          .resultLoArg(data.getResLo())
          .bitsArg(data.getBits().get(ct))
          .neg1Arg(data.getNeg1())
          .neg2Arg(data.getNeg2())
          .byte1Arg(UnsignedByte.of(data.getArg1Hi().get(ct)))
          .byte2Arg(UnsignedByte.of(data.getArg1Lo().get(ct)))
          .byte3Arg(UnsignedByte.of(data.getArg2Hi().get(ct)))
          .byte4Arg(UnsignedByte.of(data.getArg2Lo().get(ct)))
          .byte5Arg(UnsignedByte.of(data.getAdjHi().get(ct)))
          .byte6Arg(UnsignedByte.of(data.getAdjLo().get(ct)))
          .acc1Arg(data.getArg1Hi().slice(0, 1 + ct).toUnsignedBigInteger())
          .acc2Arg(data.getArg1Lo().slice(0, 1 + ct).toUnsignedBigInteger())
          .acc3Arg(data.getArg2Hi().slice(0, 1 + ct).toUnsignedBigInteger())
          .acc4Arg(data.getArg2Lo().slice(0, 1 + ct).toUnsignedBigInteger())
          .acc5Arg(data.getAdjHi().slice(0, 1 + ct).toUnsignedBigInteger())
          .acc6Arg(data.getAdjLo().slice(0, 1 + ct).toUnsignedBigInteger())
          .bit1Arg(data.getBit1())
          .bit2Arg(data.getBit2())
          .bit3Arg(data.getBit3())
          .bit4Arg(data.getBit4());
    }
  }

  @Override
  public Object commit() {
    return new WcpTrace(builder.build(), stamp);
  }

  private int maxCt(final boolean isOneLineInstruction) {
    return isOneLineInstruction ? 1 : LIMB_SIZE;
  }
}
