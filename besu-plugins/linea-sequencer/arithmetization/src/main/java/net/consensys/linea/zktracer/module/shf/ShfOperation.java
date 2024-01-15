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

import static net.consensys.linea.zktracer.module.Util.byteBits;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

import lombok.EqualsAndHashCode;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.Bytes16;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
final class ShfOperation extends ModuleOperation {
  private static final int LIMB_SIZE = 16;

  @EqualsAndHashCode.Include private final OpCode opCode;
  @EqualsAndHashCode.Include private final Bytes32 arg1;
  @EqualsAndHashCode.Include private final Bytes32 arg2;
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
    for (int i = 0; i < 15; i++) {
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
    return this.isOneLineInstruction ? 1 : LIMB_SIZE;
  }

  public void trace(Trace trace, int stamp) {
    this.compute();

    for (int i = 0; i < this.maxCt(); i++) {
      final ByteChunks arg2HiByteChunks =
          ByteChunks.fromBytes(UnsignedByte.of(this.arg2Hi().get(i)), this.mshp);
      final ByteChunks arg2LoByteChunks =
          ByteChunks.fromBytes(UnsignedByte.of(this.arg2Lo().get(i)), this.mshp);

      trace
          .acc1(this.arg1Lo().slice(0, 1 + i))
          .acc2(this.arg2Hi().slice(0, 1 + i))
          .acc3(this.arg2Lo().slice(0, 1 + i))
          .acc4(this.res.getResHi().slice(0, 1 + i))
          .acc5(this.res.getResLo().slice(0, 1 + i))
          .arg1Hi(this.arg1Hi())
          .arg1Lo(this.arg1Lo())
          .arg2Hi(this.arg2Hi())
          .arg2Lo(this.arg2Lo());

      if (this.isShiftRight) {
        trace.bit1(i >= 1).bit2(i >= 2).bit3(i >= 4).bit4(i >= 8);
      } else {
        trace.bit1(i >= (16 - 1)).bit2(i >= (16 - 2)).bit3(i >= (16 - 4)).bit4(i >= (16 - 8));
      }

      trace
          .bitB3(this.isBitB3)
          .bitB4(this.isBitB4)
          .bitB5(this.isBitB5)
          .bitB6(this.isBitB6)
          .bitB7(this.isBitB7)
          .byte1(UnsignedByte.of(this.arg1Lo().get(i)))
          .byte2(UnsignedByte.of(this.arg2Hi().get(i)))
          .byte3(UnsignedByte.of(this.arg2Lo().get(i)))
          .byte4(UnsignedByte.of(this.res.getResHi().get(i)))
          .byte5(UnsignedByte.of(this.res.getResLo().get(i)))
          .bits(this.bits.get(i))
          .counter(Bytes.of(i))
          .inst(Bytes.of(this.opCode.byteValue()))
          .known(this.isKnown)
          .neg(this.isNegative)
          .oneLineInstruction(this.isOneLineInstruction)
          .low3(Bytes.of(this.low3.toInteger()))
          .microShiftParameter(Bytes.ofUnsignedInt(this.mshp.toInteger()))
          .resHi(this.res.getResHi())
          .resLo(this.res.getResLo())
          .leftAlignedSuffixHigh(Bytes.ofUnsignedShort(arg2HiByteChunks.la().toInteger()))
          .rightAlignedPrefixHigh(Bytes.ofUnsignedInt(arg2HiByteChunks.ra().toInteger()))
          .ones(Bytes.ofUnsignedInt(arg2HiByteChunks.ones().toInteger()))
          .leftAlignedSuffixLow(Bytes.ofUnsignedInt(arg2LoByteChunks.la().toInteger()))
          .rightAlignedPrefixLow(Bytes.ofUnsignedInt(arg2LoByteChunks.ra().toInteger()))
          .shb3Hi(Bytes.ofUnsignedInt(this.shb.getShbHi()[0][i].toInteger()))
          .shb3Lo(Bytes.ofUnsignedInt(this.shb.getShbLo()[0][i].toInteger()))
          .shb4Hi(Bytes.ofUnsignedInt(this.shb.getShbHi()[4 - 3][i].toInteger()))
          .shb4Lo(Bytes.ofUnsignedInt(this.shb.getShbLo()[4 - 3][i].toInteger()))
          .shb5Hi(Bytes.ofUnsignedInt(this.shb.getShbHi()[5 - 3][i].toInteger()))
          .shb5Lo(Bytes.ofUnsignedInt(this.shb.getShbLo()[5 - 3][i].toInteger()))
          .shb6Hi(Bytes.ofUnsignedInt(this.shb.getShbHi()[6 - 3][i].toInteger()))
          .shb6Lo(Bytes.ofUnsignedInt(this.shb.getShbLo()[6 - 3][i].toInteger()))
          .shb7Hi(Bytes.ofUnsignedInt(this.shb.getShbHi()[7 - 3][i].toInteger()))
          .shb7Lo(Bytes.ofUnsignedInt(this.shb.getShbLo()[7 - 3][i].toInteger()))
          .shiftDirection(this.isShiftRight)
          .isData(stamp != 0)
          .shiftStamp(Bytes.ofUnsignedInt(stamp))
          .validateRow();
    }
  }

  @Override
  protected int computeLineCount() {
    return this.maxCt();
  }
}
