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

package net.consensys.linea.zktracer.module.shf;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.Bytes16;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Module;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class Shf implements Module {
  private static final int LIMB_SIZE = 16;

  final Trace.TraceBuilder builder = Trace.builder();
  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "shf";
  }

  @Override
  public final List<OpCode> supportedOpCodes() {
    return List.of(OpCode.SHR, OpCode.SHL, OpCode.SAR);
  }

  @Override
  public void trace(MessageFrame frame) {
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

    final Bytes16 arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
    final Bytes16 arg1Lo = Bytes16.wrap(arg1.slice(16));
    final Bytes16 arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
    final Bytes16 arg2Lo = Bytes16.wrap(arg2.slice(16));

    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    final boolean isOneLineInstruction = isOneLineInstruction(opCode, arg1Hi);
    final boolean isNegative = Long.compareUnsigned(arg2Hi.get(0), 128) >= 0;
    final boolean isShiftRight = opCode.isElementOf(OpCode.SAR, OpCode.SHR);
    final boolean isKnown = isKnown(opCode, arg1Hi, arg1Lo);

    final UnsignedByte msb = UnsignedByte.of(arg2Hi.get(0));
    final UnsignedByte lsb = UnsignedByte.of(arg1Lo.get(15));
    final UnsignedByte low3 = lsb.shiftLeft(5).shiftRight(5);

    final UnsignedByte mshp;

    if (isShiftRight) {
      mshp = low3;
    } else {
      mshp = UnsignedByte.of(8 - low3.toInteger());
    }

    final Boolean[] lsbBits = byteBits(lsb);
    final Boolean[] msbBits = byteBits(msb);

    final List<Boolean> bits = new ArrayList<>(lsbBits.length + msbBits.length);
    Collections.addAll(bits, msbBits);
    Collections.addAll(bits, lsbBits);

    final Shb shb = Shb.create(opCode, arg2, lsb);
    final Res res = Res.create(opCode, arg1, arg2);

    final boolean isBitB3 = lsbBits[4];
    final boolean isBitB4 = lsbBits[3];
    final boolean isBitB5 = lsbBits[2];
    final boolean isBitB6 = lsbBits[1];
    final boolean isBitB7 = lsbBits[0];

    stamp++;
    for (int i = 0; i < maxCt(isOneLineInstruction); i++) {
      builder
          .acc1Arg(arg1Lo.slice(0, 1 + i).toUnsignedBigInteger())
          .acc2Arg(arg2Hi.slice(0, 1 + i).toUnsignedBigInteger())
          .acc3Arg(arg2Lo.slice(0, 1 + i).toUnsignedBigInteger())
          .acc4Arg(res.getResHi().slice(0, 1 + i).toUnsignedBigInteger())
          .acc5Arg(res.getResLo().slice(0, 1 + i).toUnsignedBigInteger())
          .arg1HiArg(arg1Hi.toUnsignedBigInteger())
          .arg1LoArg(arg1Lo.toUnsignedBigInteger())
          .arg2HiArg(arg2Hi.toUnsignedBigInteger())
          .arg2LoArg(arg2Lo.toUnsignedBigInteger());

      if (isShiftRight) {
        builder.bit1Arg(i >= 1).bit2Arg(i >= 2).bit3Arg(i >= 4).bit4Arg(i >= 8);
      } else {
        builder
            .bit1Arg(i >= (16 - 1))
            .bit2Arg(i >= (16 - 2))
            .bit3Arg(i >= (16 - 4))
            .bit4Arg(i >= (16 - 8));
      }

      builder
          .bitB3Arg(isBitB3)
          .bitB4Arg(isBitB4)
          .bitB5Arg(isBitB5)
          .bitB6Arg(isBitB6)
          .bitB7Arg(isBitB7);

      builder
          .byte1Arg(UnsignedByte.of(arg1Lo.get(i)))
          .byte2Arg(UnsignedByte.of(arg2Hi.get(i)))
          .byte3Arg(UnsignedByte.of(arg2Lo.get(i)))
          .byte4Arg(UnsignedByte.of(res.getResHi().get(i)))
          .byte5Arg(UnsignedByte.of(res.getResLo().get(i)));

      builder
          .bitsArg(bits.get(i))
          .counterArg(i)
          .instArg(UnsignedByte.of(opCode.value))
          .knownArg(isKnown)
          .negArg(isNegative)
          .oneLineInstructionArg(isOneLineInstruction)
          .low3Arg(low3)
          .microShiftParameterArg(mshp)
          .resHiArg(res.getResHi().toUnsignedBigInteger())
          .resLoArg(res.getResLo().toUnsignedBigInteger());

      final ByteChunks arg2HiByteChunks =
          ByteChunks.fromBytes(UnsignedByte.of(arg2Hi.get(i)), mshp);

      builder
          .leftAlignedSuffixHighArg(arg2HiByteChunks.la())
          .rightAlignedPrefixHighArg(arg2HiByteChunks.ra())
          .onesArg(arg2HiByteChunks.ones());

      final ByteChunks arg2LoByteChunks =
          ByteChunks.fromBytes(UnsignedByte.of(arg2Lo.get(i)), mshp);

      builder
          .leftAlignedSuffixLowArg(arg2LoByteChunks.la())
          .rightAlignedPrefixLowArg(arg2LoByteChunks.ra())
          .shb3HiArg(shb.getShbHi()[0][i])
          .shb3LoArg(shb.getShbLo()[0][i])
          .shb4HiArg(shb.getShbHi()[4 - 3][i])
          .shb4LoArg(shb.getShbLo()[4 - 3][i])
          .shb5HiArg(shb.getShbHi()[5 - 3][i])
          .shb5LoArg(shb.getShbLo()[5 - 3][i])
          .shb6HiArg(shb.getShbHi()[6 - 3][i])
          .shb6LoArg(shb.getShbLo()[6 - 3][i])
          .shb7HiArg(shb.getShbHi()[7 - 3][i])
          .shb7LoArg(shb.getShbLo()[7 - 3][i])
          .shiftDirectionArg(isShiftRight)
          .isDataArg(stamp != 0)
          .shiftStampArg(stamp);
    }
  }

  @Override
  public Object commit() {
    return new ShfTrace(builder.build(), stamp);
  }

  private int maxCt(final boolean isOneLineInstruction) {
    return isOneLineInstruction ? 1 : LIMB_SIZE;
  }

  private Boolean[] byteBits(final UnsignedByte b) {
    final Boolean[] bits = new Boolean[8];

    for (int i = 0; i < 8; i++) {
      bits[7 - i] = b.shiftRight(i).mod(2).toInteger() == 1;
    }

    return bits;
  }

  private boolean isOneLineInstruction(final OpCode opCode, final Bytes16 arg1Hi) {
    return opCode.isElementOf(OpCode.SHR, OpCode.SHL) && !arg1Hi.isZero();
  }

  private boolean isKnown(final OpCode opCode, final Bytes16 arg1Hi, final Bytes16 arg1Lo) {
    if (opCode.equals(OpCode.SAR) && !arg1Hi.isZero()) {
      return true;
    }

    return !allButLastByteZero(arg1Lo);
  }

  private static boolean allButLastByteZero(final Bytes16 bytes) {
    for (int i = 0; i < 15; i++) {
      if (bytes.get(i) != 0) {
        return false;
      }
    }

    return true;
  }
}
