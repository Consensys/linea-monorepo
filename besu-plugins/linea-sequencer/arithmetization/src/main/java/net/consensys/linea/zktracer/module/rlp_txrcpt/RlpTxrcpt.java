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

package net.consensys.linea.zktracer.module.rlp_txrcpt;

import static net.consensys.linea.zktracer.module.rlputils.Pattern.bitDecomposition;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.byteCounting;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.outerRlpSize;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.leftPadTo;
import static net.consensys.linea.zktracer.types.Conversions.rightPadTo;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;
import java.util.function.Function;

import com.google.common.base.Preconditions;
import lombok.Getter;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.rlputils.BitDecOutput;
import net.consensys.linea.zktracer.module.rlputils.ByteCountAndPowerOutput;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.log.LogsBloomFilter;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class RlpTxrcpt implements Module {
  private static final int LLARGE = Trace.LLARGE;
  private static final Bytes BYTES_RLP_INT_SHORT = Bytes.minimalBytes(Trace.INT_SHORT);
  private static final int INT_RLP_INT_SHORT = Trace.INT_SHORT;
  private static final int INT_RLP_INT_LONG = Trace.INT_LONG;
  private static final Bytes BYTES_RLP_LIST_SHORT = Bytes.minimalBytes(Trace.LIST_SHORT);
  private static final int INT_RLP_LIST_SHORT = Trace.LIST_SHORT;
  private static final int INT_RLP_LIST_LONG = Trace.LIST_LONG;

  private int absLogNum = 0;
  @Getter public StackedList<RlpTxrcptChunk> chunkList = new StackedList<>();

  @Override
  public String moduleKey() {
    return "RLP_TXRCPT";
  }

  @Override
  public void enterTransaction() {
    this.chunkList.enter();
  }

  @Override
  public void popTransaction() {
    this.chunkList.pop();
  }

  @Override
  public void traceEndTx(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logList,
      long gasUsed) {
    RlpTxrcptChunk chunk = new RlpTxrcptChunk(tx.getType(), status, gasUsed, logList);
    this.chunkList.add(chunk);
  }

  public void traceChunk(final RlpTxrcptChunk chunk, int absTxNum, int absLogNumMax, Trace trace) {
    RlpTxrcptColumns traceValue = new RlpTxrcptColumns();
    traceValue.txrcptSize = txRcptSize(chunk);
    traceValue.absTxNum = absTxNum;
    traceValue.absLogNumMax = absLogNumMax;

    // PHASE 1: RLP Prefix.
    phase1(traceValue, chunk.txType(), trace);

    // PHASE 2: Status code Rz.
    phase2(traceValue, chunk.status(), trace);

    // PHASE 3: Cumulative gas Ru.
    phase3(traceValue, chunk.gasUsed(), trace);

    // PHASE 4: Bloom Filter Rb.
    phase4(traceValue, chunk.logs(), trace);

    // Phase 5: log series Rl.
    phase5(traceValue, chunk.logs(), trace);
  }

  private void phase1(RlpTxrcptColumns traceValue, TransactionType txType, Trace trace) {
    final int phase = 1;
    // byte TYPE concatenation
    traceValue.partialReset(phase, 1);
    traceValue.isPrefix = true;

    if (txType == TransactionType.FRONTIER) {
      traceValue.lcCorrection = true;
    } else {
      traceValue.limbConstructed = true;
      traceValue.input1 = bigIntegerToBytes(BigInteger.valueOf(txType.getSerializedType()));
      traceValue.limb = traceValue.input1;
      traceValue.nBytes = 1;
    }

    traceRow(traceValue, trace);

    // RLP prefix of the txRcpt list.
    rlpByteString(
        phase, traceValue.txrcptSize, true, false, false, false, true, false, 0, traceValue, trace);
  }

  private void phase2(RlpTxrcptColumns traceValue, Boolean status, Trace trace) {
    final int phase = 2;
    traceValue.partialReset(phase, 1);
    traceValue.limbConstructed = true;

    if (status) {
      traceValue.input1 = bigIntegerToBytes(BigInteger.ONE);
      traceValue.limb = traceValue.input1;
    } else {
      traceValue.input1 = Bytes.ofUnsignedShort(0);
      traceValue.limb = BYTES_RLP_INT_SHORT;
    }

    traceValue.nBytes = 1;
    traceValue.phaseEnd = true;

    traceRow(traceValue, trace);
  }

  private void phase3(RlpTxrcptColumns traceValue, Long cumulativeGasUsed, Trace trace) {
    final int phase = 3;
    Preconditions.checkArgument(cumulativeGasUsed != 0, "Cumulative Gas Used can't be 0");
    rlpInt(
        1, phase, cumulativeGasUsed, false, false, false, true, false, false, 0, traceValue, trace);
  }

  public static void insertLog(LogsBloomFilter.Builder bloomBuilder, final Log log) {
    bloomBuilder.insertBytes(log.getLogger());

    for (var topic : log.getTopics()) {
      bloomBuilder.insertBytes(topic);
    }
  }

  private void phase4(RlpTxrcptColumns traceValue, List<Log> logList, Trace trace) {
    final int phase = 4;
    // RLP prefix
    traceValue.partialReset(phase, 1);
    traceValue.isPrefix = true;
    traceValue.phaseSize = 256;
    traceValue.limbConstructed = true;
    traceValue.limb =
        Bytes.concatenate(
            bigIntegerToBytes(BigInteger.valueOf(INT_RLP_INT_LONG + 2)),
            bigIntegerToBytes(BigInteger.valueOf(256)));
    traceValue.nBytes = 3;
    traceRow(traceValue, trace);

    // Concatenation of Byte slice of the bloom Filter.
    LogsBloomFilter.Builder bloomFilterBuilder = LogsBloomFilter.builder();
    for (Log log : logList) {
      insertLog(bloomFilterBuilder, log);
    }
    final LogsBloomFilter bloomFilter = bloomFilterBuilder.build();
    for (int i = 0; i < 4; i++) {
      traceValue.partialReset(phase, LLARGE);

      traceValue.input1 = bloomFilter.slice(64 * i, LLARGE);
      traceValue.input2 = bloomFilter.slice(64 * i + LLARGE, LLARGE);
      traceValue.input3 = bloomFilter.slice(64 * i + 2 * LLARGE, LLARGE);
      traceValue.input4 = bloomFilter.slice(64 * i + 3 * LLARGE, LLARGE);

      for (int ct = 0; ct < LLARGE; ct++) {
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
            traceValue.limb = traceValue.input1;
            traceValue.nBytes = LLARGE;
          }
          case 13 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input2;
            traceValue.nBytes = LLARGE;
          }
          case 14 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input3;
            traceValue.nBytes = LLARGE;
          }
          case 15 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input4;
            traceValue.nBytes = LLARGE;
            traceValue.phaseEnd = (i == 3);
          }
          default -> {
            traceValue.limbConstructed = false;
            traceValue.limb = Bytes.ofUnsignedShort(0);
            traceValue.nBytes = 0;
          }
        }
        traceRow(traceValue, trace);
        // Update INDEX_LOCAL after the row when LIMB is constructed.
        if (traceValue.limbConstructed) {
          traceValue.indexLocal += 1;
        }
      }
    }
    // Put to 0 INDEX_LOCAL at the end of the phase.
    traceValue.indexLocal = 0;
  }

  private void phase5(RlpTxrcptColumns traceValue, List<Log> logList, Trace trace) {
    final int phase = 5;
    // Trivial case, there are no log entries.
    if (logList.isEmpty()) {
      traceValue.logEntrySize = 1;
      traceEmptyList(traceValue, phase, true, true, trace);
    } else {
      // RLP prefix of the list of log entries.
      int nbLog = logList.size();
      for (Log log : logList) {
        traceValue.phaseSize += outerRlpSize(logSize(log));
      }
      traceValue.partialReset(phase, 8);
      rlpByteString(
          phase,
          traceValue.phaseSize,
          true,
          true,
          false,
          false,
          false,
          false,
          0,
          traceValue,
          trace);

      // Trace each Log Entry.
      for (int i = 0; i < nbLog; i++) {
        // Update ABS_LOG_NUM.
        this.absLogNum += 1;

        // Log Entry RLP Prefix.
        traceValue.logEntrySize = logSize(logList.get(i));
        rlpByteString(
            phase,
            traceValue.logEntrySize,
            true,
            true,
            true,
            false,
            false,
            false,
            0,
            traceValue,
            trace);

        // Logger's Address.
        // Common values for CT=0 tp CT=2
        traceValue.partialReset(phase, 3);
        traceValue.depth1 = true;
        traceValue.input1 = logList.get(i).getLogger().slice(0, 4);
        traceValue.input2 = logList.get(i).getLogger().slice(4, LLARGE);
        traceValue.limbConstructed = true;

        traceValue.counter = 0;
        traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_RLP_INT_SHORT + 20));
        traceValue.nBytes = 1;
        traceRow(traceValue, trace);

        traceValue.counter = 1;
        traceValue.limb = traceValue.input1;
        traceValue.nBytes = 4;
        traceRow(traceValue, trace);

        traceValue.counter = 2;
        traceValue.limb = traceValue.input2;
        traceValue.nBytes = LLARGE;
        traceRow(traceValue, trace);

        // Log Topic's RLP prefix.
        traceValue.partialReset(phase, 1);
        traceValue.depth1 = true;
        traceValue.isPrefix = true;
        traceValue.isTopic = true;
        traceValue.localSize = 33 * logList.get(i).getTopics().size();
        traceValue.limbConstructed = true;

        if (logList.get(i).getTopics().isEmpty() || logList.get(i).getTopics().size() == 1) {
          traceValue.limb =
              bigIntegerToBytes(
                  BYTES_RLP_LIST_SHORT
                      .toUnsignedBigInteger()
                      .add(BigInteger.valueOf(traceValue.localSize)));
          traceValue.nBytes = 1;
        } else {
          traceValue.limb =
              Bytes.concatenate(
                  bigIntegerToBytes(BigInteger.valueOf(INT_RLP_LIST_LONG + 1)),
                  bigIntegerToBytes(BigInteger.valueOf(traceValue.localSize)));
          traceValue.nBytes = 2;
        }
        traceRow(traceValue, trace);

        // RLP Log Topic (if exist).
        if (!logList.get(i).getTopics().isEmpty()) {
          for (int j = 0; j < logList.get(i).getTopics().size(); j++) {
            traceValue.partialReset(phase, 3);
            traceValue.depth1 = true;
            traceValue.isTopic = true;
            traceValue.indexLocal += 1;
            traceValue.input1 = logList.get(i).getTopics().get(j).slice(0, LLARGE);
            traceValue.input2 = logList.get(i).getTopics().get(j).slice(LLARGE, LLARGE);
            traceValue.limbConstructed = true;

            traceValue.counter = 0;
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_RLP_INT_SHORT + 32));
            traceValue.nBytes = 1;
            traceValue.localSize -= traceValue.nBytes;
            traceRow(traceValue, trace);

            traceValue.counter = 1;
            traceValue.limb = traceValue.input1;
            traceValue.nBytes = LLARGE;
            traceValue.localSize -= traceValue.nBytes;
            traceRow(traceValue, trace);

            traceValue.counter = 2;
            traceValue.limb = traceValue.input2;
            traceValue.nBytes = LLARGE;
            traceValue.localSize -= traceValue.nBytes;
            traceRow(traceValue, trace);
          }
        }
        // Reset the value of IndexLocal at the end of the Topic
        final int indexLocalEndTopic = traceValue.indexLocal;
        traceValue.indexLocal = 0;

        // RLP Prefix of the Data.
        // Common to all the cases:
        traceValue.localSize = logList.get(i).getData().size();
        // There are three cases for tracing the data RLP prefix:
        switch (logList.get(i).getData().size()) {
          case 0 -> { // Case no data:
            traceValue.partialReset(phase, 1);
            // In INPUT_2 is stored the number of topics, stored in INDEX_LOCAL at the
            // previous row
            traceValue.input2 = Bytes.ofUnsignedShort(indexLocalEndTopic);
            traceValue.depth1 = true;
            traceValue.isPrefix = true;
            traceValue.isData = true;
            traceValue.limbConstructed = true;
            traceValue.limb = BYTES_RLP_INT_SHORT;
            traceValue.nBytes = 1;
            traceValue.phaseEnd = (i == nbLog - 1);
            traceRow(traceValue, trace);
          }
          case 1 -> // Case with data is made of one byte
          rlpInt(
              3,
              phase,
              logList.get(i).getData().toUnsignedBigInteger().longValueExact(),
              true,
              true,
              true,
              false,
              true,
              true,
              indexLocalEndTopic,
              traceValue,
              trace);

          default -> // Default case, data is made of >= 2 bytes
          rlpByteString(
              phase,
              logList.get(i).getData().size(),
              false,
              true,
              true,
              true,
              false,
              true,
              indexLocalEndTopic,
              traceValue,
              trace);
        }

        // Tracing the Data
        if (!logList.get(i).getData().isEmpty()) {
          int nbDataSlice = 1 + (logList.get(i).getData().size() - 1) / 16;

          int sizeDataLastSlice = logList.get(i).getData().size() - LLARGE * (nbDataSlice - 1);

          if (sizeDataLastSlice == 0) {
            sizeDataLastSlice = LLARGE;
          }
          traceValue.partialReset(phase, nbDataSlice);
          traceValue.localSize = logList.get(i).getData().size();
          traceValue.depth1 = true;
          traceValue.isData = true;
          traceValue.limbConstructed = true;

          for (int ct = 0; ct < nbDataSlice; ct++) {
            traceValue.counter = ct;
            traceValue.indexLocal = ct;

            if (ct != nbDataSlice - 1) {
              traceValue.input1 = logList.get(i).getData().slice(LLARGE * ct, LLARGE);
              traceValue.limb = traceValue.input1;
              traceValue.nBytes = LLARGE;
              traceValue.localSize -= LLARGE;
            } else {
              traceValue.input1 =
                  rightPadTo(
                      logList.get(i).getData().slice(LLARGE * ct, sizeDataLastSlice), LLARGE);
              traceValue.limb = traceValue.input1;
              traceValue.nBytes = sizeDataLastSlice;
              traceValue.localSize -= sizeDataLastSlice;
              traceValue.phaseEnd = (i == nbLog - 1);
            }

            traceRow(traceValue, trace);
          }
          // set to 0 index local at the end of the data phase
          traceValue.indexLocal = 0;
        }
      }
    }
  }

  private void traceEmptyList(
      RlpTxrcptColumns traceValue, int phase, boolean isPrefix, boolean endPhase, Trace trace) {
    traceValue.partialReset(phase, 1);
    traceValue.limbConstructed = true;
    traceValue.limb = BYTES_RLP_LIST_SHORT;
    traceValue.nBytes = 1;
    traceValue.isPrefix = isPrefix;
    traceValue.phaseEnd = endPhase;
    traceRow(traceValue, trace);
  }

  private void rlpByteString(
      int phase,
      long length,
      boolean isList,
      boolean isPrefix,
      boolean depth1,
      boolean isData,
      boolean endPhase,
      boolean writeInput2,
      int valueInput2,
      RlpTxrcptColumns traceValue,
      Trace trace) {
    int lengthSize =
        Bytes.ofUnsignedLong(length).size()
            - Bytes.ofUnsignedLong(length).numberOfLeadingZeroBytes();

    ByteCountAndPowerOutput byteCountingOutput = byteCounting(lengthSize, 8);

    traceValue.partialReset(phase, 8);
    traceValue.input1 = bigIntegerToBytes(BigInteger.valueOf(length));
    traceValue.isPrefix = isPrefix;
    traceValue.depth1 = depth1;
    traceValue.isData = isData;
    if (writeInput2) {
      traceValue.input2 = Bytes.minimalBytes(valueInput2);
    }

    Bytes input1RightShift = leftPadTo(traceValue.input1, traceValue.nStep);
    long acc2LastRow;

    if (length >= 56) {
      acc2LastRow = length - 56;
    } else {
      acc2LastRow = 55 - length;
    }

    Bytes acc2LastRowShift =
        leftPadTo(bigIntegerToBytes(BigInteger.valueOf(acc2LastRow)), traceValue.nStep);
    for (int ct = 0; ct < 8; ct++) {
      traceValue.counter = ct;
      traceValue.accSize = byteCountingOutput.accByteSizeList().get(ct);
      traceValue.power = byteCountingOutput.powerList().get(ct);
      traceValue.byte1 = input1RightShift.get(ct);
      traceValue.acc1 = input1RightShift.slice(0, ct + 1);
      traceValue.byte2 = acc2LastRowShift.get(ct);
      traceValue.acc2 = acc2LastRowShift.slice(0, ct + 1);

      if (length >= 56) {
        if (ct == 6) {
          traceValue.limbConstructed = true;
          traceValue.nBytes = 1;
          if (isList) {
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_RLP_LIST_LONG + lengthSize));
          } else {
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_RLP_INT_LONG + lengthSize));
          }
        }

        if (ct == 7) {
          traceValue.limbConstructed = true;
          traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(length));
          traceValue.nBytes = lengthSize;
          traceValue.bit = true;
          traceValue.bitAcc = 1;
          traceValue.phaseEnd = endPhase;
        }
      } else {
        if (ct == 7) {
          traceValue.limbConstructed = true;
          if (isList) {
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_RLP_LIST_SHORT + length));
          } else {
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_RLP_INT_SHORT + length));
          }
          traceValue.nBytes = 1;
          traceValue.phaseEnd = endPhase;
        }
      }

      traceRow(traceValue, trace);
    }
  }

  private void rlpInt(
      int inputToWrite,
      int phase,
      long input,
      boolean isPrefix,
      boolean depth1,
      boolean isData,
      boolean endPhase,
      boolean onlyPrefix,
      boolean writeInput2,
      int valueInput2,
      RlpTxrcptColumns traceValue,
      Trace trace) {

    final Bytes inputBytes = bigIntegerToBytes(BigInteger.valueOf(input));

    traceValue.partialReset(phase, 8);

    traceValue.isPrefix = isPrefix;
    traceValue.depth1 = depth1;
    traceValue.isData = isData;
    switch (inputToWrite) {
      case 1 -> {
        traceValue.input1 = inputBytes;
      }
      case 3 -> {
        traceValue.input1 = Bytes.minimalBytes(1);
        traceValue.input3 = inputBytes;
      }
      default -> throw new IllegalArgumentException(
          "should be called only to write Input1 or Input3, not Input" + inputToWrite);
    }
    if (writeInput2) {
      traceValue.input2 = Bytes.minimalBytes(valueInput2);
    }

    final int inputSize = inputBytes.size();
    ByteCountAndPowerOutput byteCountingOutput = byteCounting(inputSize, 8);

    Bytes inputBytesPadded = leftPadTo(inputBytes, 8);
    BitDecOutput bitDecOutput =
        bitDecomposition(0xff & inputBytesPadded.get(inputBytesPadded.size() - 1), 8);

    for (int ct = 0; ct < 8; ct++) {
      traceValue.counter = ct;
      traceValue.byte1 = inputBytesPadded.get(ct);
      traceValue.acc1 = inputBytesPadded.slice(0, ct + 1);
      traceValue.power = byteCountingOutput.powerList().get(ct);
      traceValue.accSize = byteCountingOutput.accByteSizeList().get(ct);
      traceValue.bit = bitDecOutput.bitDecList().get(ct);
      traceValue.bitAcc = bitDecOutput.bitAccList().get(ct);

      if (input >= 128 && ct == 6) {
        traceValue.limbConstructed = true;
        traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_RLP_INT_SHORT + inputSize));
        traceValue.nBytes = 1;
      }

      if (ct == 7) {
        if (onlyPrefix) {
          traceValue.lcCorrection = true;
          traceValue.limbConstructed = false;
          traceValue.limb = bigIntegerToBytes(BigInteger.ZERO);
          traceValue.nBytes = 0;
        } else {
          traceValue.limbConstructed = true;
          traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(input));
          traceValue.nBytes = inputSize;
          traceValue.phaseEnd = endPhase;
        }
      }

      traceRow(traceValue, trace);
    }
  }

  private void traceRow(RlpTxrcptColumns traceValue, Trace trace) {
    // Decrements sizes
    if (traceValue.limbConstructed) {
      if (traceValue.phase != 1) {
        traceValue.txrcptSize -= traceValue.nBytes;
      }
      if ((traceValue.phase == 4 && !traceValue.isPrefix)
          || (traceValue.phase == 5 && traceValue.depth1)) {
        traceValue.phaseSize -= traceValue.nBytes;
      }
      if ((traceValue.phase == 5 && traceValue.depth1 && (!traceValue.isPrefix))
          || traceValue.isTopic
          || traceValue.isData) {
        traceValue.logEntrySize -= traceValue.nBytes;
      }
    }

    trace
        .absLogNum(Bytes.ofUnsignedLong(this.absLogNum))
        .absLogNumMax(Bytes.ofUnsignedLong(traceValue.absLogNumMax))
        .absTxNum(Bytes.ofUnsignedLong(traceValue.absTxNum))
        .absTxNumMax(Bytes.ofUnsignedLong(this.chunkList.size()))
        .acc1(traceValue.acc1)
        .acc2(traceValue.acc2)
        .acc3(traceValue.acc3)
        .acc4(traceValue.acc4)
        .accSize(Bytes.ofUnsignedLong(traceValue.accSize))
        .bit(traceValue.bit)
        .bitAcc(UnsignedByte.of(traceValue.bitAcc))
        .byte1(UnsignedByte.of(traceValue.byte1))
        .byte2(UnsignedByte.of(traceValue.byte2))
        .byte3(UnsignedByte.of(traceValue.byte3))
        .byte4(UnsignedByte.of(traceValue.byte4))
        .counter(Bytes.ofUnsignedInt(traceValue.counter))
        .depth1(traceValue.depth1)
        .done(traceValue.counter == traceValue.nStep - 1)
        .index(Bytes.ofUnsignedInt(traceValue.index))
        .indexLocal(Bytes.ofUnsignedInt(traceValue.indexLocal))
        .input1(traceValue.input1)
        .input2(traceValue.input2)
        .input3(traceValue.input3)
        .input4(traceValue.input4)
        .isData(traceValue.isData)
        .isPrefix(traceValue.isPrefix)
        .isTopic(traceValue.isTopic)
        .lcCorrection(traceValue.lcCorrection)
        .limb(rightPadTo(traceValue.limb, LLARGE))
        .limbConstructed(traceValue.limbConstructed)
        .localSize(Bytes.ofUnsignedInt(traceValue.localSize))
        .logEntrySize(Bytes.ofUnsignedInt(traceValue.logEntrySize))
        .nBytes(UnsignedByte.of(traceValue.nBytes))
        .nStep(Bytes.ofUnsignedInt(traceValue.nStep));

    List<Function<Boolean, Trace>> phaseColumns =
        List.of(trace::phase1, trace::phase2, trace::phase3, trace::phase4, trace::phase5);

    for (int i = 0; i < phaseColumns.size(); i++) {
      phaseColumns.get(i).apply(i + 1 == traceValue.phase);
    }

    trace
        .phaseEnd(traceValue.phaseEnd)
        .phaseSize(Bytes.ofUnsignedInt(traceValue.phaseSize))
        .power(bigIntegerToBytes(traceValue.power))
        .txrcptSize(Bytes.ofUnsignedInt(traceValue.txrcptSize));

    trace.validateRow();

    // Increments Index.
    if (traceValue.limbConstructed) {
      traceValue.index += 1;
    }
  }

  /**
   * Calculates the size of the RLP of a transaction receipt WITHOUT its RLP prefix.
   *
   * @param chunk an instance of {@link RlpTxrcptChunk} containing information pertaining to a
   *     transaction execution
   * @return the size of the RLP of a transaction receipt WITHOUT its RLP prefix
   */
  private int txRcptSize(RlpTxrcptChunk chunk) {

    // The encoded status code is always of size 1.
    int size = 1;

    // As the cumulative gas is Gtransaction=21000, its size is >1.
    size += outerRlpSize(Bytes.minimalBytes(chunk.gasUsed()).size());

    // RLP(Rb) is always 259 (256+3) long.
    size += 259;

    // Add the size of the RLP(Log).
    int nbLog = chunk.logs().size();
    if (nbLog == 0) {
      size += 1;
    } else {
      int tmp = 0;
      for (int i = 0; i < nbLog; i++) {
        tmp += outerRlpSize(logSize(chunk.logs().get(i)));
      }
      size += outerRlpSize(tmp);
    }

    return size;
  }

  // Gives the byte size of the RLP-isation of a log entry WITHOUT its RLP prefix.
  private int logSize(Log log) {
    // The size of RLP(Oa) is always 21.
    int logSize = 21;

    // RLP(Topic) is of size 1 for 0 topic, 33+1 for 1 topic, 2 + 33*nTOPIC for 2 <=
    // nTOPIC <=4.
    logSize += outerRlpSize(33 * log.getTopics().size());

    // RLP(Od) is of size OuterRlpSize(datasize) except if the Data is made of one
    // byte.
    if (log.getData().size() == 1) {
      // If the byte is of value >= 128, its RLP is 2 byte, else 1 byte (no RLP
      // prefix).
      if (log.getData().toUnsignedBigInteger().compareTo(BigInteger.valueOf(128)) >= 0) {
        logSize += 2;
      } else {
        logSize += 1;
      }
    } else {
      logSize += outerRlpSize(log.getData().size());
    }

    return logSize;
  }

  public int ChunkRowSize(RlpTxrcptChunk chunk) {
    // Phase 0 is always 1+8=9 row long, Phase 1, 1 row long, Phase 2 8 row long,
    // Phase 3 65 = 1 +
    // 64 row long
    int rowSize = 83;

    // add the number of rows for Phase 4 : Log entry
    if (chunk.logs().isEmpty()) {
      rowSize += 1;
    } else {
      // Rlp prefix of the list of log entries is always 8 rows long
      rowSize += 8;

      for (int i = 0; i < chunk.logs().size(); i++) {
        // Rlp prefix of a log entry is always 8, Log entry address is always 3 row
        // long, Log topics
        // rlp prefix always 1
        rowSize += 12;

        // Each log Topics is 3 rows long
        rowSize += 3 * chunk.logs().get(i).getTopics().size();

        // Row size of data is 1 if empty
        if (chunk.logs().get(i).getData().isEmpty()) {
          rowSize += 1;
        }
        // Row size of the data is 8 (RLP prefix)+ integer part (data-size - 1 /16) +1
        else {
          rowSize += 8 + (chunk.logs().get(i).getData().size() - 1) / 16 + 1;
        }
      }
    }

    return rowSize;
  }

  @Override
  public int lineCount() {
    int traceRowSize = 0;
    for (RlpTxrcptChunk chunk : this.chunkList) {
      traceRowSize += ChunkRowSize(chunk);
    }
    return traceRowSize;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int absLogNumMax = 0;
    for (RlpTxrcptChunk chunk : this.chunkList) {
      absLogNumMax += chunk.logs().size();
    }

    int absTxNum = 0;
    for (RlpTxrcptChunk chunk : this.chunkList) {
      absTxNum += 1;
      traceChunk(chunk, absTxNum, absLogNumMax, trace);
    }
  }
}
