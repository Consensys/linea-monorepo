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

package net.consensys.linea.zktracer.module.rlp_txrcpt;

import static net.consensys.linea.zktracer.bytes.conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.module.rlpPatterns.pattern.bitDecomposition;
import static net.consensys.linea.zktracer.module.rlpPatterns.pattern.byteCounting;
import static net.consensys.linea.zktracer.module.rlpPatterns.pattern.outerRlpSize;
import static net.consensys.linea.zktracer.module.rlpPatterns.pattern.padToGivenSizeWithLeftZero;
import static net.consensys.linea.zktracer.module.rlpPatterns.pattern.padToGivenSizeWithRightZero;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Function;

import com.google.common.base.Preconditions;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.rlpPatterns.RlpBitDecOutput;
import net.consensys.linea.zktracer.module.rlpPatterns.RlpByteCountAndPowerOutput;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.log.LogsBloomFilter;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class RlpTxrcpt implements Module {
  public static final int llarge = RlpTxrcptTrace.LLARGE.intValue();
  public static final Bytes bytesRlpIntShort = bigIntegerToBytes(RlpTxrcptTrace.INT_SHORT);
  public static final int intRlpIntShort = bytesRlpIntShort.toUnsignedBigInteger().intValueExact();
  public static final Bytes bytesRlpIntLong = bigIntegerToBytes(RlpTxrcptTrace.INT_LONG);
  public static final int intRlpIntLong = bytesRlpIntLong.toUnsignedBigInteger().intValueExact();
  public static final Bytes bytesRlpListShort = bigIntegerToBytes(RlpTxrcptTrace.LIST_SHORT);
  public static final int intRlpListShort =
      bytesRlpListShort.toUnsignedBigInteger().intValueExact();
  public static final Bytes bytesRlpListLong = bigIntegerToBytes(RlpTxrcptTrace.LIST_LONG);
  public static final int intRlpListLong = bytesRlpListLong.toUnsignedBigInteger().intValueExact();
  private int absLogNumMax = 0;
  private int absLogNum = 0;
  private final Trace.TraceBuilder builder = Trace.builder();
  List<RlpTxrcptChunk> chunkList = new ArrayList<>();

  @Override
  public String jsonKey() {
    return "rlpTxRcpt";
  }

  @Override
  public void traceEndTx(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logList,
      long gasUsed) {

    this.absLogNumMax += logList.size();
    RlpTxrcptChunk chunk = new RlpTxrcptChunk(tx.getType(), status, gasUsed, logList);
    this.chunkList.add(chunk);
  }

  public void traceChunk(final RlpTxrcptChunk chunk, int absTxNum) {
    RlpTxrcptColumns traceValue = new RlpTxrcptColumns();
    traceValue.txrcptSize = txRcptSize(chunk);
    traceValue.absTxNum = absTxNum;

    // PHASE 0: RLP Prefix.
    phase0(traceValue, chunk.txType());

    // PHASE 1: Status code Rz.
    phase1(traceValue, chunk.status());

    // PHASE 2: Cumulative gas Ru.
    phase2(traceValue, chunk.gasUsed());

    // PHASE 3: Bloom Filter Rb.
    phase3(traceValue, chunk.logs());

    // Phase 4: log series Rl.
    phase4(traceValue, chunk.logs());
  }

  private void phase0(RlpTxrcptColumns traceValue, TransactionType txType) {

    // byte TYPE concatenation
    traceValue.partialReset(0, 1);
    traceValue.isPrefix = true;

    if (txType == TransactionType.FRONTIER) {
      traceValue.lcCorrection = true;
    } else {
      traceValue.limbConstructed = true;
      traceValue.input1 = bigIntegerToBytes(BigInteger.valueOf(txType.getSerializedType()));
      traceValue.limb = traceValue.input1;
      traceValue.nBytes = 1;
    }

    traceRow(traceValue);

    // RLP prefix of the txRcpt list.
    rlpByteString(0, traceValue.txrcptSize, true, false, false, false, true, traceValue);
  }

  private void phase1(RlpTxrcptColumns traceValue, Boolean status) {
    traceValue.partialReset(1, 1);
    traceValue.limbConstructed = true;

    if (status) {
      traceValue.input1 = bigIntegerToBytes(BigInteger.ONE);
      traceValue.limb = traceValue.input1;
    } else {
      traceValue.input1 = Bytes.ofUnsignedShort(0);
      traceValue.limb = bytesRlpIntShort;
    }

    traceValue.nBytes = 1;
    traceValue.phaseEnd = true;

    traceRow(traceValue);
  }

  private void phase2(RlpTxrcptColumns traceValue, Long cumulativeGasUsed) {
    Preconditions.checkArgument(cumulativeGasUsed != 0, "Cumulative Gas Used can't be 0");
    rlpInt(2, cumulativeGasUsed, false, false, false, true, false, traceValue);
  }

  public static void insertLog(LogsBloomFilter.Builder bloomBuilder, final Log log) {
    bloomBuilder.insertBytes(log.getLogger());

    for (var topic : log.getTopics()) {
      bloomBuilder.insertBytes(topic);
    }
  }

  private void phase3(RlpTxrcptColumns traceValue, List<Log> logList) {
    // RLP prefix
    traceValue.partialReset(3, 1);
    traceValue.isPrefix = true;
    traceValue.phaseSize = 256;
    traceValue.limbConstructed = true;
    traceValue.limb =
        Bytes.concatenate(
            bigIntegerToBytes(BigInteger.valueOf(intRlpIntLong + 2)),
            bigIntegerToBytes(BigInteger.valueOf(256)));
    traceValue.nBytes = 3;
    traceRow(traceValue);

    // Concatenation of Byte slice of the bloom Filter.
    LogsBloomFilter.Builder bloomFilterBuilder = LogsBloomFilter.builder();
    for (Log log : logList) {
      insertLog(bloomFilterBuilder, log);
    }
    final LogsBloomFilter bloomFilter = bloomFilterBuilder.build();
    for (int i = 0; i < 4; i++) {
      traceValue.partialReset(3, llarge);

      traceValue.input1 = bloomFilter.slice(64 * i, llarge);
      traceValue.input2 = bloomFilter.slice(64 * i + llarge, llarge);
      traceValue.input3 = bloomFilter.slice(64 * i + 2 * llarge, llarge);
      traceValue.input4 = bloomFilter.slice(64 * i + 3 * llarge, llarge);

      for (int ct = 0; ct < llarge; ct++) {
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
            traceValue.nBytes = llarge;
          }
          case 13 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input2;
            traceValue.nBytes = llarge;
          }
          case 14 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input3;
            traceValue.nBytes = llarge;
          }
          case 15 -> {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input4;
            traceValue.nBytes = llarge;
            traceValue.phaseEnd = (i == 3);
          }
          default -> {
            traceValue.limbConstructed = false;
            traceValue.limb = Bytes.ofUnsignedShort(0);
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

  private void phase4(RlpTxrcptColumns traceValue, List<Log> logList) {
    // Trivial case, there are no log entries.
    if (logList.isEmpty()) {
      traceValue.logEntrySize = 1;
      traceEmptyList(traceValue, 4, true, true);
    } else {
      // RLP prefix of the list of log entries.
      int nbLog = logList.size();
      for (Log log : logList) {
        traceValue.phaseSize += outerRlpSize(logSize(log));
      }
      traceValue.partialReset(4, 8);
      rlpByteString(4, traceValue.phaseSize, true, true, false, false, false, traceValue);

      // Trace each Log Entry.
      for (int i = 0; i < nbLog; i++) {
        // Update ABS_LOG_NUM.
        this.absLogNum += 1;

        // Log Entry RLP Prefix.
        traceValue.logEntrySize = logSize(logList.get(i));
        rlpByteString(4, traceValue.logEntrySize, true, true, true, false, false, traceValue);

        // Logger's Address.
        // Common values for CT=0 tp CT=2
        traceValue.partialReset(4, 3);
        traceValue.depth1 = true;
        traceValue.input1 = logList.get(i).getLogger().slice(0, 4);
        traceValue.input2 = logList.get(i).getLogger().slice(4, llarge);
        traceValue.limbConstructed = true;

        traceValue.counter = 0;
        traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(intRlpIntShort + 20));
        traceValue.nBytes = 1;
        traceRow(traceValue);

        traceValue.counter = 1;
        traceValue.limb = traceValue.input1;
        traceValue.nBytes = 4;
        traceRow(traceValue);

        traceValue.counter = 2;
        traceValue.limb = traceValue.input2;
        traceValue.nBytes = llarge;
        traceRow(traceValue);

        // Log Topic's RLP prefix.
        traceValue.partialReset(4, 1);
        traceValue.depth1 = true;
        traceValue.isPrefix = true;
        traceValue.isTopic = true;
        traceValue.localSize = 33 * logList.get(i).getTopics().size();
        traceValue.limbConstructed = true;

        if (logList.get(i).getTopics().isEmpty() || logList.get(i).getTopics().size() == 1) {
          traceValue.limb =
              bigIntegerToBytes(
                  bytesRlpListShort
                      .toUnsignedBigInteger()
                      .add(BigInteger.valueOf(traceValue.localSize)));
          traceValue.nBytes = 1;
        } else {
          traceValue.limb =
              Bytes.concatenate(
                  bigIntegerToBytes(BigInteger.valueOf(intRlpListLong + 1)),
                  bigIntegerToBytes(BigInteger.valueOf(traceValue.localSize)));
          traceValue.nBytes = 2;
        }
        traceRow(traceValue);

        // RLP Log Topic (if exist).
        if (!logList.get(i).getTopics().isEmpty()) {
          for (int j = 0; j < logList.get(i).getTopics().size(); j++) {
            traceValue.partialReset(4, 3);
            traceValue.depth1 = true;
            traceValue.isTopic = true;
            traceValue.indexLocal += 1;
            traceValue.input1 = logList.get(i).getTopics().get(j).slice(0, llarge);
            traceValue.input2 = logList.get(i).getTopics().get(j).slice(llarge, llarge);
            traceValue.limbConstructed = true;

            traceValue.counter = 0;
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(intRlpIntShort + 32));
            traceValue.nBytes = 1;
            traceValue.localSize -= traceValue.nBytes;
            traceRow(traceValue);

            traceValue.counter = 1;
            traceValue.limb = traceValue.input1;
            traceValue.nBytes = llarge;
            traceValue.localSize -= traceValue.nBytes;
            traceRow(traceValue);

            traceValue.counter = 2;
            traceValue.limb = traceValue.input2;
            traceValue.nBytes = llarge;
            traceValue.localSize -= traceValue.nBytes;
            traceRow(traceValue);
          }
        }
        // Reset the value of IndexLocal at the end of the Topic
        int indexLocalEndTopic = traceValue.indexLocal;
        traceValue.indexLocal = 0;

        // RLP Prefix of the Data.
        // Common to all the cases:
        traceValue.localSize = logList.get(i).getData().size();
        // There are three cases for tracing the data RLP prefix:
        switch (logList.get(i).getData().size()) {
          case 0 -> { // Case  no data:
            traceValue.partialReset(4, 1);
            // In INPUT_2 is stored the number of topics, stored in INDEX_LOCAL at the previous row
            traceValue.input2 = Bytes.ofUnsignedShort(indexLocalEndTopic);
            traceValue.depth1 = true;
            traceValue.isPrefix = true;
            traceValue.isData = true;
            traceValue.limbConstructed = true;
            traceValue.limb = bytesRlpIntShort;
            traceValue.nBytes = 1;
            traceValue.phaseEnd = (i == nbLog - 1);
            traceRow(traceValue);
          }
          case 1 -> { // Case with data is made of one byte
            rlpInt(
                4,
                logList.get(i).getData().toUnsignedBigInteger().longValueExact(),
                true,
                true,
                true,
                false,
                true,
                traceValue);
            for (int k = 0; k < 8; k++) {
              // In INPUT_2 is stored the number of topics, stored in INDEX_LOCAL at the previous
              // row
              this.builder.setInput2Relative(BigInteger.valueOf(indexLocalEndTopic), k);
              // In Input_1 is the Datasize, and in Input_3 the only byte of Data
              this.builder.setInput1Relative(BigInteger.ONE, k);
              this.builder.setInput3Relative(logList.get(i).getData().toUnsignedBigInteger(), k);
            }
          }
          default -> { // Default case, data is made of >= 2 bytes
            rlpByteString(
                4, logList.get(i).getData().size(), false, true, true, true, false, traceValue);
            // In INPUT_2 is stored the number of topics, stored in INDEX_LOCAL at the previous row
            for (int k = 0; k < 8; k++) {
              this.builder.setInput2Relative(BigInteger.valueOf(indexLocalEndTopic), k);
            }
          }
        }

        // Tracing the Data
        if (!logList.get(i).getData().isEmpty()) {
          int nbDataSlice = 1 + (logList.get(i).getData().size() - 1) / 16;

          int sizeDataLastSlice = logList.get(i).getData().size() - llarge * (nbDataSlice - 1);

          if (sizeDataLastSlice == 0) {
            sizeDataLastSlice = llarge;
          }
          traceValue.partialReset(4, nbDataSlice);
          traceValue.localSize = logList.get(i).getData().size();
          traceValue.depth1 = true;
          traceValue.isData = true;
          traceValue.limbConstructed = true;

          for (int ct = 0; ct < nbDataSlice; ct++) {
            traceValue.counter = ct;
            traceValue.indexLocal = ct;

            if (ct != nbDataSlice - 1) {
              traceValue.input1 = logList.get(i).getData().slice(llarge * ct, llarge);
              traceValue.limb = traceValue.input1;
              traceValue.nBytes = llarge;
              traceValue.localSize -= llarge;
            } else {
              traceValue.input1 =
                  padToGivenSizeWithRightZero(
                      logList.get(i).getData().slice(llarge * ct, sizeDataLastSlice), llarge);
              traceValue.limb = traceValue.input1;
              traceValue.nBytes = sizeDataLastSlice;
              traceValue.localSize -= sizeDataLastSlice;
              traceValue.phaseEnd = (i == nbLog - 1);
            }

            traceRow(traceValue);
          }
          // set to 0 index local at the end of the data phase
          traceValue.indexLocal = 0;
        }
      }
    }
  }

  private void traceEmptyList(
      RlpTxrcptColumns traceValue, int phase, boolean isPrefix, boolean endPhase) {
    traceValue.partialReset(phase, 1);
    traceValue.limbConstructed = true;
    traceValue.limb = bytesRlpListShort;
    traceValue.nBytes = 1;
    traceValue.isPrefix = isPrefix;
    traceValue.phaseEnd = endPhase;
    traceRow(traceValue);
  }

  private void rlpByteString(
      int phase,
      long length,
      boolean isList,
      boolean isPrefix,
      boolean depth1,
      boolean isData,
      boolean endPhase,
      RlpTxrcptColumns traceValue) {
    int lengthSize =
        Bytes.ofUnsignedLong(length).size()
            - Bytes.ofUnsignedLong(length).numberOfLeadingZeroBytes();

    RlpByteCountAndPowerOutput byteCountingOutput = byteCounting(lengthSize, 8);

    traceValue.partialReset(phase, 8);
    traceValue.input1 = bigIntegerToBytes(BigInteger.valueOf(length));
    traceValue.isPrefix = isPrefix;
    traceValue.depth1 = depth1;
    traceValue.isData = isData;

    Bytes input1RightShift = padToGivenSizeWithLeftZero(traceValue.input1, traceValue.nStep);
    long acc2LastRow;

    if (length >= 56) {
      acc2LastRow = length - 56;
    } else {
      acc2LastRow = 55 - length;
    }

    Bytes acc2LastRowShift =
        padToGivenSizeWithLeftZero(
            bigIntegerToBytes(BigInteger.valueOf(acc2LastRow)), traceValue.nStep);
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
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(intRlpListLong + lengthSize));
          } else {
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(intRlpIntLong + lengthSize));
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
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(intRlpListShort + length));
          } else {
            traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(intRlpIntShort + length));
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
      boolean isData,
      boolean endPhase,
      boolean onlyPrefix,
      RlpTxrcptColumns traceValue) {

    traceValue.partialReset(phase, 8);

    traceValue.isPrefix = isPrefix;
    traceValue.depth1 = depth1;
    traceValue.isData = isData;
    traceValue.input1 = bigIntegerToBytes(BigInteger.valueOf(input));

    int inputSize = traceValue.input1.size();
    RlpByteCountAndPowerOutput byteCountingOutput = byteCounting(inputSize, 8);

    Bytes inputBytes = padToGivenSizeWithLeftZero(traceValue.input1, 8);
    RlpBitDecOutput bitDecOutput =
        bitDecomposition(0xff & inputBytes.get(inputBytes.size() - 1), 8);

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
        traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(intRlpIntShort + inputSize));
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
        .absLogNumMax(BigInteger.valueOf(this.absLogNumMax))
        .absTxNum(BigInteger.valueOf(traceValue.absTxNum))
        .absTxNumMax(BigInteger.valueOf(this.chunkList.size()))
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
        .counter(BigInteger.valueOf(traceValue.counter))
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
        .limb(padToGivenSizeWithRightZero(traceValue.limb, llarge).toUnsignedBigInteger())
        .limbConstructed(traceValue.limbConstructed)
        .localSize(BigInteger.valueOf(traceValue.localSize))
        .logEntrySize(BigInteger.valueOf(traceValue.logEntrySize))
        .nBytes(UnsignedByte.of(traceValue.nBytes))
        .nStep(BigInteger.valueOf(traceValue.nStep));

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

    // RLP(Topic) is of size 1 for 0 topic, 33+1 for 1 topic, 2 + 33*nTOPIC for 2 <= nTOPIC <=4.
    logSize += outerRlpSize(33 * log.getTopics().size());

    // RLP(Od) is of size OuterRlpSize(datasize) except if the Data is made of one byte.
    if (log.getData().size() == 1) {
      // If the byte is of value >= 128, its RLP is 2 byte, else 1 byte (no RLP prefix).
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
    // Phase 0 is always 1+8=9 row long, Phase 1, 1 row long, Phase 2 8 row long, Phase 3 65 = 1 +
    // 64 row long
    int rowSize = 83;

    // add the number of rows for Phase 4 : Log entry
    if (chunk.logs().isEmpty()) {
      rowSize += 1;
    } else {
      // Rlp prefix of the list of log entries is always 8 rows long
      rowSize += 8;

      for (int i = 0; i < chunk.logs().size(); i++) {
        // Rlp prefix of a log entry is always 8, Log entry address is always 3 row long, Log topics
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
  public Object commit() {
    int absTxNum = 0;
    int estTraceSize = 0;
    for (RlpTxrcptChunk chunk : this.chunkList) {
      absTxNum += 1;
      traceChunk(chunk, absTxNum);
      estTraceSize += ChunkRowSize(chunk);
      if (this.builder.size() != estTraceSize) {
        throw new RuntimeException(
            "ChunkSize is not the right one, chunk n°: "
                + absTxNum
                + " estimated size ="
                + estTraceSize
                + " trace size ="
                + this.builder.size());
      }
    }
    return new RlpTxrcptTrace(builder.build());
  }
}
