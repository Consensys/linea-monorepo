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

package net.consensys.linea.zktracer.module.wcp;

import static net.consensys.linea.zktracer.module.Util.byteBits;
import static net.consensys.linea.zktracer.types.Conversions.reallyToSignedBigInteger;

import java.math.BigInteger;
import java.security.InvalidParameterException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Objects;
import java.util.Set;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.Bytes16;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

@Slf4j
public class WcpOperation {
  private static final int LIMB_SIZE = 16;

  private final OpCode opCode;
  private final Bytes32 arg1;
  private final Bytes32 arg2;

  private final boolean isOneLineInstruction;

  private Bytes16 arg1Hi;
  private Bytes16 arg1Lo;
  private Bytes16 arg2Hi;
  private Bytes16 arg2Lo;

  private Bytes16 adjHi;
  private Bytes16 adjLo;
  private Boolean neg1;

  private Boolean neg2;
  private Boolean bit1 = true;
  private Boolean bit2 = true;
  private Boolean bit3;
  private Boolean bit4;
  private Boolean resLo;

  final List<Boolean> bits = new ArrayList<>(16);

  public WcpOperation(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    this.opCode = opCode;
    this.arg1 = arg1;
    this.arg2 = arg2;

    this.isOneLineInstruction = isOneLineInstruction(opCode);
  }

  private void compute() {
    this.arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
    this.arg1Lo = Bytes16.wrap(arg1.slice(16));
    this.arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
    this.arg2Lo = Bytes16.wrap(arg2.slice(16));

    // Calculate Result Low
    resLo = calculateResLow(opCode, arg1, arg2);

    // Initiate negatives
    UnsignedByte msb1 = UnsignedByte.of(this.arg1Hi.get(0));
    UnsignedByte msb2 = UnsignedByte.of(this.arg2Hi.get(0));
    Boolean[] msb1Bits = byteBits(msb1);
    Boolean[] msb2Bits = byteBits(msb2);
    this.neg1 = msb1Bits[0];
    this.neg2 = msb2Bits[0];

    // Initiate bits
    Collections.addAll(bits, msb1Bits);
    Collections.addAll(bits, msb2Bits);

    // Set bit 1 and 2
    for (int i = 0; i < 16; i++) {
      if (arg1Hi.get(i) != arg2Hi.get(i)) {
        bit1 = false;
      }
      if (arg1Lo.get(i) != arg2Lo.get(i)) {
        bit2 = false;
      }
    }

    // Set bit 3 and AdjHi
    final BigInteger firstHi = arg1Hi.toUnsignedBigInteger();
    final BigInteger secondHi = arg2Hi.toUnsignedBigInteger();
    bit3 = firstHi.compareTo(secondHi) > 0;
    this.adjHi = calculateAdj(bit3, firstHi, secondHi);

    // Set bit 4 and AdjLo
    final BigInteger firstLo = arg1Lo.toUnsignedBigInteger();
    final BigInteger secondLo = arg2Lo.toUnsignedBigInteger();
    bit4 = firstLo.compareTo(secondLo) > 0;
    this.adjLo = calculateAdj(bit4, firstLo, secondLo);
  }

  @Override
  public int hashCode() {
    return Objects.hash(this.opCode, this.arg1, this.arg2);
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (o == null || getClass() != o.getClass()) return false;
    final WcpOperation that = (WcpOperation) o;
    return Objects.equals(opCode, that.opCode)
        && Objects.equals(arg1, that.arg1)
        && Objects.equals(arg2, that.arg2);
  }

  public Boolean getResHi() {
    return false;
  }

  private boolean isOneLineInstruction(final OpCode opCode) {
    return Set.of(OpCode.EQ, OpCode.ISZERO).contains(opCode);
  }

  private boolean calculateResLow(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    return switch (opCode) {
      case LT -> arg1.compareTo(arg2) < 0;
      case GT -> arg1.compareTo(arg2) > 0;
      case SLT -> reallyToSignedBigInteger(arg1).compareTo(reallyToSignedBigInteger(arg2)) < 0;
      case SGT -> reallyToSignedBigInteger(arg1).compareTo(reallyToSignedBigInteger(arg2)) > 0;
      case EQ -> arg1.compareTo(arg2) == 0;
      case ISZERO -> arg1.isZero();
      default -> throw new InvalidParameterException("Invalid opcode");
    };
  }

  private Bytes16 calculateAdj(boolean cmp, BigInteger arg1, BigInteger arg2) {
    BigInteger adjHi;
    if (cmp) {
      adjHi = arg1.subtract(arg2).subtract(BigInteger.ONE);
    } else {
      adjHi = arg2.subtract(arg1);
    }
    var bytes32 = Bytes32.leftPad(Bytes.of(adjHi.toByteArray()));

    return Bytes16.wrap(bytes32.slice(16));
  }

  void trace(Trace trace, int stamp) {
    this.compute();

    final Bytes resHi = this.getResHi() ? Bytes.of(1) : Bytes.EMPTY;
    final Bytes resLo = this.resLo ? Bytes.of(1) : Bytes.EMPTY;

    for (int i = 0; i < this.maxCt(); i++) {
      trace
          .wordComparisonStamp(Bytes.ofUnsignedInt(stamp))
          .oneLineInstruction(this.isOneLineInstruction)
          .counter(Bytes.of(i))
          .inst(Bytes.of(this.opCode.byteValue()))
          .argument1Hi(this.arg1Hi)
          .argument1Lo(this.arg1Lo)
          .argument2Hi(this.arg2Hi)
          .argument2Lo(this.arg2Lo)
          .resultHi(resHi)
          .resultLo(resLo)
          .bits(bits.get(i))
          .neg1(neg1)
          .neg2(neg2)
          .byte1(UnsignedByte.of(this.arg1Hi.get(i)))
          .byte2(UnsignedByte.of(this.arg1Lo.get(i)))
          .byte3(UnsignedByte.of(this.arg2Hi.get(i)))
          .byte4(UnsignedByte.of(this.arg2Lo.get(i)))
          .byte5(UnsignedByte.of(adjHi.get(i)))
          .byte6(UnsignedByte.of(adjLo.get(i)))
          .acc1(this.arg1Hi.slice(0, 1 + i))
          .acc2(this.arg1Lo.slice(0, 1 + i))
          .acc3(this.arg2Hi.slice(0, 1 + i))
          .acc4(this.arg2Lo.slice(0, 1 + i))
          .acc5(adjHi.slice(0, 1 + i))
          .acc6(adjLo.slice(0, 1 + i))
          .bit1(bit1)
          .bit2(bit2)
          .bit3(bit3)
          .bit4(bit4)
          .validateRow();
    }
  }

  int maxCt() {
    return this.isOneLineInstruction ? 1 : LIMB_SIZE;
  }
}
