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

package net.consensys.linea.zktracer.module.loginfo;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.rlptxrcpt.RlpTxnRcpt;
import net.consensys.linea.zktracer.module.rlptxrcpt.RlpTxrcptChunk;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.log.Log;

public class LogInfo implements Module {
  private final RlpTxnRcpt rlpTxnRcpt;

  public LogInfo(RlpTxnRcpt rlpTxnRcpt) {
    this.rlpTxnRcpt = rlpTxnRcpt;
  }

  private static final int LOG0 = 0xa0; // TODO why I don't get it from the .lisp ?

  @Override
  public String moduleKey() {
    return "LOG_INFO";
  }

  @Override
  public void enterTransaction() {}

  @Override
  public void popTransaction() {}

  @Override
  public int lineCount() {
    int rowSize = 0;
    for (RlpTxrcptChunk chunk : this.rlpTxnRcpt.getChunkList()) {
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
    for (RlpTxrcptChunk tx : this.rlpTxnRcpt.chunkList) {
      absLogNumMax += tx.logs().size();
    }

    int absTxNum = 0;
    int absLogNum = 0;
    for (RlpTxrcptChunk tx : this.rlpTxnRcpt.chunkList) {
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
        .absTxnNumMax(this.rlpTxnRcpt.chunkList.size())
        .absTxnNum(absTxNum)
        .txnEmitsLogs(false)
        .absLogNumMax(absLogNumMax)
        .absLogNum(absLogNum)
        .ctMax(UnsignedByte.ZERO)
        .ct(UnsignedByte.ZERO)
        .addrHi(0L)
        .addrLo(Bytes.EMPTY)
        .topicHi1(Bytes.EMPTY)
        .topicLo1(Bytes.EMPTY)
        .topicHi2(Bytes.EMPTY)
        .topicLo2(Bytes.EMPTY)
        .topicHi3(Bytes.EMPTY)
        .topicLo3(Bytes.EMPTY)
        .topicHi4(Bytes.EMPTY)
        .topicLo4(Bytes.EMPTY)
        .dataSize(0L)
        .inst(UnsignedByte.ZERO)
        .isLogX0(false)
        .isLogX1(false)
        .isLogX2(false)
        .isLogX3(false)
        .isLogX4(false)
        .phase(LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_NO_LOG_ENTRY)
        .dataHi(Bytes.EMPTY)
        .dataLo(Bytes.EMPTY)
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
          .absTxnNumMax(this.rlpTxnRcpt.chunkList.size())
          .absTxnNum(absTxNum)
          .txnEmitsLogs(true)
          .absLogNumMax(absLogNumMax)
          .absLogNum(absLogNum)
          .ctMax(UnsignedByte.of(ctMax))
          .ct(UnsignedByte.of(ct))
          .addrHi(log.getLogger().slice(0, 4).toLong())
          .addrLo(log.getLogger().slice(4, 16))
          .topicHi1(topic1.slice(0, 16))
          .topicLo1(topic1.slice(16, 16))
          .topicHi2(topic2.slice(0, 16))
          .topicLo2(topic2.slice(16, 16))
          .topicHi3(topic3.slice(0, 16))
          .topicLo3(topic3.slice(16, 16))
          .topicHi4(topic4.slice(0, 16))
          .topicLo4(topic4.slice(16, 16))
          .dataSize(log.getData().size())
          .inst(UnsignedByte.of(LOG0 + nbTopic))
          .isLogX0(nbTopic == 0)
          .isLogX1(nbTopic == 1)
          .isLogX2(nbTopic == 2)
          .isLogX3(nbTopic == 3)
          .isLogX4(nbTopic == 4);

      switch (ct) {
        case 0 -> {
          trace
              .phase(LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_DATA_SIZE)
              .dataHi(Bytes.ofUnsignedInt(log.getData().size()))
              .dataLo(Bytes.ofUnsignedInt(nbTopic))
              .validateRow();
        }
        case 1 -> {
          trace
              .phase(LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_ADDR)
              .dataHi(log.getLogger().slice(0, 4))
              .dataLo(log.getLogger().slice(4, 16))
              .validateRow();
        }
        case 2 -> {
          trace
              .phase(
                  LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE
                      + LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA)
              .dataHi(topic1.slice(0, 16))
              .dataLo(topic1.slice(16, 16))
              .validateRow();
        }
        case 3 -> {
          trace
              .phase(
                  LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE
                      + 2 * LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA)
              .dataHi(topic2.slice(0, 16))
              .dataLo(topic2.slice(16, 16))
              .validateRow();
        }
        case 4 -> {
          trace
              .phase(
                  LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE
                      + 3 * LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA)
              .dataHi(topic3.slice(0, 16))
              .dataLo(topic3.slice(16, 16))
              .validateRow();
        }
        case 5 -> {
          trace
              .phase(
                  LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE
                      + 4 * LogInfoTrace.RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA)
              .dataHi(topic4.slice(0, 16))
              .dataLo(topic4.slice(16, 16))
              .validateRow();
        }
        default -> throw new IllegalArgumentException(
            "ct = " + ct + " greater than ctMax =" + ctMax);
      }
    }
  }

  private int ctMax(Log log) {
    return log.getTopics().size() + 1;
  }
}
