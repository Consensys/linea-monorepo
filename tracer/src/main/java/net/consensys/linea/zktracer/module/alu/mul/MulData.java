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

package net.consensys.linea.zktracer.module.alu.mul;

import static net.consensys.linea.zktracer.module.Util.boolToByte;
import static net.consensys.linea.zktracer.module.Util.byteBits;
import static net.consensys.linea.zktracer.module.Util.getBit;
import static net.consensys.linea.zktracer.module.Util.getOverflow;

import java.lang.reflect.Array;
import java.math.BigInteger;
import java.util.Arrays;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.Bytes16;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

public class MulData {
  private static final int MMEDIUM = 8;
  final OpCode opCode;
  final Bytes32 arg1;
  final Bytes32 arg2;

  final Bytes16 arg1Hi;
  final Bytes16 arg1Lo;
  final Bytes16 arg2Hi;
  final Bytes16 arg2Lo;

  final boolean tinyBase;
  final boolean tinyExponent;

  UInt256 resAcc =
      UInt256.ZERO; // accumulator which converges in a series of "square and multiply"'s
  UInt256 expAcc =
      UInt256.ZERO; // accumulator for doubles and adds of the exponent, resets at some point

  final BaseTheta aBytes;
  final BaseTheta bBytes;
  BaseTheta cBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  BaseTheta hBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
  boolean snm = false;
  int index;
  Boolean[] bits = new Boolean[8];
  String exponentBits = "0";

  BaseBytes res;

  @SuppressWarnings("checkstyle:WhitespaceAround")
  public MulData(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    this.opCode = opCode;
    this.arg1 = arg1;
    this.arg2 = arg2;
    this.aBytes = BaseTheta.fromBytes32(arg1);
    this.bBytes = BaseTheta.fromBytes32(arg2);

    arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
    arg1Lo = Bytes16.wrap(arg1.slice(16));
    arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
    arg2Lo = Bytes16.wrap(arg2.slice(16));

    this.res = getRes(opCode, arg1, arg2); // TODO can we get this from the EVM

    final BigInteger arg1BigInt = UInt256.fromBytes(arg1).toUnsignedBigInteger();
    final BigInteger arg2BigInt = UInt256.fromBytes(arg2).toUnsignedBigInteger();

    this.tinyBase = isTiny(arg1BigInt);
    this.tinyExponent = isTiny(arg2BigInt);

    // initialize bits
    Arrays.fill(bits, false);

    final Regime regime = getRegime();
    switch (regime) {
      case TRIVIAL_MUL -> {}
      case NON_TRIVIAL_MUL -> cBytes = BaseTheta.fromBytes32(res.getBytes32());
      case EXPONENT_ZERO_RESULT -> setArraysForZeroResultCase();
      case EXPONENT_NON_ZERO_RESULT -> {
        this.exponentBits = arg2.toBigInteger().toString();
        snm = false;
      }
      case IOTA -> throw new RuntimeException("alu/mul regime was never set");
      default -> throw new IllegalStateException("[MUL module] Unexpected regime value: " + regime);
    }
  }

  private static BaseBytes getRes(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    return switch (opCode) {
      case MUL -> BaseBytes.fromBytes32(UInt256.fromBytes(arg1).multiply(UInt256.fromBytes(arg2)));
      case EXP -> BaseBytes.fromBytes32(UInt256.fromBytes(arg1).pow(UInt256.fromBytes(arg2)));
      default -> throw new IllegalStateException("[MUL module] Unexpected opcode: " + opCode);
    };
  }

  private void setArraysForZeroResultCase() {
    int nu = twoAdicity(arg1);

    if (nu >= 128) {
      return;
    }

    Bytes ones = Bytes.repeat((byte) 1, 8);
    Bytes bytes;

    if (128 > nu && nu >= 64) {
      bytes = aBytes.get(1);
    } else {
      for (int i = 0; i < 8; i++) {
        cBytes.setChunk(0, ones);
      }
      bytes = aBytes.get(0);
    }

    int nuQuo = (nu / 8) % 8;
    int nuRem = nu % 8;
    byte pivotByte = bytes.get(7 - nuQuo);

    for (int i = 0; i < 8; i++) {
      cBytes.set(1, i, pivotByte);
      cBytes.set(2, i, boolToByte(i > 7 - nuRem));
      cBytes.set(3, i, boolToByte(i > 7 - nuQuo));
      hBytes.set(2, i, callFunc(i, 7 - nuRem));
      hBytes.set(3, i, callFunc(i, 7 - nuQuo));
    }

    bits = byteBits(UnsignedByte.of(pivotByte));

    int lowerBoundOnTwoAdicity = 8 * (int) (hBytes.get(3, 7)) + (int) (hBytes.get(2, 7));

    if (nu >= 64) {
      lowerBoundOnTwoAdicity += 64;
    }

    // our lower bound should coincide with the 2-adicity
    if (lowerBoundOnTwoAdicity != nu) {
      String s =
          String.format(
              "2-adicity nu = %d != %d = lower bound on 2-adicity", nu, lowerBoundOnTwoAdicity);
      throw new RuntimeException(s);
    }
    if (lowerBoundOnTwoAdicity == 0) {
      throw new RuntimeException("lower bound on 2 adicity == 0 in the zero result case");
    }

    final UInt256 twoFiftySix = UInt256.valueOf(256);
    if (arg2.compareTo(twoFiftySix) >= 0) {
      // arg2 = exponent >= 256
      hBytes.set(1, 6, (byte) ((lowerBoundOnTwoAdicity - 1) / 256));
      hBytes.set(1, 7, (byte) ((lowerBoundOnTwoAdicity - 1) % 256));
    } else {
      // exponent < 256
      int exponent = arg2.toUnsignedBigInteger().intValue();
      int target = exponent * lowerBoundOnTwoAdicity - 256;

      if (target < 0) {
        throw new RuntimeException("lower bound on 2-adicity is wrong");
      }

      if (target > 255 * (8 * 7 + 7 + 64)) {
        throw new RuntimeException("something went awfully wrong ...");
      }

      final BaseTheta thing = BaseTheta.fromBytes32(UInt256.valueOf(target));
      hBytes.setChunk(1, thing.get(0));
    }
  }

  public static byte callFunc(final int x, final int k) {
    if (x < k) {
      return 0;
    }

    return (byte) (x - k);
  }

  public boolean exponentBit() {
    return '1' == exponentBits.charAt(index);
  }

  public boolean exponentSource() {
    return this.index + 128 >= exponentBits.length();
  }

  public static int twoAdicity(final Bytes32 x) {
    if (x.isZero()) {
      // panic("twoAdicity was called on zero")
      return 256;
    }

    String baseStringBase2 = x.toBigInteger().toString(2);

    for (int i = 0; i < baseStringBase2.length(); i++) {
      int j = baseStringBase2.length() - i - 1;
      char zeroAscii = '0';
      if (baseStringBase2.charAt(j) != zeroAscii) {
        return i;
      }
    }

    return 0;
  }
  //
  //  private boolean largeExponent() {
  //    return exponentBits.length() > 128;
  //  }

  public boolean isOneLineInstruction() {
    return tinyBase || tinyExponent;
  }

  public Regime getRegime() {
    if (isOneLineInstruction()) {
      return Regime.TRIVIAL_MUL;
    }

    if (OpCode.MUL.equals(opCode)) {
      return Regime.NON_TRIVIAL_MUL;
    }

    if (OpCode.EXP.equals(opCode)) {
      if (res.isZero()) {
        return Regime.EXPONENT_ZERO_RESULT;
      } else {
        return Regime.EXPONENT_NON_ZERO_RESULT;
      }
    }

    return Regime.IOTA;
  }

  public static boolean isTiny(BigInteger arg) {
    return arg.compareTo(BigInteger.valueOf(1)) <= 0;
  }

  public boolean carryOn() {
    // first round is special
    if (index == 0 && !snm) {
      snm = true;
      resAcc = UInt256.valueOf(1); // TODO assuming this is what SetOne() does
      cBytes = BaseTheta.fromBytes32(arg1);

      return true;
    }

    if (snm == exponentBit()) {
      hiToLoExponentBitAccumulatorReset();
      index++;
      snm = false;
      return index != exponentBits.length();
    } else {
      snm = true;
    }

    return true;
  }

  public int getBitNum() {
    return bitNum(index, exponentBits.length());
  }

  private int bitNum(int i, int length) {
    if (length <= 128) {
      return i;
    } else {
      if (i + 128 < length) {
        return i;
      } else {
        return i + 128 - length;
      }
    }
  }

  public void update() {
    final BigInteger arg1BigInt = UInt256.fromBytes(arg1).toUnsignedBigInteger();
    if (!snm) {
      // squaring
      setHsAndBits(resAcc, resAcc);
      expAcc = expAcc.add(expAcc);
      resAcc = resAcc.multiply(resAcc);
    } else {
      // multiplying by base
      setHsAndBits(UInt256.valueOf(arg1BigInt), resAcc);
      expAcc = expAcc.add(UInt256.ONE);
      resAcc = UInt256.valueOf(arg1BigInt).multiply(resAcc);
    }
    cBytes = BaseTheta.fromBytes32(resAcc);
  }

  public void setHsAndBits(UInt256 a, UInt256 b) {
    setHsAndBitsFromBaseThetas(BaseTheta.fromBytes32(a), BaseTheta.fromBytes32(b));
  }

  @SuppressWarnings("checkstyle:VariableDeclarationUsageDistance")
  public void setHsAndBitsFromBaseThetas(BaseTheta aBaseTheta, BaseTheta bBaseTheta) {
    UInt256[] aBaseThetaInts = (UInt256[]) Array.newInstance(UInt256.class, 4);
    UInt256[] bBaseThetaInts = (UInt256[]) Array.newInstance(UInt256.class, 4);

    for (int i = 0; i < 4; i++) {
      aBaseThetaInts[i] = UInt256.fromBytes(aBaseTheta.get(i));
      bBaseThetaInts[i] = UInt256.fromBytes(bBaseTheta.get(i));
    }

    UInt256 prod = aBaseThetaInts[1].multiply(bBaseThetaInts[0]);
    UInt256 sum = prod; // sum := a1 * b0
    prod = aBaseThetaInts[0].multiply(bBaseThetaInts[1]);
    sum = sum.add(prod); // sum += a0 * b1

    BaseTheta sumBaseTheta = BaseTheta.fromBytes32(sum);
    hBytes.setChunk(0, sumBaseTheta.get(0));
    hBytes.setChunk(1, sumBaseTheta.get(1));

    long alpha = getOverflow(sum, 1, "alpha OOB");

    sum = aBaseThetaInts[3].multiply(bBaseThetaInts[0]); // sum := a3 * b0
    prod = aBaseThetaInts[2].multiply(bBaseThetaInts[1]);
    sum = sum.add(prod); // sum += a2 * b1
    prod = aBaseThetaInts[1].multiply(bBaseThetaInts[2]);
    sum = sum.add(prod); // sum += a1 * b2
    prod = aBaseThetaInts[0].multiply(bBaseThetaInts[3]);
    sum = sum.add(prod); // sum += a0 * b3

    sumBaseTheta = BaseTheta.fromBytes32(sum);
    hBytes.setChunk(2, sumBaseTheta.get(0));
    hBytes.setChunk(3, sumBaseTheta.get(1));

    long beta = getOverflow(sum, 3, "beta OOB");
    prod = aBaseThetaInts[0].multiply(bBaseThetaInts[0]);
    sum = prod; // sum := a0 * b0

    prod = UInt256.fromBytes(hBytes.get(0)).shiftLeft(64);
    sum = sum.add(prod); // sum += (h0 << 64)

    long eta = getOverflow(sum, 1, "eta OOB");

    sum = UInt256.valueOf(eta); // sum := eta
    sum = sum.add(UInt256.fromBytes(hBytes.get(1))); // sum += h1
    prod = UInt256.valueOf(alpha).shiftLeft(64);
    sum = sum.add(prod); // sum += (alpha << 64)
    prod = aBaseThetaInts[2].multiply(bBaseThetaInts[0]);
    sum = sum.add(prod); // sum += a2 * b0
    prod = aBaseThetaInts[1].multiply(bBaseThetaInts[1]);
    sum = sum.add(prod); // sum += a1 * b1
    prod = aBaseThetaInts[0].multiply(bBaseThetaInts[2]);
    sum = sum.add(prod); // sum += a0 * b2
    sum = sum.add(UInt256.fromBytes(hBytes.get(2)).shiftLeft(64)); // sum += (h2 << 64)

    long mu = getOverflow(sum, 3, "mu OOB");

    bits[0] = false;
    bits[1] = false;
    bits[2] = getBit(alpha, 0);
    bits[3] = getBit(beta, 0);
    bits[4] = getBit(beta, 1);
    bits[5] = getBit(eta, 0);
    bits[6] = getBit(mu, 0);
    bits[7] = getBit(mu, 1);
  }

  // hiToLoExponentBitAccumulatorReset resets the exponent bit accumulator
  // under the following conditions:
  //   - we are dealing with the high part of the exponent bits, i.e. md.exponentBit() = 0
  //   - SQUARE_AND_MULTIPLY == EXPONENT_BIT
  //   - the exponent bit accumulator coincides with the high part of the exponent
  private void hiToLoExponentBitAccumulatorReset() {
    if (!exponentSource()) {
      if (snm == exponentBit()) { // note: when called this is already assumed
        Bytes32 arg2Copy = arg2.copy();
        if (arg2Copy.shiftRight(128).equals(expAcc)) {
          expAcc = UInt256.MIN_VALUE;
        }
      }
    }
  }

  public int maxCt() {
    return isOneLineInstruction() ? 1 : MMEDIUM;
  }
}
