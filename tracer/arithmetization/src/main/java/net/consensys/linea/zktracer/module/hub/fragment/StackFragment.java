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

import java.util.List;
import java.util.Optional;
import java.util.function.Function;

import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.DeploymentExceptions;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.signals.AbortingConditions;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.opcode.InstructionFamily;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.gas.MxpType;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjection;
import net.consensys.linea.zktracer.runtime.stack.Action;
import net.consensys.linea.zktracer.runtime.stack.Stack;
import net.consensys.linea.zktracer.runtime.stack.StackOperation;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Accessors(fluent = true)
public final class StackFragment implements TraceFragment {
  private final Stack stack;
  @Getter private final List<StackOperation> stackOps;
  private final Exceptions exceptions;
  @Setter private DeploymentExceptions contextExceptions;
  private final long staticGas;
  private EWord hashInfoKeccak = EWord.ZERO;
  private final long hashInfoSize;
  private final boolean hashInfoFlag;
  @Getter private final OpCode opCode;

  private StackFragment(
      final Hub hub,
      Stack stack,
      List<StackOperation> stackOps,
      Exceptions exceptions,
      AbortingConditions aborts,
      DeploymentExceptions contextExceptions,
      GasProjection gp,
      boolean isDeploying) {
    this.stack = stack;
    this.stackOps = stackOps;
    this.exceptions = exceptions;
    this.contextExceptions = contextExceptions;
    this.opCode = stack.getCurrentOpcodeData().mnemonic();
    this.hashInfoFlag =
        switch (this.opCode) {
          case SHA3 -> exceptions.none() && gp.messageSize() > 0;
          case RETURN -> exceptions.none() && gp.messageSize() > 0 && isDeploying;
          case CREATE2 -> exceptions.none()
              && contextExceptions.none()
              && aborts.none()
              && gp.messageSize() > 0;
          default -> false;
        };
    this.hashInfoSize = this.hashInfoFlag ? gp.messageSize() : 0;
    this.staticGas = gp.staticGas();
    if (this.opCode == OpCode.RETURN && exceptions.none()) {
      this.hashInfoKeccak =
          EWord.of(org.hyperledger.besu.crypto.Hash.keccak256(hub.transients().op().returnData()));
    }
  }

  public static StackFragment prepare(
      final Hub hub,
      final Stack stack,
      final List<StackOperation> stackOperations,
      final Exceptions exceptions,
      final AbortingConditions aborts,
      final GasProjection gp,
      boolean isDeploying) {
    return new StackFragment(
        hub,
        stack,
        stackOperations,
        exceptions,
        aborts,
        DeploymentExceptions.empty(),
        gp,
        isDeploying);
  }

  public void feedHashedValue(MessageFrame frame) {
    if (hashInfoFlag) {
      switch (this.opCode) {
        case SHA3 -> this.hashInfoKeccak = EWord.of(frame.getStackItem(0));
        case CREATE2 -> {
          Address newAddress = EWord.of(frame.getStackItem(0)).toAddress();
          // zero address indicates a failed deployment
          if (!newAddress.isZero()) {
            this.hashInfoKeccak =
                EWord.of(
                    Optional.ofNullable(frame.getWorldUpdater().get(newAddress))
                        .map(AccountState::getCodeHash)
                        .orElse(Hash.EMPTY));
          }
        }
        case RETURN -> {
          /* already set at opcode invocation */
        }
        default -> throw new IllegalStateException("unexpected opcode");
      }
    }
  }

  @Override
  public Trace trace(Trace trace) {
    final List<Function<Bytes, Trace>> valHiTracers =
        List.of(
            trace::pStackStackItemValueHi1,
            trace::pStackStackItemValueHi2,
            trace::pStackStackItemValueHi3,
            trace::pStackStackItemValueHi4);

    final List<Function<Bytes, Trace>> valLoTracers =
        List.of(
            trace::pStackStackItemValueLo1,
            trace::pStackStackItemValueLo2,
            trace::pStackStackItemValueLo3,
            trace::pStackStackItemValueLo4);

    final List<Function<Boolean, Trace>> popTracers =
        List.of(
            trace::pStackStackItemPop1,
            trace::pStackStackItemPop2,
            trace::pStackStackItemPop3,
            trace::pStackStackItemPop4);

    final List<Function<Bytes, Trace>> heightTracers =
        List.of(
            trace::pStackStackItemHeight1,
            trace::pStackStackItemHeight2,
            trace::pStackStackItemHeight3,
            trace::pStackStackItemHeight4);

    final List<Function<Bytes, Trace>> stampTracers =
        List.of(
            trace::pStackStackItemStamp1,
            trace::pStackStackItemStamp2,
            trace::pStackStackItemStamp3,
            trace::pStackStackItemStamp4);

    EWord pushValue = EWord.ZERO;
    var it = stackOps.listIterator();
    while (it.hasNext()) {
      var i = it.nextIndex();
      var op = it.next();
      final EWord eValue = EWord.of(op.value());
      if (this.stack.getCurrentOpcodeData().isPush()) {
        pushValue = eValue;
      }

      heightTracers.get(i).apply(Bytes.ofUnsignedShort(op.height()));
      valLoTracers.get(i).apply(eValue.lo());
      valHiTracers.get(i).apply(eValue.hi());
      popTracers.get(i).apply(op.action() == Action.POP);
      stampTracers.get(i).apply(Bytes.ofUnsignedLong(op.stackStamp()));
    }

    return trace
        .peekAtStack(true)
        // Instruction details
        .pStackAlpha(Bytes.ofUnsignedInt(this.stack.getCurrentOpcodeData().stackSettings().alpha()))
        .pStackDelta(Bytes.ofUnsignedInt(this.stack.getCurrentOpcodeData().stackSettings().delta()))
        .pStackNbAdded(
            Bytes.ofUnsignedInt(this.stack.getCurrentOpcodeData().stackSettings().nbAdded()))
        .pStackNbRemoved(
            Bytes.ofUnsignedInt(this.stack.getCurrentOpcodeData().stackSettings().nbRemoved()))
        .pStackInstruction(Bytes.of(this.stack.getCurrentOpcodeData().value()))
        .pStackStaticGas(Bytes.ofUnsignedInt(staticGas))
        .pStackPushValueHi(pushValue.hi())
        .pStackPushValueLo(pushValue.lo())
        .pStackDecFlag1(this.stack.getCurrentOpcodeData().stackSettings().flag1())
        .pStackDecFlag2(this.stack.getCurrentOpcodeData().stackSettings().flag2())
        .pStackDecFlag3(this.stack.getCurrentOpcodeData().stackSettings().flag3())
        .pStackDecFlag4(this.stack.getCurrentOpcodeData().stackSettings().flag4())
        // Exception flag
        .pStackOpcx(exceptions.invalidOpcode())
        .pStackSux(exceptions.stackUnderflow())
        .pStackSox(exceptions.stackOverflow())
        .pStackOogx(exceptions.outOfGas())
        .pStackMxpx(exceptions.outOfMemoryExpansion())
        .pStackRdcx(exceptions.returnDataCopyFault())
        .pStackJumpx(exceptions.jumpFault())
        .pStackStaticx(exceptions.staticFault())
        .pStackSstorex(exceptions.outOfSStore())
        .pStackIcpx(contextExceptions.invalidCodePrefix())
        .pStackMaxcsx(contextExceptions.codeSizeOverflow())
        // Opcode families
        .pStackAddFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.ADD)
        .pStackModFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.MOD)
        .pStackMulFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.MUL)
        .pStackExtFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.EXT)
        .pStackWcpFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.WCP)
        .pStackBinFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.BIN)
        .pStackShfFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.SHF)
        .pStackKecFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.KEC)
        .pStackConFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.CONTEXT)
        .pStackAccFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.ACCOUNT)
        .pStackCopyFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.COPY)
        .pStackTxnFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.TRANSACTION)
        .pStackBtcFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.BATCH)
        .pStackStackramFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.STACK_RAM)
        .pStackStoFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.STORAGE)
        .pStackJumpFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.JUMP)
        .pStackPushpopFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.PUSH_POP)
        .pStackDupFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.DUP)
        .pStackSwapFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.SWAP)
        .pStackLogFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.LOG)
        .pStackCreateFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.CREATE)
        .pStackCallFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.CALL)
        .pStackHaltFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.HALT)
        .pStackInvalidFlag(
            this.stack.getCurrentOpcodeData().instructionFamily() == InstructionFamily.INVALID)
        .pStackMxpFlag(
            Optional.ofNullable(this.stack.getCurrentOpcodeData().billing())
                .map(b -> b.type() != MxpType.NONE)
                .orElse(false))
        .pStackStaticFlag(this.stack.getCurrentOpcodeData().stackSettings().forbiddenInStatic())
        // Hash data
        .pStackHashInfoSize(Bytes.ofUnsignedInt(hashInfoSize))
        .pStackHashInfoKeccakHi(this.hashInfoKeccak.hi())
        .pStackHashInfoKeccakLo(this.hashInfoKeccak.lo())
        .pStackHashInfoFlag(this.hashInfoFlag);
  }
}
