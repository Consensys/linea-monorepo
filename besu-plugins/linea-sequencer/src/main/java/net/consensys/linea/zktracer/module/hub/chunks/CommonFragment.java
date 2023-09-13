/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.hub.chunks;

import java.math.BigInteger;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.hub.Exceptions;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.TxState;
import net.consensys.linea.zktracer.opcode.InstructionFamily;

@Accessors(fluent = true, chain = false)
@AllArgsConstructor
public final class CommonFragment implements TraceFragment {
  private final int txNumber;
  private final int batchNumber;
  private final TxState txState;
  private final int stamp;
  @Setter private int txEndStamp;
  private final boolean txReverts;
  private final InstructionFamily instructionFamily;
  private final Exceptions exceptions;
  private final boolean abortFlag;
  private final boolean failureConditionFlag;
  @Getter private final int contextNumber;
  @Setter private int newContextNumber;
  private final int revertStamp;
  private final boolean willRevert;
  private final boolean getsReverted;
  private final boolean selfReverts;
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
  private final long gasRefund;
  @Getter @Setter private boolean twoLinesInstruction;
  @Getter @Setter private boolean twoLinesInstructionCounter;
  @Getter @Setter private int numberOfNonStackRows;
  @Getter @Setter private int nonStackRowsCounter;

  @Override
  public Trace.TraceBuilder trace(Trace.TraceBuilder trace) {
    return trace
        .absoluteTransactionNumber(BigInteger.valueOf(this.txNumber))
        .batchNumber(BigInteger.valueOf(this.batchNumber))
        .txSkip(this.txState == TxState.TX_SKIP)
        .txWarm(this.txState == TxState.TX_WARM)
        .txInit(this.txState == TxState.TX_INIT)
        .txExec(this.txState == TxState.TX_STATE)
        .txFinl(this.txState == TxState.TX_FINAL)
        .hubStamp(BigInteger.valueOf(this.stamp))
        .transactionEndStamp(BigInteger.valueOf(txEndStamp))
        .transactionReverts(BigInteger.valueOf(txReverts ? 1 : 0))
        .contextMayChangeFlag(
            (instructionFamily == InstructionFamily.CALL
                    || instructionFamily == InstructionFamily.CREATE
                    || instructionFamily == InstructionFamily.HALT
                    || instructionFamily == InstructionFamily.INVALID)
                || exceptions.any())
        .exceptionAhoyFlag(exceptions.any())
        .abortFlag(abortFlag)
        .failureConditionFlag(failureConditionFlag)

        // Context data
        .contextNumber(BigInteger.valueOf(contextNumber))
        .contextNumberNew(BigInteger.valueOf(newContextNumber))
        .contextRevertStamp(BigInteger.valueOf(revertStamp))
        .contextWillRevertFlag(willRevert)
        .contextGetsRevrtdFlag(getsReverted)
        .contextSelfRevrtsFlag(selfReverts)
        .programCounter(BigInteger.valueOf(pc))
        .programCounterNew(BigInteger.valueOf(newPc))

        // Bytecode metadata
        .codeAddressHi(codeAddress.hiBigInt())
        .codeAddressLo(codeAddress.loBigInt())
        .codeDeploymentNumber(BigInteger.valueOf(codeDeploymentNumber))
        .codeDeploymentStatus(codeDeploymentStatus)
        .callerContextNumber(BigInteger.valueOf(callerContextNumber))
        .gasExpected(BigInteger.valueOf(gasExpected))
        .gasActual(BigInteger.valueOf(gasActual))
        .gasCost(BigInteger.valueOf(gasCost))
        .gasNext(BigInteger.valueOf(gasNext))
        .gasRefund(BigInteger.valueOf(gasRefund))
        .twoLineInstruction(twoLinesInstruction)
        .counterTli(twoLinesInstructionCounter)
        .numberOfNonStackRows(BigInteger.valueOf(numberOfNonStackRows))
        .counterNsr(BigInteger.valueOf(nonStackRowsCounter));
  }
}
