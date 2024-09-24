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

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.BLOCKHASH_MAX_HISTORY;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.OperationSetModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedSet;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostOpcodeDefer;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

@Getter
@Accessors(fluent = true)
public class Blockhash implements OperationSetModule<BlockhashOperation>, PostOpcodeDefer {
  private final Hub hub;
  private final Wcp wcp;
  private final ModuleOperationStackedSet<BlockhashOperation> operations =
      new ModuleOperationStackedSet<>();

  List<BlockhashOperation> sortedOperations;

  /* Stores the result of BLOCKHASH if the result of the opcode is not 0 */
  private final Map<Bytes32, Bytes32> blockHashMap = new HashMap<>();
  /* Store the number of call (capped to 2) of BLOCKHASH of a BLOCK_NUMBER*/
  private final Map<Bytes32, Integer> numberOfCall = new HashMap<>();

  private long absoluteBlockNumber;
  private short relativeBlock;

  private Bytes32 opcodeArgument;
  private boolean lowerBound;
  private boolean upperBound;

  public Blockhash(Hub hub, Wcp wcp) {
    this.hub = hub;
    this.wcp = wcp;
    this.relativeBlock = 0;
  }

  @Override
  public String moduleKey() {
    return "BLOCK_HASH";
  }

  @Override
  public void traceStartBlock(final ProcessableBlockHeader processableBlockHeader) {
    relativeBlock += 1;
    absoluteBlockNumber = processableBlockHeader.getNumber();
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    checkArgument(opCode == OpCode.BLOCKHASH, "Expected BLOCKHASH opcode");

    opcodeArgument = Bytes32.leftPad(frame.getStackItem(0));
    lowerBound =
        this.wcp.callGEQ(
            opcodeArgument, Bytes.ofUnsignedLong(this.absoluteBlockNumber - BLOCKHASH_MAX_HISTORY));
    upperBound = this.wcp.callLT(opcodeArgument, Bytes.ofUnsignedLong(this.absoluteBlockNumber));

    hub.defers().scheduleForPostExecution(this);

    /* To prove the lex order of BLOCK_NUMBER_HI/LO, we call WCP at endConflation, so we need to add rows in WCP now.
    If a BLOCK_NUMBER is already called at least two times, no need for additional rows in WCP*/
    final int numberOfCall = this.numberOfCall.getOrDefault(this.opcodeArgument, 0);
    if (numberOfCall < 2) {
      wcp.additionalRows.add(
          Math.max(Math.min(LLARGE, this.opcodeArgument.trimLeadingZeros().size()), 1));
      this.numberOfCall.replace(this.opcodeArgument, numberOfCall, numberOfCall + 1);
    }
  }

  @Override
  public void resolvePostExecution(
      Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    if (opCode == OpCode.BLOCKHASH) {
      final Bytes32 result = Bytes32.leftPad(frame.getStackItem(0));
      operations.add(
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
  }

  @Override
  public void traceEndConflation(WorldView state) {
    OperationSetModule.super.traceEndConflation(state);
    sortedOperations = new ArrayList<>(operations.getAll());
    if (!sortedOperations.isEmpty()) {
      final BlockhashComparator BLOCKHASH_COMPARATOR = new BlockhashComparator();
      sortedOperations.sort(BLOCKHASH_COMPARATOR);
      wcp.callGEQ(sortedOperations.getFirst().opcodeArgument(), Bytes32.ZERO);
      for (int i = 1; i < sortedOperations.size(); i++) {
        wcp.callGEQ(
            sortedOperations.get(i).opcodeArgument(), sortedOperations.get(i - 1).opcodeArgument());
      }
    }
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    for (BlockhashOperation op : sortedOperations) {
      final Bytes32 hash =
          op.result() == Bytes32.ZERO
              ? this.blockHashMap.getOrDefault(op.opcodeArgument(), Bytes32.ZERO)
              : op.result();

      op.trace(trace, hash);
    }
  }
}
