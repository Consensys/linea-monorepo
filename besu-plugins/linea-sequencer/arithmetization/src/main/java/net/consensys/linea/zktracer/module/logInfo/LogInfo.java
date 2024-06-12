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

package net.consensys.linea.zktracer.module.logInfo;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.rlp.txrcpt.RlpTxrcpt;
import net.consensys.linea.zktracer.module.rlp.txrcpt.RlpTxrcptChunk;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.log.Log;

public class LogInfo implements Module {
  private final RlpTxrcpt rlpTxrcpt;

  public LogInfo(RlpTxrcpt rlpTxrcpt) {
    this.rlpTxrcpt = rlpTxrcpt;
  }

  private static final int LOG0 = 0xa0; // TODO why I don't get it from the .lisp ?

  @Override
  public String moduleKey() {
    return "PUB_LOG_INFO";
  }

  @Override
  public void enterTransaction() {}

  @Override
  public void popTransaction() {}

  @Override
  public int lineCount() {
    int rowSize = 0;
    for (RlpTxrcptChunk chunk : this.rlpTxrcpt.getChunkList()) {
      rowSize += txRowSize(chunk);
    }
    return rowSize;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int absLogNumMax = 0;
    for (RlpTxrcptChunk tx : this.rlpTxrcpt.chunkList) {
      absLogNumMax += tx.logs().size();
    }

    int absTxNum = 0;
    int absLogNum = 0;
    for (RlpTxrcptChunk tx : this.rlpTxrcpt.chunkList) {
      absTxNum += 1;
      if (tx.logs().isEmpty()) {
        traceTxWoLog(absTxNum, absLogNum, absLogNumMax, trace);
      } else {
        for (Log log : tx.logs()) {
          absLogNum += 1;
          traceLog(log, absTxNum, absLogNum, absLogNumMax, trace);
        }
      }
    }
  }

  private int txRowSize(RlpTxrcptChunk tx) {
    int txRowSize = 0;
    if (tx.logs().isEmpty()) {
      return 1;
    } else {
      for (Log log : tx.logs()) {
        txRowSize += ctMax(log) + 1;
      }
      return txRowSize;
    }
  }

  public void traceTxWoLog(
      final int absTxNum, final int absLogNum, final int absLogNumMax, Trace trace) {
    trace
        .absTxnNumMax(BigInteger.valueOf(this.rlpTxrcpt.chunkList.size()))
        .absTxnNum(BigInteger.valueOf(absTxNum))
        .txnEmitsLogs(false)
        .absLogNumMax(BigInteger.valueOf(absLogNumMax))
        .absLogNum(BigInteger.valueOf(absLogNum))
        .ctMax(BigInteger.ZERO)
        .ct(BigInteger.ZERO)
        .addrHi(BigInteger.ZERO)
        .addrLo(BigInteger.ZERO)
        .topicHi1(BigInteger.ZERO)
        .topicLo1(BigInteger.ZERO)
        .topicHi2(BigInteger.ZERO)
        .topicLo2(BigInteger.ZERO)
        .topicHi3(BigInteger.ZERO)
        .topicLo3(BigInteger.ZERO)
        .topicHi4(BigInteger.ZERO)
        .topicLo4(BigInteger.ZERO)
        .dataSize(BigInteger.ZERO)
        .inst(BigInteger.ZERO)
        .isLogX0(false)
        .isLogX1(false)
        .isLogX2(false)
        .isLogX3(false)
        .isLogX4(false)
        .phase(BigInteger.valueOf(LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_NO_LOG_ENTRY))
        .dataHi(BigInteger.ZERO)
        .dataLo(BigInteger.ZERO)
        .validateRow();
  }

  public void traceLog(
      final Log log, final int absTxNum, final int absLogNum, final int absLogNumMax, Trace trace) {
    final int ctMax = ctMax(log);
    final int nbTopic = log.getTopics().size();
    final Bytes32 topic1 = nbTopic >= 1 ? log.getTopics().get(0) : Bytes32.ZERO;
    final Bytes32 topic2 = nbTopic >= 2 ? log.getTopics().get(1) : Bytes32.ZERO;
    final Bytes32 topic3 = nbTopic >= 3 ? log.getTopics().get(2) : Bytes32.ZERO;
    final Bytes32 topic4 = nbTopic >= 4 ? log.getTopics().get(3) : Bytes32.ZERO;
    for (int ct = 0; ct < ctMax + 1; ct++) {
      trace
          .absTxnNumMax(BigInteger.valueOf(this.rlpTxrcpt.chunkList.size()))
          .absTxnNum(BigInteger.valueOf(absTxNum))
          .txnEmitsLogs(true)
          .absLogNumMax(BigInteger.valueOf(absLogNumMax))
          .absLogNum(BigInteger.valueOf(absLogNum))
          .ctMax(BigInteger.valueOf(ctMax))
          .ct(BigInteger.valueOf(ct))
          .addrHi(log.getLogger().slice(0, 4).toUnsignedBigInteger())
          .addrLo(log.getLogger().slice(4, 16).toUnsignedBigInteger())
          .topicHi1(topic1.slice(0, 16).toUnsignedBigInteger())
          .topicLo1(topic1.slice(16, 16).toUnsignedBigInteger())
          .topicHi2(topic2.slice(0, 16).toUnsignedBigInteger())
          .topicLo2(topic2.slice(16, 16).toUnsignedBigInteger())
          .topicHi3(topic3.slice(0, 16).toUnsignedBigInteger())
          .topicLo3(topic3.slice(16, 16).toUnsignedBigInteger())
          .topicHi4(topic4.slice(0, 16).toUnsignedBigInteger())
          .topicLo4(topic4.slice(16, 16).toUnsignedBigInteger())
          .dataSize(BigInteger.valueOf(log.getData().size()))
          .inst(BigInteger.valueOf(LOG0 + nbTopic))
          .isLogX0(nbTopic == 0)
          .isLogX1(nbTopic == 1)
          .isLogX2(nbTopic == 2)
          .isLogX3(nbTopic == 3)
          .isLogX4(nbTopic == 4);

      switch (ct) {
        case 0 -> {
          trace
              .phase(BigInteger.valueOf(LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_DATA_SIZE))
              .dataHi(BigInteger.valueOf(log.getData().size()))
              .dataLo(BigInteger.valueOf(nbTopic))
              .validateRow();
        }
        case 1 -> {
          trace
              .phase(BigInteger.valueOf(LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_ADDR))
              .dataHi(log.getLogger().slice(0, 4).toUnsignedBigInteger())
              .dataLo(log.getLogger().slice(4, 16).toUnsignedBigInteger())
              .validateRow();
        }
        case 2 -> {
          trace
              .phase(
                  BigInteger.valueOf(
                      LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE
                          + LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA))
              .dataHi(topic1.slice(0, 16).toUnsignedBigInteger())
              .dataLo(topic1.slice(16, 16).toUnsignedBigInteger())
              .validateRow();
        }
        case 3 -> {
          trace
              .phase(
                  BigInteger.valueOf(
                      LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE
                          + 2 * LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA))
              .dataHi(topic2.slice(0, 16).toUnsignedBigInteger())
              .dataLo(topic2.slice(16, 16).toUnsignedBigInteger())
              .validateRow();
        }
        case 4 -> {
          trace
              .phase(
                  BigInteger.valueOf(
                      LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE
                          + 3 * LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA))
              .dataHi(topic3.slice(0, 16).toUnsignedBigInteger())
              .dataLo(topic3.slice(16, 16).toUnsignedBigInteger())
              .validateRow();
        }
        case 5 -> {
          trace
              .phase(
                  BigInteger.valueOf(
                      LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE
                          + 4 * LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA))
              .dataHi(topic4.slice(0, 16).toUnsignedBigInteger())
              .dataLo(topic4.slice(16, 16).toUnsignedBigInteger())
              .validateRow();
        }
        default ->
            throw new IllegalArgumentException("ct = " + ct + " greater than ctMax =" + ctMax);
      }
    }
  }

  private int ctMax(Log log) {
    return log.getTopics().size() + 1;
  }
}
