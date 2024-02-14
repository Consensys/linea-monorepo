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

package net.consensys.linea.zktracer.module.romLex;

import static net.consensys.linea.zktracer.types.AddressUtils.getCreate2Address;
import static net.consensys.linea.zktracer.types.AddressUtils.getCreateAddress;
import static org.hyperledger.besu.evm.internal.Words.clampedToLong;

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Optional;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class RomLex implements Module {
  private static final int LLARGE = 16;
  private static final RomChunkComparator romChunkComparator = new RomChunkComparator();

  private final Hub hub;
  public int codeIdentifierBeforeLexOrder = 0;

  public final StackedSet<RomChunk> chunks = new StackedSet<>();
  public final List<RomChunk> sortedChunks = new ArrayList<>();
  private Bytes byteCode = Bytes.EMPTY;
  private Address address = Address.ZERO;

  public int nextCfiBeforeReordering() {
    return this.codeIdentifierBeforeLexOrder + 1;
  }

  static class RomChunkComparator implements Comparator<RomChunk> {
    // Initialize the ChunkList
    public int compare(RomChunk chunk1, RomChunk chunk2) {
      // First sort by Address
      int addressComparison = chunk1.address().compareTo(chunk2.address());
      if (addressComparison != 0) {
        return addressComparison;
      } else {
        // Second, sort by Deployment Number
        int deploymentNumberComparison = chunk1.deploymentNumber() - chunk2.deploymentNumber();
        if (deploymentNumberComparison != 0) {
          return deploymentNumberComparison;
        } else {
          // Third sort by Deployment Status (true greater)
          if (chunk1.deploymentStatus() == chunk2.deploymentStatus()) {
            return 0;
          } else {
            return chunk1.deploymentStatus() ? -1 : 1;
          }
        }
      }
    }
  }

  @Override
  public String moduleKey() {
    return "ROM_LEX";
  }

  public RomLex(Hub hub) {
    this.hub = hub;
  }

  @Override
  public void enterTransaction() {
    this.chunks.enter();
  }

  @Override
  public void popTransaction() {
    this.chunks.pop();
  }

  public int getSortedCfiByCfi(int value) {
    if (this.sortedChunks.isEmpty()) {
      throw new RuntimeException("Chunks have not been sorted yet");
    }

    if (value == 0) {
      return 0;
    }

    for (int i = 0; i < this.sortedChunks.size(); i++) {
      if (this.sortedChunks.get(i).id() == value) {
        return i + 1;
      }
    }

    throw new RuntimeException("RomChunk n°" + value + " not found");
  }

  public int getCfiByMetadata(
      final Address address, final int deploymentNumber, final boolean deploymentStatus) {
    if (this.sortedChunks.isEmpty()) {
      throw new RuntimeException("Chunks have not been sorted yet");
    }

    for (int i = 0; i < this.sortedChunks.size(); i++) {
      final RomChunk c = this.sortedChunks.get(i);
      if (c.address().equals(address)
          && c.deploymentNumber() == deploymentNumber
          && c.deploymentStatus() == deploymentStatus) {
        return i + 1;
      }
    }

    throw new RuntimeException("RomChunk not found");
  }

  public Optional<RomChunk> getChunkByMetadata(
      final Address address, final int deploymentNumber, final boolean deploymentStatus) {
    if (this.sortedChunks.isEmpty()) {
      throw new RuntimeException("Chunks have not been sorted yet");
    }

    for (RomChunk c : this.chunks) {
      if (c.address().equals(address)
          && c.deploymentNumber() == deploymentNumber
          && c.deploymentStatus() == deploymentStatus) {
        return Optional.of(c);
      }
    }

    return Optional.empty();
  }

  @Override
  public void traceStartTx(WorldView worldView, Transaction tx) {
    // Contract creation with InitCode
    if (tx.getInit().isPresent() && !tx.getInit().orElseThrow().isEmpty()) {
      codeIdentifierBeforeLexOrder += 1;
      this.chunks.add(
          new RomChunk(
              codeIdentifierBeforeLexOrder,
              Address.contractAddress(tx.getSender(), tx.getNonce()),
              1,
              true,
              false,
              false,
              tx.getInit().get()));
    }

    // Call to an account with bytecode
    tx.getTo()
        .map(worldView::get)
        .map(AccountState::getCode)
        .ifPresent(
            code -> {
              codeIdentifierBeforeLexOrder += 1;
              int depNumber =
                  hub.transients().conflation().deploymentInfo().number(tx.getTo().get());
              boolean depStatus =
                  hub.transients().conflation().deploymentInfo().isDeploying(tx.getTo().get());

              this.chunks.add(
                  new RomChunk(
                      codeIdentifierBeforeLexOrder,
                      tx.getTo().get(),
                      depNumber,
                      depStatus,
                      true,
                      false,
                      code));
            });
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    OpCode opcode = this.hub.opCode();

    switch (opcode) {
      case CREATE -> {
        this.address = getCreateAddress(frame);
        final long offset = clampedToLong(frame.getStackItem(1));
        final long length = clampedToLong(frame.getStackItem(2));
        this.byteCode = frame.shadowReadMemory(offset, length);
        if (!this.byteCode.isEmpty()) {
          codeIdentifierBeforeLexOrder += 1;
        }
      }

      case CREATE2 -> {
        final long offset = clampedToLong(frame.getStackItem(1));
        final long length = clampedToLong(frame.getStackItem(2));
        this.byteCode = frame.shadowReadMemory(offset, length);

        if (!this.byteCode.isEmpty()) {
          codeIdentifierBeforeLexOrder += 1;
          this.address = getCreate2Address(frame);
        }
      }

      case RETURN -> {
        final long offset = clampedToLong(frame.getStackItem(0));
        final long length = clampedToLong(frame.getStackItem(1));
        final Bytes code = frame.shadowReadMemory(offset, length);
        final boolean depStatus =
            hub.transients().conflation().deploymentInfo().isDeploying(frame.getContractAddress());
        if (!code.isEmpty() && depStatus) {
          codeIdentifierBeforeLexOrder += 1;
          int depNumber =
              hub.transients().conflation().deploymentInfo().number(frame.getContractAddress());
          this.chunks.add(
              new RomChunk(
                  codeIdentifierBeforeLexOrder,
                  frame.getContractAddress(),
                  depNumber,
                  depStatus,
                  true,
                  false,
                  code));
        }
      }

      case CALL, CALLCODE, DELEGATECALL, STATICCALL -> {
        final Address calledAddress = Words.toAddress(frame.getStackItem(1));
        final boolean depStatus =
            hub.transients().conflation().deploymentInfo().isDeploying(frame.getContractAddress());
        final int depNumber =
            hub.transients().conflation().deploymentInfo().number(frame.getContractAddress());
        Optional.ofNullable(frame.getWorldUpdater().get(calledAddress))
            .map(AccountState::getCode)
            .ifPresent(
                byteCode -> {
                  codeIdentifierBeforeLexOrder += 1;
                  this.chunks.add(
                      new RomChunk(
                          codeIdentifierBeforeLexOrder,
                          calledAddress,
                          depNumber,
                          depStatus,
                          true,
                          false,
                          byteCode));
                });
      }

      case EXTCODECOPY -> {
        final Address calledAddress = Words.toAddress(frame.getStackItem(0));
        final long length = Words.clampedToLong(frame.getStackItem(3));
        final boolean isDeploying =
            hub.transients().conflation().deploymentInfo().isDeploying(frame.getContractAddress());
        if (length == 0 || isDeploying) {
          return;
        }
        final int depNumber =
            hub.transients().conflation().deploymentInfo().number(frame.getContractAddress());
        Optional.ofNullable(frame.getWorldUpdater().get(calledAddress))
            .map(AccountState::getCode)
            .ifPresent(
                byteCode -> {
                  if (!byteCode.isEmpty()) {
                    codeIdentifierBeforeLexOrder += 1;
                    this.chunks.add(
                        new RomChunk(
                            codeIdentifierBeforeLexOrder,
                            calledAddress,
                            depNumber,
                            isDeploying,
                            true,
                            false,
                            byteCode));
                  }
                });
      }
    }
  }

  @Override
  public void tracePostOpcode(MessageFrame frame) {
    OpCode opcode = OpCode.of(frame.getCurrentOperation().getOpcode());
    switch (opcode) {
      case CREATE, CREATE2 -> {
        final int depNumber = hub.transients().conflation().deploymentInfo().number(this.address);
        final boolean depStatus =
            hub.transients().conflation().deploymentInfo().isDeploying(this.address);

        this.chunks.add(
            new RomChunk(
                codeIdentifierBeforeLexOrder,
                this.address,
                depNumber,
                depStatus,
                true,
                false,
                this.byteCode));
      }
    }
  }

  private void traceChunk(
      final RomChunk chunk, int cfi, int codeFragmentIndexInfinity, Trace trace) {
    trace
        .codeFragmentIndex(Bytes.ofUnsignedInt(cfi))
        .codeFragmentIndexInfty(Bytes.ofUnsignedInt(codeFragmentIndexInfinity))
        .codeSize(Bytes.ofUnsignedInt(chunk.byteCode().size()))
        .addrHi(chunk.address().slice(0, 4))
        .addrLo(chunk.address().slice(4, LLARGE))
        .commitToState(chunk.commitToTheState())
        .depNumber(Bytes.ofUnsignedInt(chunk.deploymentNumber()))
        .depStatus(chunk.deploymentStatus())
        .readFromState(chunk.readFromTheState())
        .validateRow();
  }

  @Override
  public void traceEndConflation(final WorldView state) {
    this.sortedChunks.addAll(this.chunks);
    this.sortedChunks.sort(romChunkComparator);
  }

  @Override
  public int lineCount() {
    // WARN: the line count for the RomLex is the *number of code fragments*, not their actual line
    // count – that's for the ROM.
    return this.chunks.size();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    final int codeFragmentIndexInfinity = chunks.size();

    int cfi = 0;
    for (RomChunk chunk : sortedChunks) {
      cfi += 1;
      traceChunk(chunk, cfi, codeFragmentIndexInfinity, trace);
    }
  }
}
