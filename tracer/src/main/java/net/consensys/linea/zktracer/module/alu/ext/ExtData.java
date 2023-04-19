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
package net.consensys.linea.zktracer.module.alu.ext;

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.zktracer.module.Util.getBit;
import static net.consensys.linea.zktracer.module.Util.getOverflow;

import java.math.BigInteger;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.bytestheta.BytesArray;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

@SuppressWarnings("UnusedVariable")
public class ExtData {
  private final OpCode opCode;
  private final boolean oli;
  private final BaseBytes arg1;
  private final BaseBytes arg2;

  private final BaseBytes arg3;
  private BaseBytes result;
  private BaseTheta aBytes;

  private BaseTheta bBytes;
  private BaseTheta cBytes;
  private BaseTheta deltaBytes;

  private BytesArray hBytes;
  private BaseTheta rBytes;
  private BytesArray iBytes;

  private BytesArray jBytes;
  private BytesArray qBytes;

  private boolean[] cmp = new boolean[8];

  boolean[] overflowH = new boolean[8];
  boolean[] overflowI = new boolean[8];


  public ExtData(OpCode opCode, Bytes32 arg1, Bytes32 arg2, Bytes32 arg3) {

    this.opCode = opCode;
    this.arg1 = BaseBytes.fromBytes32(arg1);
    this.arg2 = BaseBytes.fromBytes32(arg2);
    this.arg3 = BaseBytes.fromBytes32(arg3);
    this.aBytes = BaseTheta.fromBytes32(arg1);
    this.bBytes = BaseTheta.fromBytes32(arg2);
    this.cBytes = BaseTheta.fromBytes32(arg3);

    this.hBytes = new BytesArray(8);
    this.iBytes = new BytesArray(7);

    this.deltaBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
    // this.jBytes = BaseTheta.fromBytes32(Bytes32.ZERO);

    UInt256 result = ExtCalculator.performExtendedModularArithmetic(opCode, arg1, arg2, arg3);
    this.result = BaseTheta.fromBytes32(result);
    rBytes = BaseTheta.fromBytes32(result);

    this.oli = isOneLineInstruction();
    if (this.oli) {
      return;
    }

    computeDeltasAndComparisonFlags(cBytes, rBytes);
    setJsAndQs();
    setHsAndOverflowH();
    setIsAndOverflowI();
    setOverflowJ();
    setOverflowRes();
  }

  public boolean bit1() {
    return this.opCode == OpCode.MULMOD && this.arg1.isZero();
  }

  public boolean bit2() {
    return this.opCode == OpCode.MULMOD && this.arg2.isZero();
  }

  public boolean bit3() {
    UInt256 uInt256 = UInt256.fromBytes(this.arg3.getBytes32());
    return UInt256.ONE.compareTo(uInt256) >= 0;
  }

  /** Returns true if any of the bit1, bit2, or bit3 flags are set. */
  private boolean isOneLineInstruction() {
    return bit1() || bit2() || bit3();
  }

  public boolean isOli() {
    return oli;
  }

  /**
   * Computes and stores delta values and comparison flags for pairs of elements in two arrays. The
   * comparison flag indicates whether the value in cBytes is greater than, less than, or equal to
   * the value in rBytes. The delta value is the absolute difference between the two values, minus
   * one if cBytes is greater than rBytes. The method computes the delta and comparison flag for
   * each pair of elements and stores the delta values as byte arrays in deltaBytes.
   */
  private void computeDeltasAndComparisonFlags(BaseTheta cBytes, BaseTheta rBytes) {
    ExtCalculator extCalculator =
        new ExtCalculator(arg1.getBytes32(), arg2.getBytes32(), arg3.getBytes32());
    cmp = extCalculator.computeComparisonFlags(cBytes, rBytes);
    deltaBytes = extCalculator.computeDeltas(cBytes, rBytes);
  }

  /**
   * Sets the values of jBytes and qBytes based on the opcode and arguments in this extData object.
   * If the opcode is MULMOD, computes the product of arg1 and arg2, divides the product by arg3,
   * and stores the result in qBytes. If the product is less than 2^256, stores the quotient in
   * qBytes. If the product is greater than or equal to 2^256, stores the quotient in qBytes and the
   * product in jBytes. If the opcode is ADDMOD, computes the sum of arg1 and arg2, divides the sum
   * by arg3, and stores the result in qBytes. If the sum is less than 2^256, stores the quotient in
   * qBytes. If the sum is greater than or equal to 2^256, stores the quotient in qBytes and the sum
   * in jBytes.
   */
  private void setJsAndQs() {
    if (this.opCode == OpCode.MULMOD) {
      setJsAndQsForMulModOpcode();
    } else if (this.opCode == OpCode.ADDMOD) {
      setJsAndQsForAddMod();
    }
  }

  private void setJsAndQsForMulModOpcode() {
    ExtCalculator extCalculator =
        new ExtCalculator(arg1.getBytes32(), arg2.getBytes32(), arg3.getBytes32());
    jBytes = extCalculator.computeJs();
    qBytes = extCalculator.computeQs();
  }

  private void setJsAndQsForAddMod() {

    /* BigInteger sum = this.arg1.add(this.arg2);
    mod.BaseTheta sumBaseTheta = new mod.BaseTheta(sum.toByteArray());
    for (int k = 0; k < sumBaseTheta.length; k++) {
      ed.jBytes[k] = sumBaseTheta.get(k);
    }

    if (!sum.bitLengthOverflowed()) {
      BigInteger quot = sum.divide(new BigInteger(1, ed.arg3.getBytes32()));
      mod.BaseTheta quotBaseTheta = new mod.BaseTheta(quot.toByteArray());

      for (int i = 0; i < quotBaseTheta.length; i++) {
        ed.qBytes[i] = quotBaseTheta.get(i);
      }
      return;
    }

    ed.jBytes[4] = uint64ToBytes(1);

    long[] sumUint64 = new long[5];
    for (int k = 0; k < 4; k++) {
      sumUint64[k] = sum.longValue();
      sum = sum.shiftRight(64);
    }
    sumUint64[4] = 1;

    long[] quot = new long[8];
    i_udivrem(quot, sumUint64, ed.arg3);

    for (int k = 0; k < quot.length; k++) {
      ed.qBytes[k] = uint64ToBytes(quot[k]);
    }*/
  }

  @SuppressWarnings("UnusedMethod")
  private boolean isUInt256(BigInteger number) {
    return number.bitLength() <= 256;
  }

  /**
   * Converts a 64-bit unsigned integer into an 8-byte array
   *
   * @param x The 64-bit unsigned integer to be converted.
   * @return An 8-byte array representing the input value.
   */
  public static byte[] uInt64ToBytes(long x) {
    byte[] xBytes = new byte[8];
    for (int k = 0; k < 8; k++) {
      xBytes[7 - k] = (byte) (x % 256);
      x >>= 8;
    }
    return xBytes;
  }

  private void setHsAndOverflowH() {

    UInt256 prodA0xB1 = UInt256.fromBytes(aBytes.get(0)).multiply(UInt256.fromBytes(bBytes.get(1)));
    UInt256 prodA1xB0 = UInt256.fromBytes(aBytes.get(1)).multiply(UInt256.fromBytes(bBytes.get(0)));
    UInt256 sum1 = prodA0xB1.add(prodA1xB0);
    var truc = BaseTheta.fromBytes32(sum1);
    hBytes.set(0, truc.get(0));
    hBytes.set(1, truc.get(1));

    long alpha = getOverflow(sum1, 1, "alpha OOB");

    UInt256 prodA0xB3 = UInt256.fromBytes(aBytes.get(0)).multiply(UInt256.fromBytes(bBytes.get(3)));
    UInt256 prodA1xB2 = UInt256.fromBytes(aBytes.get(1)).multiply(UInt256.fromBytes(bBytes.get(2)));
    UInt256 prodA2xB1 = UInt256.fromBytes(aBytes.get(2)).multiply(UInt256.fromBytes(bBytes.get(1)));
    UInt256 prodA3xB0 = UInt256.fromBytes(aBytes.get(3)).multiply(UInt256.fromBytes(bBytes.get(0)));

    UInt256 sum2 = prodA0xB3.add(prodA1xB2).add(prodA2xB1).add(prodA3xB0);
    var truc2 = BaseTheta.fromBytes32(sum2);
    hBytes.set(2, truc2.get(0));
    hBytes.set(3, truc2.get(1));

    long beta = getOverflow(sum2, 3, "beta OOB");

    UInt256 prodA2xB3 = UInt256.fromBytes(aBytes.get(2)).multiply(UInt256.fromBytes(bBytes.get(3)));
    UInt256 prodA3xB2 = UInt256.fromBytes(aBytes.get(3)).multiply(UInt256.fromBytes(bBytes.get(2)));
    UInt256 sum3 = prodA2xB3.add(prodA3xB2);

    var truc3 = BaseTheta.fromBytes32(sum3);
    hBytes.set(4, truc3.get(0));
    hBytes.set(5, truc3.get(1));

    long gamma = getOverflow(sum2, 1, "gamma OOB");

    hBytes.set(2, truc.get(0));
    hBytes.set(3, truc.get(1));

    overflowH[0] = getBit(alpha, 0);
    overflowH[1] = getBit(beta, 0);
    overflowH[2] = getBit(beta, 1);
    overflowH[3] = getBit(gamma, 0);
  }

  private void setIsAndOverflowI() {

    UInt256 sumSigma = multiplyRange(qBytes.getBytesRange(0, 1), cBytes.getBytesRange(0, 1));
    BaseTheta thetaSigma = BaseTheta.fromBytes32(sumSigma);
    iBytes.set(0, thetaSigma.get(0));
    iBytes.set(1, thetaSigma.get(1));
    long sigma = getOverflow(sumSigma, 1, "sigma OOB");

    UInt256 sumTau = multiplyRange(qBytes.getBytesRange(0, 3), cBytes.getBytesRange(0,3));
    BaseTheta thetaTau = BaseTheta.fromBytes32(sumTau);
    iBytes.set(2, thetaTau.get(0));
    iBytes.set(3, thetaTau.get(1));
    long tau = getOverflow(UInt256.fromBytes(sumTau), 3, "tau OOB");

    UInt256 sumRho = multiplyRange(qBytes.getBytesRange(2, 5), cBytes.getBytesRange(0,3));
    BaseTheta thetaRho = BaseTheta.fromBytes32(sumRho);
    iBytes.set(4, thetaRho.get(0));
    iBytes.set(5, thetaRho.get(1));
    long rho = getOverflow(sumRho, 3, "rho OOB");

    UInt256 lastSum = multiplyRange(qBytes.getBytesRange(4, 7), cBytes.getBytesRange(0,3));
    BaseTheta lastTheta = BaseTheta.fromBytes32(lastSum);
    iBytes.set(6, lastTheta.get(0));

    overflowI[0] = getBit(sigma, 0);
    overflowI[1] = getBit(tau, 0);
    overflowI[2] = getBit(tau, 1);
    overflowI[4] = getBit(rho, 0);
    overflowI[5] = getBit(rho, 1);
  }

  private UInt256 multiplyRange(Bytes[] range1, Bytes[] range2) {
    checkArgument(range1.length == range2.length);
    UInt256 sum = UInt256.ZERO;
    for (int i = 0; i < range1.length; i++) {
      UInt256 prod =
          UInt256.fromBytes(range1[0]).multiply(UInt256.fromBytes(range2[range2.length - i - 1]));
      sum = sum.add(prod);
    }
    return sum;
  }

  /*

    // setHsAndOverflowH sets
  //   - hBytes
  //   - OF_H
  func (ed *extData) setHsAndOverflowH() {

  	a, b, prod, sum := new(uint256.Int), new(uint256.Int), new(uint256.Int), new(uint256.Int)
  	var truc mod.BaseTheta

  	sum.Set(prod.Mul(a.SetBytes(ed.aBytes[0][:]), b.SetBytes(ed.bBytes[1][:])))      // sum := A0 * B1
  	sum.Add(sum, prod.Mul(a.SetBytes(ed.aBytes[1][:]), b.SetBytes(ed.bBytes[0][:]))) // sum += A1 * B0

  	truc.Set(sum)
  	ed.hBytes[0] = truc[0]
  	ed.hBytes[1] = truc[1]

  	alpha := common.GetOverflow(sum, 1, "alpha OOB")

  	sum.Set(prod.Mul(a.SetBytes(ed.aBytes[0][:]), b.SetBytes(ed.bBytes[3][:])))      // sum := A0 * B3
  	sum.Add(sum, prod.Mul(a.SetBytes(ed.aBytes[1][:]), b.SetBytes(ed.bBytes[2][:]))) // sum += A1 * B2
  	sum.Add(sum, prod.Mul(a.SetBytes(ed.aBytes[2][:]), b.SetBytes(ed.bBytes[1][:]))) // sum += A2 * B1
  	sum.Add(sum, prod.Mul(a.SetBytes(ed.aBytes[3][:]), b.SetBytes(ed.bBytes[0][:]))) // sum += A3 * B0

  	truc.Set(sum)
  	ed.hBytes[2] = truc[0]
  	ed.hBytes[3] = truc[1]

  	beta := common.GetOverflow(sum, 3, "beta OOB")

  	sum.Set(prod.Mul(a.SetBytes(ed.aBytes[3][:]), b.SetBytes(ed.bBytes[2][:])))      // sum := A3 * B2
  	sum.Add(sum, prod.Mul(a.SetBytes(ed.aBytes[2][:]), b.SetBytes(ed.bBytes[3][:]))) // sum := A2 * B3

  	truc.Set(sum)
  	ed.hBytes[4] = truc[0]
  	ed.hBytes[5] = truc[1]

  	gamma := common.GetOverflow(sum, 1, "gamma OOB")

  	ed.overflowH[0] = common.GetBit(alpha, 0)
  	ed.overflowH[1] = common.GetBit(beta, 0)
  	ed.overflowH[2] = common.GetBit(beta, 1)
  	ed.overflowH[3] = common.GetBit(gamma, 0)
  }

     */
  private void setOverflowJ() {}

  private void setOverflowRes() {}

  public BaseTheta getABytes() {
    return aBytes;
  }

  public BaseTheta getBBytes() {
    return bBytes;
  }

  public BaseTheta getCBytes() {
    return cBytes;
  }

  public BaseTheta getDeltaBytes() {
    return deltaBytes;
  }

  public BytesArray getHBytes() {
    return hBytes;
  }

  public BytesArray getIBytes() {
    return iBytes;
  }

  public BytesArray getJBytes() {
    return jBytes;
  }

  public BytesArray getQBytes() {
    return qBytes;
  }

  public BaseTheta getRBytes() {
    return rBytes;
  }

  public BaseBytes getArg1() {
    return arg1;
  }

  public BaseBytes getArg2() {
    return arg2;
  }

  public BaseBytes getArg3() {
    return arg3;
  }

  public BaseBytes getResult() {
    return result;
  }

  public boolean[] getCmp() {
    return cmp;
  }

  public boolean[] getOverflowH() {
    return overflowH;
  }
}
