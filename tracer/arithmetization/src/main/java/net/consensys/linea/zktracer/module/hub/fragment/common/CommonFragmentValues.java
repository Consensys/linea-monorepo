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

import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.TX_EXEC;

import java.math.BigInteger;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.HubProcessingPhase;
import net.consensys.linea.zktracer.module.hub.State;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
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

  public CommonFragmentValues(Hub hub) {
    final boolean noStackException = !Exceptions.stackException(hub.pch().exceptions());
    final InstructionFamily instructionFamily = hub.opCode().getData().instructionFamily();

    this.txMetadata = hub.txStack().current();
    this.hubProcessingPhase = hub.state().getProcessingPhase();
    this.hubStamp = hub.stamp();
    this.callStack = hub.callStack();
    this.stamps = hub.state().stamps();
    this.callFrame = hub.currentFrame();
    this.exceptionAhoy = Exceptions.any(hub.pch().exceptions());
    // this.contextNumberNew = hub.contextNumberNew(callFrame);
    this.pc = hubProcessingPhase == TX_EXEC ? hub.currentFrame().pc() : 0;
    this.pcNew = computePcNew(hub, pc, noStackException, hub.state.getProcessingPhase() == TX_EXEC);
    this.height = (short) callFrame.stack().getHeight();
    this.heightNew = (short) callFrame.stack().getHeightNew();

    // TODO: partial solution, will not work in general
    this.gasExpected = hub.expectedGas();
    this.gasActual = hub.remainingGas();

    this.contextMayChange =
        hubProcessingPhase == HubProcessingPhase.TX_EXEC
            && ((instructionFamily == InstructionFamily.CALL
                    || instructionFamily == InstructionFamily.CREATE
                    || instructionFamily == InstructionFamily.HALT
                    || instructionFamily == InstructionFamily.INVALID)
                || exceptionAhoy);
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
