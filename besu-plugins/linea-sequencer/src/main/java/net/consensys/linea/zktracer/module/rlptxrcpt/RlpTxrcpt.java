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

package net.consensys.linea.zktracer.module.rlptxrcpt;

import java.math.BigInteger;
import java.util.List;
import java.util.function.Function;

import com.google.common.base.Preconditions;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.rlppatterns.RlpBitDecOutput;
import net.consensys.linea.zktracer.module.rlppatterns.RlpByteCountAndPowerOutput;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.data.Log;
import org.hyperledger.besu.plugin.data.TransactionReceipt;

public class RlpTxrcpt implements Module {
  public static final int LLARGE_INTEGER = RlpTxrcptTrace.LLARGE.intValue();
  public static final int INT_SHORT_INTEGER = RlpTxrcptTrace.INT_SHORT.intValue();
  public static final int INT_LONG_INTEGER = RlpTxrcptTrace.INT_LONG.intValue();
  public static final int LIST_SHORT_INTEGER = RlpTxrcptTrace.LIST_SHORT.intValue();
  public static final int LIST_LONG_INTEGER = RlpTxrcptTrace.LIST_LONG.intValue();
  private int absTxNum = 0;
  private int absLogNum = 0;
  private final Trace.TraceBuilder builder = Trace.builder();

  @Override
  public String jsonKey() {
    return "rlpTxRcpt";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of();
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    for (Transaction tx : blockBody.getTransactions()) {
      // this.traceTransaction(tx);
    }
  }

  public void traceTransaction(TransactionReceipt txrcpt, TransactionType txType) {
    this.absTxNum += 1;
    RlpTxrcptColumns traceValue = new RlpTxrcptColumns();
    traceValue.txrcptSize = txRcptSize(txrcpt);

    // PHASE 0: RLP Prefix.
    phase0(traceValue, txType);

    // PHASE 1: Status code Rz.
    phase1(traceValue, txrcpt);

    // PHASE 2: Cumulative gas Ru.
    phase2(traceValue, txrcpt);

    // PHASE 3: Bloom Filter Rb.
    phase3(traceValue, txrcpt);

    // Phase 4: log series Rl.
    phase4(traceValue, txrcpt);
  }

  @Override
  public void traceEndConflation() {
    // Rewrite the ABS_TX_NUM_MAX and ABS_LOG_NUM_MAX columns.
    for (int i = 0; i < this.builder.size(); i++) {
      this.builder.setAbsTxNumMaxAt(BigInteger.valueOf(this.absTxNum), i);
      this.builder.setAbsLogNumMaxAt(BigInteger.valueOf(this.absLogNum), i);
    }
  }

  private void phase0(RlpTxrcptColumns traceValue, TransactionType txType) {

    // byte TYPE concatenation
    traceValue.partialReset(0, 1);
    traceValue.isPrefix = true;

    if (txType == TransactionType.FRONTIER) {
      traceValue.lcCorrection = true;
    } else {
      traceValue.limbConstructed = true;
      traceValue.input1 = Bytes.of(txType.getSerializedType());
      traceValue.limb = BigInteger.valueOf(txType.getSerializedType());
      traceValue.nBytes = 1;
    }

    traceRow(traceValue);

    // RLP prefix of the txRcpt list.
    rlpByteString(0, traceValue.txrcptSize, true, false, false, false, false, true, traceValue);
  }

  private void phase1(RlpTxrcptColumns traceValue, TransactionReceipt txrcpt) {
    traceValue.partialReset(1, 1);
    traceValue.limbConstructed = true;

    if (txrcpt.getStatus() == 0) {
      traceValue.input1 = Bytes.ofUnsignedShort(0);
      traceValue.limb = BigInteger.valueOf(INT_SHORT_INTEGER);
    } else {
      traceValue.input1 = Bytes.ofUnsignedShort(1);
      traceValue.limb = BigInteger.ONE;
    }

    traceValue.nBytes = 1;
    traceValue.phaseEnd = true;

    traceRow(traceValue);
  }

  private void phase2(RlpTxrcptColumns traceValue, TransactionReceipt txrcpt) {
    rlpInt(2, txrcpt.getCumulativeGasUsed(), false, false, false, false, true, false, traceValue);
  }

  private void phase3(RlpTxrcptColumns traceValue, TransactionReceipt txrcpt) {
    // RLP prefix
    traceValue.partialReset(3, 1);
    traceValue.isPrefix = true;
    traceValue.phaseSize = 256;
    traceValue.limbConstructed = true;
    traceValue.limb =
        BigInteger.valueOf(INT_LONG_INTEGER + 2)
            .multiply(BigInteger.valueOf(256))
            .multiply(BigInteger.valueOf(256))
            .add(BigInteger.valueOf(256));
    traceValue.nBytes = 3;

    traceRow(traceValue);

    // Concatenation of Byte slice of the bloom Filter.
    for (int i = 0; i < 4; i++) {
      traceValue.partialReset(3, LLARGE_INTEGER);

      Bytes bloomFilter = txrcpt.getBloomFilter();
      traceValue.input1 = bloomFilter.slice(64 * i, LLARGE_INTEGER);
      traceValue.input2 = bloomFilter.slice(64 * i + LLARGE_INTEGER, LLARGE_INTEGER);
      traceValue.input3 = bloomFilter.slice(64 * i + 2 * LLARGE_INTEGER, LLARGE_INTEGER);
      traceValue.input4 = bloomFilter.slice(64 * i + 3 * LLARGE_INTEGER, LLARGE_INTEGER);

      for (int ct = 0; ct < LLARGE_INTEGER; ct++) {
        traceValue.counter = ct;
        traceValue.byte1 = traceValue.input1.get(ct);
        traceValue.acc1 = traceValue.input1.slice(0, ct + 1);
        traceValue.byte2 = traceValue.input2.get(ct);
        traceValue.acc2 = traceValue.input2.slice(0, ct + 1);
        traceValue.byte3 = traceValue.input3.get(ct);
        traceValue.acc3 = traceValue.input3.slice(0, ct + 1);
        traceValue.byte4 = traceValue.input4.get(ct);
        traceValue.acc4 = traceValue.input4.slice(0, ct + 1);

        switch (ct) {
          case 12 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input1.toUnsignedBigInteger();
            traceValue.nBytes = LLARGE_INTEGER;
          }
          case 13 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input2.toUnsignedBigInteger();
            traceValue.nBytes = LLARGE_INTEGER;
          }
          case 14 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input3.toUnsignedBigInteger();
            traceValue.nBytes = LLARGE_INTEGER;
          }
          case 15 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input4.toUnsignedBigInteger();
            traceValue.nBytes = LLARGE_INTEGER;
            traceValue.phaseEnd = (i == 3);
          }
          default -> {
            traceValue.limbConstructed = false;
            traceValue.limb = BigInteger.ZERO;
            traceValue.nBytes = 0;
          }
        }
        traceRow(traceValue);
        // Update INDEX_LOCAL after the row when LIMB is constructed.
        if (traceValue.limbConstructed) {
          traceValue.indexLocal += 1;
        }
      }
    }
    // Put to 0 INDEX_LOCAL at the end of the phase.
    traceValue.indexLocal = 0;
  }

  private void phase4(RlpTxrcptColumns traceValue, TransactionReceipt txrcpt) {
    // Trivial case, there are no log entries.
    if (txrcpt.getLogs().isEmpty()) {
      traceValue.partialReset(4, 1);
      traceValue.isPrefix = true;
      traceValue.limbConstructed = true;
      traceValue.limb = BigInteger.valueOf(LIST_SHORT_INTEGER);
      traceValue.nBytes = 1;
      traceValue.phaseEnd = true;

      traceRow(traceValue);
    } else {
      // RLP prefix of the list of log entries.
      int nbLog = txrcpt.getLogs().size();
      for (int i = 0; i < nbLog; i++) {
        traceValue.phaseSize += outerRlpSize(logSize(txrcpt.getLogs().get(i)));
      }
      traceValue.partialReset(4, 8);
      rlpByteString(4, traceValue.phaseSize, true, true, false, false, false, false, traceValue);

      // Trace each Log Entry.
      for (int i = 0; i < nbLog; i++) {
        // Update ABS_LOG_NUM.
        this.absLogNum += 1;

        // Log Entry RLP Prefix.
        traceValue.partialReset(4, 8);
        traceValue.logEntrySize = logSize(txrcpt.getLogs().get(i));
        rlpByteString(
            4, traceValue.logEntrySize, true, true, true, false, false, false, traceValue);

        // Logger's Address.
        traceValue.partialReset(4, 3);
        traceValue.depth1 = true;
        traceValue.input1 = txrcpt.getLogs().get(i).getLogger().slice(0, 4);
        traceValue.input2 = txrcpt.getLogs().get(i).getLogger().slice(4, LLARGE_INTEGER);
        traceValue.limbConstructed = true;

        traceValue.counter = 0;
        traceValue.limb = BigInteger.valueOf(INT_SHORT_INTEGER + 20);
        traceValue.nBytes = 1;
        traceRow(traceValue);

        traceValue.counter = 1;
        traceValue.limb = txrcpt.getLogs().get(i).getLogger().slice(0, 4).toUnsignedBigInteger();
        traceValue.nBytes = 4;
        traceRow(traceValue);

        traceValue.counter = 2;
        traceValue.limb =
            txrcpt.getLogs().get(i).getLogger().slice(4, LLARGE_INTEGER).toUnsignedBigInteger();
        traceValue.nBytes = LLARGE_INTEGER;
        traceRow(traceValue);

        // Log Topic's RLP prefix.
        traceValue.partialReset(4, 1);
        traceValue.depth1 = true;
        traceValue.isPrefix = true;
        traceValue.isTopic = true;
        traceValue.localSize = 33 * txrcpt.getLogs().get(i).getTopics().size();
        traceValue.limbConstructed = true;

        if (txrcpt.getLogs().get(i).getTopics().isEmpty()
            || txrcpt.getLogs().get(i).getTopics().size() == 1) {
          traceValue.limb = BigInteger.valueOf(LIST_SHORT_INTEGER + traceValue.localSize);
          traceValue.nBytes = 1;
        } else {
          traceValue.limb =
              BigInteger.valueOf(256L * (LIST_LONG_INTEGER + 1) + traceValue.localSize);
          traceValue.nBytes = 2;
        }

        traceRow(traceValue);

        // RLP Log Topic (if exist).
        if (!txrcpt.getLogs().get(i).getTopics().isEmpty()) {
          for (int j = 0; j < txrcpt.getLogs().get(i).getTopics().size(); j++) {
            traceValue.partialReset(4, 3);
            traceValue.depth1 = true;
            traceValue.isTopic = true;
            traceValue.indexLocal += 1;
            traceValue.input1 = txrcpt.getLogs().get(i).getTopics().get(j).slice(0, LLARGE_INTEGER);
            traceValue.input2 =
                txrcpt.getLogs().get(i).getTopics().get(j).slice(LLARGE_INTEGER, LLARGE_INTEGER);
            traceValue.limbConstructed = true;

            traceValue.counter = 0;
            traceValue.limb = BigInteger.valueOf(INT_SHORT_INTEGER + 32);
            traceValue.nBytes = 1;
            traceValue.localSize -= traceValue.nBytes;
            traceRow(traceValue);

            traceValue.counter = 1;
            traceValue.limb = traceValue.input1.toUnsignedBigInteger();
            traceValue.nBytes = LLARGE_INTEGER;
            traceValue.localSize -= traceValue.nBytes;
            traceRow(traceValue);

            traceValue.counter = 2;
            traceValue.limb = traceValue.input2.toUnsignedBigInteger();
            traceValue.nBytes = LLARGE_INTEGER;
            traceValue.localSize -= traceValue.nBytes;
            traceRow(traceValue);
          }
        }

        // RLP Prefix of the Data.
        // In INPUT_2 is stored the number of topics, stored in INDEX_LOCAL at the previous row
        // (needed for LookUp)
        traceValue.input2 = Bytes.ofUnsignedShort(traceValue.indexLocal);
        traceValue.indexLocal = 0;

        switch (txrcpt.getLogs().get(i).getData().size()) {
          case 0:
            traceValue.partialReset(4, 1);
            traceValue.depth1 = true;
            traceValue.isPrefix = true;
            traceValue.isData = true;
            traceValue.input1 = Bytes.ofUnsignedInt(txrcpt.getLogs().get(i).getData().size());
            traceValue.localSize = txrcpt.getLogs().get(i).getData().size();
            traceValue.limbConstructed = true;
            traceValue.limb = BigInteger.valueOf(INT_SHORT_INTEGER);
            traceValue.nBytes = 1;
            traceValue.phaseEnd = (i == nbLog - 1);
            traceRow(traceValue);
            break;
          case 1:
            traceValue.partialReset(4, 8);
            rlpInt(
                4,
                txrcpt.getLogs().get(i).getData().get(0),
                true,
                true,
                false,
                true,
                false,
                true,
                traceValue);

            for (int k = 0; k < 8; k++) {
              this.builder.setInput1Relative(BigInteger.ONE, k);
              this.builder.setInput2Relative(
                  BigInteger.valueOf(txrcpt.getLogs().get(i).getTopics().size()), k);
              this.builder.setInput3Relative(
                  BigInteger.valueOf(txrcpt.getLogs().get(i).getData().get(0)), k);
            }
            break;
          default:
            traceValue.partialReset(4, 8);
            traceValue.localSize = txrcpt.getLogs().get(i).getData().size();
            rlpByteString(
                4,
                txrcpt.getLogs().get(i).getData().size(),
                false,
                true,
                true,
                false,
                true,
                false,
                traceValue);
            for (int k = 0; k < 8; k++) {
              this.builder.setInput2Relative(
                  BigInteger.valueOf(txrcpt.getLogs().get(i).getTopics().size()), k);
            }
            break;
        }

        // Tracing the Data
        if (!txrcpt.getLogs().get(i).getData().isEmpty()) {
          int nbDataSlice = 1 + (txrcpt.getLogs().get(i).getData().size() - 1) / 16;

          int sizeDataLastSlice =
              txrcpt.getLogs().get(i).getData().size() - LLARGE_INTEGER * (nbDataSlice - 1);

          if (sizeDataLastSlice == 0) {
            sizeDataLastSlice = LLARGE_INTEGER;
          }
          traceValue.partialReset(4, nbDataSlice);
          traceValue.depth1 = true;
          traceValue.isData = true;
          traceValue.limbConstructed = true;

          for (int ct = 0; ct < nbDataSlice; ct++) {
            traceValue.counter = ct;
            traceValue.indexLocal = ct;

            if (!(ct == nbDataSlice - 1)) {
              traceValue.input1 =
                  txrcpt.getLogs().get(i).getData().slice(LLARGE_INTEGER * ct, LLARGE_INTEGER);
              traceValue.limb = traceValue.input1.toUnsignedBigInteger();
              traceValue.nBytes = LLARGE_INTEGER;
              traceValue.localSize -= LLARGE_INTEGER;
            } else {
              traceValue.input1 =
                  txrcpt
                      .getLogs()
                      .get(i)
                      .getData()
                      .slice(LLARGE_INTEGER * ct, sizeDataLastSlice)
                      .shiftLeft(LLARGE_INTEGER - sizeDataLastSlice);
              traceValue.limb = traceValue.input1.toUnsignedBigInteger();
              traceValue.nBytes = sizeDataLastSlice;
              traceValue.localSize -= sizeDataLastSlice;
              traceValue.phaseEnd = (i == nbLog - 1);
            }

            traceRow(traceValue);
          }
        }
      }
    }
  }

  private void rlpByteString(
      int phase,
      long length,
      boolean isList,
      boolean isPrefix,
      boolean depth1,
      boolean isTopic,
      boolean isData,
      boolean endPhase,
      RlpTxrcptColumns traceValue) {
    int lengthSize =
        Bytes.ofUnsignedLong(length).size()
            - Bytes.ofUnsignedLong(length).numberOfLeadingZeroBytes();

    RlpByteCountAndPowerOutput byteCountingOutput = byteCounting(lengthSize, 8);

    traceValue.partialReset(phase, 8);
    traceValue.input1 = Bytes.ofUnsignedInt(length);
    traceValue.isPrefix = isPrefix;
    traceValue.depth1 = depth1;
    traceValue.isTopic = isTopic;
    traceValue.isData = isData;

    Bytes input1RightShift = padToGivenSizeWithLeftZero(traceValue.input1, 8);
    long acc2LastRow = 0;

    if (length >= 56) {
      acc2LastRow = length - 56;
    } else {
      acc2LastRow = 55 - length;
    }

    Bytes acc2LastRowShift = padToGivenSizeWithLeftZero(Bytes.ofUnsignedInt(acc2LastRow), 8);
    for (int ct = 0; ct < 8; ct++) {
      traceValue.counter = ct;
      traceValue.accSize = byteCountingOutput.getAccByteSizeList().get(ct);
      traceValue.power = byteCountingOutput.getPowerList().get(ct);
      traceValue.byte1 = input1RightShift.get(ct);
      traceValue.acc1 = input1RightShift.slice(0, ct + 1);
      traceValue.byte2 = acc2LastRowShift.get(ct);
      traceValue.acc2 = acc2LastRowShift.slice(0, ct + 1);

      if (length >= 56) {
        if (ct == 6) {
          traceValue.limbConstructed = true;
          traceValue.nBytes = 1;
          if (isList) {
            traceValue.limb = BigInteger.valueOf(LIST_LONG_INTEGER + lengthSize);
          } else {
            traceValue.limb = BigInteger.valueOf(INT_LONG_INTEGER + lengthSize);
          }
        }

        if (ct == 7) {
          traceValue.limbConstructed = true;
          traceValue.limb = BigInteger.valueOf(length);
          traceValue.nBytes = lengthSize;
          traceValue.bit = true;
          traceValue.bitAcc = 1;
          traceValue.phaseEnd = endPhase;
        }
      } else {
        if (ct == 7) {
          traceValue.limbConstructed = true;
          if (isList) {
            traceValue.limb = BigInteger.valueOf(LIST_SHORT_INTEGER + length);
          } else {
            traceValue.limb = BigInteger.valueOf(INT_SHORT_INTEGER + length);
          }
          traceValue.nBytes = 1;
          traceValue.phaseEnd = endPhase;
        }
      }

      traceRow(traceValue);
    }
  }

  private void rlpInt(
      int phase,
      long input,
      boolean isPrefix,
      boolean depth1,
      boolean isTopic,
      boolean isData,
      boolean endPhase,
      boolean onlyPrefix,
      RlpTxrcptColumns traceValue) {

    traceValue.partialReset(phase, 8);

    traceValue.isPrefix = isPrefix;
    traceValue.depth1 = depth1;
    traceValue.isTopic = isTopic;
    traceValue.isData = isData;

    int inputSize =
        Bytes.ofUnsignedInt(input).size() - Bytes.ofUnsignedInt(input).numberOfLeadingZeroBytes();
    RlpByteCountAndPowerOutput byteCountingOutput = byteCounting(inputSize, 8);

    Bytes inputBytes = padToGivenSizeWithLeftZero(Bytes.ofUnsignedInt(input), 8);
    RlpBitDecOutput bitDecOutput =
        bitDecomposition(0xff & inputBytes.get(inputBytes.size() - 1), 8);

    traceValue.input1 = Bytes.ofUnsignedInt(input);

    for (int ct = 0; ct < 8; ct++) {
      traceValue.counter = ct;
      traceValue.byte1 = inputBytes.get(ct);
      traceValue.acc1 = inputBytes.slice(0, ct + 1);
      traceValue.power = byteCountingOutput.getPowerList().get(ct);
      traceValue.accSize = byteCountingOutput.getAccByteSizeList().get(ct);
      traceValue.bit = bitDecOutput.getBitDecList().get(ct);
      traceValue.bitAcc = bitDecOutput.getBitAccList().get(ct);

      if (input >= 128 && ct == 6) {
        traceValue.limbConstructed = true;
        traceValue.limb = BigInteger.valueOf(INT_SHORT_INTEGER + inputSize);
        traceValue.nBytes = 1;
      }

      if (ct == 7) {
        if (onlyPrefix) {
          traceValue.lcCorrection = true;
          traceValue.limbConstructed = false;
          traceValue.limb = BigInteger.ZERO;
          traceValue.nBytes = 0;
        } else {
          traceValue.limbConstructed = true;
          traceValue.limb = BigInteger.valueOf(input);
          traceValue.nBytes = inputSize;
          traceValue.phaseEnd = endPhase;
        }
      }

      traceRow(traceValue);
    }
  }

  private void traceRow(RlpTxrcptColumns traceValue) {
    // Decrements sizes
    if (traceValue.limbConstructed) {
      if (traceValue.phase != 0) {
        traceValue.txrcptSize -= traceValue.nBytes;
      }
      if ((traceValue.phase == 3 && !traceValue.isPrefix)
          || (traceValue.phase == 4 && traceValue.depth1)) {
        traceValue.phaseSize -= traceValue.nBytes;
      }
      if ((traceValue.phase == 4 && traceValue.depth1 && (!traceValue.isPrefix))
          || traceValue.isTopic
          || traceValue.isData) {
        traceValue.logEntrySize -= traceValue.nBytes;
      }
    }

    builder
        .absLogNum(BigInteger.valueOf(this.absLogNum))
        .absLogNumMax(BigInteger.ONE)
        .absTxNum(BigInteger.valueOf(this.absTxNum))
        .absTxNumMax(BigInteger.ONE)
        .acc1(traceValue.acc1.toUnsignedBigInteger())
        .acc2(traceValue.acc2.toUnsignedBigInteger())
        .acc3(traceValue.acc3.toUnsignedBigInteger())
        .acc4(traceValue.acc4.toUnsignedBigInteger())
        .accSize(BigInteger.valueOf(traceValue.accSize))
        .bit(traceValue.bit)
        .bitAcc(UnsignedByte.of(traceValue.bitAcc))
        .byte1(UnsignedByte.of(traceValue.byte1))
        .byte2(UnsignedByte.of(traceValue.byte2))
        .byte3(UnsignedByte.of(traceValue.byte3))
        .byte4(UnsignedByte.of(traceValue.byte4))
        .counter(UnsignedByte.of(traceValue.counter))
        .depth1(traceValue.depth1)
        .done(traceValue.counter == traceValue.nStep - 1)
        .index(BigInteger.valueOf(traceValue.index))
        .indexLocal(BigInteger.valueOf(traceValue.indexLocal))
        .input1(traceValue.input1.toUnsignedBigInteger())
        .input2(traceValue.input2.toUnsignedBigInteger())
        .input3(traceValue.input3.toUnsignedBigInteger())
        .input4(traceValue.input4.toUnsignedBigInteger())
        .isData(traceValue.isData)
        .isPrefix(traceValue.isPrefix)
        .isTopic(traceValue.isTopic)
        .lcCorrection(traceValue.lcCorrection)
        .limb(traceValue.limb.shiftLeft(8 * (LLARGE_INTEGER - traceValue.nBytes)))
        .limbConstructed(traceValue.limbConstructed)
        .localSize(BigInteger.valueOf(traceValue.localSize))
        .logEntrySize(BigInteger.valueOf(traceValue.logEntrySize))
        .nBytes(UnsignedByte.of(traceValue.nBytes))
        .nStep(UnsignedByte.of(traceValue.nStep));

    List<Function<Boolean, Trace.TraceBuilder>> phaseColumns =
        List.of(
            this.builder::phase0,
            this.builder::phase1,
            this.builder::phase2,
            this.builder::phase3,
            this.builder::phase4);

    for (int i = 0; i < phaseColumns.size(); i++) {
      phaseColumns.get(i).apply(i == traceValue.phase);
    }

    this.builder
        .phaseEnd(traceValue.phaseEnd)
        .phaseSize(BigInteger.valueOf(traceValue.phaseSize))
        .power(traceValue.power)
        .txrcptSize(BigInteger.valueOf(traceValue.txrcptSize));

    this.builder.validateRow();

    // Increments Index.
    if (traceValue.limbConstructed) {
      traceValue.index += 1;
    }
  }

  /**
   * Returns the size of RLP(something) where something is of size inputSize (!=1) (it can be ZERO
   * though).
   */
  public static int outerRlpSize(int inputSize) {
    int rlpSize = inputSize;
    if (inputSize == 1) {
      // TODO panic
    } else {
      rlpSize += 1;
      if (inputSize >= 56) {
        rlpSize += Bytes.ofUnsignedShort(inputSize).size();
      }
    }
    return rlpSize;
  }

  /**
   * Calculates the size of the RLP of a transaction receipt WITHOUT its RLP prefix.
   *
   * @param txrcpt an instance of {@link TransactionReceipt} containing information pertaining to a
   *     transaction execution
   * @return the size of the RLP of a transaction receipt WITHOUT its RLP prefix
   */
  private int txRcptSize(TransactionReceipt txrcpt) {

    // The encoded status code is always of size 1.
    int size = 1;

    // As the cumulative gas is Gtransaction=21000, its size is >1.
    size +=
        outerRlpSize(
            Bytes.ofUnsignedInt(txrcpt.getCumulativeGasUsed()).size()
                - Bytes.ofUnsignedInt(txrcpt.getCumulativeGasUsed()).numberOfLeadingZeroBytes());

    // RLP(Rb) is always 259 (256+3) long.
    size += 259;

    // Add the size of the RLP(Log).
    int nbLog = txrcpt.getLogs().size();
    if (nbLog == 0) {
      size += 1;
    } else {
      int tmp = 0;
      for (int i = 0; i < nbLog; i++) {
        tmp += outerRlpSize(logSize(txrcpt.getLogs().get(i)));
      }
      size += outerRlpSize(tmp);
    }

    return size;
  }

  /** Gives the byte size of the RLP-isation of a log entry WITHOUT its RLP prefix. */
  private int logSize(Log log) {
    // The size of RLP(Oa) is always 21.
    int logSize = 21;

    // RLP(Topic) is of size 1 for 0 topic, 33+1 for 1 topic, 2 + 33*nTOPIC for 2 <= nTOPIC <=4.
    logSize += outerRlpSize(33 * log.getTopics().size());

    // RLP(Od) is of size OuterRlpSize(datasize) except if the Data is made of one byte.
    if (log.getData().size() == 1) {
      // If the byte is of value >= 128, its RLP is 2 byte, else 1 byte (no RLP prefix).
      if (log.getData().bitLength() == 8) {
        logSize += 2;
      } else {
        logSize += 1;
      }
    } else {
      logSize += outerRlpSize(log.getData().size());
    }

    return logSize;
  }

  /**
   * Add zeroes to the left of the {@link Bytes} to create {@link Bytes} of the given size. The
   * wantedSize must be at least the size of the Bytes.
   *
   * @param input
   * @param wantedSize
   * @return
   */
  public static Bytes padToGivenSizeWithLeftZero(Bytes input, int wantedSize) {
    Preconditions.checkArgument(
        wantedSize >= input.size(), "wantedSize can't be shorter than the input size");
    byte nullByte = 0;

    return Bytes.concatenate(Bytes.repeat(nullByte, wantedSize - input.size()), input);
  }

  public static Bytes bigIntegerToBytes(BigInteger big) {
    byte[] byteArray;
    byteArray = big.toByteArray();
    Bytes bytes;
    if (byteArray[0] == 0) {
      Bytes tmp = Bytes.wrap(byteArray);
      bytes = Bytes.wrap(tmp.slice(1, tmp.size() - 1));
    } else {
      bytes = Bytes.wrap(byteArray);
    }
    return bytes;
  }

  public static Bytes padToGivenSizeWithRightZero(Bytes input, int wantedSize) {
    Preconditions.checkArgument(
        wantedSize >= input.size(), "wantedSize can't be shorter than the input size");
    byte nullByte = 0;

    return Bytes.concatenate(input, Bytes.repeat(nullByte, wantedSize - input.size()));
  }

  /**
   * Create the Power and AccSize list of the ByteCountAndPower RLP pattern.
   *
   * @param inputByteLen represents the number of meaningful bytes of inputByte, i.e. without the
   *     zero left padding
   * @param nbStep
   * @return
   */
  public static RlpByteCountAndPowerOutput byteCounting(int inputByteLen, int nbStep) {
    RlpByteCountAndPowerOutput output = new RlpByteCountAndPowerOutput();

    BigInteger power;
    int accByteSize = 0;
    int offset = 16 - nbStep;

    if (inputByteLen == nbStep) {
      power = BigInteger.valueOf(256).pow(offset);
      accByteSize = 1;
    } else {
      offset += 1;
      power = BigInteger.valueOf(256).pow(offset);
    }

    output.getPowerList().add(0, power);
    output.getAccByteSizeList().add(0, accByteSize);

    for (int i = 1; i < nbStep; i++) {
      if (inputByteLen + i < nbStep) {
        power = power.multiply(BigInteger.valueOf(256));
      } else {
        accByteSize += 1;
      }
      output.getPowerList().add(i, power);
      output.getAccByteSizeList().add(i, accByteSize);
    }
    return output;
  }

  /**
   * Create the Bit and BitDec list of the RLP pattern of an int.
   *
   * @param input
   * @param nbStep
   * @return
   */
  public static RlpBitDecOutput bitDecomposition(int input, int nbStep) {
    Preconditions.checkArgument(nbStep >= 8, "Number of steps must be at least 8");

    RlpBitDecOutput output = new RlpBitDecOutput();
    // Set to zero first value
    for (int i = 0; i < nbStep; i++) {
      output.getBitAccList().add(i, 0);
      output.getBitDecList().add(i, false);
    }

    int bitAcc = 0;
    boolean bitDec = false;
    double div = 0;

    for (int i = 7; i >= 0; i--) {
      div = Math.pow(2, i);
      bitAcc *= 2;

      if (input >= div) {
        bitDec = true;
        bitAcc += 1;
        input -= (int) div;
      } else {
        bitDec = false;
      }

      output.getBitDecList().add(nbStep - i - 1, bitDec);
      output.getBitAccList().add(nbStep - i - 1, bitAcc);
    }
    return output;
  }

  @Override
  public Object commit() {
    return new RlpTxrcptTrace(builder.build());
  }
}
