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

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Optional;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.worldstate.WorldView;

@Accessors(fluent = true)
public class RomLex implements Module {
  private static final int LLARGE = 16;
  private static final RomChunkComparator romChunkComparator = new RomChunkComparator();

  private final Hub hub;

  public final StackedSet<RomChunk> chunks = new StackedSet<>();
  public final List<RomChunk> sortedChunks = new ArrayList<>();
  private Bytes byteCode = Bytes.EMPTY;
  private Address address = Address.ZERO;

  @Getter private final DeferRegistry returnDefers = new DeferRegistry();
  @Getter private final DeferRegistry createDefers = new DeferRegistry();

  static class RomChunkComparator implements Comparator<RomChunk> {
    // Initialize the ChunkList
    public int compare(RomChunk chunk1, RomChunk chunk2) {
      // First sort by Address
      int addressComparison = chunk1.metadata().address().compareTo(chunk2.metadata().address());
      if (addressComparison != 0) {
        return addressComparison;
      } else {
        // Second, sort by Deployment Number
        int deploymentNumberComparison =
            chunk1.metadata().deploymentNumber() - chunk2.metadata().deploymentNumber();
        if (deploymentNumberComparison != 0) {
          return deploymentNumberComparison;
        } else {
          // Third sort by Deployment Status (true greater)
          if (chunk1.metadata().underDeployment() == chunk2.metadata().underDeployment()) {
            return 0;
          } else {
            return chunk1.metadata().underDeployment() ? -1 : 1;
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

  public int getCfiByMetadata(final ContractMetadata metadata) {
    if (this.sortedChunks.isEmpty()) {
      throw new RuntimeException("Chunks have not been sorted yet");
    }

    for (int i = 0; i < this.sortedChunks.size(); i++) {
      final RomChunk c = this.sortedChunks.get(i);
      if (c.metadata().equals(metadata)) {
        return i + 1;
      }
    }

    throw new RuntimeException("RomChunk not found");
  }

  public Optional<RomChunk> getChunkByMetadata(final ContractMetadata metadata) {
    if (this.sortedChunks.isEmpty()) {
      throw new RuntimeException("Chunks have not been sorted yet");
    }

    for (RomChunk c : this.chunks) {
      if (c.metadata().equals(metadata)) {
        return Optional.of(c);
      }
    }

    return Optional.empty();
  }

  @Override
  public void traceStartTx(WorldView worldView, Transaction tx) {
    // Contract creation with InitCode
    if (tx.getInit().isPresent() && !tx.getInit().orElseThrow().isEmpty()) {
      this.chunks.add(
          new RomChunk(
              ContractMetadata.underDeployment(
                  Address.contractAddress(tx.getSender(), tx.getNonce()), 1),
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
              int depNumber =
                  hub.transients().conflation().deploymentInfo().number(tx.getTo().get());
              boolean depStatus =
                  hub.transients().conflation().deploymentInfo().isDeploying(tx.getTo().get());

              this.chunks.add(
                  new RomChunk(
                      ContractMetadata.make(tx.getTo().get(), depNumber, depStatus),
                      true,
                      false,
                      code));
            });
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    switch (this.hub.opCode()) {
      case CREATE -> {
        this.address = getCreateAddress(frame);
        this.byteCode = this.hub.transients().op().callData();
      }

      case CREATE2 -> {
        this.byteCode = this.hub.transients().op().callData();
        if (!this.byteCode.isEmpty()) {
          this.address = getCreate2Address(frame);
        }
      }

      case RETURN -> {
        final Bytes code = hub.transients().op().returnData();
        final boolean depStatus =
            hub.transients().conflation().deploymentInfo().isDeploying(frame.getContractAddress());
        if (!code.isEmpty() && depStatus) {
          int depNumber =
              hub.transients().conflation().deploymentInfo().number(frame.getContractAddress());
          final ContractMetadata contractMetadata =
              ContractMetadata.underDeployment(frame.getContractAddress(), depNumber);

          this.chunks.add(new RomChunk(contractMetadata, true, false, code));
          this.returnDefers.trigger(contractMetadata);
        }
        this.returnDefers.clear();
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
                byteCode ->
                    this.chunks.add(
                        new RomChunk(
                            ContractMetadata.make(calledAddress, depNumber, depStatus),
                            true,
                            false,
                            byteCode)));
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
                    this.chunks.add(
                        new RomChunk(
                            ContractMetadata.make(calledAddress, depNumber, false),
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
    switch (hub.opCode()) {
      case CREATE, CREATE2 -> {
        final int depNumber = hub.transients().conflation().deploymentInfo().number(this.address);
        final boolean depStatus =
            hub.transients().conflation().deploymentInfo().isDeploying(this.address);
        final ContractMetadata contractMetadata =
            ContractMetadata.make(this.address, depNumber, depStatus);
        this.chunks.add(new RomChunk(contractMetadata, true, false, this.byteCode));
        this.createDefers.trigger(contractMetadata);
      }
    }
  }

  private void traceChunk(
      final RomChunk chunk, int cfi, int codeFragmentIndexInfinity, Trace trace) {
    trace
        .codeFragmentIndex(cfi)
        .codeFragmentIndexInfty(codeFragmentIndexInfinity)
        .codeSize(chunk.byteCode().size())
        .addressHi(chunk.metadata().address().slice(0, 4).toLong())
        .addressLo(chunk.metadata().address().slice(4, LLARGE))
        .commitToState(chunk.commitToTheState())
        .deploymentNumber(chunk.metadata().deploymentNumber())
        .deploymentStatus(chunk.metadata().underDeployment())
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
