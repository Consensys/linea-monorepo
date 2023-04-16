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
package net.consensys.linea.zktracer.module.alu.mod;

import static com.google.common.base.Preconditions.checkElementIndex;
import static net.consensys.linea.zktracer.module.Util.byteBits;

import java.math.BigInteger;
import java.util.Arrays;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.apache.tuweni.units.bigints.UInt64;

public class ModData {
  private final OpCode opCode;
  private final boolean oli;
  private final BaseBytes arg1;
  private final BaseBytes arg2;
  private BaseBytes result = BaseBytes.fromBytes32(Bytes32.ZERO);
  private BaseTheta A_Bytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta B_Bytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta Q_Bytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta R_Bytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta H_Bytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private BaseTheta D_Bytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  private final boolean[] cmp1 = new boolean[8];
  private final boolean[] cmp2 = new boolean[8];
  private Boolean[] msb1 = new Boolean[8];
  private Boolean[] msb2 = new Boolean[8];

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
      UInt256 b = absoluteValueIfSignedInst(arg2);
      UInt256 q = a.divide(b);
      UInt256 r = a.mod(b);
      this.A_Bytes = BaseTheta.fromBytes32(a);
      this.B_Bytes = BaseTheta.fromBytes32(b);
      this.Q_Bytes = BaseTheta.fromBytes32(q);
      this.R_Bytes = BaseTheta.fromBytes32(r);
      this.D_Bytes = BaseTheta.fromBytes32(Bytes32.ZERO);
      this.setCmp12();
      this.setDeltas();
      this.setAlphaBetasH012();

      UnsignedByte msb_1 = UnsignedByte.of(this.arg1.getHigh().get(0));
      UnsignedByte msb_2 = UnsignedByte.of(this.arg2.getHigh().get(0));

      this.msb1 = byteBits(msb_1);
      this.msb2 = byteBits(msb_2);
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

  private UInt256 b(int k) {
    checkElementIndex(k, 4);
    return UInt256.fromBytes(B_Bytes.get(k));
  }

  private UInt256 q(int k) {
    checkElementIndex(k, 4);
    return UInt256.fromBytes(Q_Bytes.get(k));
  }

  private UInt256 r(int k) {
    checkElementIndex(k, 4);
    return UInt256.fromBytes(R_Bytes.get(k));
  }

  private UInt256 h(int k) {
    checkElementIndex(k, 3);
    return UInt256.fromBytes(H_Bytes.get(k));
  }

  private void setCmp12() {
    for (int k = 0; k < 4; k++) {
      cmp1[k] = b(k).compareTo(r(k)) > 0;
      cmp2[k] = b(k).compareTo(r(k)) == 0;
    }
  }

  private void setDeltas() {
    for (int k = 0; k < 4; k++) {
      UInt256 delta;
      if (this.cmp1[k]) {
        delta = b(k).subtract(r(k)).subtract(UInt256.ONE);
      } else {
        delta = r(k).subtract(b(k));
      }
      D_Bytes.setBytes(k * 8, delta.slice(24, 8));
    }
  }

  private void setAlphaBetasH012() {
    UInt256 theta = UInt256.ONE;
    UInt256 thetaSquared = UInt256.ONE;

    theta = theta.shiftLeft(64);
    thetaSquared = thetaSquared.shiftLeft(128);

    UInt256 sum = b(0).multiply(q(1)).add(b(1).multiply(q(0)));
    this.H_Bytes = BaseTheta.fromBytes32(sum);

    // alpha
    cmp2[4] = sum.compareTo(thetaSquared) >= 0;

    sum =
        b(0).multiply(q(3))
            .add(b(1).multiply(q(2)))
            .add(b(2).multiply(q(1)))
            .add(b(3).multiply(q(0)));

    if (sum.bitLength() > 64) {
      throw new RuntimeException("b[0]q[3] + b[1]q[2] + b[2]q[1] + b[3]q[0] >= (1 << 64)");
    }

    H_Bytes.setBytes(2 * 8, sum.slice(24, 8));

    sum = q(0).multiply(b(0));
    sum = sum.add(h(0).multiply(theta));
    sum = sum.add(UInt256.fromBytes(R_Bytes.getLow()));

    UInt256 beta = sum.divide(thetaSquared);
    if (beta.compareTo(UInt256.valueOf(2)) > 0) {
      throw new RuntimeException("b[0]q[0] + theta.h[0] + rLo = [beta|...] with beta > 2");
    }

    UInt64 betaUInt64 = UInt64.valueOf(beta.toUnsignedBigInteger());
    cmp2[5] = betaUInt64.mod(UInt64.valueOf(2)).compareTo(UInt64.ONE) == 0; // beta_0
    cmp2[6] = betaUInt64.divide(UInt64.valueOf(2)).compareTo(UInt64.ONE) == 0; // beta_1

    BigInteger sumInt = sum.mod(thetaSquared).toUnsignedBigInteger();
    BigInteger aLo = this.A_Bytes.getLow().toUnsignedBigInteger();
    if (sumInt.compareTo(aLo) != 0) {
      throw new RuntimeException("b[0]q[0] + theta.h[0] + rLo = [beta|xxx] and xxx != aLo");
    }
  }

  public OpCode getOpCode() {
    return opCode;
  }

  public boolean isOli() {
    return oli;
  }

  public BaseBytes getArg1() {
    return arg1;
  }

  public BaseBytes getArg2() {
    return arg2;
  }

  public BaseBytes getResult() {
    return result;
  }

  public BaseTheta getBBytes() {
    return B_Bytes;
  }

  public BaseTheta getQBytes() {
    return Q_Bytes;
  }

  public BaseTheta getRBytes() {
    return R_Bytes;
  }

  public BaseTheta getHBytes() {
    return H_Bytes;
  }

  public BaseTheta getDeltaBytes() {
    return D_Bytes;
  }

  public Boolean[] getMsb1() {
    return msb1;
  }

  public Boolean[] getMsb2() {
    return msb2;
  }

  public boolean[] getCmp1() {
    return cmp1;
  }

  public boolean[] getCmp2() {
    return cmp2;
  }

  public boolean isSigned() {
    return this.opCode == OpCode.SDIV || this.opCode == OpCode.SMOD;
  }

  public boolean isDiv() {
    return this.opCode == OpCode.DIV || this.opCode == OpCode.SDIV;
  }
}
