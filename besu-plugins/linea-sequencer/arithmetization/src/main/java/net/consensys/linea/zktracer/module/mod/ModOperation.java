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

package net.consensys.linea.zktracer.module.mod;

import static com.google.common.base.Preconditions.checkElementIndex;
import static net.consensys.linea.zktracer.module.Util.byteBits;

import java.math.BigInteger;
import java.util.Arrays;
import java.util.Objects;

import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.apache.tuweni.units.bigints.UInt64;

public class ModOperation {
  private static final int MMEDIUM = 8;

  private final OpCode opCode;
  private final boolean oli;
  private final Bytes32 rawArg1;
  private final Bytes32 rawArg2;
  private BaseBytes arg1;
  private BaseBytes arg2;

  private BaseBytes result = BaseBytes.fromBytes32(Bytes32.ZERO);
  private BaseTheta aBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta bBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta qBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta rBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta hBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta dBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private final boolean[] cmp1 = new boolean[8];
  private final boolean[] cmp2 = new boolean[8];
  private Boolean[] msb1 = new Boolean[8];
  private Boolean[] msb2 = new Boolean[8];

  /**
   * This custom hash function ensures that all identical operations are only traced once per
   * conflation block.
   */
  @Override
  public int hashCode() {
    return Objects.hash(this.opCode, this.rawArg1, this.rawArg2);
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (o == null || getClass() != o.getClass()) return false;
    final ModOperation that = (ModOperation) o;
    return Objects.equals(opCode, that.opCode)
        && Objects.equals(rawArg1, that.rawArg1)
        && Objects.equals(rawArg2, that.rawArg2);
  }

  public ModOperation(OpCodeData opCodeData, Bytes32 arg1, Bytes32 arg2) {
    this(opCodeData.mnemonic(), arg1, arg2);
  }

  public ModOperation(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    this.rawArg1 = arg1;
    this.rawArg2 = arg2;

    this.opCode = opCode;
    this.oli = arg2.isZero();
  }

  private void compute() {
    this.arg1 = BaseBytes.fromBytes32(this.rawArg1);
    this.arg2 = BaseBytes.fromBytes32(this.rawArg2);

    Arrays.fill(msb1, false);
    Arrays.fill(msb2, false);

    if (!this.oli) {
      this.result = getRes(opCode, this.rawArg1, this.rawArg2);

      UInt256 a = absoluteValueIfSignedInst(this.rawArg1);
      this.aBytes = BaseTheta.fromBytes32(a);

      UInt256 b = absoluteValueIfSignedInst(this.rawArg2);
      this.bBytes = BaseTheta.fromBytes32(b);

      UInt256 q = a.divide(b);
      this.qBytes = BaseTheta.fromBytes32(q);

      UInt256 r = a.mod(b);
      this.rBytes = BaseTheta.fromBytes32(r);

      this.dBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
      this.setCmp12();
      this.setDeltas();
      this.setAlphaBetasH012();

      UnsignedByte msb1 = UnsignedByte.of(this.arg1.getHigh().get(0));
      UnsignedByte msb2 = UnsignedByte.of(this.arg2.getHigh().get(0));

      this.msb1 = byteBits(msb1);
      this.msb2 = byteBits(msb2);
    }
  }

  private static BaseBytes getRes(OpCode op, Bytes32 arg1, Bytes32 arg2) {
    return switch (op) {
      case DIV -> BaseBytes.fromBytes32(UInt256.fromBytes(arg1).divide(UInt256.fromBytes(arg2)));
      case SDIV -> BaseBytes.fromBytes32(UInt256.fromBytes(arg1).sdiv0(UInt256.fromBytes(arg2)));
      case MOD -> BaseBytes.fromBytes32(UInt256.fromBytes(arg1).mod(UInt256.fromBytes(arg2)));
      case SMOD -> BaseBytes.fromBytes32(UInt256.fromBytes(arg1).smod0(UInt256.fromBytes(arg2)));
      default -> throw new IllegalArgumentException("Modular arithmetic was given wrong opcode");
    };
  }

  private UInt256 absoluteValueIfSignedInst(Bytes32 arg) {
    if (isSigned()) {
      return UInt256.valueOf(arg.toBigInteger().abs());
    }
    return UInt256.fromBytes(arg);
  }

  private UInt256 bVar(int k) {
    checkElementIndex(k, 4);
    return UInt256.fromBytes(bBytes.get(k));
  }

  private UInt256 qVar(int k) {
    checkElementIndex(k, 4);
    return UInt256.fromBytes(qBytes.get(k));
  }

  private UInt256 rVar(int k) {
    checkElementIndex(k, 4);
    return UInt256.fromBytes(rBytes.get(k));
  }

  private UInt256 hVar(int k) {
    checkElementIndex(k, 3);
    return UInt256.fromBytes(hBytes.get(k));
  }

  private void setCmp12() {
    for (int k = 0; k < 4; k++) {
      cmp1[k] = bVar(k).compareTo(rVar(k)) > 0;
      cmp2[k] = bVar(k).compareTo(rVar(k)) == 0;
    }
  }

  private void setDeltas() {
    for (int k = 0; k < 4; k++) {
      UInt256 delta;
      if (this.cmp1[k]) {
        delta = bVar(k).subtract(rVar(k)).subtract(UInt256.ONE);
      } else {
        delta = rVar(k).subtract(bVar(k));
      }
      dBytes.set(k, delta.slice(24, 8));
    }
  }

  private void setAlphaBetasH012() {
    UInt256 theta = UInt256.ONE;
    UInt256 thetaSquared = UInt256.ONE;

    theta = theta.shiftLeft(64);
    thetaSquared = thetaSquared.shiftLeft(128);

    UInt256 sum = bVar(0).multiply(qVar(1)).add(bVar(1).multiply(qVar(0)));
    this.hBytes = BaseTheta.fromBytes32(sum);

    // alpha
    cmp2[4] = sum.compareTo(thetaSquared) >= 0;

    sum =
        bVar(0)
            .multiply(qVar(3))
            .add(bVar(1).multiply(qVar(2)))
            .add(bVar(2).multiply(qVar(1)))
            .add(bVar(3).multiply(qVar(0)));

    if (sum.bitLength() > 64) {
      throw new RuntimeException("b[0]q[3] + b[1]q[2] + b[2]q[1] + b[3]q[0] >= (1 << 64)");
    }

    hBytes.set(2, sum.slice(24, 8));

    sum = qVar(0).multiply(bVar(0));
    sum = sum.add(hVar(0).multiply(theta));
    sum = sum.add(UInt256.fromBytes(rBytes.getLow()));

    UInt256 beta = sum.divide(thetaSquared);
    if (beta.compareTo(UInt256.valueOf(2)) > 0) {
      throw new RuntimeException("b[0]q[0] + theta.h[0] + rLo = [beta|...] with beta > 2");
    }

    UInt64 betaUInt64 = UInt64.valueOf(beta.toUnsignedBigInteger());
    cmp2[5] = betaUInt64.mod(UInt64.valueOf(2)).compareTo(UInt64.ONE) == 0; // beta_0
    cmp2[6] = betaUInt64.divide(UInt64.valueOf(2)).compareTo(UInt64.ONE) == 0; // beta_1

    BigInteger sumInt = sum.mod(thetaSquared).toUnsignedBigInteger();
    BigInteger aLo = this.aBytes.getLow().toUnsignedBigInteger();
    if (sumInt.compareTo(aLo) != 0) {
      throw new RuntimeException("b[0]q[0] + theta.h[0] + rLo = [beta|xxx] and xxx != aLo");
    }
  }

  boolean isSigned() {
    return this.opCode == OpCode.SDIV || this.opCode == OpCode.SMOD;
  }

  boolean isDiv() {
    return this.opCode == OpCode.DIV || this.opCode == OpCode.SDIV;
  }

  int maxCounter() {
    if (this.oli) {
      return 1;
    } else {
      return MMEDIUM;
    }
  }

  public void trace(Trace trace, int stamp) {
    this.compute();

    for (int i = 0; i < this.maxCounter(); i++) {
      final int accLength = i + 1;
      trace
          .stamp(Bytes.ofUnsignedLong(stamp))
          .oli(this.oli)
          .ct(Bytes.of(i))
          .inst(Bytes.of(this.opCode.byteValue()))
          .decSigned(this.isSigned())
          .decOutput(this.isDiv())
          .arg1Hi(this.arg1.getHigh())
          .arg1Lo(this.arg1.getLow())
          .arg2Hi(this.arg2.getHigh())
          .arg2Lo(this.arg2.getLow())
          .resHi(this.result.getHigh())
          .resLo(this.result.getLow())
          .acc12(this.arg1.getBytes32().slice(8, i + 1))
          .acc13(this.arg1.getBytes32().slice(0, i + 1))
          .acc22(this.arg2.getBytes32().slice(8, i + 1))
          .acc23(this.arg2.getBytes32().slice(0, i + 1))
          .accB0(this.bBytes.get(0).slice(0, accLength))
          .accB1(this.bBytes.get(1).slice(0, accLength))
          .accB2(this.bBytes.get(2).slice(0, accLength))
          .accB3(this.bBytes.get(3).slice(0, accLength))
          .accR0(this.rBytes.get(0).slice(0, accLength))
          .accR1(this.rBytes.get(1).slice(0, accLength))
          .accR2(this.rBytes.get(2).slice(0, accLength))
          .accR3(this.rBytes.get(3).slice(0, accLength))
          .accQ0(this.qBytes.get(0).slice(0, accLength))
          .accQ1(this.qBytes.get(1).slice(0, accLength))
          .accQ2(this.qBytes.get(2).slice(0, accLength))
          .accQ3(this.qBytes.get(3).slice(0, accLength))
          .accDelta0(this.dBytes.get(0).slice(0, accLength))
          .accDelta1(this.dBytes.get(1).slice(0, accLength))
          .accDelta2(this.dBytes.get(2).slice(0, accLength))
          .accDelta3(this.dBytes.get(3).slice(0, accLength))
          .byte22(UnsignedByte.of(this.arg2.getByte(i + 8)))
          .byte23(UnsignedByte.of(this.arg2.getByte(i)))
          .byte12(UnsignedByte.of(this.arg1.getByte(i + 8)))
          .byte13(UnsignedByte.of(this.arg1.getByte(i)))
          .byteB0(UnsignedByte.of(this.bBytes.get(0).get(i)))
          .byteB1(UnsignedByte.of(this.bBytes.get(1).get(i)))
          .byteB2(UnsignedByte.of(this.bBytes.get(2).get(i)))
          .byteB3(UnsignedByte.of(this.bBytes.get(3).get(i)))
          .byteR0(UnsignedByte.of(this.rBytes.get(0).get(i)))
          .byteR1(UnsignedByte.of(this.rBytes.get(1).get(i)))
          .byteR2(UnsignedByte.of(this.rBytes.get(2).get(i)))
          .byteR3(UnsignedByte.of(this.rBytes.get(3).get(i)))
          .byteQ0(UnsignedByte.of(this.qBytes.get(0).get(i)))
          .byteQ1(UnsignedByte.of(this.qBytes.get(1).get(i)))
          .byteQ2(UnsignedByte.of(this.qBytes.get(2).get(i)))
          .byteQ3(UnsignedByte.of(this.qBytes.get(3).get(i)))
          .byteDelta0(UnsignedByte.of(this.dBytes.get(0).get(i)))
          .byteDelta1(UnsignedByte.of(this.dBytes.get(1).get(i)))
          .byteDelta2(UnsignedByte.of(this.dBytes.get(2).get(i)))
          .byteDelta3(UnsignedByte.of(this.dBytes.get(3).get(i)))
          .byteH0(UnsignedByte.of(this.hBytes.get(0).get(i)))
          .byteH1(UnsignedByte.of(this.hBytes.get(1).get(i)))
          .byteH2(UnsignedByte.of(this.hBytes.get(2).get(i)))
          .accH0(Bytes.wrap(this.hBytes.get(0)).slice(0, i + 1))
          .accH1(Bytes.wrap(this.hBytes.get(1)).slice(0, i + 1))
          .accH2(Bytes.wrap(this.hBytes.get(2)).slice(0, i + 1))
          .cmp1(this.cmp1[i])
          .cmp2(this.cmp2[i])
          .msb1(this.msb1[i])
          .msb2(this.msb2[i])
          .validateRow();
    }
  }
}
