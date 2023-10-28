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
import java.util.Objects;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.bytes.Bytes16;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;

@Accessors(fluent = true)
final class ShfOperation {
  private static final int LIMB_SIZE = 16;

  @Getter private final OpCode opCode;
  @Getter private final Bytes32 arg1;
  @Getter private final Bytes32 arg2;
  @Getter private final boolean isOneLineInstruction;
  @Getter private final boolean isNegative;
  @Getter private final boolean isShiftRight;
  @Getter private final boolean isKnown;
  @Getter private final UnsignedByte msb;
  @Getter private final UnsignedByte lsb;
  @Getter private final UnsignedByte low3;
  @Getter private final UnsignedByte mshp;
  @Getter private final Boolean[] lsbBits;
  @Getter private final Boolean[] msbBits;
  @Getter private final List<Boolean> bits;
  @Getter private final Shb shb;
  @Getter private final Res res;
  @Getter private final boolean isBitB3;
  @Getter private final boolean isBitB4;
  @Getter private final boolean isBitB5;
  @Getter private final boolean isBitB6;
  @Getter private final boolean isBitB7;

  public ShfOperation(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    this.opCode = opCode;
    this.arg1 = arg1;
    this.arg2 = arg2;

    this.isOneLineInstruction = isOneLineInstruction(opCode, arg1Hi());
    this.isNegative = Long.compareUnsigned(arg2Hi().get(0), 128) >= 0;
    this.isShiftRight = List.of(OpCode.SAR, OpCode.SHR).contains(opCode);
    this.isKnown = isKnown(opCode, arg1Hi(), arg1Lo());

    this.msb = UnsignedByte.of(arg2Hi().get(0));
    this.lsb = UnsignedByte.of(arg1Lo().get(15));
    this.low3 = lsb.shiftLeft(5).shiftRight(5);

    if (isShiftRight) {
      this.mshp = low3;
    } else {
      this.mshp = UnsignedByte.of(8 - low3.toInteger());
    }

    this.lsbBits = byteBits(lsb);
    this.msbBits = byteBits(msb);

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

  @Override
  public int hashCode() {
    return Objects.hash(this.opCode, this.arg1, this.arg2);
  }

  public int maxCt() {
    return this.isOneLineInstruction ? 1 : LIMB_SIZE;
  }
}
