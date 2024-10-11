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

package net.consensys.linea.zktracer.module.hub.fragment.common;

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.TX_EXEC;
import static net.consensys.linea.zktracer.module.hub.signals.Exceptions.*;
import static net.consensys.linea.zktracer.module.hub.signals.TracedException.*;
import static net.consensys.linea.zktracer.opcode.InstructionFamily.*;

import java.math.BigInteger;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.HubProcessingPhase;
import net.consensys.linea.zktracer.module.hub.State;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.module.hub.signals.TracedException;
import net.consensys.linea.zktracer.opcode.InstructionFamily;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;

@Accessors(fluent = true)
@RequiredArgsConstructor
public class CommonFragmentValues {
  public final TransactionProcessingMetadata txMetadata;
  public final HubProcessingPhase hubProcessingPhase;
  public final int hubStamp;
  public final CallStack callStack;
  public final State.TxState.Stamps stamps; // for MMU and MXP stamps
  @Setter public int logStamp = -1;
  @Getter final CallFrame callFrame;
  public final boolean exceptionAhoy;
  @Setter public int contextNumberNew;
  public final int pc;
  public final int pcNew;
  final short height;
  final short heightNew;
  public final long gasExpected;
  public final long gasActual;
  public final boolean contextMayChange;
  @Setter long gasCost; // Set at Post Execution
  @Setter long gasNext; // Set at Post Execution
  @Setter public long refundDelta = 0; // 0 is default Value, can be modified only by SSTORE section
  @Setter public long gasRefund; // Set at commit time
  @Setter public long gasRefundNew; // Set at commit time
  @Setter public int numberOfNonStackRows;
  @Setter public boolean TLI;
  @Setter public int codeFragmentIndex = -1;
  @Getter private TracedException tracedException = UNDEFINED;

  public CommonFragmentValues(Hub hub) {
    final short exceptions = hub.pch().exceptions();
    final boolean noStackException = !stackException(exceptions);

    this.txMetadata = hub.txStack().current();
    this.hubProcessingPhase = hub.state().getProcessingPhase();
    this.hubStamp = hub.stamp();
    this.callStack = hub.callStack();
    this.stamps = hub.state().stamps();
    this.callFrame = hub.currentFrame();
    this.exceptionAhoy = any(exceptions);
    // this.contextNumberNew = hub.contextNumberNew(callFrame);
    this.pc = hubProcessingPhase == TX_EXEC ? hub.currentFrame().pc() : 0;
    this.pcNew = computePcNew(hub, pc, noStackException, hub.state.getProcessingPhase() == TX_EXEC);
    this.height = callFrame.stack().getHeight();
    this.heightNew = callFrame.stack().getHeightNew();

    // TODO: partial solution, will not work in general
    this.gasExpected = hub.expectedGas();
    this.gasActual = hub.remainingGas();

    final InstructionFamily instructionFamily = hub.opCode().getData().instructionFamily();
    this.contextMayChange =
        hubProcessingPhase == HubProcessingPhase.TX_EXEC
            && ((instructionFamily == CALL
                    || instructionFamily == CREATE
                    || instructionFamily == HALT
                    || instructionFamily == INVALID)
                || exceptionAhoy);

    if (none(exceptions)) {
      tracedException = TracedException.NONE;
      return;
    }

    final OpCode opCode = hub.opCode();

    if (Exceptions.staticFault(exceptions)) {
      checkArgument(opCode.mayTriggerStaticException());
      setTracedException(TracedException.STATIC_FAULT);
      return;
    }

    // RETURNDATACOPY opcode specific exception
    if (opCode == OpCode.RETURNDATACOPY && Exceptions.returnDataCopyFault(exceptions)) {
      setTracedException(TracedException.RETURN_DATA_COPY_FAULT);
      return;
    }

    // SSTORE opcode specific exception
    if (opCode == OpCode.SSTORE && Exceptions.outOfSStore(exceptions)) {
      setTracedException(TracedException.OUT_OF_SSTORE);
      return;
    }

    // For RETURN, in case none of the above exceptions is the traced one,
    // we have a complex logic to determine the traced exception that is
    // implemented in the instruction processing
    if (opCode == OpCode.RETURN) {
      return;
    }

    if (Exceptions.memoryExpansionException(exceptions)) {
      checkArgument(opCode.mayTriggerMemoryExpansionException());
      setTracedException(TracedException.MEMORY_EXPANSION_EXCEPTION);
      return;
    }

    if (Exceptions.outOfGasException(exceptions)) {
      setTracedException(TracedException.OUT_OF_GAS_EXCEPTION);
      return;
    }

    // JUMP instruction family specific exception
    if (instructionFamily == InstructionFamily.JUMP && Exceptions.jumpFault(exceptions)) {
      setTracedException(TracedException.JUMP_FAULT);
    }
  }

  public void setTracedException(TracedException tracedException) {
    checkArgument(
        this.tracedException == UNDEFINED); // || this.tracedException == tracedException);
    this.tracedException = tracedException;
  }

  static int computePcNew(
      final Hub hub, final int pc, boolean noStackException, boolean hubInExecPhase) {
    OpCode opCode = hub.opCode();
    if (!(noStackException && hubInExecPhase)) {
      return 0;
    }

    if (opCode.getData().isPush()) {
      return pc + opCode.byteValue() - OpCode.PUSH1.byteValue() + 2;
    }

    if (opCode.isJump()) {
      final BigInteger prospectivePcNew =
          hub.currentFrame().frame().getStackItem(0).toUnsignedBigInteger();
      final BigInteger codeSize = BigInteger.valueOf(hub.currentFrame().code().getSize());

      final int attemptedPcNew =
          codeSize.compareTo(prospectivePcNew) > 0 ? prospectivePcNew.intValueExact() : 0;

      if (opCode.equals(OpCode.JUMP)) {
        return attemptedPcNew;
      }

      if (opCode.equals(OpCode.JUMPI)) {
        BigInteger condition = hub.currentFrame().frame().getStackItem(1).toUnsignedBigInteger();
        if (!condition.equals(BigInteger.ZERO)) {
          return attemptedPcNew;
        }
      }
    }
    ;

    return pc + 1;
  }
}
