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

package net.consensys.linea.zktracer.module.mod;

import java.math.BigInteger;

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.ModuleTrace;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class Mod implements Module {
  private int stamp = 0;

  final Trace.TraceBuilder builder = Trace.builder();

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

  public void traceModOperation(ModOperation op) {
    this.stamp++;

    for (int i = 0; i < op.maxCounter(); i++) {
      final int accLength = i + 1;
      builder
          .stamp(BigInteger.valueOf(stamp))
          .oli(op.isOli())
          .ct(BigInteger.valueOf(i))
          .inst(BigInteger.valueOf(op.getOpCode().getData().value()))
          .decSigned(op.isSigned())
          .decOutput(op.isDiv())
          .arg1Hi(op.getArg1().getHigh().toUnsignedBigInteger())
          .arg1Lo(op.getArg1().getLow().toUnsignedBigInteger())
          .arg2Hi(op.getArg2().getHigh().toUnsignedBigInteger())
          .arg2Lo(op.getArg2().getLow().toUnsignedBigInteger())
          .resHi(op.getResult().getHigh().toUnsignedBigInteger())
          .resLo(op.getResult().getLow().toUnsignedBigInteger())
          .acc12(op.getArg1().getBytes32().slice(8, i + 1).toUnsignedBigInteger())
          .acc13(op.getArg1().getBytes32().slice(0, i + 1).toUnsignedBigInteger())
          .acc22(op.getArg2().getBytes32().slice(8, i + 1).toUnsignedBigInteger())
          .acc23(op.getArg2().getBytes32().slice(0, i + 1).toUnsignedBigInteger())
          .accB0(op.getBBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accB1(op.getBBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accB2(op.getBBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accB3(op.getBBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accR0(op.getRBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accR1(op.getRBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accR2(op.getRBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accR3(op.getRBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accQ0(op.getQBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accQ1(op.getQBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accQ2(op.getQBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accQ3(op.getQBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .accDelta0(op.getDBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .accDelta1(op.getDBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .accDelta2(op.getDBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .accDelta3(op.getDBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
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
          .accH0(Bytes.wrap(op.getHBytes().get(0)).slice(0, i + 1).toUnsignedBigInteger())
          .accH1(Bytes.wrap(op.getHBytes().get(1)).slice(0, i + 1).toUnsignedBigInteger())
          .accH2(Bytes.wrap(op.getHBytes().get(2)).slice(0, i + 1).toUnsignedBigInteger())
          .cmp1(op.getCmp1()[i])
          .cmp2(op.getCmp2()[i])
          .msb1(op.getMsb1()[i])
          .msb2(op.getMsb2()[i])
          .validateRow();
    }
  }

  @Override
  public ModuleTrace commit() {
    for (ModOperation op : this.chunks) {
      this.traceModOperation(op);
    }
    return new ModTrace(builder.build());
  }

  @Override
  public int lineCount() {
    return this.chunks.stream().mapToInt(ModOperation::maxCounter).sum();
  }

  /**
   * Performs a context-free call to the DIV opcode in the current trace.
   *
   * @param arg1 the divider
   * @param arg2 the dividend
   */
  public void callDiv(Bytes32 arg1, Bytes32 arg2) {
    ModOperation data = new ModOperation(OpCode.DIV, arg1, arg2);
    this.traceModOperation(data);
  }

  /**
   * Performs a context-free call to the MOD opcode in the current trace.
   *
   * @param arg1 the number
   * @param arg2 the module
   */
  public void callMod(Bytes32 arg1, Bytes32 arg2) {
    ModOperation data = new ModOperation(OpCode.MOD, arg1, arg2);
    this.traceModOperation(data);
  }
}
