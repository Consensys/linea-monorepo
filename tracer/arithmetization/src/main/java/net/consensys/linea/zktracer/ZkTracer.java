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

package net.consensys.linea.zktracer;

import java.io.IOException;
import java.io.RandomAccessFile;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.toml.Toml;
import org.apache.tuweni.toml.TomlTable;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.gascalculator.LondonGasCalculator;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

@Slf4j
public class ZkTracer implements ZkBlockAwareOperationTracer {
  /** The {@link GasCalculator} used in this version of the arithmetization */
  public static final GasCalculator gasCalculator = new LondonGasCalculator();

  @Getter private final Hub hub = new Hub();
  private final Map<String, Integer> spillings = new HashMap<>();
  private Hash hashOfLastTransactionTraced = Hash.EMPTY;

  public ZkTracer() {
    // Load opcodes configured in src/main/resources/opcodes.yml.
    OpCodes.load();

    // Load spillings configured in src/main/resources/spillings.toml.
    try {
      final TomlTable table =
          Toml.parse(getClass().getClassLoader().getResourceAsStream("spillings.toml"))
              .getTable("spillings");
      table.toMap().keySet().forEach(k -> spillings.put(k, Math.toIntExact(table.getLong(k))));
    } catch (final Exception e) {
      throw new RuntimeException(e);
    }
  }

  public Path writeToTmpFile() {
    try {
      final Path traceFile = Files.createTempFile(null, ".lt");
      this.writeToFile(traceFile);
      return traceFile;
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  public Path writeToTmpFile(final Path rootDir) {
    try {
      final Path traceFile = Files.createTempFile(rootDir, null, ".lt");
      this.writeToFile(traceFile);
      return traceFile;
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  @Override
  public void writeToFile(final Path filename) {
    final List<Module> modules = this.hub.getModulesToTrace();
    final List<ColumnHeader> traceMap =
        modules.stream().flatMap(m -> m.columnsHeaders().stream()).toList();
    final int headerSize = traceMap.stream().mapToInt(ColumnHeader::headerSize).sum() + 4;

    try (RandomAccessFile file = new RandomAccessFile(filename.toString(), "rw")) {
      file.setLength(traceMap.stream().mapToLong(ColumnHeader::cumulatedSize).sum());
      MappedByteBuffer header =
          file.getChannel().map(FileChannel.MapMode.READ_WRITE, 0, headerSize);

      header.putInt(traceMap.size());
      for (ColumnHeader h : traceMap) {
        final String name = h.name();
        header.putShort((short) name.length());
        header.put(name.getBytes());
        header.put((byte) h.bytesPerElement());
        header.putInt(h.length());
      }
      long offset = headerSize;
      for (Module m : modules) {
        List<MappedByteBuffer> buffers = new ArrayList<>();
        for (ColumnHeader columnHeader : m.columnsHeaders()) {
          final int columnLength = columnHeader.dataSize();
          buffers.add(file.getChannel().map(FileChannel.MapMode.READ_WRITE, offset, columnLength));
          offset += columnLength;
        }
        m.commit(buffers);
      }
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  @Override
  public void traceStartConflation(final long numBlocksInConflation) {
    hub.traceStartConflation(numBlocksInConflation);
  }

  @Override
  public void traceEndConflation() {
    this.hub.traceEndConflation();
  }

  @Override
  public void traceStartBlock(final ProcessableBlockHeader processableBlockHeader) {
    this.hub.traceStartBlock(processableBlockHeader);
  }

  @Override
  public void traceStartBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.hub.traceStartBlock(blockHeader);
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.hub.traceEndBlock(blockHeader, blockBody);
  }

  @Override
  public void traceStartTransaction(WorldView worldView, Transaction transaction) {
    hashOfLastTransactionTraced = transaction.getHash();
    this.hub.traceStartTx(worldView, transaction);
  }

  @Override
  public void traceEndTransaction(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed,
      long timeNs) {
    this.hub.traceEndTx(worldView, tx, status, output, logs, gasUsed);
  }

  @Override
  public void tracePreExecution(final MessageFrame frame) {
    if (frame.getCode().getSize() > 0) {
      this.hub.tracePreOpcode(frame);
    }
  }

  @Override
  public void tracePostExecution(MessageFrame frame, Operation.OperationResult operationResult) {
    if (frame.getCode().getSize() > 0) {
      this.hub.tracePostExecution(frame, operationResult);
    }
  }

  @Override
  public void traceContextEnter(MessageFrame frame) {
    // We only want to trigger on creation of new contexts, not on re-entry in
    // existing contexts
    if (frame.getState() == MessageFrame.State.NOT_STARTED) {
      this.hub.traceContextEnter(frame);
    }
  }

  @Override
  public void traceContextReEnter(MessageFrame frame) {
    this.hub.traceContextReEnter(frame);
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    this.hub.traceContextExit(frame);
  }

  /** When called, erase all tracing related to the last included transaction. */
  public void popTransaction(final PendingTransaction pendingTransaction) {
    if (hashOfLastTransactionTraced.equals(pendingTransaction.getTransaction().getHash())) {
      hub.popTransaction();
    }
  }

  public Map<String, Integer> getModulesLineCount() {
    final HashMap<String, Integer> modulesLineCount = new HashMap<>();
    hub.getModulesToCount()
        .forEach(
            m ->
                modulesLineCount.put(
                    m.moduleKey(),
                    m.lineCount()
                        + Optional.ofNullable(this.spillings.get(m.moduleKey()))
                            .orElseThrow(
                                () ->
                                    new IllegalStateException(
                                        "Module "
                                            + m.moduleKey()
                                            + " not found in spillings.toml"))));
    modulesLineCount.put("BLOCK_TX", hub.tx().number());
    return modulesLineCount;
  }
}
