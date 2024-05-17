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
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_EQ;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_GT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_ISZERO;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_LT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_SGT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_SLT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGEMO;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.WCP_INST_GEQ;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.WCP_INST_LEQ;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.reallyToSignedBigInteger;

import java.math.BigInteger;
import java.security.InvalidParameterException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

import lombok.EqualsAndHashCode;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.types.Bytes16;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

@Slf4j
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public class WcpOperation extends ModuleOperation {
  public static final byte LEQbv = (byte) WCP_INST_LEQ;
  public static final byte GEQbv = (byte) WCP_INST_GEQ;
  static final byte LTbv = (byte) EVM_INST_LT;
  static final byte GTbv = (byte) EVM_INST_GT;
  private static final byte SLTbv = (byte) EVM_INST_SLT;
  private static final byte SGTbv = (byte) EVM_INST_SGT;
  static final byte EQbv = (byte) EVM_INST_EQ;
  static final byte ISZERObv = (byte) EVM_INST_ISZERO;

  @EqualsAndHashCode.Include private final byte wcpInst;
  @EqualsAndHashCode.Include private final Bytes32 arg1;
  @EqualsAndHashCode.Include private final Bytes32 arg2;
  final int ctMax;

  private Bytes arg1Hi;
  private Bytes arg1Lo;
  private Bytes arg2Hi;
  private Bytes arg2Lo;

  private Bytes adjHi;
  private Bytes adjLo;
  private Boolean neg1;

  private Boolean neg2;
  private boolean bit1;
  private Boolean bit2;
  private Boolean bit3;
  private Boolean bit4;
  private Boolean resLo;

  final List<Boolean> bits = new ArrayList<>(LLARGE);

  public WcpOperation(final byte wcpInst, final Bytes32 arg1, final Bytes32 arg2) {
    this.wcpInst = wcpInst;
    this.arg1 = arg1;
    this.arg2 = arg2;

    this.ctMax = maxCt();
  }

  private void compute() {
    final int length = this.isOli() ? LLARGE : this.ctMax + 1;
    final int offset = LLARGE - length;
    this.arg1Hi = arg1.slice(offset, length);
    this.arg1Lo = arg1.slice(LLARGE + offset, length);
    this.arg2Hi = arg2.slice(offset, length);
    this.arg2Lo = arg2.slice(LLARGE + offset, length);

    // Calculate Result Low
    resLo = calculateResLow(wcpInst, arg1, arg2);

    // Set bit 3 and AdjHi
    final BigInteger firstHi = this.arg1.slice(0, LLARGE).toUnsignedBigInteger();
    final BigInteger secondHi = this.arg2.slice(0, LLARGE).toUnsignedBigInteger();
    this.bit3 = firstHi.compareTo(secondHi) > 0;
    this.adjHi = calculateAdj(bit3, firstHi, secondHi).slice(offset, length);

    // Set bit 4 and AdjLo
    final BigInteger firstLo = this.arg1.slice(LLARGE, LLARGE).toUnsignedBigInteger();
    final BigInteger secondLo = this.arg2.slice(LLARGE, LLARGE).toUnsignedBigInteger();
    this.bit4 = firstLo.compareTo(secondLo) > 0;
    this.adjLo = calculateAdj(bit4, firstLo, secondLo).slice(offset, length);

    // Initiate negatives and BITS
    if (this.ctMax == LLARGEMO && (this.wcpInst == SLTbv || this.wcpInst == SGTbv)) {
      // meaningful only for signed OpCode with LLARGE argument
      UnsignedByte msb1 = UnsignedByte.of(this.arg1Hi.get(0));
      UnsignedByte msb2 = UnsignedByte.of(this.arg2Hi.get(0));
      Boolean[] msb1Bits = byteBits(msb1);
      Boolean[] msb2Bits = byteBits(msb2);
      this.neg1 = msb1Bits[0];
      this.neg2 = msb2Bits[0];
      Collections.addAll(bits, msb1Bits);
      Collections.addAll(bits, msb2Bits);
    } else {
      this.neg1 = false;
      this.neg2 = false;
      for (int ct = 0; ct <= this.ctMax; ct++) {
        bits.add(ct, false);
      }
    }

    // Set bit 1 and 2
    this.bit1 = this.arg1Hi.compareTo(this.arg2Hi) == 0;
    this.bit2 = this.arg1Lo.compareTo(this.arg2Lo) == 0;
  }

  private boolean calculateResLow(byte opCode, Bytes32 arg1, Bytes32 arg2) {
    return switch (opCode) {
      case EQbv -> arg1.compareTo(arg2) == 0;
      case ISZERObv -> arg1.isZero();
      case SLTbv -> reallyToSignedBigInteger(arg1).compareTo(reallyToSignedBigInteger(arg2)) < 0;
      case SGTbv -> reallyToSignedBigInteger(arg1).compareTo(reallyToSignedBigInteger(arg2)) > 0;
      case LTbv -> arg1.compareTo(arg2) < 0;
      case GTbv -> arg1.compareTo(arg2) > 0;
      case LEQbv -> arg1.compareTo(arg2) <= 0;
      case GEQbv -> arg1.compareTo(arg2) >= 0;
      default -> throw new InvalidParameterException("Invalid opcode");
    };
  }

  private Bytes16 calculateAdj(boolean cmp, BigInteger arg1, BigInteger arg2) {
    return cmp
        ? Bytes16.leftPad(bigIntegerToBytes(arg1.subtract(arg2).subtract(BigInteger.ONE)))
        : Bytes16.leftPad(bigIntegerToBytes(arg2.subtract(arg1)));
  }

  void trace(Trace trace, int stamp) {
    this.compute();

    final boolean resLo = this.resLo;
    final boolean oli = isOli();
    final boolean vli = isVli();
    final UnsignedByte inst = UnsignedByte.of(this.wcpInst);

    for (int ct = 0; ct <= this.maxCt(); ct++) {
      trace
          .wordComparisonStamp(stamp)
          .oneLineInstruction(oli)
          .variableLengthInstruction(vli)
          .counter(UnsignedByte.of(ct))
          .ctMax(UnsignedByte.of(this.ctMax))
          .inst(inst)
          .isEq(this.wcpInst == EQbv)
          .isIszero(this.wcpInst == ISZERObv)
          .isSlt(this.wcpInst == SLTbv)
          .isSgt(this.wcpInst == SGTbv)
          .isLt(this.wcpInst == LTbv)
          .isGt(this.wcpInst == GTbv)
          .isLeq(this.wcpInst == LEQbv)
          .isGeq(this.wcpInst == GEQbv)
          .argument1Hi(this.arg1Hi)
          .argument1Lo(this.arg1Lo)
          .argument2Hi(this.arg2Hi)
          .argument2Lo(this.arg2Lo)
          .result(resLo)
          .bits(bits.get(ct))
          .neg1(neg1)
          .neg2(neg2)
          .byte1(UnsignedByte.of(this.arg1Hi.get(ct)))
          .byte2(UnsignedByte.of(this.arg1Lo.get(ct)))
          .byte3(UnsignedByte.of(this.arg2Hi.get(ct)))
          .byte4(UnsignedByte.of(this.arg2Lo.get(ct)))
          .byte5(UnsignedByte.of(adjHi.get(ct)))
          .byte6(UnsignedByte.of(adjLo.get(ct)))
          .acc1(this.arg1Hi.slice(0, 1 + ct))
          .acc2(this.arg1Lo.slice(0, 1 + ct))
          .acc3(this.arg2Hi.slice(0, 1 + ct))
          .acc4(this.arg2Lo.slice(0, 1 + ct))
          .acc5(adjHi.slice(0, 1 + ct))
          .acc6(adjLo.slice(0, 1 + ct))
          .bit1(bit1)
          .bit2(bit2)
          .bit3(bit3)
          .bit4(bit4)
          .validateRow();
    }
  }

  private boolean isOli() {
    return switch (this.wcpInst) {
      case ISZERObv, EQbv -> true;
      case SLTbv, SGTbv, LTbv, GTbv, LEQbv, GEQbv -> false;
      default -> throw new IllegalStateException("Unexpected value: " + this.wcpInst);
    };
  }

  private boolean isVli() {
    return switch (this.wcpInst) {
      case LTbv, GTbv, LEQbv, GEQbv, SLTbv, SGTbv -> true;
      case ISZERObv, EQbv -> false;
      default -> throw new IllegalStateException("Unexpected value: " + this.wcpInst);
    };
  }

  private int maxCt() {
    switch (this.wcpInst) {
      case ISZERObv, EQbv -> {
        return 0;
      }
      case LTbv, GTbv, LEQbv, GEQbv, SLTbv, SGTbv -> {
        if (this.arg1.isZero() && this.arg2.isZero()) {
          return 0;
        } else {
          final ArrayList<Integer> sizes = new ArrayList<>(4);
          sizes.add(this.arg1.slice(0, LLARGE).trimLeadingZeros().size());
          sizes.add(this.arg2.slice(0, LLARGE).trimLeadingZeros().size());
          sizes.add(this.arg1.slice(LLARGE, LLARGE).trimLeadingZeros().size());
          sizes.add(this.arg2.slice(LLARGE, LLARGE).trimLeadingZeros().size());
          return Collections.max(sizes) - 1;
        }
      }
      default -> throw new IllegalStateException("Unexpected value: " + this.wcpInst);
    }
  }

  @Override
  protected int computeLineCount() {
    return this.ctMax + 1;
  }
}
