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

package net.consensys.linea.zktracer.module.romlex;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;
import static net.consensys.linea.zktracer.types.AddressUtils.getDeploymentAddress;
import static net.consensys.linea.zktracer.types.AddressUtils.highPart;
import static net.consensys.linea.zktracer.types.AddressUtils.lowPart;

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostOpcodeDefer;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;

@Accessors(fluent = true)
public class RomLex implements Module, PostOpcodeDefer {
  private static final RomChunkComparator ROM_CHUNK_COMPARATOR = new RomChunkComparator();

  private final Hub hub;

  @Getter private final StackedSet<RomChunk> chunks = new StackedSet<>();
  @Getter private final List<RomChunk> sortedChunks = new ArrayList<>();
  private Bytes byteCode = Bytes.EMPTY;
  private Address address = Address.ZERO;

  @Getter private final DeferRegistry createDefers = new DeferRegistry();

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

  public int getCodeFragmentIndexByMetadata(
      final Address address, final int deploymentNumber, final boolean depStatus) {
    return getCodeFragmentIndexByMetadata(
        ContractMetadata.make(address, deploymentNumber, depStatus));
  }

  public int getCodeFragmentIndexByMetadata(final ContractMetadata metadata) {
    if (this.sortedChunks.isEmpty()) {
      throw new RuntimeException("Chunks have not been sorted yet");
    }

    for (int i = 0; i < this.sortedChunks.size(); i++) {
      final RomChunk c = this.sortedChunks.get(i);
      if (c.metadata().equals(metadata)) {
        return i + 1;
      }
    }

    throw new RuntimeException(
        String.format(
            "RomChunk with address %s, deployment number %s and deploymentStatus %s not found",
            metadata.address(), metadata.deploymentNumber(), !metadata.underDeployment()));
  }

  public Optional<RomChunk> getChunkByMetadata(final ContractMetadata metadata) {
    for (RomChunk c : this.chunks) {
      if (c.metadata().equals(metadata)) {
        return Optional.of(c);
      }
    }

    return Optional.empty();
  }

  @Override
  public void traceStartTx(WorldView worldView, TransactionProcessingMetadata txMetaData) {
    final Transaction tx = txMetaData.getBesuTransaction();
    // Contract creation with InitCode
    if (tx.getInit().isPresent() && !tx.getInit().get().isEmpty()) {
      final Address calledAddress = Address.contractAddress(tx.getSender(), tx.getNonce());
      final RomChunk chunk =
          new RomChunk(
              ContractMetadata.underDeployment(calledAddress, 1), false, false, tx.getInit().get());

      this.chunks.add(chunk);
    }

    // Call to an account with bytecode
    tx.getTo()
        .map(worldView::get)
        .map(AccountState::getCode)
        .ifPresent(
            code -> {
              if (!code.isEmpty()) {

                final Address calledAddress = tx.getTo().get();
                final int depNumber =
                    hub.transients().conflation().deploymentInfo().number(calledAddress);
                final boolean depStatus =
                    hub.transients().conflation().deploymentInfo().isDeploying(calledAddress);

                final RomChunk chunk =
                    new RomChunk(
                        ContractMetadata.make(calledAddress, depNumber, depStatus),
                        true,
                        false,
                        code);

                this.chunks.add(chunk);
              }
            });
  }

  public void callRomLex(final MessageFrame frame) {
    switch (OpCode.of(frame.getCurrentOperation().getOpcode())) {
      case CREATE, CREATE2 -> {
        hub.defers().scheduleForPostExecution(this);
        final long offset = Words.clampedToLong(frame.getStackItem(1));
        final long length = Words.clampedToLong(frame.getStackItem(2));
        byteCode = frame.shadowReadMemory(offset, length);
        address = getDeploymentAddress(frame);
      }

      case RETURN -> {
        Preconditions.checkArgument(frame.getType() == MessageFrame.Type.CONTRACT_CREATION);

        final Bytes code = hub.transients().op().outputData();

        if (code.isEmpty()) {
          return;
        }

        final boolean depStatus =
            hub.transients().conflation().deploymentInfo().isDeploying(frame.getContractAddress());
        if (depStatus) {
          int depNumber =
              hub.transients().conflation().deploymentInfo().number(frame.getContractAddress());
          final ContractMetadata contractMetadata =
              ContractMetadata.underDeployment(frame.getContractAddress(), depNumber);

          final RomChunk chunk = new RomChunk(contractMetadata, true, false, code);
          this.chunks.add(chunk);
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
                  if (!byteCode.isEmpty()) {
                    final RomChunk chunk =
                        new RomChunk(
                            ContractMetadata.make(calledAddress, depNumber, depStatus),
                            true,
                            false,
                            byteCode);
                    chunks.add(chunk);
                  }
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
                    final RomChunk chunk =
                        new RomChunk(
                            ContractMetadata.make(calledAddress, depNumber, false),
                            true,
                            false,
                            byteCode);

                    this.chunks.add(chunk);
                  }
                });
      }
    }
  }

  @Override
  public void resolvePostExecution(
      Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {
    Preconditions.checkArgument(hub.opCode().isCreate());
    final int depNumber = hub.transients().conflation().deploymentInfo().number(this.address);
    final ContractMetadata contractMetadata =
        ContractMetadata.underDeployment(this.address, depNumber);

    final RomChunk chunk = new RomChunk(contractMetadata, true, false, this.byteCode);
    this.chunks.add(chunk);
    this.createDefers.trigger(contractMetadata);
  }

  // This is the tracing for ROMLEX module
  private void traceChunk(
      final RomChunk chunk, final int cfi, final int codeFragmentIndexInfinity, Trace trace) {
    final Hash codeHash =
        chunk.metadata().underDeployment() ? Hash.EMPTY : Hash.hash(chunk.byteCode());
    trace
        .codeFragmentIndex(cfi)
        .codeFragmentIndexInfty(codeFragmentIndexInfinity)
        .codeSize(chunk.byteCode().size())
        .addressHi(highPart(chunk.metadata().address()))
        .addressLo(lowPart(chunk.metadata().address()))
        .commitToState(chunk.commitToTheState())
        .deploymentNumber(chunk.metadata().deploymentNumber())
        .deploymentStatus(chunk.metadata().underDeployment())
        .readFromState(chunk.readFromTheState())
        .codeHashHi(codeHash.slice(0, LLARGE))
        .codeHashLo(codeHash.slice(LLARGE, LLARGE))
        .validateRow();
  }

  public void determineCodeFragmentIndex() {
    this.sortedChunks.addAll(this.chunks);
    this.sortedChunks.sort(ROM_CHUNK_COMPARATOR);
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
