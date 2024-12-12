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

package net.consensys.linea.zktracer.module.hub.fragment.common;

import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.*;
import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.TX_EXEC;

import java.util.function.Supplier;

import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.module.hub.signals.TracedException;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.EVM;
import org.hyperledger.besu.evm.EvmSpecVersion;
import org.hyperledger.besu.evm.internal.EvmConfiguration;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.operation.OperationRegistry;
import org.hyperledger.besu.evm.operation.SelfDestructOperation;
import org.hyperledger.besu.evm.worldstate.WorldView;

@Accessors(fluent = true, chain = false)
@RequiredArgsConstructor
public final class CommonFragment implements TraceFragment {

  private final CommonFragmentValues commonFragmentValues;
  private final int nonStackRowsCounter;
  private final boolean twoLineInstructionCounter;
  private final int mmuStamp;
  private final int mxpStamp;

  public CommonFragment(
      CommonFragmentValues commonValues,
      int stackLineCounter,
      int nonStackLineCounter,
      int mmuStamp,
      int mxpStamp) {
    this.commonFragmentValues = commonValues;
    this.twoLineInstructionCounter = stackLineCounter == 1;
    this.nonStackRowsCounter = nonStackLineCounter;
    this.mmuStamp = mmuStamp;
    this.mxpStamp = mxpStamp;
  }

  private boolean isUnexceptional() {
    return Exceptions.none(commonFragmentValues.exceptions);
  }

  public boolean txReverts() {
    return commonFragmentValues.txMetadata.statusCode();
  }

  public Trace trace(Trace trace) {
    final CallFrame frame = commonFragmentValues.callFrame;
    final TransactionProcessingMetadata tx = commonFragmentValues.txMetadata;
    final boolean isExec = commonFragmentValues.hubProcessingPhase == TX_EXEC;
    final boolean oogx =
        commonFragmentValues.tracedException() == TracedException.OUT_OF_GAS_EXCEPTION;
    final boolean nonOogException = Exceptions.any(commonFragmentValues.exceptions) && !oogx;
    return trace
        .absoluteTransactionNumber(tx.getAbsoluteTransactionNumber())
        .relativeBlockNumber(tx.getRelativeBlockNumber())
        .txSkip(commonFragmentValues.hubProcessingPhase == TX_SKIP)
        .txWarm(commonFragmentValues.hubProcessingPhase == TX_WARM)
        .txInit(commonFragmentValues.hubProcessingPhase == TX_INIT)
        .txExec(commonFragmentValues.hubProcessingPhase == TX_EXEC)
        .txFinl(commonFragmentValues.hubProcessingPhase == TX_FINL)
        .hubStamp(commonFragmentValues.hubStamp)
        .hubStampTransactionEnd(tx.getHubStampTransactionEnd())
        .contextMayChange(commonFragmentValues.contextMayChange)
        .exceptionAhoy(Exceptions.any(commonFragmentValues.exceptions) && isExec)
        .logInfoStamp(commonFragmentValues.logStamp)
        .mmuStamp(mmuStamp)
        .mxpStamp(mxpStamp)
        // nontrivial dom / sub are traced in storage or account fragments only
        .contextNumber(isExec ? frame.contextNumber() : 0)
        .contextNumberNew(commonFragmentValues.contextNumberNew)
        .callerContextNumber(
            commonFragmentValues.callStack.getById(frame.parentId()).contextNumber())
        .contextWillRevert(frame.willRevert() && isExec)
        .contextGetsReverted(frame.getsReverted() && isExec)
        .contextSelfReverts(frame.selfReverts() && isExec)
        .contextRevertStamp(isExec ? frame.revertStamp() : 0)
        .codeFragmentIndex(commonFragmentValues.codeFragmentIndex)
        .programCounter(commonFragmentValues.pc)
        .programCounterNew(commonFragmentValues.pcNew)
        .height(isExec ? commonFragmentValues.height : 0)
        .heightNew(isExec ? commonFragmentValues.heightNew : 0)
        // peeking flags are traced in the respective fragments
        .gasExpected(Bytes.ofUnsignedLong(commonFragmentValues.gasExpected))
        .gasActual(Bytes.ofUnsignedLong(commonFragmentValues.gasActual))
        .gasCost(Bytes.ofUnsignedLong(commonFragmentValues.gasCostToTrace()))
        .gasNext(
            Bytes.ofUnsignedLong(isExec && isUnexceptional() ? commonFragmentValues.gasNext : 0))
        .refundCounter(
            (commonFragmentValues.hubProcessingPhase == TX_EXEC)
                ? commonFragmentValues.gasRefund
                : 0)
        .refundCounterNew(
            (commonFragmentValues.hubProcessingPhase == TX_EXEC)
                ? commonFragmentValues.gasRefundNew
                : 0)
        .twoLineInstruction(commonFragmentValues.TLI)
        .counterTli(twoLineInstructionCounter)
        .nonStackRows((short) commonFragmentValues.numberOfNonStackRows)
        .counterNsr((short) nonStackRowsCounter);
  }

  static long computeGasCost(Hub hub, WorldView world) {

    switch (hub.opCodeData().instructionFamily()) {
      case ADD, MOD, SHF, BIN, WCP, EXT, BATCH, MACHINE_STATE, PUSH_POP, DUP, SWAP, INVALID -> {
        if (Exceptions.outOfGasException(hub.pch().exceptions())
            || Exceptions.none(hub.pch().exceptions())) {
          return hub.opCode().getData().stackSettings().staticGas().cost();
        }
        return 0;
      }
      case STORAGE -> {
        switch (hub.opCode()) {
          case SSTORE -> {
            return gasCostSstore(hub, world);
          }
          case SLOAD -> {
            return gasCostSload(hub, world);
          }
          default -> throw new RuntimeException(
              "Gas cost not covered for " + hub.opCode().toString());
        }
      }
      case HALT -> {
        switch (hub.opCode()) {
          case STOP -> {
            return 0;
          }
          case RETURN, REVERT -> {
            Bytes offset = hub.messageFrame().getStackItem(0);
            Bytes size = hub.messageFrame().getStackItem(0);
            return Exceptions.memoryExpansionException(hub.pch().exceptions())
                ? 0
                : ZkTracer.gasCalculator.memoryExpansionGasCost(
                    hub.messageFrame(), offset.toLong(), size.toLong());
          }
          case SELFDESTRUCT -> {
            SelfDestructOperation op = new SelfDestructOperation(ZkTracer.gasCalculator);
            Operation.OperationResult operationResult =
                op.execute(
                    hub.messageFrame(),
                    new EVM(
                        new OperationRegistry(),
                        ZkTracer.gasCalculator,
                        EvmConfiguration.DEFAULT,
                        EvmSpecVersion.LONDON));
            long gasCost = operationResult.getGasCost();
            Address recipient = Address.extract((Bytes32) hub.messageFrame().getStackItem(0));
            Wei inheritance = world.get(hub.messageFrame().getRecipientAddress()).getBalance();
            return ZkTracer.gasCalculator.selfDestructOperationGasCost(
                world.get(recipient), inheritance);
          }
        }
        return 0;
      }
      default -> {
        throw new RuntimeException("Gas cost not covered for " + hub.opCode().toString());
      }
    }
  }

  static long gasCostSstore(Hub hub, WorldView world) {

    final Address address = hub.currentFrame().accountAddress();
    final Bytes32 storageKey = EWord.of(hub.messageFrame().getStackItem(0));

    final UInt256 storageKeyUint256 = UInt256.fromBytes(hub.messageFrame().getStackItem(0));
    final UInt256 valueNextUint256 = UInt256.fromBytes(hub.messageFrame().getStackItem(1));

    final Supplier<UInt256> valueCurrentSupplier =
        () -> world.get(address).getStorageValue(storageKeyUint256);
    final Supplier<UInt256> valueOriginalSupplier =
        () -> world.get(address).getOriginalStorageValue(storageKeyUint256);

    final long storageCost =
        ZkTracer.gasCalculator.calculateStorageCost(
            valueNextUint256, valueCurrentSupplier, valueOriginalSupplier);
    final boolean storageSlotWarmth =
        hub.currentFrame().frame().getWarmedUpStorage().contains(address, storageKey);

    return storageCost + (storageSlotWarmth ? 0L : ZkTracer.gasCalculator.getColdSloadCost());
  }

  static long gasCostSload(Hub hub, WorldView world) {
    final Address address = hub.currentFrame().accountAddress();
    final Bytes32 storageKey = EWord.of(hub.messageFrame().getStackItem(0));
    final boolean storageSlotWarmth =
        hub.currentFrame().frame().getWarmedUpStorage().contains(address, storageKey);

    return ZkTracer.gasCalculator.getSloadOperationGasCost()
        + (storageSlotWarmth
            ? ZkTracer.gasCalculator.getWarmStorageReadCost()
            : ZkTracer.gasCalculator.getColdSloadCost());
  }
}
