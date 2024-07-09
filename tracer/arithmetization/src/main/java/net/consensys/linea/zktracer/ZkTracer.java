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
import java.io.PrintStream;
import java.io.PrintWriter;
import java.io.RandomAccessFile;
import java.io.StringWriter;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.plugins.config.LineaL1L2BridgeConfiguration;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.Pin55;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.toml.Toml;
import org.apache.tuweni.toml.TomlTable;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.ethereum.mainnet.feemarket.FeeMarket;
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
public class ZkTracer implements ConflationAwareOperationTracer {
  /** The {@link GasCalculator} used in this version of the arithmetization */
  public static final GasCalculator gasCalculator = new LondonGasCalculator();

  public static final FeeMarket feeMarket = FeeMarket.london(-1);

  @Getter private final Hub hub;
  private final Optional<Pin55> pin55;
  private final Map<String, Integer> spillings = new HashMap<>();
  private Hash hashOfLastTransactionTraced = Hash.EMPTY;

  /** Accumulate all the exceptions that happened at tracing time. */
  @Getter private final List<Exception> tracingExceptions = new FiniteList<>(50);

  public ZkTracer() {
    this(LineaL1L2BridgeConfiguration.EMPTY);
  }

  public ZkTracer(final LineaL1L2BridgeConfiguration bridgeConfiguration) {
    this.hub = new Hub(bridgeConfiguration.contract(), bridgeConfiguration.topic());

    // Load opcodes configured in src/main/resources/opcodes.yml.
    OpCodes.load();
    // Load spillings configured in src/main/resources/spillings.toml.
    try {
      final TomlTable table =
          Toml.parse(getClass().getClassLoader().getResourceAsStream("spillings.toml"))
              .getTable("spillings");
      table.toMap().keySet().forEach(k -> spillings.put(k, Math.toIntExact(table.getLong(k))));

      for (Module m : this.hub.getModulesToCount()) {
        if (!this.spillings.containsKey(m.moduleKey())) {
          throw new IllegalStateException(
              "Spilling for module " + m.moduleKey() + " not defined in spillings.toml");
        }
      }
    } catch (final Exception e) {
      throw new RuntimeException(e);
    }

    // >>>> CHANGE ME >>>>
    // >>>> CHANGE ME >>>>
    // >>>> CHANGE ME >>>>
    final Pin55.PinLevel debugLevel = new Pin55.PinLevel();
    // <<<< CHANGE ME <<<<
    // <<<< CHANGE ME <<<<
    // <<<< CHANGE ME <<<<
    this.pin55 =
        debugLevel.none() ? Optional.empty() : Optional.of(new Pin55(debugLevel, this.hub));
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

  public void writeToFile(final Path filename) {
    maybeThrowTracingExceptions();

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
    try {
      hub.traceStartConflation(numBlocksInConflation);
      this.pin55.ifPresent(x -> x.traceStartConflation(numBlocksInConflation));
    } catch (final Exception e) {
      this.tracingExceptions.add(e);
    }
  }

  @Override
  public void traceEndConflation(final WorldView state) {
    try {
      this.hub.traceEndConflation(state);
      this.pin55.ifPresent(Pin55::traceEndConflation);
    } catch (final Exception e) {
      this.tracingExceptions.add(e);
    }

    if (!this.tracingExceptions.isEmpty()) {
      throw new TracingExceptions(this.tracingExceptions);
    }
  }

  @Override
  public void traceStartBlock(final ProcessableBlockHeader processableBlockHeader) {
    try {
      this.hub.traceStartBlock(processableBlockHeader);
      this.pin55.ifPresent(Pin55::traceEndConflation);
    } catch (final Exception e) {
      this.tracingExceptions.add(e);
    }
  }

  @Override
  public void traceStartBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    try {
      this.hub.traceStartBlock(blockHeader);
      this.pin55.ifPresent(x -> x.traceStartBlock(blockHeader, blockBody));
    } catch (final Exception e) {
      this.tracingExceptions.add(e);
    }
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    try {
      this.hub.traceEndBlock(blockHeader, blockBody);
      this.pin55.ifPresent(Pin55::traceEndBlock);
    } catch (final Exception e) {
      this.tracingExceptions.add(e);
    }
  }

  //  @Override
  public void tracePrepareTransaction(WorldView worldView, Transaction transaction) {
    try {
      hashOfLastTransactionTraced = transaction.getHash();
      this.pin55.ifPresent(x -> x.tracePrepareTx(worldView, transaction));
      this.hub.traceStartTx(worldView, transaction);
    } catch (final Exception e) {
      this.tracingExceptions.add(e);
    }
  }

  public void traceEndTransaction(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed,
      Set<Address> selfDestructs,
      long timeNs) {
    try {
      this.pin55.ifPresent(x -> x.traceEndTx(worldView, tx, status, output, logs, gasUsed));
      this.hub.traceEndTx(worldView, tx, status, output, logs, gasUsed);
    } catch (final Exception e) {
      this.tracingExceptions.add(e);
    }
  }

  @Override
  public void tracePreExecution(final MessageFrame frame) {
    if (frame.getCode().getSize() > 0) {
      try {
        this.hub.tracePreOpcode(frame);
        this.pin55.ifPresent(x -> x.tracePreOpcode(frame));
      } catch (final Exception e) {
        this.tracingExceptions.add(e);
      }
    }
  }

  @Override
  public void tracePostExecution(MessageFrame frame, Operation.OperationResult operationResult) {
    if (frame.getCode().getSize() > 0) {
      try {
        this.hub.tracePostExecution(frame, operationResult);
        this.pin55.ifPresent(x -> x.tracePostOpcode(frame, operationResult));
      } catch (final Exception e) {
        this.tracingExceptions.add(e);
      }
    }
  }

  @Override
  public void traceContextEnter(MessageFrame frame) {
    // We only want to trigger on creation of new contexts, not on re-entry in
    // existing contexts
    if (frame.getState() == MessageFrame.State.NOT_STARTED) {
      try {
        this.hub.traceContextEnter(frame);
        this.pin55.ifPresent(x -> x.traceContextEnter(frame));
      } catch (final Exception e) {
        this.tracingExceptions.add(e);
      }
    }
  }

  @Override
  public void traceContextReEnter(MessageFrame frame) {
    try {
      this.hub.traceContextReEnter(frame);
      this.pin55.ifPresent(x -> x.traceContextReEnter(frame));
    } catch (final Exception e) {
      this.tracingExceptions.add(e);
    }
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    try {
      this.hub.traceContextExit(frame);
      this.pin55.ifPresent(x -> x.traceContextExit(frame));
    } catch (final Exception e) {
      this.tracingExceptions.add(e);
    }
  }

  /** When called, erase all tracing related to the last included transaction. */
  public void popTransaction(final PendingTransaction pendingTransaction) {
    if (hashOfLastTransactionTraced.equals(pendingTransaction.getTransaction().getHash())) {
      hub.popTransaction();
    }
  }

  private void maybeThrowTracingExceptions() {
    if (!this.tracingExceptions.isEmpty()) {
      throw new TracingExceptions(this.tracingExceptions);
    }
  }

  public Map<String, Integer> getModulesLineCount() {
    maybeThrowTracingExceptions();
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
    modulesLineCount.put("BLOCK_TRANSACTIONS", hub.cumulatedTxCount());
    return modulesLineCount;
  }

  /** Gather all and any exception that happened during tracing under a common umbrella. */
  @Getter
  @RequiredArgsConstructor
  private static class TracingExceptions extends RuntimeException {
    private final List<Exception> tracingExceptions;

    @Override
    public String getMessage() {
      final StringBuilder msg = new StringBuilder("Exceptions triggered while tracing:\n");
      for (final Exception e : tracingExceptions) {
        msg.append("  - ").append(e.getMessage()).append("\n");
      }
      return msg.toString();
    }

    @Override
    public void printStackTrace(PrintStream s) {
      for (final Exception e : this.tracingExceptions) {
        e.printStackTrace(s);
      }
    }

    @Override
    public String toString() {
      StringWriter stringWriter = new StringWriter();
      PrintWriter s = new PrintWriter(stringWriter);
      for (final Exception e : this.tracingExceptions) {
        s.append("\n");
        e.printStackTrace(s);
      }
      return stringWriter.toString();
    }
  }

  /** An {@link ArrayList} with an upper bound on the number of element it can store. */
  @RequiredArgsConstructor
  private static class FiniteList<T> extends ArrayList<T> {
    /** The maximal number of elements in this list. */
    private final int maxLength;

    @Override
    public void add(int index, T element) {
      throw new UnsupportedOperationException();
    }

    @Override
    public boolean addAll(Collection<? extends T> c) {
      throw new UnsupportedOperationException();
    }

    @Override
    public boolean addAll(int index, Collection<? extends T> c) {
      throw new UnsupportedOperationException();
    }

    @Override
    public boolean add(T t) {
      if (this.size() < this.maxLength) {
        return super.add(t);
      }
      return false;
    }
  }
}
