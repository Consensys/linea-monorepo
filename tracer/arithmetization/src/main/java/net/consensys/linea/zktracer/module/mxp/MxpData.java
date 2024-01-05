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

package net.consensys.linea.zktracer.module.mxp;

import static net.consensys.linea.zktracer.module.Util.max;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static org.hyperledger.besu.evm.internal.Words.clampedAdd;
import static org.hyperledger.besu.evm.internal.Words.clampedMultiply;

import java.math.BigInteger;
import java.util.Arrays;

import com.google.common.base.Preconditions;
import lombok.Getter;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.gas.BillingRate;
import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import net.consensys.linea.zktracer.opcode.gas.MxpType;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@Getter
public class MxpData extends ModuleOperation {
  public static final BigInteger TWO_POW_128 = BigInteger.ONE.shiftLeft(128);
  public static final BigInteger TWO_POW_32 = BigInteger.ONE.shiftLeft(32);

  // constants from protocol_params.go

  private final OpCodeData opCodeData;
  private final int contextNumber;
  private EWord offset1 = EWord.ZERO;
  private EWord offset2 = EWord.ZERO;
  private EWord size1 = EWord.ZERO;
  private EWord size2 = EWord.ZERO;
  private BigInteger maxOffset1 = BigInteger.ZERO;
  private BigInteger maxOffset2 = BigInteger.ZERO;
  private BigInteger maxOffset = BigInteger.ZERO;
  private boolean mxpx;
  private boolean roob;
  private boolean noOperation;
  private boolean comp;
  private BigInteger acc1 = BigInteger.ZERO;
  private BigInteger acc2 = BigInteger.ZERO;
  private BigInteger acc3 = BigInteger.ZERO;
  private BigInteger acc4 = BigInteger.ZERO;
  private BigInteger accA = BigInteger.ZERO;
  private BigInteger accW = BigInteger.ZERO;
  private BigInteger accQ = BigInteger.ZERO;
  private UnsignedByte[] byte1;
  private UnsignedByte[] byte2;
  private UnsignedByte[] byte3;
  private UnsignedByte[] byte4;
  private UnsignedByte[] byteA;
  private UnsignedByte[] byteW;
  private UnsignedByte[] byteQ;
  private UnsignedByte[] byteQQ;
  private UnsignedByte[] byteR;
  private boolean expands;
  private final MxpType typeMxp;
  private final long words;
  private long wordsNew;
  private final long cMem;
  private long cMemNew;
  private long quadCost = 0;
  private long linCost = 0;
  private final boolean deploys;

  public MxpData(final MessageFrame frame, final Hub hub) {
    this.opCodeData = hub.opCodeData();
    this.contextNumber = hub.currentFrame().contextNumber();
    this.typeMxp = opCodeData.billing().type();

    this.words = frame.memoryWordSize();
    this.wordsNew = frame.memoryWordSize();
    this.cMem = memoryCost(frame.memoryWordSize());
    this.cMemNew = memoryCost(frame.memoryWordSize());
    this.deploys = hub.currentFrame().underDeployment();

    setOffsetsAndSizes(frame);
    setRoob();
    setNoOperation();
    setMaxOffset1and2();
    setMaxOffsetAndMxpx();
    setInitializeByteArrays();
    setAccAAndFirstTwoBytesOfByteR();
    setExpands();
    setWordsNew(frame);
  }

  @Override
  protected int computeLineCount() {
    return this.maxCt();
  }

  void compute() {
    setCMemNew();
    setComp();
    setAcc1and2();
    setAcc3();
    setAcc4();
    setAccWAndLastTwoBytesOfByteR();
    setAccQAndByteQQ();
    setBytes();
    setCosts();
  }

  private void setInitializeByteArrays() {
    byte1 = new UnsignedByte[maxCt()];
    byte2 = new UnsignedByte[maxCt()];
    byte3 = new UnsignedByte[maxCt()];
    byte4 = new UnsignedByte[maxCt()];
    byteA = new UnsignedByte[maxCt()];
    byteW = new UnsignedByte[maxCt()];
    byteQ = new UnsignedByte[maxCt()];
    byteQQ = new UnsignedByte[maxCt()];
    byteR = new UnsignedByte[maxCt()];
    Arrays.fill(byte1, UnsignedByte.of(0));
    Arrays.fill(byte2, UnsignedByte.of(0));
    Arrays.fill(byte3, UnsignedByte.of(0));
    Arrays.fill(byte4, UnsignedByte.of(0));
    Arrays.fill(byteA, UnsignedByte.of(0));
    Arrays.fill(byteW, UnsignedByte.of(0));
    Arrays.fill(byteQ, UnsignedByte.of(0));
    Arrays.fill(byteQQ, UnsignedByte.of(0));
    Arrays.fill(byteR, UnsignedByte.of(0));
  }

  private void setOffsetsAndSizes(final MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    switch (opCode) {
      case SHA3, LOG0, LOG1, LOG2, LOG3, LOG4, RETURN, REVERT -> {
        offset1 = EWord.of(frame.getStackItem(0));
        size1 = EWord.of(frame.getStackItem(1));
      }
      case MSIZE -> {}
      case CALLDATACOPY, CODECOPY, RETURNDATACOPY -> {
        offset1 = EWord.of(frame.getStackItem(0));
        size1 = EWord.of(frame.getStackItem(2));
      }
      case EXTCODECOPY -> {
        offset1 = EWord.of(frame.getStackItem(1));
        size1 = EWord.of(frame.getStackItem(3));
      }
      case MLOAD, MSTORE, MSTORE8 -> offset1 = EWord.of(frame.getStackItem(0));
      case CREATE, CREATE2 -> {
        offset1 = EWord.of(frame.getStackItem(1));
        size1 = EWord.of(frame.getStackItem(2));
      }
      case CALL, CALLCODE -> {
        offset1 = EWord.of(frame.getStackItem(3));
        size1 = EWord.of(frame.getStackItem(4));
        offset2 = EWord.of(frame.getStackItem(5));
        size2 = EWord.of(frame.getStackItem(6));
      }
      case DELEGATECALL, STATICCALL -> {
        offset1 = EWord.of(frame.getStackItem(2));
        size1 = EWord.of(frame.getStackItem(3));
        offset2 = EWord.of(frame.getStackItem(4));
        size2 = EWord.of(frame.getStackItem(5));
      }
      default -> throw new IllegalStateException("Unexpected value: " + opCode);
    }
  }

  /** get ridiculously out of bounds. */
  protected void setRoob() {
    roob =
        switch (typeMxp) {
          case TYPE_2, TYPE_3 -> offset1.toBigInteger().compareTo(TWO_POW_128) >= 0;
          case TYPE_4 -> size1.toBigInteger().compareTo(TWO_POW_128) >= 0
              || (offset1.toBigInteger().compareTo(TWO_POW_128) >= 0
                  && !size1.toBigInteger().equals(BigInteger.ZERO));
          case TYPE_5 -> size1.toBigInteger().compareTo(TWO_POW_128) >= 0
              || (offset1.toBigInteger().compareTo(TWO_POW_128) >= 0
                  && !size1.toBigInteger().equals(BigInteger.ZERO))
              || (size2.toBigInteger().compareTo(TWO_POW_128) >= 0
                  || (offset2.toBigInteger().compareTo(TWO_POW_128) >= 0
                      && !size2.toBigInteger().equals(BigInteger.ZERO)));
          default -> false;
        };
  }

  /** get no op. */
  protected void setNoOperation() {
    if (roob) {
      noOperation = false;
    }
    noOperation =
        switch (typeMxp) {
          case TYPE_1 -> true;
          case TYPE_4 -> size1.isZero();
          case TYPE_5 -> size1.isZero() && size2.isZero();
          default -> false;
        };
  }

  /** set max offsets 1 and 2. */
  protected void setMaxOffset1and2() {
    if (getMxpExecutionPath() != mxpExecutionPath.TRIVIAL) {
      switch (typeMxp) {
        case TYPE_2 -> maxOffset1 = offset1.toBigInteger().add(BigInteger.valueOf(31));
        case TYPE_3 -> maxOffset1 = offset1.toBigInteger();
        case TYPE_4 -> {
          if (!size1.toBigInteger().equals(BigInteger.ZERO)) {
            maxOffset1 = offset1.toBigInteger().add(size1.toBigInteger()).subtract(BigInteger.ONE);
          }
        }
        case TYPE_5 -> {
          if (!size1.toBigInteger().equals(BigInteger.ZERO)) {
            maxOffset1 = offset1.toBigInteger().add(size1.toBigInteger()).subtract(BigInteger.ONE);
          }
          if (!size2.toBigInteger().equals(BigInteger.ZERO)) {
            maxOffset2 = offset2.toBigInteger().add(size2.toBigInteger()).subtract(BigInteger.ONE);
          }
        }
      }
    }
  }

  /** set max offset and mxpx. */
  protected void setMaxOffsetAndMxpx() {
    if (roob || noOperation) {
      mxpx = roob;
    } else {
      // choose the max value
      maxOffset = max(maxOffset1, maxOffset2);
      mxpx = maxOffset.compareTo(TWO_POW_32) >= 0;
    }
  }

  public void setExpands() {
    if (!roob && !noOperation && !mxpx) {
      expands = accA.compareTo(BigInteger.valueOf(words)) > 0;
    }
  }

  // This is a copy and past from FrontierGasCalculator.java
  private static long memoryCost(final long length) {
    final long lengthSquare = clampedMultiply(length, length);
    final long base =
        (lengthSquare == Long.MAX_VALUE)
            ? clampedMultiply(length / 512, length)
            : lengthSquare / 512;

    return clampedAdd(clampedMultiply(GasConstants.G_MEMORY.cost(), length), base);
  }

  private long getLinCost(OpCodeData opCodeData, long sizeInBytes) {
    if (getMxpExecutionPath() == mxpExecutionPath.NON_TRIVIAL) {
      if (opCodeData.billing().billingRate() == BillingRate.BY_BYTE) {
        return opCodeData.billing().perUnit().cost() * sizeInBytes;
      }
      if (opCodeData.billing().billingRate() == BillingRate.BY_WORD) {
        long sizeInWords = (sizeInBytes + 31) / 32;
        return opCodeData.billing().perUnit().cost() * sizeInWords;
      }
    }
    return 0;
  }

  protected void setComp() {
    comp = maxOffset1.compareTo(maxOffset2) >= 0;
  }

  // This should translate into code page 12, point 5?
  protected void setAccAAndFirstTwoBytesOfByteR() {
    if (this.getMxpExecutionPath() == mxpExecutionPath.NON_TRIVIAL) {
      BigInteger maxOffsetPlusOne = maxOffset.add(BigInteger.ONE);
      accA = divRoundedUpBigInteger(maxOffsetPlusOne, BigInteger.valueOf(32));
      BigInteger diff = accA.multiply(BigInteger.valueOf(32)).subtract(maxOffsetPlusOne);
      Bytes32 diffBytes = Bytes32.leftPad(UInt256.valueOf(diff));

      UnsignedByte r = UnsignedByte.of(diffBytes.get(31));
      int rAsInt = r.toInteger();
      byteR[0] = UnsignedByte.of(rAsInt + 224);
      byteR[1] = UnsignedByte.of(rAsInt);
    }
  }

  public BigInteger divRoundedUpBigInteger(BigInteger a, BigInteger b) {
    return a.add(b.subtract(BigInteger.ONE)).divide(b);
  }

  protected void setAcc1and2() {
    if (roob) {
      return;
    }
    if (mxpx) {
      if (maxOffset1.compareTo(TWO_POW_32) >= 0) {
        acc1 = maxOffset1.subtract(TWO_POW_32);
      } else {
        Preconditions.checkArgument(maxOffset2.compareTo(TWO_POW_32) >= 0);
        acc2 = maxOffset2.subtract(TWO_POW_32);
      }
    } else {
      acc1 = maxOffset1;
      acc2 = maxOffset2;
    }
  }

  protected void setAcc3() {
    if (comp) {
      acc3 = maxOffset1.subtract(maxOffset2);
    } else {
      acc3 = maxOffset2.subtract(maxOffset1);
      acc3 = acc3.subtract(BigInteger.ONE);
    }
  }

  protected void setAcc4() {
    if (this.getMxpExecutionPath() == mxpExecutionPath.NON_TRIVIAL) {
      if (expands) {
        acc4 = accA.subtract(BigInteger.valueOf(words + 1));
      } else {
        acc4 = BigInteger.valueOf(words).subtract(accA);
      }
    }
  }

  protected void setAccWAndLastTwoBytesOfByteR() {
    // @TODO: do it also for other cases
    if (this.getMxpExecutionPath() == mxpExecutionPath.NON_TRIVIAL) {
      if (typeMxp != MxpType.TYPE_4) {
        return;
      }

      accW = size1.toBigInteger().add(BigInteger.valueOf(31)).divide(BigInteger.valueOf(32));

      BigInteger r = accW.multiply(BigInteger.valueOf(32)).subtract(size1.toBigInteger());

      // r in [0,31]
      UnsignedByte rByte = UnsignedByte.of(r.toByteArray()[r.toByteArray().length - 1]);
      int rByteAsInt = rByte.toInteger();
      byteR[2] = UnsignedByte.of(rByteAsInt + 224);
      byteR[3] = rByte;
    }
  }

  protected enum mxpExecutionPath {
    TRIVIAL,
    NON_TRIVIAL_BUT_MXPX,
    NON_TRIVIAL
  }

  private mxpExecutionPath getMxpExecutionPath() {
    if (this.isRoob() || this.isNoOperation()) {
      return mxpExecutionPath.TRIVIAL;
    }
    if (this.isMxpx()) {
      return mxpExecutionPath.NON_TRIVIAL_BUT_MXPX;
    }
    return mxpExecutionPath.NON_TRIVIAL;
  }

  public int maxCt() {
    return switch (this.getMxpExecutionPath()) {
      case TRIVIAL -> 1;
      case NON_TRIVIAL_BUT_MXPX -> 17;
      case NON_TRIVIAL -> 4;
    };
  }

  protected void setAccQAndByteQQ() {
    // accQ and byteQQ all equal to 0 by default
    if (this.getMxpExecutionPath() == mxpExecutionPath.NON_TRIVIAL) {
      setAccQAndByteQQNonTrivialCase();
    }
  }

  protected void setAccQAndByteQQNonTrivialCase() {
    long square = wordsNew * wordsNew; // ACC_A

    long quotient = square / 512;
    long remainder = square % 512;

    accQ = BigInteger.valueOf(quotient % (1L << 32));

    Bytes32 quotientBytes = UInt256.valueOf(quotient); // q'
    Bytes32 remainderBytes = UInt256.valueOf(remainder); // r'

    byteQQ[0] = UnsignedByte.of(quotientBytes.get(quotientBytes.size() - 6));
    byteQQ[1] = UnsignedByte.of(quotientBytes.get(quotientBytes.size() - 5));
    byteQQ[2] = UnsignedByte.of(remainderBytes.get(quotientBytes.size() - 2));
    byteQQ[3] = UnsignedByte.of(remainderBytes.get(quotientBytes.size() - 1));
  }

  protected void setBytes() {
    int maxCt = maxCt();
    Bytes32 b1 = UInt256.valueOf(acc1);
    Bytes32 b2 = UInt256.valueOf(acc2);
    Bytes32 b3 = UInt256.valueOf(acc3);
    Bytes32 b4 = UInt256.valueOf(acc4);
    Bytes32 bA = UInt256.valueOf(accA);
    Bytes32 bW = UInt256.valueOf(accW);
    Bytes32 bQ = UInt256.valueOf(accQ);
    for (int i = 0; i < maxCt; i++) {
      byte1[i] = UnsignedByte.of(b1.get(b1.size() - 1 - maxCt + i));
      byte2[i] = UnsignedByte.of(b2.get(b2.size() - 1 - maxCt + i));
      byte3[i] = UnsignedByte.of(b3.get(b3.size() - 1 - maxCt + i));
      byte4[i] = UnsignedByte.of(b4.get(b4.size() - 1 - maxCt + i));
      byteA[i] = UnsignedByte.of(bA.get(bA.size() - 1 - maxCt + i));
      byteW[i] = UnsignedByte.of(bW.get(bW.size() - 1 - maxCt + i));
      byteQ[i] = UnsignedByte.of(bQ.get(bQ.size() - 1 - maxCt + i));
    }
  }

  private void setWordsNew(final MessageFrame frame) {
    if (getMxpExecutionPath() == MxpData.mxpExecutionPath.NON_TRIVIAL && expands) {
      switch (getTypeMxp()) {
        case TYPE_1 -> wordsNew = frame.calculateMemoryExpansion(Words.clampedToLong(offset1), 0);
        case TYPE_2 -> wordsNew = frame.calculateMemoryExpansion(Words.clampedToLong(offset1), 32);
        case TYPE_3 -> wordsNew = frame.calculateMemoryExpansion(Words.clampedToLong(offset1), 1);
        case TYPE_4 -> wordsNew =
            frame.calculateMemoryExpansion(
                Words.clampedToLong(offset1), Words.clampedToLong(size1));
        case TYPE_5 -> {
          long wordsNew1 =
              frame.calculateMemoryExpansion(
                  Words.clampedToLong(offset1), Words.clampedToLong(size1));
          long wordsNew2 =
              frame.calculateMemoryExpansion(
                  Words.clampedToLong(offset2), Words.clampedToLong(size2));
          wordsNew = Math.max(wordsNew1, wordsNew2);
        }
      }
    }
  }

  private void setCMemNew() {
    if (getMxpExecutionPath() == MxpData.mxpExecutionPath.NON_TRIVIAL && expands) {
      cMemNew = memoryCost(wordsNew);
    }
  }

  private void setCosts() {
    if (getMxpExecutionPath() == mxpExecutionPath.NON_TRIVIAL) {
      quadCost = cMemNew - cMem;
      linCost = getLinCost(opCodeData, Words.clampedToLong(size1));
    }
  }

  long getEffectiveLinCost() {
    if (opCodeData.mnemonic() != OpCode.RETURN) {
      return getLinCost(opCodeData, Words.clampedToLong(size1));
    } else {
      if (deploys) {
        return getLinCost(opCodeData, Words.clampedToLong(size1));
      } else {
        return 0;
      }
    }
  }

  final void trace(int stamp, Trace trace) {
    this.compute();

    Bytes32 acc1Bytes32 = Bytes32.leftPad(bigIntegerToBytes(this.getAcc1()));
    Bytes32 acc2Bytes32 = Bytes32.leftPad(bigIntegerToBytes(this.getAcc2()));
    Bytes32 acc3Bytes32 = Bytes32.leftPad(bigIntegerToBytes(this.getAcc3()));
    Bytes32 acc4Bytes32 = Bytes32.leftPad(bigIntegerToBytes(this.getAcc4()));
    Bytes32 accABytes32 = Bytes32.leftPad(bigIntegerToBytes(this.getAccA()));
    Bytes32 accWBytes32 = Bytes32.leftPad(bigIntegerToBytes(this.getAccW()));
    Bytes32 accQBytes32 = Bytes32.leftPad(bigIntegerToBytes(this.getAccQ()));
    final EWord eOffset1 = EWord.of(this.offset1);
    final EWord eOffset2 = EWord.of(this.offset2);
    final EWord eSize1 = EWord.of(this.size1);
    final EWord eSize2 = EWord.of(this.size2);

    int maxCt = this.maxCt();
    int maxCtComplement = 32 - maxCt;

    for (int i = 0; i < maxCt; i++) {
      trace
          .stamp(Bytes.ofUnsignedLong(stamp))
          .cn(Bytes.ofUnsignedLong(this.getContextNumber()))
          .ct(Bytes.of(i))
          .roob(this.isRoob())
          .noop(this.isNoOperation())
          .mxpx(this.isMxpx())
          .inst(Bytes.of(this.getOpCodeData().value()))
          .mxpType1(this.getOpCodeData().billing().type() == MxpType.TYPE_1)
          .mxpType2(this.getOpCodeData().billing().type() == MxpType.TYPE_2)
          .mxpType3(this.getOpCodeData().billing().type() == MxpType.TYPE_3)
          .mxpType4(this.getOpCodeData().billing().type() == MxpType.TYPE_4)
          .mxpType5(this.getOpCodeData().billing().type() == MxpType.TYPE_5)
          .gword(
              Bytes.ofUnsignedLong(
                  this.getOpCodeData().billing().billingRate() == BillingRate.BY_WORD
                      ? this.getOpCodeData().billing().perUnit().cost()
                      : 0))
          .gbyte(
              Bytes.ofUnsignedLong(
                  this.getOpCodeData().billing().billingRate() == BillingRate.BY_BYTE
                      ? this.getOpCodeData().billing().perUnit().cost()
                      : 0))
          .deploys(this.isDeploys())
          .offset1Hi(eOffset1.hi())
          .offset1Lo(eOffset1.lo())
          .offset2Hi(eOffset2.hi())
          .offset2Lo(eOffset2.lo())
          .size1Hi(eSize1.hi())
          .size1Lo(eSize1.lo())
          .size2Hi(eSize2.hi())
          .size2Lo(eSize2.lo())
          .maxOffset1(bigIntegerToBytes(this.getMaxOffset1()))
          .maxOffset2(bigIntegerToBytes(this.getMaxOffset2()))
          .maxOffset(bigIntegerToBytes(this.getMaxOffset()))
          .comp(this.isComp())
          .acc1(acc1Bytes32.slice(maxCtComplement, 1 + i))
          .acc2(acc2Bytes32.slice(maxCtComplement, 1 + i))
          .acc3(acc3Bytes32.slice(maxCtComplement, 1 + i))
          .acc4(acc4Bytes32.slice(maxCtComplement, 1 + i))
          .accA(accABytes32.slice(maxCtComplement, 1 + i))
          .accW(accWBytes32.slice(maxCtComplement, 1 + i))
          .accQ(accQBytes32.slice(maxCtComplement, 1 + i))
          .byte1(UnsignedByte.of(acc1Bytes32.get(maxCtComplement + i)))
          .byte2(UnsignedByte.of(acc2Bytes32.get(maxCtComplement + i)))
          .byte3(UnsignedByte.of(acc3Bytes32.get(maxCtComplement + i)))
          .byte4(UnsignedByte.of(acc4Bytes32.get(maxCtComplement + i)))
          .byteA(UnsignedByte.of(accABytes32.get(maxCtComplement + i)))
          .byteW(UnsignedByte.of(accWBytes32.get(maxCtComplement + i)))
          .byteQ(UnsignedByte.of(accQBytes32.get(maxCtComplement + i)))
          .byteQq(Bytes.ofUnsignedLong(this.getByteQQ()[i].toInteger()))
          .byteR(Bytes.ofUnsignedLong(this.getByteR()[i].toInteger()))
          .words(Bytes.ofUnsignedLong(this.getWords()))
          .wordsNew(
              Bytes.ofUnsignedLong(
                  this.getWordsNew())) // TODO: Could (should?) be set in tracePostOp?
          .cMem(Bytes.ofUnsignedLong(this.getCMem())) // Returns current memory size in EVM words
          .cMemNew(Bytes.ofUnsignedLong(this.getCMemNew()))
          .quadCost(Bytes.ofUnsignedLong(this.getQuadCost()))
          .linCost(Bytes.ofUnsignedLong(this.getLinCost()))
          .gasMxp(Bytes.ofUnsignedLong(this.getQuadCost() + this.getEffectiveLinCost()))
          .expands(this.isExpands())
          .validateRow();
    }
  }
}
