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
package net.consensys.linea.zktracer.module.alu.mul;

import org.hyperledger.besu.evm.frame.MessageFrame;

import java.math.BigInteger;
import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.Bytes16;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

public class MulTracer implements ModuleTracer {
  private static final int MMEDIUM = 8;

  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "mul";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of(OpCode.MUL, OpCode.EXP);
  }

  @SuppressWarnings("UnusedVariable")
  @Override
  public Object trace(MessageFrame frame) {
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

    final Bytes16 arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
    final Bytes16 arg1Lo = Bytes16.wrap(arg1.slice(16));
    final Bytes16 arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
    final Bytes16 arg2Lo = Bytes16.wrap(arg2.slice(16));

    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    final MulData data = new MulData(opCode, arg1, arg2);
    final MulTrace.Trace.Builder builder = MulTrace.Trace.Builder.newInstance();
    final boolean isOneLineInstruction = data.isOneLineInstruction();

    stamp++;
    for (int i = 0; i < maxCt(isOneLineInstruction); i++) {
      builder.appendStamp(stamp);
      builder.appendCounter(i);

      builder
          .appendOneLineInstruction(isOneLineInstruction)
          .appendTinyBase(data.tinyBase)
          .appendTinyExponent(data.tinyExponent)
          .appendResultVanishes(data.res.isZero());

      builder
          .appendInst(UnsignedByte.of(opCode.value))
          .appendArg1Hi(arg1Hi.toUnsignedBigInteger())
          .appendArg1Lo(arg1Lo.toUnsignedBigInteger())
          .appendArg2Hi(arg2Hi.toUnsignedBigInteger())
          .appendArg2Lo(arg2Lo.toUnsignedBigInteger());

      builder
          .appendResHi(data.res.getResHi().toUnsignedBigInteger())
          .appendResLo(data.res.getResLo().toUnsignedBigInteger());

      //      builder.appendBits(bits.get(i)).appendCounter(i); // TODO

      builder
          .appendByteA3(UnsignedByte.of(data.aBytes.get(3, i)))
          .appendByteA2(UnsignedByte.of(data.aBytes.get(2, i)))
          .appendByteA1(UnsignedByte.of(data.aBytes.get(1, i)))
          .appendByteA0(UnsignedByte.of(data.aBytes.get(0, i)));
      builder
          .appendAccA3(Bytes.of(data.aBytes.getRange(3, 0, i + 1)).toUnsignedBigInteger())
          .appendAccA2(Bytes.of(data.aBytes.getRange(2, 0, i + 1)).toUnsignedBigInteger())
          .appendAccA1(Bytes.of(data.aBytes.getRange(1, 0, i + 1)).toUnsignedBigInteger())
          .appendAccA0(Bytes.of(data.aBytes.getRange(0, 0, i + 1)).toUnsignedBigInteger());

      builder
          .appendByteB3(UnsignedByte.of(data.bBytes.get(3, i)))
          .appendByteB2(UnsignedByte.of(data.bBytes.get(2, i)))
          .appendByteB1(UnsignedByte.of(data.bBytes.get(1, i)))
          .appendByteB0(UnsignedByte.of(data.bBytes.get(0, i)));
      builder
          .appendAccB3(Bytes.of(data.bBytes.getRange(3, 0, i + 1)).toUnsignedBigInteger())
          .appendAccB2(Bytes.of(data.bBytes.getRange(2, 0, i + 1)).toUnsignedBigInteger())
          .appendAccB1(Bytes.of(data.bBytes.getRange(1, 0, i + 1)).toUnsignedBigInteger())
          .appendAccB0(Bytes.of(data.bBytes.getRange(0, 0, i + 1)).toUnsignedBigInteger());
      builder
          .appendByteC3(UnsignedByte.of(data.cBytes.get(3, i)))
          .appendByteC2(UnsignedByte.of(data.cBytes.get(2, i)))
          .appendByteC1(UnsignedByte.of(data.cBytes.get(1, i)))
          .appendByteC0(UnsignedByte.of(data.cBytes.get(0, i)));
    }
    builder.setStamp(stamp);

    return builder.build();
  }


  private int maxCt(final boolean isOneLineInstruction) {
    return isOneLineInstruction ? 1 : MMEDIUM;
  }
}
