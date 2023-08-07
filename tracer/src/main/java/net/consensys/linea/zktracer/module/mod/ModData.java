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

package net.consensys.linea.zktracer.module.mod;

import static com.google.common.base.Preconditions.checkElementIndex;
import static net.consensys.linea.zktracer.module.Util.byteBits;

import java.math.BigInteger;
import java.util.Arrays;

import lombok.Getter;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.apache.tuweni.units.bigints.UInt64;

public class ModData {
  @Getter private final OpCode opCode;
  @Getter private final boolean oli;
  @Getter private final BaseBytes arg1;
  @Getter private final BaseBytes arg2;
  @Getter private BaseBytes result = BaseBytes.fromBytes32(Bytes32.ZERO);
  @Getter private BaseTheta aBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  @Getter private BaseTheta bBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  @Getter private BaseTheta qBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  @Getter private BaseTheta rBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  @Getter private BaseTheta hBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  @Getter private BaseTheta dBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  @Getter private final boolean[] cmp1 = new boolean[8];
  @Getter private final boolean[] cmp2 = new boolean[8];
  @Getter private Boolean[] msb1 = new Boolean[8];
  @Getter private Boolean[] msb2 = new Boolean[8];

  public ModData(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    this.arg1 = BaseBytes.fromBytes32(arg1);
    this.arg2 = BaseBytes.fromBytes32(arg2);

    this.opCode = opCode;
    this.oli = arg2.isZero();

    Arrays.fill(msb1, false);
    Arrays.fill(msb2, false);

    if (!this.oli) {
      this.result = getRes(opCode, arg1, arg2);

      UInt256 a = absoluteValueIfSignedInst(arg1);
      this.aBytes = BaseTheta.fromBytes32(a);

      UInt256 b = absoluteValueIfSignedInst(arg2);
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
      default -> throw new RuntimeException("Modular arithmetic was given wrong opcode");
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

  public boolean isSigned() {
    return this.opCode == OpCode.SDIV || this.opCode == OpCode.SMOD;
  }

  public boolean isDiv() {
    return this.opCode == OpCode.DIV || this.opCode == OpCode.SDIV;
  }
}
