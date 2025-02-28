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

package net.consensys.linea.zktracer.module.bin;

import static net.consensys.linea.zktracer.Trace.LLARGE;
import static net.consensys.linea.zktracer.Trace.LLARGEMO;
import static net.consensys.linea.zktracer.types.Utils.bitDecomposition;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

@Getter
@Accessors(fluent = true)
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
@RequiredArgsConstructor
public class BinOperation extends ModuleOperation {
  @EqualsAndHashCode.Include private final OpCode opCode;
  @EqualsAndHashCode.Include private final BaseBytes arg1;
  @EqualsAndHashCode.Include private final BaseBytes arg2;

  private int ctMax;
  private List<Boolean> lastEightBits = List.of(false);
  private boolean bit4 = false;
  private int low4 = 0;
  private boolean isSmall = false;
  private int pivotThreshold = 0;
  private int pivot = 0;

  @Override
  protected int computeLineCount() {
    ctMax = computeCtMax();
    return ctMax + 1;
  }

  private int computeCtMax() {
    return switch (opCode) {
      case NOT -> LLARGEMO;
      case BYTE, SIGNEXTEND -> arg1.getHigh().isZero() ? LLARGEMO : 0;
      case AND, OR, XOR -> Math.max(
          0,
          Math.max(
                  Math.max(
                      arg1.getHigh().trimLeadingZeros().size(),
                      arg2.getHigh().trimLeadingZeros().size()),
                  Math.max(
                      arg1.getLow().trimLeadingZeros().size(),
                      arg2.getLow().trimLeadingZeros().size()))
              - 1);
      default -> throw new IllegalStateException("BIN doesn't support OpCode" + opCode);
    };
  }

  private boolean isSmall() {
    return arg1.getBytes32().trimLeadingZeros().bitLength() < 6;
  }

  private int getPivotThreshold() {
    return switch (opCode) {
      case AND, OR, XOR, NOT -> 16;
      case BYTE -> low4;
      case SIGNEXTEND -> 15 - low4;
      default -> throw new IllegalStateException("BIN doesn't support OpCode" + opCode);
    };
  }

  private BaseBytes getResult() {
    return switch (opCode) {
      case AND -> arg1.and(arg2);
      case OR -> arg1.or(arg2);
      case XOR -> BaseBytes.fromBytes32(arg1.getBytes32().xor(arg2.getBytes32()));
      case NOT -> arg1.not();
      case BYTE -> byteResult();
      case SIGNEXTEND -> signExtensionResult();
      default -> throw new IllegalStateException("Bin doesn't support OpCode" + opCode);
    };
  }

  private BaseBytes signExtensionResult() {
    if (!isSmall) {
      return arg2;
    }
    final int indexLeadingByte = 31 - arg1.getByte(31) & 0xff;
    final byte toSet = (byte) (arg2().getByte(indexLeadingByte) < 0 ? 0xff : 0x00);
    return BaseBytes.fromBytes32(
        Bytes32.leftPad(arg2.getBytes32().slice(indexLeadingByte, 32 - indexLeadingByte), toSet));
  }

  private BaseBytes byteResult() {
    final int result = isSmall ? pivot : 0;
    return BaseBytes.fromBytes32(Bytes32.leftPad(Bytes.ofUnsignedShort(result)));
  }

  private List<Boolean> getLastEightBits() {
    final int leastByteOfArg1 = arg1().getByte(31) & 0xff;
    return bitDecomposition(leastByteOfArg1, 8).bitDecList();
  }

  private boolean getBit4() {
    return getLastEightBits().get(3);
  }

  private int getLow4() {
    int r = 0;
    for (int k = 0; k < 4; k++) {
      if (lastEightBits.get(7 - k)) {
        r += (int) Math.pow(2, k);
      }
    }
    return r;
  }

  private List<Boolean> getBit1() {
    return plateau(pivotThreshold);
  }

  private List<Boolean> plateau(final int threshold) {
    ArrayList<Boolean> output = new ArrayList<>(16);
    for (int ct = 0; ct < 16; ct++) {
      output.add(ct, ct >= threshold);
    }
    return output;
  }

  private int getPivot() {
    switch (opCode) {
      case AND, OR, XOR, NOT -> {
        return 0;
      }
      case BYTE -> {
        if (low4 == 0) {
          return !bit4 ? arg2.getHigh().get(0) & 0xff : arg2.getLow().get(0) & 0xff;
        } else {
          return !bit4
              ? arg2.getHigh().get(pivotThreshold) & 0xff
              : arg2.getLow().get(pivotThreshold) & 0xff;
        }
      }
      case SIGNEXTEND -> {
        if (low4 == 15) {
          return !bit4 ? arg2.getLow().get(0) & 0xff : arg2.getHigh().get(0) & 0xff;
        } else {
          return !bit4
              ? arg2.getLow().get(pivotThreshold) & 0xff
              : arg2.getHigh().get(pivotThreshold) & 0xff;
        }
      }
      default -> throw new IllegalStateException("Bin doesn't support OpCode" + opCode);
    }
  }

  private List<Boolean> getFirstEightBits() {
    return bitDecomposition(pivot, 8).bitDecList();
  }

  private void compute() {
    this.lastEightBits = getLastEightBits();
    this.bit4 = getBit4();
    this.low4 = getLow4();
    this.isSmall = isSmall();
    this.pivotThreshold = getPivotThreshold();
    this.pivot = getPivot();
  }

  private boolean isTrivialOperation() {
    return (opCode == OpCode.BYTE || opCode == OpCode.SIGNEXTEND) && ctMax == 0;
  }

  public void traceBinOperation(int stamp, Trace.Bin trace) {
    this.compute();

    final int length = isTrivialOperation() ? LLARGE : ctMax + 1;
    final int offset = LLARGE - length;

    final Bytes arg1Hi = this.arg1.getHigh().slice(offset, length);
    final Bytes arg1Lo = this.arg1.getLow().slice(offset, length);
    final Bytes arg2Hi = this.arg2.getHigh().slice(offset, length);
    final Bytes arg2Lo = this.arg2.getLow().slice(offset, length);
    final Bytes resHi = this.getResult().getHigh().slice(offset, length);
    final Bytes resLo = this.getResult().getLow().slice(offset, length);
    final List<Boolean> bit1 = this.getBit1();
    final List<Boolean> bits =
        Stream.concat(this.getFirstEightBits().stream(), this.lastEightBits.stream()).toList();
    for (int ct = 0; ct <= this.ctMax; ct++) {
      trace
          .stamp(stamp)
          .ctMax(UnsignedByte.of(ctMax))
          .counter(UnsignedByte.of(ct))
          .inst(UnsignedByte.of(this.opCode().byteValue()))
          .argument1Hi(arg1Hi)
          .argument1Lo(arg1Lo)
          .argument2Hi(arg2Hi)
          .argument2Lo(arg2Lo)
          .resultHi(resHi)
          .resultLo(resLo)
          .isAnd(this.opCode() == OpCode.AND)
          .isOr(this.opCode() == OpCode.OR)
          .isXor(this.opCode() == OpCode.XOR)
          .isNot(this.opCode() == OpCode.NOT)
          .isByte(this.opCode() == OpCode.BYTE)
          .isSignextend(this.opCode() == OpCode.SIGNEXTEND)
          .small(this.isSmall)
          .bits(bits.get(ct))
          .bitB4(this.bit4)
          .low4(UnsignedByte.of(this.low4))
          .neg(bits.get(0))
          .bit1(bit1.get(ct))
          .pivot(UnsignedByte.of(this.pivot))
          .byte1(UnsignedByte.of(arg1Hi.get(ct)))
          .byte2(UnsignedByte.of(arg1Lo.get(ct)))
          .byte3(UnsignedByte.of(arg2Hi.get(ct)))
          .byte4(UnsignedByte.of(arg2Lo.get(ct)))
          .byte5(UnsignedByte.of(resHi.get(ct)))
          .byte6(UnsignedByte.of(resLo.get(ct)))
          .acc1(arg1Hi.slice(0, ct + 1))
          .acc2(arg1Lo.slice(0, ct + 1))
          .acc3(arg2Hi.slice(0, ct + 1))
          .acc4(arg2Lo.slice(0, ct + 1))
          .acc5(resHi.slice(0, ct + 1))
          .acc6(resLo.slice(0, ct + 1))
          .xxxByteHi(UnsignedByte.of(resHi.get(ct)))
          .xxxByteLo(UnsignedByte.of(resLo.get(ct)))
          .validateRow();
    }
  }
}
