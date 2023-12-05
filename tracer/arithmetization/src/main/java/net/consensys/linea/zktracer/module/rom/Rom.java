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

package net.consensys.linea.zktracer.module.rom;

import static net.consensys.linea.zktracer.module.rlputils.Pattern.padToGivenSizeWithRightZero;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.romLex.RomChunk;
import net.consensys.linea.zktracer.module.romLex.RomLex;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

public class Rom implements Module {
  private static final int LLARGE = 16;
  private static final Bytes BYTES_LLARGE = Bytes.of(LLARGE);
  private static final int LLARGE_MO = 15;
  private static final Bytes BYTES_LLARGE_MO = Bytes.of(LLARGE_MO);
  private static final int EVM_WORD_MO = 31;
  private static final Bytes BYTES_EVW_WORD_MO = Bytes.of(EVM_WORD_MO);
  private static final int PUSH_1 = 0x60;
  private static final int PUSH_32 = 0x7f;
  private static final UnsignedByte INVALID = UnsignedByte.of(0xFE);
  private static final int JUMPDEST = 0x5b;

  private final RomLex romLex;

  public Rom(RomLex _romLex) {
    this.romLex = _romLex;
  }

  @Override
  public String moduleKey() {
    return "ROM";
  }

  @Override
  public void enterTransaction() {}

  @Override
  public void popTransaction() {}

  @Override
  public int lineCount() {
    int traceRowSize = 0;
    for (RomChunk chunk : this.romLex.chunks) {
      traceRowSize += chunkRowSize(chunk);
    }
    return traceRowSize;
  }

  public int chunkRowSize(RomChunk chunk) {
    final int nPaddingRow = 32;
    final int codeSize = chunk.byteCode().size();
    final int nbSlice = (codeSize + (LLARGE - 1)) / LLARGE;

    return LLARGE * nbSlice + nPaddingRow;
  }

  private void traceChunk(RomChunk chunk, int cfi, int cfiInfty, Trace trace) {
    final int chunkRowSize = chunkRowSize(chunk);
    final int codeSize = chunk.byteCode().size();
    final int nLimbSlice = (codeSize + (LLARGE - 1)) / LLARGE;
    final Bytes dataPadded = padToGivenSizeWithRightZero(chunk.byteCode(), chunkRowSize);
    int nBytesLastRow = codeSize % LLARGE;
    if (nBytesLastRow == 0) {
      nBytesLastRow = LLARGE;
    }

    int pushParameter = 0;
    int ctPush = 0;
    Bytes pushValueHigh = Bytes.minimalBytes(0);
    Bytes pushValueLow = Bytes.minimalBytes(0);

    for (int i = 0; i < chunkRowSize; i++) {
      boolean codeSizeReached = i >= codeSize;
      int sliceNumber = i / 16;

      // Fill Generic columns
      trace
          .codeFragmentIndex(Bytes.ofUnsignedInt(cfi))
          .codeFragmentIndexInfty(Bytes.ofUnsignedInt(cfiInfty))
          .programmeCounter(Bytes.ofUnsignedInt(i))
          .limb(dataPadded.slice(sliceNumber * LLARGE, LLARGE))
          .codeSize(Bytes.ofUnsignedInt(codeSize))
          .paddedBytecodeByte(UnsignedByte.of(dataPadded.get(i)))
          .acc(dataPadded.slice(sliceNumber * LLARGE, (i % LLARGE) + 1))
          .codesizeReached(codeSizeReached)
          .index(Bytes.ofUnsignedInt(sliceNumber));

      // Fill CT, CTmax nBYTES, nBYTES_ACC
      if (sliceNumber < nLimbSlice) {
        trace.counter(Bytes.of(i % LLARGE)).counterMax(BYTES_LLARGE_MO);
        if (sliceNumber < nLimbSlice - 1) {
          trace.nBytes(BYTES_LLARGE).nBytesAcc(Bytes.of((i % LLARGE) + 1));
        }
        if (sliceNumber == nLimbSlice - 1) {
          trace
              .nBytes(Bytes.of(nBytesLastRow))
              .nBytesAcc(Bytes.of(Math.min(nBytesLastRow, (i % LLARGE) + 1)));
        }
      } else if (sliceNumber == nLimbSlice || sliceNumber == nLimbSlice + 1) {
        trace
            .counter(Bytes.of(i - nLimbSlice * LLARGE))
            .counterMax(BYTES_EVW_WORD_MO)
            .nBytes(Bytes.EMPTY)
            .nBytesAcc(Bytes.EMPTY);
      }

      // Deal when not in a PUSH instruction
      if (pushParameter == 0) {
        UnsignedByte opCode = UnsignedByte.of(dataPadded.get(i));
        final boolean isPush = PUSH_1 <= opCode.toInteger() && opCode.toInteger() <= PUSH_32;

        // The OpCode is a PUSH instruction
        if (isPush) {
          pushParameter = opCode.toInteger() - PUSH_1 + 1;
          if (pushParameter > LLARGE) {
            pushValueHigh = dataPadded.slice(i + 1, pushParameter - LLARGE);
            pushValueLow = dataPadded.slice(i + 1 + pushParameter - LLARGE, LLARGE);
          } else {
            pushValueLow = dataPadded.slice(i + 1, pushParameter);
          }
        }

        trace
            .isPush(isPush)
            .isPushData(false)
            .opcode(opCode)
            .pushParameter(Bytes.ofUnsignedShort(pushParameter))
            .counterPush(Bytes.EMPTY)
            .pushValueAcc(Bytes.EMPTY)
            .pushValueHigh(pushValueHigh)
            .pushValueLow(pushValueLow)
            .pushFunnelBit(false)
            .validJumpDestination(opCode.toInteger() == JUMPDEST);
      }
      // Deal when in a PUSH instruction
      else {
        ctPush += 1;
        trace
            .isPush(false)
            .isPushData(true)
            .opcode(INVALID)
            .pushParameter(Bytes.ofUnsignedShort(pushParameter))
            .pushValueHigh(pushValueHigh)
            .pushValueLow(pushValueLow)
            .counterPush(Bytes.of(ctPush))
            .pushFunnelBit(pushParameter > LLARGE && ctPush > pushParameter - LLARGE)
            .validJumpDestination(false);

        if (pushParameter <= LLARGE) {
          trace.pushValueAcc(pushValueLow.slice(0, ctPush));
        } else {
          if (ctPush <= pushParameter - LLARGE) {
            trace.pushValueAcc(pushValueHigh.slice(0, ctPush));
          } else {
            trace.pushValueAcc(pushValueLow.slice(0, ctPush + LLARGE - pushParameter));
          }
        }

        // reinitialise push constant data
        if (ctPush == pushParameter) {
          ctPush = 0;
          pushParameter = 0;
          pushValueHigh = Bytes.minimalBytes(0);
          pushValueLow = Bytes.minimalBytes(0);
        }
      }

      trace.validateRow();
    }
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int cfi = 0;
    final int cfiInfty = this.romLex.sortedChunks.size();
    for (RomChunk chunk : this.romLex.sortedChunks) {
      cfi += 1;
      traceChunk(chunk, cfi, cfiInfty, trace);
    }
  }
}
