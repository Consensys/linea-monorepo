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

package net.consensys.linea.zktracer.module.hub.fragment;

import lombok.Builder;
import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.TransactionStack;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.opcode.InstructionFamily;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.TxState;
import org.apache.tuweni.bytes.Bytes;

@Accessors(fluent = true, chain = false)
@Builder
public final class CommonFragment implements TraceFragment {
  private final Hub hub;
  private final int txId;
  private final int batchNumber;
  private final TxState txState;
  private final int stamp;
  private final InstructionFamily instructionFamily;
  private final Exceptions exceptions;
  private final int callFrameId;
  @Getter private final int contextNumber;
  @Setter private int newContextNumber;
  private final int revertStamp;
  @Getter private final int pc;
  @Setter private int newPc;
  private final EWord codeAddress;
  private int codeDeploymentNumber;
  private final boolean codeDeploymentStatus;
  private final int callerContextNumber;
  private final long gasExpected;
  private final long gasActual;
  private final long gasCost;
  private final long gasNext;
  @Getter private final long refundDelta;
  @Setter private long gasRefund;
  @Getter @Setter private boolean twoLinesInstruction;
  @Getter @Setter private boolean twoLinesInstructionCounter;
  @Getter @Setter private int numberOfNonStackRows;
  @Getter @Setter private int nonStackRowsCounter;

  public static CommonFragment fromHub(
      final Hub hub, final CallFrame frame, boolean tliCounter, int nonStackRowsCounter) {
    long refund = 0;
    if (hub.pch().exceptions().noStackException()) {
      refund = Hub.gp.of(frame.frame(), hub.opCode()).refund();
    }

    return CommonFragment.builder()
        .hub(hub)
        .txId(hub.transients().tx().id())
        .batchNumber(hub.transients().conflation().number())
        .txState(hub.transients().tx().state())
        .stamp(hub.stamp())
        .instructionFamily(hub.opCodeData().instructionFamily())
        .exceptions(hub.pch().exceptions().snapshot())
        .callFrameId(frame.id())
        .contextNumber(frame.contextNumber())
        .newContextNumber(frame.contextNumber())
        .pc(frame.pc())
        .codeAddress(frame.addressAsEWord())
        .codeDeploymentNumber(frame.codeDeploymentNumber())
        .codeDeploymentStatus(frame.underDeployment())
        .callerContextNumber(hub.callStack().getParentOf(frame.id()).contextNumber())
        .refundDelta(refund)
        .twoLinesInstruction(hub.opCodeData().stackSettings().twoLinesInstruction())
        .twoLinesInstructionCounter(tliCounter)
        .nonStackRowsCounter(nonStackRowsCounter)
        .build();
  }

  public boolean txReverts() {
    return hub.txStack().getById(this.txId).status();
  }

  @Override
  public Trace trace(Trace trace) {
    CallFrame frame = this.hub.callStack().getById(this.callFrameId);
    TransactionStack.MetaTransaction tx = hub.txStack().getById(this.txId);
    final boolean selfReverts = frame.selfReverts();
    final boolean getsReverted = frame.getsReverted();

    return trace
        .absoluteTransactionNumber(Bytes.ofUnsignedInt(tx.absNumber()))
        .batchNumber(Bytes.ofUnsignedInt(this.batchNumber))
        .txSkip(this.txState == TxState.TX_SKIP)
        .txWarm(this.txState == TxState.TX_WARM)
        .txInit(this.txState == TxState.TX_INIT)
        .txExec(this.txState == TxState.TX_EXEC)
        .txFinl(this.txState == TxState.TX_FINAL)
        .hubStamp(Bytes.ofUnsignedInt(this.stamp))
        .hubStampTransactionEnd(Bytes.ofUnsignedLong(tx.endStamp()))
        .transactionReverts(tx.status())
        .contextMayChangeFlag(
            (instructionFamily == InstructionFamily.CALL
                    || instructionFamily == InstructionFamily.CREATE
                    || instructionFamily == InstructionFamily.HALT
                    || instructionFamily == InstructionFamily.INVALID)
                || exceptions.any())
        .exceptionAhoyFlag(exceptions.any())

        // Context data
        .contextNumber(Bytes.ofUnsignedInt(contextNumber))
        .contextNumberNew(Bytes.ofUnsignedInt(newContextNumber))
        .contextRevertStamp(Bytes.ofUnsignedInt(revertStamp))
        .contextWillRevertFlag(getsReverted || selfReverts)
        .contextGetsRevertedFlag(getsReverted)
        .contextSelfRevertsFlag(selfReverts)
        .programCounter(Bytes.ofUnsignedInt(pc))
        .programCounterNew(Bytes.ofUnsignedInt(newPc))

        // Bytecode metadata
        .codeAddressHi(codeAddress.hi())
        .codeAddressLo(codeAddress.lo())
        .codeDeploymentNumber(Bytes.ofUnsignedInt(codeDeploymentNumber))
        .codeDeploymentStatus(codeDeploymentStatus)
        .callerContextNumber(Bytes.ofUnsignedInt(callerContextNumber))
        .gasExpected(Bytes.ofUnsignedLong(gasExpected))
        .gasActual(Bytes.ofUnsignedLong(gasActual))
        .gasCost(Bytes.ofUnsignedLong(gasCost))
        .gasNext(Bytes.ofUnsignedLong(gasNext))
        .gasRefund(Bytes.ofUnsignedLong(gasRefund))
        .twoLineInstruction(twoLinesInstruction)
        .counterTli(twoLinesInstructionCounter)
        .numberOfNonStackRows(Bytes.ofUnsignedShort(numberOfNonStackRows))
        .counterNsr(Bytes.ofUnsignedShort(nonStackRowsCounter));
  }
}
