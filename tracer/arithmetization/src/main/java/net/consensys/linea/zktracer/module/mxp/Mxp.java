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

import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.gas.BillingRate;
import net.consensys.linea.zktracer.opcode.gas.MxpType;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

/** Implementation of a {@link Module} for memory expansion. */
public class Mxp implements Module {
  /** A list of the operations to trace */
  private final StackedList<MxpData> chunks = new StackedList<>();

  private Hub hub;

  @Override
  public String moduleKey() {
    return "MXP";
  }

  public Mxp(Hub hub) {
    this.hub = hub;
  }

  // TODO: update tests and eliminate this constructor
  public Mxp() {}

  @Override
  public void tracePreOpcode(MessageFrame frame) { // This will be renamed to tracePreOp
    // create a data object to do the work
    this.chunks.add(new MxpData(frame, hub));

    // sanity check
    //    sanityCheck(opCode, scope, mxpData);
  }

  final void traceChunk(final MxpData chunk, int stamp, Trace trace) {
    Bytes32 acc1Bytes32 = Bytes32.leftPad(bigIntegerToBytes(chunk.getAcc1()));
    Bytes32 acc2Bytes32 = Bytes32.leftPad(bigIntegerToBytes(chunk.getAcc2()));
    Bytes32 acc3Bytes32 = Bytes32.leftPad(bigIntegerToBytes(chunk.getAcc3()));
    Bytes32 acc4Bytes32 = Bytes32.leftPad(bigIntegerToBytes(chunk.getAcc4()));
    Bytes32 accABytes32 = Bytes32.leftPad(bigIntegerToBytes(chunk.getAccA()));
    Bytes32 accWBytes32 = Bytes32.leftPad(bigIntegerToBytes(chunk.getAccW()));
    Bytes32 accQBytes32 = Bytes32.leftPad(bigIntegerToBytes(chunk.getAccQ()));

    int maxCt = chunk.maxCt();
    int maxCtComplement = 32 - maxCt;

    for (int i = 0; i < maxCt; i++) {
      trace
          .stamp(Bytes.ofUnsignedLong(stamp))
          .cn(Bytes.ofUnsignedLong(chunk.getContextNumber()))
          .ct(Bytes.of(i))
          .roob(chunk.isRoob())
          .noop(chunk.isNoOperation())
          .mxpx(chunk.isMxpx())
          .inst(Bytes.of(chunk.getOpCodeData().value()))
          .mxpType1(chunk.getOpCodeData().billing().type() == MxpType.TYPE_1)
          .mxpType2(chunk.getOpCodeData().billing().type() == MxpType.TYPE_2)
          .mxpType3(chunk.getOpCodeData().billing().type() == MxpType.TYPE_3)
          .mxpType4(chunk.getOpCodeData().billing().type() == MxpType.TYPE_4)
          .mxpType5(chunk.getOpCodeData().billing().type() == MxpType.TYPE_5)
          .gword(
              Bytes.ofUnsignedLong(
                  chunk.getOpCodeData().billing().billingRate() == BillingRate.BY_WORD
                      ? chunk.getOpCodeData().billing().perUnit().cost()
                      : 0))
          .gbyte(
              Bytes.ofUnsignedLong(
                  chunk.getOpCodeData().billing().billingRate() == BillingRate.BY_BYTE
                      ? chunk.getOpCodeData().billing().perUnit().cost()
                      : 0))
          .deploys(chunk.isDeploys())
          .offset1Hi(chunk.getOffset1().hi())
          .offset1Lo(chunk.getOffset1().lo())
          .offset2Hi(chunk.getOffset2().hi())
          .offset2Lo(chunk.getOffset2().lo())
          .size1Hi(chunk.getSize1().hi())
          .size1Lo(chunk.getSize1().lo())
          .size2Hi(chunk.getSize2().hi())
          .size2Lo(chunk.getSize2().lo())
          .maxOffset1(bigIntegerToBytes(chunk.getMaxOffset1()))
          .maxOffset2(bigIntegerToBytes(chunk.getMaxOffset2()))
          .maxOffset(bigIntegerToBytes(chunk.getMaxOffset()))
          .comp(chunk.isComp())
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
          .byteQq(Bytes.ofUnsignedLong(chunk.getByteQQ()[i].toInteger()))
          .byteR(Bytes.ofUnsignedLong(chunk.getByteR()[i].toInteger()))
          .words(Bytes.ofUnsignedLong(chunk.getWords()))
          .wordsNew(
              Bytes.ofUnsignedLong(
                  chunk.getWordsNew())) // TODO: Could (should?) be set in tracePostOp?
          .cMem(Bytes.ofUnsignedLong(chunk.getCMem())) // Returns current memory size in EVM words
          .cMemNew(Bytes.ofUnsignedLong(chunk.getCMemNew()))
          .quadCost(Bytes.ofUnsignedLong(chunk.getQuadCost()))
          .linCost(Bytes.ofUnsignedLong(chunk.getLinCost()))
          .gasMxp(Bytes.ofUnsignedLong(chunk.getQuadCost() + chunk.getEffectiveLinCost()))
          .expands(chunk.isExpands())
          .validateRow();
    }
  }

  @Override
  public void enterTransaction() {
    this.chunks.enter();
  }

  @Override
  public void popTransaction() {
    this.chunks.pop();
  }

  @Override
  public int lineCount() {
    return this.chunks.stream().mapToInt(MxpData::maxCt).sum();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    for (int i = 0; i < this.chunks.size(); i++) {
      this.traceChunk(this.chunks.get(i), i + 1, trace);
    }
  }

  @Override
  public void tracePostOp(MessageFrame frame) { // This is paired with tracePreOp
  }
}
