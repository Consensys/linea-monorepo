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
import static net.consensys.linea.zktracer.runtime.callstack.CallFrame.getOpCode;
import static net.consensys.linea.zktracer.types.AddressUtils.getDeploymentAddress;
import static net.consensys.linea.zktracer.types.AddressUtils.highPart;
import static net.consensys.linea.zktracer.types.AddressUtils.lowPart;

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.OperationSetModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedSet;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.ContextExitDefer;
import net.consensys.linea.zktracer.module.hub.defer.ImmediateContextEntryDefer;
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
import org.hyperledger.besu.evm.worldstate.WorldView;

@Accessors(fluent = true)
@RequiredArgsConstructor
public class RomLex
    implements OperationSetModule<RomOperation>, ImmediateContextEntryDefer, ContextExitDefer {

  private final Hub hub;

  @Getter
  private final ModuleOperationStackedSet<RomOperation> operations =
      new ModuleOperationStackedSet<>();

  @Getter private List<RomOperation> sortedOperations;
  private Bytes byteCode = Bytes.EMPTY;
  private Address address = Address.ZERO;

  @Getter private final DeferRegistry createDefers = new DeferRegistry();

  @Override
  public String moduleKey() {
    return "ROM_LEX";
  }

  public int getCodeFragmentIndexByMetadata(
      final Address address, final int deploymentNumber, final boolean depStatus) {
    return getCodeFragmentIndexByMetadata(
        ContractMetadata.make(address, deploymentNumber, depStatus));
  }

  public int getCodeFragmentIndexByMetadata(final ContractMetadata metadata) {
    if (sortedOperations.isEmpty()) {
      throw new RuntimeException("Chunks have not been sorted yet");
    }

    for (int i = 0; i < sortedOperations.size(); i++) {
      final RomOperation c = sortedOperations.get(i);
      if (c.metadata().equals(metadata)) {
        return i + 1;
      }
    }

    throw new RuntimeException(
        "RomChunk with:"
            + String.format("\n\t\taddress = %s", metadata.address())
            + String.format("\n\t\tdeployment number = %s", metadata.deploymentNumber())
            + String.format("\n\t\tdeployment status = %s", metadata.underDeployment())
            + "\n\tnot found");
  }

  public Optional<RomOperation> getChunkByMetadata(final ContractMetadata metadata) {
    // First search in the chunk added in the current transaction
    for (RomOperation c : operations.operationsInTransaction()) {
      if (c.metadata().equals(metadata)) {
        return Optional.of(c);
      }
    }

    // If not found, search in the chunk added since the beginning of the conflation
    for (RomOperation c : operations.operationsCommitedToTheConflation()) {
      if (c.metadata().equals(metadata)) {
        return Optional.of(c);
      }
    }

    return Optional.empty();
  }

  public Bytes getCodeByMetadata(final ContractMetadata metadata) {
    return getChunkByMetadata(metadata).map(RomOperation::byteCode).orElse(Bytes.EMPTY);
  }

  // TODO: it would maybe make more sense to only implement traceContextEnter
  //  and distinguish between depth == 0 and depth > 0. Why? So as to not have
  //  to manually tinker with deployment numbers / statuses.
  @Override
  public void traceStartTx(WorldView worldView, TransactionProcessingMetadata txMetaData) {
    final Transaction tx = txMetaData.getBesuTransaction();
    // Contract creation with InitCode
    if (tx.getInit().isPresent() && !tx.getInit().get().isEmpty()) {
      final Address deploymentAddress = Address.contractAddress(tx.getSender(), tx.getNonce());
      final RomOperation operation =
          new RomOperation(
              ContractMetadata.canonical(hub, deploymentAddress), false, false, tx.getInit().get());

      operations.add(operation);
    }

    // Call to an account with bytecode
    tx.getTo()
        .map(worldView::get)
        .map(AccountState::getCode)
        .ifPresent(
            code -> {
              if (!code.isEmpty()) {

                final Address calledAddress = tx.getTo().get();
                final RomOperation operation =
                    new RomOperation(
                        ContractMetadata.canonical(hub, calledAddress), true, false, code);

                operations.add(operation);
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
        final int currentDeploymentNumber = hub.deploymentNumberOfBytecodeAddress();
        final boolean currentDeploymentStatus = hub.deploymentStatusOfBytecodeAddress();

        final long offset = Words.clampedToLong(frame.getStackItem(0));
        final long length = Words.clampedToLong(frame.getStackItem(1));

        checkArgument(
            frame.getType() == MessageFrame.Type.CONTRACT_CREATION
                && currentDeploymentNumber > 0
                && currentDeploymentStatus,
            "callRomLex for RETURN expects the byte code address to be under deployment, yet:"
                + String.format("\n\t\tframe.getType() = %s", frame.getType())
                + String.format(
                    "\n\t\tMessageFrame contract address = %s", frame.getContractAddress())
                + String.format(
                    "\n\t\tCallFrame    bytecode address = %s",
                    hub.currentFrame().byteCodeAddress())
                + String.format("\n\t\tdeployment number = %s", currentDeploymentNumber)
                + String.format("\n\t\tdeployment status = %s", currentDeploymentStatus));
        checkArgument(length > 0, "callRomLex for RETURN expects positive size");

        byteCode = frame.shadowReadMemory(offset, length);
        hub.defers().scheduleForContextExit(this, hub.currentFrame().id());
      }

      case CALL, CALLCODE, DELEGATECALL, STATICCALL -> {
        final Address calleeAddress = Words.toAddress(frame.getStackItem(1));

        Optional.ofNullable(frame.getWorldUpdater().get(calleeAddress))
            .map(AccountState::getCode)
            .ifPresent(
                byteCode -> {
                  if (!byteCode.isEmpty()) {
                    final RomOperation operation =
                        new RomOperation(
                            ContractMetadata.canonical(hub, calleeAddress), true, false, byteCode);
                    operations.add(operation);
                  }
                });
      }

      case EXTCODECOPY -> {
        final Address foreignCodeAddress = Words.toAddress(frame.getStackItem(0));
        final long length = Words.clampedToLong(frame.getStackItem(3));

        checkArgument(
            length > 0,
            "EXTCODECOPY should only trigger a ROM_LEX chunk if nonzero size parameter");
        checkArgument(
            !hub.deploymentStatusOf(foreignCodeAddress),
            "EXTCODECOPY should only trigger a ROM_LEX chunk if its target isn't currently deploying");
        checkArgument(
            !frame.getWorldUpdater().getAccount(foreignCodeAddress).isEmpty()
                && frame.getWorldUpdater().getAccount(foreignCodeAddress).hasCode());

        Optional.ofNullable(frame.getWorldUpdater().get(foreignCodeAddress))
            .map(AccountState::getCode)
            .ifPresent(
                byteCode -> {
                  if (!byteCode.isEmpty()) {
                    final RomOperation operation =
                        new RomOperation(
                            ContractMetadata.canonical(hub, foreignCodeAddress),
                            true,
                            false,
                            byteCode);

                    operations.add(operation);
                  }
                });
      }

      default -> throw new RuntimeException(
          String.format("%s does not trigger the creation of ROM_LEX", getOpCode(frame)));
    }
  }

  @Override
  public void resolveUponContextEntry(Hub hub) {
    checkArgument(hub.messageFrame().getType() == MessageFrame.Type.CONTRACT_CREATION);
    checkArgument(
        hub.deploymentStatusOf(address), "After a CREATE the deployment status should be true");

    final ContractMetadata contractMetadata = ContractMetadata.canonical(hub, address);

    final RomOperation operation = new RomOperation(contractMetadata, true, false, byteCode);
    operations.add(operation);
    createDefers.trigger(contractMetadata);
  }

  // This is the tracing for ROMLEX module
  private void traceOperation(
      final RomOperation operation,
      final int cfi,
      final int codeFragmentIndexInfinity,
      Trace trace) {
    final Hash codeHash =
        operation.metadata().underDeployment() ? Hash.EMPTY : Hash.hash(operation.byteCode());
    trace
        .codeFragmentIndex(cfi)
        .codeFragmentIndexInfty(codeFragmentIndexInfinity)
        .codeSize(operation.byteCode().size())
        .addressHi(highPart(operation.metadata().address()))
        .addressLo(lowPart(operation.metadata().address()))
        .commitToState(operation.commitToTheState())
        .deploymentNumber(operation.metadata().deploymentNumber())
        .deploymentStatus(operation.metadata().underDeployment())
        .readFromState(operation.readFromTheState())
        .codeHashHi(codeHash.slice(0, LLARGE))
        .codeHashLo(codeHash.slice(LLARGE, LLARGE))
        .validateRow();
  }

  public void determineCodeFragmentIndex() {
    operations.finishConflation();
    sortedOperations = new ArrayList<>(operations.getAll());
    final RomOperationComparator ROM_CHUNK_COMPARATOR = new RomOperationComparator();
    sortedOperations.sort(ROM_CHUNK_COMPARATOR);
  }

  @Override
  public int lineCount() {
    // WARN: the line count for the RomLex is the *number of code fragments*, not their actual line
    // count – that's for the ROM.
    return operations.size();
  }

  @Override
  public void traceEndConflation(final WorldView state) {
    Preconditions.checkArgument(
        operations.conflationFinished(), "Conflation is done before traceEndConflation for RomLex");
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {

    final Trace trace = new Trace(buffers);
    final int codeFragmentIndexInfinity = operations.size();

    int cfi = 0;
    for (RomOperation operation : sortedOperations) {
      traceOperation(operation, ++cfi, codeFragmentIndexInfinity, trace);
    }
  }

  @Override
  public void resolveUponContextExit(Hub hub, CallFrame frame) {

    checkArgument(hub.opCode() == RETURN);
    checkArgument(!hub.deploymentStatusOfBytecodeAddress());

    final ContractMetadata contractMetadata =
        ContractMetadata.canonical(hub, hub.messageFrame().getContractAddress());

    final RomOperation chunk = new RomOperation(contractMetadata, false, true, byteCode);
    operations.add(chunk);
  }
}
