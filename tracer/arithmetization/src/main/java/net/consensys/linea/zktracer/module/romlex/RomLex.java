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

import static com.google.common.base.Preconditions.*;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;
import static net.consensys.linea.zktracer.opcode.OpCode.*;
import static net.consensys.linea.zktracer.types.AddressUtils.getDeploymentAddress;
import static net.consensys.linea.zktracer.types.AddressUtils.highPart;
import static net.consensys.linea.zktracer.types.AddressUtils.lowPart;

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.ContextExitDefer;
import net.consensys.linea.zktracer.module.hub.defer.ImmediateContextEntryDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostOpcodeDefer;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
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
public class RomLex
    implements Module, PostOpcodeDefer, ImmediateContextEntryDefer, ContextExitDefer {
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
    chunks.enter();
  }

  @Override
  public void popTransaction() {
    chunks.pop();
  }

  public int getCodeFragmentIndexByMetadata(
      final Address address, final int deploymentNumber, final boolean depStatus) {
    return getCodeFragmentIndexByMetadata(
        ContractMetadata.make(address, deploymentNumber, depStatus));
  }

  public int getCodeFragmentIndexByMetadata(final ContractMetadata metadata) {
    if (sortedChunks.isEmpty()) {
      throw new RuntimeException("Chunks have not been sorted yet");
    }

    for (int i = 0; i < sortedChunks.size(); i++) {
      final RomChunk c = sortedChunks.get(i);
      if (c.metadata().equals(metadata)) {
        return i + 1;
      }
    }

    throw new RuntimeException(
        String.format(
            "RomChunk with address %s, deployment number %s and deploymentStatus %s not found",
            metadata.address(), metadata.deploymentNumber(), metadata.underDeployment()));
  }

  public Optional<RomChunk> getChunkByMetadata(final ContractMetadata metadata) {
    for (RomChunk c : chunks) {
      if (c.metadata().equals(metadata)) {
        return Optional.of(c);
      }
    }

    return Optional.empty();
  }

  public Bytes getCodeByMetadata(final ContractMetadata metadata) {
    return getChunkByMetadata(metadata).map(RomChunk::byteCode).orElse(Bytes.EMPTY);
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

      chunks.add(chunk);
    }

    // Call to an account with bytecode
    tx.getTo()
        .map(worldView::get)
        .map(AccountState::getCode)
        .ifPresent(
            code -> {
              if (!code.isEmpty()) {

                final Address calledAddress = tx.getTo().get();
                final int depNumber = hub.deploymentNumberOf(calledAddress);
                final boolean depStatus = hub.deploymentStatusOf(calledAddress);

                final RomChunk chunk =
                    new RomChunk(
                        ContractMetadata.make(calledAddress, depNumber, depStatus),
                        true,
                        false,
                        code);

                chunks.add(chunk);
              }
            });
  }

  public void callRomLex(final MessageFrame frame) {
    switch (OpCode.of(frame.getCurrentOperation().getOpcode())) {
      case CREATE, CREATE2 -> {
        final long offset = Words.clampedToLong(frame.getStackItem(1));
        final long length = Words.clampedToLong(frame.getStackItem(2));

        checkArgument(length > 0, "callRomLex expects positive size for CREATE(2)");

        hub.defers().scheduleForImmediateContextEntry(this);
        byteCode = frame.shadowReadMemory(offset, length);
        address = getDeploymentAddress(frame);
      }

      case RETURN -> {
        final boolean currentDeploymentStatus = hub.deploymentStatusOfBytecodeAddress();
        final int currentDeploymentNumber = hub.deploymentNumberOfBytecodeAddress();

        checkArgument(frame.getType() == MessageFrame.Type.CONTRACT_CREATION);
        checkArgument(currentDeploymentStatus);
        checkArgument(currentDeploymentNumber > 0);

        final long offset = Words.clampedToLong(frame.getStackItem(0));
        final long length = Words.clampedToLong(frame.getStackItem(1));
        checkArgument(length > 0, "callRomLex expects positive size for RETURN");

        byteCode = frame.shadowReadMemory(offset, length);
        hub.defers().scheduleForContextExit(this, hub.currentFrame().id());
      }

      case CALL, CALLCODE, DELEGATECALL, STATICCALL -> {
        final Address calledAddress = Words.toAddress(frame.getStackItem(1));
        final int depNumber = hub.deploymentNumberOf(calledAddress);
        final boolean depStatus = hub.deploymentStatusOf(calledAddress);

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
        final Address foreignCodeAddress = Words.toAddress(frame.getStackItem(0));
        final long length = Words.clampedToLong(frame.getStackItem(3));

        final int foreignDeploymentNumber = hub.deploymentNumberOf(foreignCodeAddress);
        final boolean foreignDeploymentStatus = hub.deploymentStatusOf(foreignCodeAddress);

        checkArgument(
            length > 0,
            "EXTCODECOPY should only trigger a ROM_LEX chunk if nonzero size parameter");
        checkArgument(
            !foreignDeploymentStatus,
            "EXTCODECOPY should only trigger a ROM_LEX chunk if its target isn't currently deploying");
        checkArgument(
            !frame.getWorldUpdater().getAccount(foreignCodeAddress).isEmpty()
                && frame.getWorldUpdater().getAccount(foreignCodeAddress).hasCode());

        Optional.ofNullable(frame.getWorldUpdater().get(foreignCodeAddress))
            .map(AccountState::getCode)
            .ifPresent(
                byteCode -> {
                  if (!byteCode.isEmpty()) {
                    final RomChunk chunk =
                        new RomChunk(
                            ContractMetadata.make(
                                foreignCodeAddress, foreignDeploymentNumber, false),
                            true,
                            false,
                            byteCode);

                    chunks.add(chunk);
                  }
                });
      }
    }
  }

  @Override
  public void resolvePostExecution(
      Hub hub, MessageFrame frame, Operation.OperationResult operationResult) {}

  @Override
  public void resolveUponImmediateContextEntry(Hub hub) {
    checkArgument(hub.messageFrame().getType() == MessageFrame.Type.CONTRACT_CREATION);

    final int deploymentNumber = hub.deploymentNumberOf(address);
    final boolean deploymentStatus = hub.deploymentStatusOf(address);
    final ContractMetadata contractMetadata =
        ContractMetadata.underDeployment(address, deploymentNumber);

    checkArgument(deploymentStatus, "After a CREATE the deployment status should be true");

    final RomChunk chunk = new RomChunk(contractMetadata, true, false, byteCode);
    chunks.add(chunk);
    createDefers.trigger(contractMetadata);
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
    sortedChunks.addAll(chunks);
    sortedChunks.sort(ROM_CHUNK_COMPARATOR);
  }

  @Override
  public int lineCount() {
    // WARN: the line count for the RomLex is the *number of code fragments*, not their actual line
    // count – that's for the ROM.
    return chunks.size();
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

  @Override
  public void resolveUponExitingContext(Hub hub, CallFrame frame) {

    checkArgument(hub.opCode() == RETURN);

    final ContractMetadata contractMetadata =
        ContractMetadata.make(
            hub.messageFrame().getContractAddress(),
            hub.deploymentNumberOfBytecodeAddress(),
            hub.deploymentStatusOfBytecodeAddress());

    checkArgument(!hub.deploymentStatusOfBytecodeAddress());

    final RomChunk chunk = new RomChunk(contractMetadata, false, true, byteCode);
    chunks.add(chunk);
  }
}
