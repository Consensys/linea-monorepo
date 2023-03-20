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
package net.consensys.zktracer.module.shf;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.bytes.Bytes16;
import net.consensys.zktracer.bytes.UnsignedByte;
import net.consensys.zktracer.module.ModuleTracer;
import net.consensys.zktracer.module.shf.ByteChunks;
import net.consensys.zktracer.module.shf.Res;
import net.consensys.zktracer.module.shf.Shb;
import net.consensys.zktracer.module.shf.ShfTrace;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class ShfTracer implements ModuleTracer {
  private static final int LIMB_SIZE = 16;

  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "shf";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of(OpCode.SHR, OpCode.SHL, OpCode.SAR);
  }

  @Override
  public Object trace(MessageFrame frame) {
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

    final ShfTrace.Trace.Builder builder = ShfTrace.Trace.Builder.newInstance();

    stamp++;
    for (int i = 0; i < maxCt(isOneLineInstruction); i++) {
      builder
          .appendAcc1(arg1Lo.slice(0, 1 + i).toUnsignedBigInteger())
          .appendAcc2(arg2Hi.slice(0, 1 + i).toUnsignedBigInteger())
          .appendAcc3(arg2Lo.slice(0, 1 + i).toUnsignedBigInteger())
          .appendAcc4(res.getResHi().slice(0, 1 + i).toUnsignedBigInteger())
          .appendAcc5(res.getResLo().slice(0, 1 + i).toUnsignedBigInteger())
          .appendArg1Hi(arg1Hi.toUnsignedBigInteger())
          .appendArg1Lo(arg1Lo.toUnsignedBigInteger())
          .appendArg2Hi(arg2Hi.toUnsignedBigInteger())
          .appendArg2Lo(arg2Lo.toUnsignedBigInteger());

      if (isShiftRight) {
        builder.appendBit1(i >= 1).appendBit2(i >= 2).appendBit3(i >= 4).appendBit4(i >= 8);
      } else {
        builder
            .appendBit1(i >= (16 - 1))
            .appendBit2(i >= (16 - 2))
            .appendBit3(i >= (16 - 4))
            .appendBit4(i >= (16 - 8));
      }

      builder
          .appendBitB3(isBitB3)
          .appendBitB4(isBitB4)
          .appendBitB5(isBitB5)
          .appendBitB6(isBitB6)
          .appendBitB7(isBitB7);

      builder
          .appendByte1(UnsignedByte.of(arg1Lo.get(i)))
          .appendByte2(UnsignedByte.of(arg2Hi.get(i)))
          .appendByte3(UnsignedByte.of(arg2Lo.get(i)))
          .appendByte4(UnsignedByte.of(res.getResHi().get(i)))
          .appendByte5(UnsignedByte.of(res.getResLo().get(i)));

      builder.appendBits(bits.get(i)).appendCounter(i);

      builder
          .appendInst(UnsignedByte.of(opCode.value))
          .appendKnown(isKnown)
          .appendNeg(isNegative)
          .appendOneLineInstruction(isOneLineInstruction);

      builder.appendLow3(low3).appendMicroShiftParameter(mshp);

      builder
          .appendResHi(res.getResHi().toUnsignedBigInteger())
          .appendResLo(res.getResLo().toUnsignedBigInteger());

      final ByteChunks arg2HiByteChunks =
          ByteChunks.fromBytes(UnsignedByte.of(arg2Hi.get(i)), mshp);
      builder
          .appendLeftAlignedSuffixHigh(arg2HiByteChunks.la())
          .appendRightAlignedPrefixHigh(arg2HiByteChunks.ra())
          .appendOnes(arg2HiByteChunks.ones());

      final ByteChunks arg2LoByteChunks =
          ByteChunks.fromBytes(UnsignedByte.of(arg2Lo.get(i)), mshp);
      builder
          .appendLeftAlignedSuffixLow(arg2LoByteChunks.la())
          .appendRightAlignedPrefixLow(arg2LoByteChunks.ra());

      builder
          .appendShb3Hi(shb.getShbHi()[3 - 3][i])
          .appendShb3Lo(shb.getShbLo()[3 - 3][i])
          .appendShb4Hi(shb.getShbHi()[4 - 3][i])
          .appendShb4Lo(shb.getShbLo()[4 - 3][i])
          .appendShb5Hi(shb.getShbHi()[5 - 3][i])
          .appendShb5Lo(shb.getShbLo()[5 - 3][i])
          .appendShb6Hi(shb.getShbHi()[6 - 3][i])
          .appendShb6Lo(shb.getShbLo()[6 - 3][i])
          .appendShb7Hi(shb.getShbHi()[7 - 3][i])
          .appendShb7Lo(shb.getShbLo()[7 - 3][i]);

      builder.appendShiftDirection(isShiftRight).appendIsData(stamp != 0).appendShiftStamp(stamp);
    }
    builder.setStamp(stamp);

    return builder.build();
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
