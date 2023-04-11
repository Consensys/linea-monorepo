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
package net.consensys.zktracer.module.alu.mul;

import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.bytes.Bytes16;
import net.consensys.zktracer.bytes.BytesBaseTheta;
import net.consensys.zktracer.bytes.UnsignedByte;
import net.consensys.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.evm.frame.MessageFrame;

import java.math.BigInteger;
import java.util.List;

public class MulTracer implements ModuleTracer {
  private static final int MMEDIUM = 8;

  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "mul";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of(OpCode.MUL, OpCode.EXP);
  }

  @Override
  public Object trace(MessageFrame frame) {
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

    final Bytes16 arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
    final Bytes16 arg1Lo = Bytes16.wrap(arg1.slice(16));
    final Bytes16 arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
    final Bytes16 arg2Lo = Bytes16.wrap(arg2.slice(16));

    final UInt256 arg1Int = UInt256.fromBytes(arg1);
    final UInt256 arg2Int = UInt256.fromBytes(arg2);
    final BigInteger arg1BigInt = arg1Int.toUnsignedBigInteger();
    final BigInteger arg2BigInt = arg2Int.toUnsignedBigInteger();

    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    final boolean tinyBase = isTiny(arg1BigInt);
    final boolean tinyExponent = isTiny(arg2BigInt);

    final boolean isOneLineInstruction = isOneLineInstruction(tinyBase, tinyExponent);
    final Res res = Res.create(opCode, arg1, arg2);

    final MulTrace.Trace.Builder builder = MulTrace.Trace.Builder.newInstance();

    final BytesBaseTheta aBytes = new BytesBaseTheta(arg1);
    final BytesBaseTheta bBytes = new BytesBaseTheta(arg2);
    BytesBaseTheta cBytes = null;
    BytesBaseTheta hBytes ;
    boolean snm = false;

    final Regime regime = getRegime(opCode, tinyBase, tinyExponent, res);
    System.out.println(regime);
    switch (regime) {
      case TRIVIAL_MUL: break;
      case NON_TRIVIAL_MUL:
        cBytes = new BytesBaseTheta(res);
      case EXPONENT_ZERO_RESULT:
        setArraysForZeroResultCase();
      case EXPONENT_NON_ZERO_RESULT:
        setExponentBit();
        snm = false;
      case IOTA: throw new RuntimeException("alu/mul regime was never set");
    }


    stamp++;
    for (int i = 0; i < maxCt(isOneLineInstruction); i++) {
      builder.appendStamp(stamp);
      builder.appendCounter(i);

      builder
              .appendOneLineInstruction(isOneLineInstruction)
              .appendTinyBase(tinyBase)
              .appendTinyExponent(tinyExponent)
              .appendResultVanishes(res.isZero());

      builder
              .appendInst(UnsignedByte.of(opCode.value))
              .appendArg1Hi(arg1Hi.toUnsignedBigInteger())
              .appendArg1Lo(arg1Lo.toUnsignedBigInteger())
              .appendArg2Hi(arg2Hi.toUnsignedBigInteger())
              .appendArg2Lo(arg2Lo.toUnsignedBigInteger());

      builder
              .appendResHi(res.getResHi().toUnsignedBigInteger())
              .appendResLo(res.getResLo().toUnsignedBigInteger());

//      builder.appendBits(bits.get(i)).appendCounter(i); // TODO


      builder
              .appendByteA3(UnsignedByte.of(aBytes.get(3, i)))
              .appendByteA2(UnsignedByte.of(aBytes.get(2, i)))
              .appendByteA1(UnsignedByte.of(aBytes.get(1, i)))
              .appendByteA0(UnsignedByte.of(aBytes.get(0, i)));
      builder
              .appendAccA3(Bytes.of(aBytes.getRange(3, 0, i+1)).toUnsignedBigInteger())
              .appendAccA2(Bytes.of(aBytes.getRange(2, 0, i+1)).toUnsignedBigInteger())
              .appendAccA1(Bytes.of(aBytes.getRange(1, 0, i+1)).toUnsignedBigInteger())
              .appendAccA0(Bytes.of(aBytes.getRange(0, 0, i+1)).toUnsignedBigInteger());

      builder
              .appendByteB3(UnsignedByte.of(bBytes.get(3, i)))
              .appendByteB2(UnsignedByte.of(bBytes.get(2, i)))
              .appendByteB1(UnsignedByte.of(bBytes.get(1, i)))
              .appendByteB0(UnsignedByte.of(bBytes.get(0, i)));
      builder
              .appendAccB3(Bytes.of(bBytes.getRange(3, 0, i+1)).toUnsignedBigInteger())
              .appendAccB2(Bytes.of(bBytes.getRange(2, 0, i+1)).toUnsignedBigInteger())
              .appendAccB1(Bytes.of(bBytes.getRange(1, 0, i+1)).toUnsignedBigInteger())
              .appendAccB0(Bytes.of(bBytes.getRange(0, 0, i+1)).toUnsignedBigInteger());
      builder
              .appendByteC3(UnsignedByte.of(cBytes.get(3, i)))
              .appendByteC2(UnsignedByte.of(cBytes.get(2, i)))
              .appendByteC1(UnsignedByte.of(cBytes.get(1, i)))
              .appendByteC0(UnsignedByte.of(cBytes.get(0, i)));

    }
    builder.setStamp(stamp);

    return builder.build();
  }

  private void setArraysForZeroResultCase() {
    // TODO
  }
  private boolean setExponentBit()  {
    // TODO
    return false;
//    return string(exponentBits[md.index]) == "1";
  }

  public static boolean isTiny(BigInteger arg) {
    return arg.compareTo(BigInteger.valueOf(1)) <= 0;
  }

  private int maxCt(final boolean isOneLineInstruction) {
    return isOneLineInstruction ? 1 : MMEDIUM;
  }

  private boolean isOneLineInstruction(final boolean tinyBase, final boolean tinyExponent) {
    return tinyBase || tinyExponent;
  }

  private enum Regime {
    IOTA,
    TRIVIAL_MUL,
    NON_TRIVIAL_MUL,
    EXPONENT_ZERO_RESULT,
    EXPONENT_NON_ZERO_RESULT
  }
  private Regime getRegime(final OpCode opCode, final boolean tinyBase, final boolean tinyExponent, final Res res) {

    if (isOneLineInstruction(tinyBase, tinyExponent)) return Regime.TRIVIAL_MUL;


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

  public int lineCount(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {


    final UInt256 arg1Int = UInt256.fromBytes(arg1);
    final UInt256 arg2Int = UInt256.fromBytes(arg2);
    final BigInteger arg1BigInt = arg1Int.toUnsignedBigInteger();
    final BigInteger arg2BigInt = arg2Int.toUnsignedBigInteger();

    final boolean tinyBase = isTiny(arg1BigInt);
    final boolean tinyExponent = isTiny(arg2BigInt);

    return maxCt(isOneLineInstruction(tinyBase, tinyExponent));
  }
}
