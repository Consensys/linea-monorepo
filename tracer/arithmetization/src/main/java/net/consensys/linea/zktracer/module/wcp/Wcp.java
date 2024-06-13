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

package net.consensys.linea.zktracer.module.wcp;

import static net.consensys.linea.zktracer.module.wcp.WcpOperation.EQbv;
import static net.consensys.linea.zktracer.module.wcp.WcpOperation.GEQbv;
import static net.consensys.linea.zktracer.module.wcp.WcpOperation.GTbv;
import static net.consensys.linea.zktracer.module.wcp.WcpOperation.ISZERObv;
import static net.consensys.linea.zktracer.module.wcp.WcpOperation.LEQbv;
import static net.consensys.linea.zktracer.module.wcp.WcpOperation.LTbv;

import java.nio.MappedByteBuffer;
import java.util.ArrayDeque;
import java.util.Deque;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class Wcp implements Module {
  private final StackedSet<WcpOperation> operations = new StackedSet<>();

  /** count the number of rows that could be added after the sequencer counts the number of line */
  public final Deque<Integer> additionalRows = new ArrayDeque<>();

  private final Hub hub;
  private boolean batchUnderConstruction;

  public Wcp(Hub hub) {
    this.hub = hub;
    this.batchUnderConstruction = true;
  }

  @Override
  public String moduleKey() {
    return "WCP";
  }

  @Override
  public void enterTransaction() {
    this.operations.enter();
    this.additionalRows.push(this.additionalRows.getFirst());
  }

  @Override
  public void popTransaction() {
    this.operations.pop();
    this.additionalRows.pop();
  }

  @Override
  public void traceStartConflation(final long blockCount) {
    this.additionalRows.push(0);
  }

  @Override
  public void tracePreOpcode(final MessageFrame frame) {
    final OpCode opcode = this.hub.opCode();
    final Bytes32 arg1 = Bytes32.leftPad(frame.getStackItem(0));
    final Bytes32 arg2 =
        (opcode != OpCode.ISZERO) ? Bytes32.leftPad(frame.getStackItem(1)) : Bytes32.ZERO;

    this.operations.add(new WcpOperation(opcode.byteValue(), arg1, arg2));
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int stamp = 0;
    for (WcpOperation operation : this.operations) {
      stamp++;
      operation.trace(trace, stamp);
    }
  }

  @Override
  public void traceEndConflation(final WorldView state) {
    this.batchUnderConstruction = false;
  }

  @Override
  public int lineCount() {
    return batchUnderConstruction
        ? this.operations.lineCount() + this.additionalRows.getFirst()
        : this.operations.lineCount();
  }

  public boolean callLT(final Bytes32 arg1, final Bytes32 arg2) {
    this.operations.add(new WcpOperation(LTbv, arg1, arg2));
    return arg1.compareTo(arg2) < 0;
  }

  public boolean callLT(final Bytes arg1, final Bytes arg2) {
    return this.callLT(Bytes32.leftPad(arg1), Bytes32.leftPad(arg2));
  }

  public boolean callLT(final long arg1, final long arg2) {
    return this.callLT(Bytes.ofUnsignedLong(arg1), Bytes.ofUnsignedLong(arg2));
  }

  public boolean callGT(final Bytes32 arg1, final Bytes32 arg2) {
    this.operations.add(new WcpOperation(GTbv, arg1, arg2));
    return arg1.compareTo(arg2) > 0;
  }

  public boolean callGT(final Bytes arg1, final Bytes arg2) {
    return this.callGT(Bytes32.leftPad(arg1), Bytes32.leftPad(arg2));
  }

  public boolean callGT(final int arg1, final int arg2) {
    return this.callGT(Bytes.ofUnsignedLong(arg1), Bytes.ofUnsignedLong(arg2));
  }

  public boolean callEQ(final Bytes32 arg1, final Bytes32 arg2) {
    this.operations.add(new WcpOperation(EQbv, arg1, arg2));
    return arg1.compareTo(arg2) == 0;
  }

  public boolean callEQ(final Bytes arg1, final Bytes arg2) {
    return this.callEQ(Bytes32.leftPad(arg1), Bytes32.leftPad(arg2));
  }

  public boolean callISZERO(final Bytes32 arg1) {
    this.operations.add(new WcpOperation(ISZERObv, arg1, Bytes32.ZERO));
    return arg1.isZero();
  }

  public boolean callISZERO(final Bytes arg1) {
    return this.callISZERO(Bytes32.leftPad(arg1));
  }

  public boolean callLEQ(final Bytes32 arg1, final Bytes32 arg2) {
    this.operations.add(new WcpOperation(LEQbv, arg1, arg2));
    return arg1.compareTo(arg2) <= 0;
  }

  public boolean callLEQ(final long arg1, final long arg2) {
    return this.callLEQ(Bytes.ofUnsignedLong(arg1), Bytes.ofUnsignedLong(arg2));
  }

  public boolean callLEQ(final Bytes arg1, final Bytes arg2) {
    return this.callLEQ(Bytes32.leftPad(arg1), Bytes32.leftPad(arg2));
  }

  public boolean callGEQ(final Bytes32 arg1, final Bytes32 arg2) {
    this.operations.add(new WcpOperation(GEQbv, arg1, arg2));
    return arg1.compareTo(arg2) >= 0;
  }

  public boolean callGEQ(final Bytes arg1, final Bytes arg2) {
    return this.callGEQ(Bytes32.leftPad(arg1), Bytes32.leftPad(arg2));
  }
}
