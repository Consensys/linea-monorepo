/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.blockhash;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.BLOCKHASH_MAX_HISTORY;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

public class Blockhash implements Module {
  private final Wcp wcp;

  private final StackedSet<BlockhashOperation> operations = new StackedSet<>();
  private final List<BlockhashOperation> sortedOperations = new ArrayList<>();

  /* Stores the result of BLOCKHASH if the result of the opcode is not 0 */
  private final Map<Bytes32, Bytes32> blockHashMap = new HashMap<>();
  /* Store the number of call (capped to 2) of BLOCKHASH of a BLOCK_NUMBER*/
  private final Map<Bytes32, Integer> numberOfCall = new HashMap<>();

  private long absoluteBlockNumber;
  private short relativeBlock;

  private Bytes32 opcodeArgument;
  private boolean lowerBound;
  private boolean upperBound;

  public Blockhash(Wcp wcp) {
    this.wcp = wcp;
    this.relativeBlock = 0;
  }

  @Override
  public String moduleKey() {
    return "BLOCK_HASH";
  }

  @Override
  public void enterTransaction() {
    this.operations.enter();
  }

  @Override
  public void popTransaction() {
    this.operations.pop();
  }

  @Override
  public void traceStartBlock(final ProcessableBlockHeader processableBlockHeader) {
    this.relativeBlock += 1;
    this.absoluteBlockNumber = processableBlockHeader.getNumber();
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    this.opcodeArgument = Bytes32.leftPad(frame.getStackItem(0));
    lowerBound =
        this.wcp.callGEQ(
            opcodeArgument, Bytes.ofUnsignedLong(this.absoluteBlockNumber - BLOCKHASH_MAX_HISTORY));
    upperBound = this.wcp.callLT(opcodeArgument, Bytes.ofUnsignedLong(this.absoluteBlockNumber));

    /* To prove the lex order of BLOCK_NUMBER_HI/LO, we call WCP at endConflation, so we need to add rows in WCP now.
    If a BLOCK_NUMBER is already called at least two times, no need for additional rows in WCP*/
    final int numberOfCall = this.numberOfCall.getOrDefault(this.opcodeArgument, 0);
    if (numberOfCall < 2) {
      this.wcp.additionalRows.push(
          this.wcp.additionalRows.pop()
              + Math.max(Math.min(LLARGE, this.opcodeArgument.trimLeadingZeros().size()), 1));
      this.numberOfCall.replace(this.opcodeArgument, numberOfCall, numberOfCall + 1);
    }
  }

  @Override
  public void tracePostOpcode(MessageFrame frame) {
    final Bytes32 result = Bytes32.leftPad(frame.getStackItem(0));
    this.operations.add(
        new BlockhashOperation(
            this.relativeBlock,
            this.opcodeArgument,
            this.absoluteBlockNumber,
            lowerBound,
            upperBound,
            result));
    if (result != Bytes32.ZERO) {
      blockHashMap.put(this.opcodeArgument, result);
    }
  }

  @Override
  public void traceEndConflation(WorldView state) {
    final BlockhashComparator BLOCKHASH_COMPARATOR = new BlockhashComparator();
    this.sortedOperations.addAll(this.operations);
    this.sortedOperations.sort(BLOCKHASH_COMPARATOR);
    if (!this.sortedOperations.isEmpty()) {
      this.wcp.callGEQ(this.sortedOperations.get(0).opcodeArgument(), Bytes32.ZERO);
      for (int i = 1; i < this.sortedOperations.size(); i++) {
        this.wcp.callGEQ(
            this.sortedOperations.get(i).opcodeArgument(),
            this.sortedOperations.get(i - 1).opcodeArgument());
      }
    }
  }

  @Override
  public int lineCount() {
    return this.operations.lineCount();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    for (BlockhashOperation op : this.sortedOperations) {
      final Bytes32 hash =
          op.result() == Bytes32.ZERO
              ? this.blockHashMap.getOrDefault(op.opcodeArgument(), Bytes32.ZERO)
              : op.result();

      op.trace(trace, hash);
    }
  }
}
