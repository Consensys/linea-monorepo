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

package net.consensys.linea.zktracer.module.shf;

import static net.consensys.linea.zktracer.Trace.LLARGE;
import static net.consensys.linea.zktracer.Trace.LLARGEMO;
import static net.consensys.linea.zktracer.module.Util.byteBits;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.Bytes16;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

@Accessors(fluent = true)
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
final class ShfOperation extends ModuleOperation {

  @EqualsAndHashCode.Include @Getter private final OpCode opCode;
  @EqualsAndHashCode.Include @Getter private final Bytes32 arg1;
  @EqualsAndHashCode.Include @Getter private final Bytes32 arg2;
  private final boolean isOneLineInstruction;
  private boolean isNegative;
  private boolean isShiftRight;
  private boolean isKnown;
  private UnsignedByte low3;
  private UnsignedByte mshp;
  private List<Boolean> bits;
  private Shb shb;
  private Res res;
  private boolean isBitB3;
  private boolean isBitB4;
  private boolean isBitB5;
  private boolean isBitB6;
  private boolean isBitB7;

  public ShfOperation(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    this.opCode = opCode;
    this.arg1 = arg1;
    this.arg2 = arg2;
    this.isOneLineInstruction = isOneLineInstruction(opCode, arg1Hi());
  }

  private static boolean isOneLineInstruction(final OpCode opCode, final Bytes16 arg1Hi) {
    return (opCode == OpCode.SHR || opCode == OpCode.SHL) && !arg1Hi.isZero();
  }

  private static boolean isKnown(final OpCode opCode, final Bytes16 arg1Hi, final Bytes16 arg1Lo) {
    if (opCode.equals(OpCode.SAR) && !arg1Hi.isZero()) {
      return true;
    }

    return !allButLastByteZero(arg1Lo);
  }

  private static boolean allButLastByteZero(final Bytes16 bytes) {
    for (int i = 0; i < LLARGEMO; i++) {
      if (bytes.get(i) != 0) {
        return false;
      }
    }

    return true;
  }

  private void compute() {
    this.isNegative = Long.compareUnsigned(arg2Hi().get(0), 128) >= 0;
    this.isShiftRight = List.of(OpCode.SAR, OpCode.SHR).contains(opCode);
    this.isKnown = isKnown(opCode, arg1Hi(), arg1Lo());

    UnsignedByte msb = UnsignedByte.of(arg2Hi().get(0));
    UnsignedByte lsb = UnsignedByte.of(arg1Lo().get(15));
    this.low3 = lsb.shiftLeft(5).shiftRight(5);

    if (isShiftRight) {
      this.mshp = low3;
    } else {
      this.mshp = UnsignedByte.of(8 - low3.toInteger());
    }

    Boolean[] lsbBits = byteBits(lsb);
    Boolean[] msbBits = byteBits(msb);

    this.bits = new ArrayList<>(lsbBits.length + msbBits.length);
    Collections.addAll(this.bits, msbBits);
    Collections.addAll(this.bits, lsbBits);

    this.shb = Shb.create(opCode, arg2, lsb);
    this.res = Res.create(opCode, arg1, arg2);

    this.isBitB3 = lsbBits[4];
    this.isBitB4 = lsbBits[3];
    this.isBitB5 = lsbBits[2];
    this.isBitB6 = lsbBits[1];
    this.isBitB7 = lsbBits[0];
  }

  public Bytes16 arg1Hi() {
    return Bytes16.wrap(arg1.slice(0, 16));
  }

  public Bytes16 arg1Lo() {
    return Bytes16.wrap(arg1.slice(16));
  }

  public Bytes16 arg2Hi() {
    return Bytes16.wrap(arg2.slice(0, 16));
  }

  public Bytes16 arg2Lo() {
    return Bytes16.wrap(arg2.slice(16));
  }

  public int maxCt() {
    return this.isOneLineInstruction ? 1 : LLARGE;
  }

  public void trace(Trace.Shf trace, int stamp) {
    this.compute();

    for (int i = 0; i < this.maxCt(); i++) {
      final ByteChunks arg2HiByteChunks =
          ByteChunks.fromBytes(UnsignedByte.of(this.arg2Hi().get(i)), mshp);
      final ByteChunks arg2LoByteChunks =
          ByteChunks.fromBytes(UnsignedByte.of(this.arg2Lo().get(i)), mshp);

      trace
          .acc1(this.arg1Lo().slice(0, 1 + i))
          .acc2(this.arg2Hi().slice(0, 1 + i))
          .acc3(this.arg2Lo().slice(0, 1 + i))
          .acc4(res.getResHi().slice(0, 1 + i))
          .acc5(res.getResLo().slice(0, 1 + i))
          .arg1Hi(this.arg1Hi())
          .arg1Lo(this.arg1Lo())
          .arg2Hi(this.arg2Hi())
          .arg2Lo(this.arg2Lo());

      if (isShiftRight) {
        trace.bit1(i >= 1).bit2(i >= 2).bit3(i >= 4).bit4(i >= 8);
      } else {
        trace.bit1(i >= (16 - 1)).bit2(i >= (16 - 2)).bit3(i >= (16 - 4)).bit4(i >= (16 - 8));
      }

      trace
          .bitB3(isBitB3)
          .bitB4(isBitB4)
          .bitB5(isBitB5)
          .bitB6(isBitB6)
          .bitB7(isBitB7)
          .byte1(UnsignedByte.of(this.arg1Lo().get(i)))
          .byte2(UnsignedByte.of(this.arg2Hi().get(i)))
          .byte3(UnsignedByte.of(this.arg2Lo().get(i)))
          .byte4(UnsignedByte.of(res.getResHi().get(i)))
          .byte5(UnsignedByte.of(res.getResLo().get(i)))
          .bits(bits.get(i))
          .counter((short) i)
          .inst(UnsignedByte.of(opCode.byteValue()))
          .known(isKnown)
          .neg(isNegative)
          .oneLineInstruction(isOneLineInstruction)
          .low3(Bytes.of(low3.toInteger()))
          .microShiftParameter((short) mshp.toInteger())
          .resHi(res.getResHi())
          .resLo(res.getResLo())
          .leftAlignedSuffixHigh(arg2HiByteChunks.la())
          .rightAlignedPrefixHigh(arg2HiByteChunks.ra())
          .ones(arg2HiByteChunks.ones())
          .leftAlignedSuffixLow(arg2LoByteChunks.la())
          .rightAlignedPrefixLow(arg2LoByteChunks.ra())
          .shb3Hi(shb.getShbHi()[0][i])
          .shb3Lo(shb.getShbLo()[0][i])
          .shb4Hi(shb.getShbHi()[4 - 3][i])
          .shb4Lo(shb.getShbLo()[4 - 3][i])
          .shb5Hi(shb.getShbHi()[5 - 3][i])
          .shb5Lo(shb.getShbLo()[5 - 3][i])
          .shb6Hi(shb.getShbHi()[6 - 3][i])
          .shb6Lo(shb.getShbLo()[6 - 3][i])
          .shb7Hi(shb.getShbHi()[7 - 3][i])
          .shb7Lo(shb.getShbLo()[7 - 3][i])
          .shiftDirection(isShiftRight)
          .iomf(true)
          .shiftStamp(stamp)
          .validateRow();
    }
  }

  @Override
  protected int computeLineCount() {
    return this.maxCt();
  }
}
