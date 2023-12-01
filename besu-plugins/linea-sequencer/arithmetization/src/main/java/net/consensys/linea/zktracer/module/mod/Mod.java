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

package net.consensys.linea.zktracer.module.mod;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class Mod implements Module {
  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "mod";
  }

  private final StackedSet<ModOperation> chunks = new StackedSet<>();

  @Override
  public void tracePreOpcode(final MessageFrame frame) {
    final OpCodeData opCodeData = OpCodes.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.leftPad(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.leftPad(frame.getStackItem(1));

    this.chunks.add(new ModOperation(opCodeData, arg1, arg2));
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
  public void traceStartTx(WorldView worldView, Transaction tx) {
    this.chunks.enter();
  }

  public void traceModOperation(ModOperation op, Trace trace) {
    this.stamp++;

    for (int i = 0; i < op.maxCounter(); i++) {
      final int accLength = i + 1;
      trace
          .stamp(Bytes.ofUnsignedLong(stamp))
          .oli(op.isOli())
          .ct(Bytes.of(i))
          .inst(Bytes.of(op.getOpCode().byteValue()))
          .decSigned(op.isSigned())
          .decOutput(op.isDiv())
          .arg1Hi(op.getArg1().getHigh())
          .arg1Lo(op.getArg1().getLow())
          .arg2Hi(op.getArg2().getHigh())
          .arg2Lo(op.getArg2().getLow())
          .resHi(op.getResult().getHigh())
          .resLo(op.getResult().getLow())
          .acc12(op.getArg1().getBytes32().slice(8, i + 1))
          .acc13(op.getArg1().getBytes32().slice(0, i + 1))
          .acc22(op.getArg2().getBytes32().slice(8, i + 1))
          .acc23(op.getArg2().getBytes32().slice(0, i + 1))
          .accB0(op.getBBytes().get(0).slice(0, accLength))
          .accB1(op.getBBytes().get(1).slice(0, accLength))
          .accB2(op.getBBytes().get(2).slice(0, accLength))
          .accB3(op.getBBytes().get(3).slice(0, accLength))
          .accR0(op.getRBytes().get(0).slice(0, accLength))
          .accR1(op.getRBytes().get(1).slice(0, accLength))
          .accR2(op.getRBytes().get(2).slice(0, accLength))
          .accR3(op.getRBytes().get(3).slice(0, accLength))
          .accQ0(op.getQBytes().get(0).slice(0, accLength))
          .accQ1(op.getQBytes().get(1).slice(0, accLength))
          .accQ2(op.getQBytes().get(2).slice(0, accLength))
          .accQ3(op.getQBytes().get(3).slice(0, accLength))
          .accDelta0(op.getDBytes().get(0).slice(0, accLength))
          .accDelta1(op.getDBytes().get(1).slice(0, accLength))
          .accDelta2(op.getDBytes().get(2).slice(0, accLength))
          .accDelta3(op.getDBytes().get(3).slice(0, accLength))
          .byte22(UnsignedByte.of(op.getArg2().getByte(i + 8)))
          .byte23(UnsignedByte.of(op.getArg2().getByte(i)))
          .byte12(UnsignedByte.of(op.getArg1().getByte(i + 8)))
          .byte13(UnsignedByte.of(op.getArg1().getByte(i)))
          .byteB0(UnsignedByte.of(op.getBBytes().get(0).get(i)))
          .byteB1(UnsignedByte.of(op.getBBytes().get(1).get(i)))
          .byteB2(UnsignedByte.of(op.getBBytes().get(2).get(i)))
          .byteB3(UnsignedByte.of(op.getBBytes().get(3).get(i)))
          .byteR0(UnsignedByte.of(op.getRBytes().get(0).get(i)))
          .byteR1(UnsignedByte.of(op.getRBytes().get(1).get(i)))
          .byteR2(UnsignedByte.of(op.getRBytes().get(2).get(i)))
          .byteR3(UnsignedByte.of(op.getRBytes().get(3).get(i)))
          .byteQ0(UnsignedByte.of(op.getQBytes().get(0).get(i)))
          .byteQ1(UnsignedByte.of(op.getQBytes().get(1).get(i)))
          .byteQ2(UnsignedByte.of(op.getQBytes().get(2).get(i)))
          .byteQ3(UnsignedByte.of(op.getQBytes().get(3).get(i)))
          .byteDelta0(UnsignedByte.of(op.getDBytes().get(0).get(i)))
          .byteDelta1(UnsignedByte.of(op.getDBytes().get(1).get(i)))
          .byteDelta2(UnsignedByte.of(op.getDBytes().get(2).get(i)))
          .byteDelta3(UnsignedByte.of(op.getDBytes().get(3).get(i)))
          .byteH0(UnsignedByte.of(op.getHBytes().get(0).get(i)))
          .byteH1(UnsignedByte.of(op.getHBytes().get(1).get(i)))
          .byteH2(UnsignedByte.of(op.getHBytes().get(2).get(i)))
          .accH0(Bytes.wrap(op.getHBytes().get(0)).slice(0, i + 1))
          .accH1(Bytes.wrap(op.getHBytes().get(1)).slice(0, i + 1))
          .accH2(Bytes.wrap(op.getHBytes().get(2)).slice(0, i + 1))
          .cmp1(op.getCmp1()[i])
          .cmp2(op.getCmp2()[i])
          .msb1(op.getMsb1()[i])
          .msb2(op.getMsb2()[i])
          .validateRow();
    }
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    for (ModOperation op : this.chunks) {
      this.traceModOperation(op, trace);
    }
  }

  @Override
  public int lineCount() {
    return this.chunks.stream().mapToInt(ModOperation::maxCounter).sum();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  /**
   * Performs a context-free call to the DIV opcode in the current trace.
   *
   * @param arg1 the divider
   * @param arg2 the dividend
   */
  public void callDiv(Bytes32 arg1, Bytes32 arg2) {
    this.chunks.add(new ModOperation(OpCode.DIV, arg1, arg2));
  }

  /**
   * Performs a context-free call to the MOD opcode in the current trace.
   *
   * @param arg1 the number
   * @param arg2 the module
   */
  public void callMod(Bytes32 arg1, Bytes32 arg2) {
    this.chunks.add(new ModOperation(OpCode.MOD, arg1, arg2));
  }
}
