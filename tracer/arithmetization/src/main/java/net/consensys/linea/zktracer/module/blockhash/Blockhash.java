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

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;
import static net.consensys.linea.zktracer.opcode.OpCode.*;

import java.nio.MappedByteBuffer;
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
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
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

  private short relBlock;
  private long absBlock;

  private Bytes32 blockhashArg;

  public Blockhash(Hub hub, Wcp wcp) {
    this.hub = hub;
    this.wcp = wcp;
    this.relBlock = 0;
  }

  @Override
  public String moduleKey() {
    return "BLOCK_HASH";
  }

  @Override
  public void traceStartBlock(
      final ProcessableBlockHeader processableBlockHeader, final Address miningBeneficiary) {
    relBlock += 1;
    absBlock = processableBlockHeader.getNumber();
  }

  @Override
  public void tracePreOpcode(MessageFrame frame, OpCode opcode) {
    if (opcode == BLOCKHASH) {

      blockhashArg = Bytes32.leftPad(frame.getStackItem(0));

      hub.defers().scheduleForPostExecution(this);
    }
  }

  @Override
  public void resolvePostExecution(
      Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {

    final Bytes32 blockhashRes = Bytes32.leftPad(frame.getStackItem(0));
    operations.add(new BlockhashOperation(relBlock, absBlock, blockhashArg, blockhashRes, wcp));
    // We have 4 LLARGE and one OLI call to WCP, made at the end of the conflation, so we need to
    // add line count to WCP
    wcp.additionalRows.add(4 * LLARGE + 1);
    if (blockhashRes != Bytes32.ZERO) {
      blockHashMap.put(blockhashArg, blockhashRes);
    }
  }

  /**
   * Operations are sorted wrt blockhashArg and the wcp module is called accordingly. We must call
   * the WCP module before calling {@link #commit(List<MappedByteBuffer>)} as the headers sizes must
   * be computed with the final list of operations ready.
   */
  @Override
  public void traceEndConflation(WorldView state) {
    OperationSetModule.super.traceEndConflation(state);
    sortedOperations = sortOperations(new BlockhashComparator());
    Bytes32 prevBlockhashArg = Bytes32.ZERO;
    for (BlockhashOperation op : sortedOperations) {
      op.handlePreprocessing(prevBlockhashArg);
      prevBlockhashArg = op.blockhashArg();
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
      final Bytes32 blockhashVal =
          op.blockhashRes() == Bytes32.ZERO
              ? this.blockHashMap.getOrDefault(op.blockhashArg(), Bytes32.ZERO)
              : op.blockhashRes();
      op.traceMacro(trace, blockhashVal);
      op.tracePreprocessing(trace);
    }
  }
}
