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

import java.math.BigInteger;
import java.util.List;

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
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
  public final List<OpCodeData> supportedOpCodes() {
    return OpCodes.of(OpCode.LT, OpCode.GT, OpCode.SLT, OpCode.SGT, OpCode.EQ, OpCode.ISZERO);
  }

  @Override
  public void trace(final MessageFrame frame) {
    final OpCodeData opCode = OpCodes.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 =
        (opCode.mnemonic() != OpCode.ISZERO)
            ? Bytes32.wrap(frame.getStackItem(1))
            : Bytes32.repeat((byte) 0x00);

    final WcpData data = new WcpData(opCode, arg1, arg2);

    stamp++;

    this.traceWcpData(data);
  }

  public void traceWcpData(WcpData data) {
    for (int i = 0; i < maxCt(data.isOneLineInstruction()); i++) {
      builder
          .wordComparisonStamp(BigInteger.valueOf(stamp))
          .oneLineInstruction(data.isOneLineInstruction())
          .counter(BigInteger.valueOf(i))
          .inst(BigInteger.valueOf(data.getOpCode().getData().value()))
          .argument1Hi(data.getArg1Hi().toUnsignedBigInteger())
          .argument1Lo(data.getArg1Lo().toUnsignedBigInteger())
          .argument2Hi(data.getArg2Hi().toUnsignedBigInteger())
          .argument2Lo(data.getArg2Lo().toUnsignedBigInteger())
          .resultHi(data.getResHi() ? BigInteger.ONE : BigInteger.ZERO)
          .resultLo(data.getResLo() ? BigInteger.ONE : BigInteger.ZERO)
          .bits(data.getBits().get(i))
          .neg1(data.getNeg1())
          .neg2(data.getNeg2())
          .byte1(UnsignedByte.of(data.getArg1Hi().get(i)))
          .byte2(UnsignedByte.of(data.getArg1Lo().get(i)))
          .byte3(UnsignedByte.of(data.getArg2Hi().get(i)))
          .byte4(UnsignedByte.of(data.getArg2Lo().get(i)))
          .byte5(UnsignedByte.of(data.getAdjHi().get(i)))
          .byte6(UnsignedByte.of(data.getAdjLo().get(i)))
          .acc1(data.getArg1Hi().slice(0, 1 + i).toUnsignedBigInteger())
          .acc2(data.getArg1Lo().slice(0, 1 + i).toUnsignedBigInteger())
          .acc3(data.getArg2Hi().slice(0, 1 + i).toUnsignedBigInteger())
          .acc4(data.getArg2Lo().slice(0, 1 + i).toUnsignedBigInteger())
          .acc5(data.getAdjHi().slice(0, 1 + i).toUnsignedBigInteger())
          .acc6(data.getAdjLo().slice(0, 1 + i).toUnsignedBigInteger())
          .bit1(data.getBit1())
          .bit2(data.getBit2())
          .bit3(data.getBit3())
          .bit4(data.getBit4())
          .validateRow();
    }
  }

  @Override
  public Object commit() {
    return new WcpTrace(builder.build());
  }

  private int maxCt(final boolean isOneLineInstruction) {
    return isOneLineInstruction ? 1 : LIMB_SIZE;
  }

  public void callLT(Bytes32 arg1, Bytes32 arg2) {
    WcpData data = new WcpData(OpCode.LT, arg1, arg2);
    this.traceWcpData(data);
  }

  public void callEQ(Bytes32 arg1, Bytes32 arg2) {
    WcpData data = new WcpData(OpCode.EQ, arg1, arg2);
    this.traceWcpData(data);
  }

  public void callISZERO(Bytes32 arg1) {
    Bytes32 zero = Bytes32.repeat((byte) 0x00);
    WcpData data = new WcpData(OpCode.ISZERO, arg1, zero);
    this.traceWcpData(data);
  }
}
