package net.consensys.linea.zktracer.module.alu.mul;

import java.lang.reflect.Array;
import java.math.BigInteger;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.BytesBaseTheta;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.apache.tuweni.units.bigints.UInt64;

@SuppressWarnings("UnusedVariable")
public class MulData {
  final OpCode opCode;
  final Bytes32 arg1;
  final Bytes32 arg2;
  final boolean tinyBase;
  final boolean tinyExponent;

  BigInteger resAcc; // accumulator which converges in a series of "square and multiply"'s
  UInt256 expAcc; // accumulator for doubles and adds of the exponent, resets at some point

  final BytesBaseTheta aBytes;
  final BytesBaseTheta bBytes;
  BytesBaseTheta cBytes;
  BytesBaseTheta hBytes;
  boolean snm = false;
  int index;
  boolean[] bits;
  String exponentBits;

  Res res;

  public MulData(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {

    this.opCode = opCode;
    this.arg1 = arg1;
    this.arg2 = arg2;
    this.aBytes = new BytesBaseTheta(arg1);
    this.bBytes = new BytesBaseTheta(arg2);

    // TODO what should these be initialized to (or is this not needed)
    this.cBytes = null;
    this.hBytes = null;
    this.expAcc = UInt256.MIN_VALUE;

    this.res = Res.create(opCode, arg1, arg2); // TODO can we get this from the EVM

    final BigInteger arg1BigInt = UInt256.fromBytes(arg1).toUnsignedBigInteger();
    final BigInteger arg2BigInt = UInt256.fromBytes(arg2).toUnsignedBigInteger();

    this.tinyBase = isTiny(arg1BigInt);
    this.tinyExponent = isTiny(arg2BigInt);

    final Regime regime = getRegime(opCode);
    System.out.println(regime);
    switch (regime) {
      case TRIVIAL_MUL:
        break;
      case NON_TRIVIAL_MUL:
        cBytes = new BytesBaseTheta(res);
        break;
      case EXPONENT_ZERO_RESULT:
        setArraysForZeroResultCase();
        break;
      case EXPONENT_NON_ZERO_RESULT:
        this.exponentBits = arg2.toBigInteger().toString();
        snm = false;
        break;
      case IOTA:
        throw new RuntimeException("alu/mul regime was never set");
    }
  }

  private void setArraysForZeroResultCase() {
    // TODO
  }

  public boolean exponentBit() {
    return '1' == exponentBits.charAt(index);
  }

  public boolean exponentSource() {
    return this.index + 128 >= exponentBits.length();
  }

  public static int twoAdicity(final UInt256 x) {

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

  private enum Regime {
    IOTA,
    TRIVIAL_MUL,
    NON_TRIVIAL_MUL,
    EXPONENT_ZERO_RESULT,
    EXPONENT_NON_ZERO_RESULT
  }

  public boolean isOneLineInstruction() {
    return tinyBase || tinyExponent;
  }

  private Regime getRegime(final OpCode opCode) {

    if (isOneLineInstruction()) return Regime.TRIVIAL_MUL;

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

  private void update() {

    final BigInteger arg1BigInt = UInt256.fromBytes(arg1).toUnsignedBigInteger();
    final BigInteger arg2BigInt = UInt256.fromBytes(arg2).toUnsignedBigInteger();
    if (!snm) {
      // squaring
      setHsAndBits(resAcc, resAcc);
      expAcc = expAcc.add(expAcc);
      resAcc = resAcc.multiply(resAcc);
    } else {
      // multiplying by base
      setHsAndBits(arg1BigInt, resAcc);
      expAcc = expAcc.add(UInt256.ONE);
      resAcc = arg1BigInt.multiply(resAcc);
    }
    cBytes.set(resAcc); // TODO how to get from BigInteger to Bytes32
  }

  private void setHsAndBits(BigInteger a, BigInteger b) {

    // TODO set hBytes and bits[]
    BytesBaseTheta aBaseTheta, bBaseTheta, sumBaseTheta ;

    aBaseTheta.set(a);
    bBaseTheta.set(b);

    UInt256[] aBaseThetaInts = (UInt256[]) Array.newInstance(UInt256.class, 4);
    UInt256[] bBaseThetaInts = (UInt256[]) Array.newInstance(UInt256.class, 4);

    for (int i = 0; i < 4; i++) {
      aBaseThetaInts[i] = UInt256.ZERO;
              bBaseThetaInts[i] = UInt256.ZERO;
              aBaseThetaInts[i].setBytes(aBaseTheta.getChunk(i));
      bBaseThetaInts[i].setBytes(bBaseTheta.getChunk(i));
    }

    UInt256 sum, prod;
    prod = aBaseThetaInts[1].multiply(bBaseThetaInts[0]);
    sum = UInt256.MIN_VALUE.add(prod); // sum := a1 * b0
    prod = aBaseThetaInts[0].multiply(bBaseThetaInts[1]);
    sum = sum.add(prod); // sum += a0 * b1

    sumBaseTheta.set(sum.toBigInteger());
    hBytes.set(0, sumBaseTheta.getChunk(0));
    hBytes.set(1, sumBaseTheta.getChunk(1));
    int alpha = getOverflow(sum, 1, "alpha OOB");

    prod = aBaseThetaInts[3].multiply(bBaseThetaInts[0]);
    sum = UInt256.MIN_VALUE.add(prod); // sum := a3 * b0
    prod = aBaseThetaInts[2].multiply(bBaseThetaInts[1]);
    sum = sum.add(prod); // sum += a2 * b1
    prod = aBaseThetaInts[1].multiply(bBaseThetaInts[2]);
    sum = sum.add(prod); // sum += a1 * b2
    prod = aBaseThetaInts[0].multiply(bBaseThetaInts[3]);
    sum = sum.add(prod); // sum += a0 * b3

    sumBaseTheta.set(sum.toBigInteger());
    hBytes.set(2, sumBaseTheta.getChunk(0));
    hBytes.set(3, sumBaseTheta.getChunk(1));
    int beta = getOverflow(sum, 3, "beta OOB");

    prod = aBaseThetaInts[0].multiply(bBaseThetaInts[0]);
    sum = UInt256.MIN_VALUE.add(prod); // sum := a0 * b0
    prod = hBytes.getChunk(0).shiftLeft(64);
    sum = sum.add(prod);// sum += (h0 << 64)
//    sum.Add(sum, prod.Lsh(prod.SetBytes(hs[0][:]), 64))          // sum += (h0 << 64)

    int eta = getOverflow(sum, 1, "eta OOB");

    sum = UInt256.valueOf(eta);                                          // sum := eta
    sum.Add(sum, prod.SetBytes(hs[1][:]))                        // sum += h1
    sum.Add(sum, prod.Lsh(prod.SetUint64(alpha), 64)) ;           // sum += (alpha << 64)
    prod = aBaseThetaInts[2].multiply(bBaseThetaInts[0]);
    sum = sum.add(prod); // sum += a2 * b0
    prod = aBaseThetaInts[1].multiply(bBaseThetaInts[1]);
    sum = sum.add(prod); // sum += a1 * b1
    prod = aBaseThetaInts[0].multiply(bBaseThetaInts[2]);
    sum = sum.add(prod); // sum += a0 * b2
    sum.Add(sum, prod.Lsh(prod.SetBytes(hs[2][:]), 64))          // sum += (h2 << 64)

    int mu = getOverflow(sum, 3, "mu OOB");

    bits[2] = getBit(alpha, 0);
    bits[3] = getBit(beta, 0);
    bits[4] = getBit(beta, 1);
    bits[5] = getBit(eta, 0);
    bits[6] = getBit(mu, 0);
    bits[7] = getBit(mu, 1);

    return;
  }

  public static int getOverflow(final UInt256 arg, final int maxVal, final String err) {
    UInt256 shiftRight = arg.shiftRight( 128);
    if (shiftRight.toBigInteger().compareTo (UInt64.MAX_VALUE.toBigInteger()) > 0) {
      throw new RuntimeException("getOverflow expects a small high part");
    }
    int overflow = shiftRight.toInt();
    if (overflow > maxVal) {
      throw new RuntimeException(err);
    }
    return overflow;
  }
  // GetBit returns true iff the k'th bit of x is 1
  private boolean getBit(int x, int k) {
    return (x>>k)%2 == 1;
  }
}
