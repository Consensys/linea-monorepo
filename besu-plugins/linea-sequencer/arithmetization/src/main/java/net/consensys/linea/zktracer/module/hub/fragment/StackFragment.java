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

import java.math.BigInteger;
import java.util.List;
import java.util.Optional;
import java.util.function.Function;

import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.hub.Aborts;
import net.consensys.linea.zktracer.module.hub.DeploymentExceptions;
import net.consensys.linea.zktracer.module.hub.Exceptions;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.stack.Action;
import net.consensys.linea.zktracer.module.hub.stack.Stack;
import net.consensys.linea.zktracer.module.hub.stack.StackOperation;
import net.consensys.linea.zktracer.opcode.InstructionFamily;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.gas.MxpType;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjection;
import org.hyperledger.besu.datatypes.Address;
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
  private final OpCode opCode;

  private StackFragment(
      Stack stack,
      List<StackOperation> stackOps,
      Exceptions exceptions,
      Aborts aborts,
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
  }

  public static StackFragment prepare(
      final Stack stack,
      final List<StackOperation> stackOperations,
      final Exceptions exceptions,
      final Aborts aborts,
      final GasProjection gp,
      boolean isDeploying) {
    return new StackFragment(
        stack, stackOperations, exceptions, aborts, DeploymentExceptions.empty(), gp, isDeploying);
  }

  public void feedHashedValue(MessageFrame frame) {
    if (hashInfoFlag) {
      switch (this.opCode) {
        case SHA3 -> this.hashInfoKeccak = EWord.of(frame.getStackItem(0));
        case RETURN -> this.hashInfoKeccak = EWord.ZERO; // TODO: fixme
        case CREATE2 -> {
          Address newAddress = EWord.of(frame.getStackItem(0)).toAddress();
          this.hashInfoKeccak = EWord.of(frame.getWorldUpdater().get(newAddress).getCodeHash());
        }
        default -> throw new IllegalStateException("unexpected opcode");
      }
    }
  }

  @Override
  public Trace.TraceBuilder trace(Trace.TraceBuilder trace) {
    final List<Function<BigInteger, Trace.TraceBuilder>> valHiTracers =
        List.of(
            trace::pStackStackItemValueHi1,
            trace::pStackStackItemValueHi2,
            trace::pStackStackItemValueHi3,
            trace::pStackStackItemValueHi4);

    final List<Function<BigInteger, Trace.TraceBuilder>> valLoTracers =
        List.of(
            trace::pStackStackItemValueLo1,
            trace::pStackStackItemValueLo2,
            trace::pStackStackItemValueLo3,
            trace::pStackStackItemValueLo4);

    final List<Function<Boolean, Trace.TraceBuilder>> popTracers =
        List.of(
            trace::pStackStackItemPop1,
            trace::pStackStackItemPop2,
            trace::pStackStackItemPop3,
            trace::pStackStackItemPop4);

    final List<Function<BigInteger, Trace.TraceBuilder>> heightTracers =
        List.of(
            trace::pStackStackItemHeight1,
            trace::pStackStackItemHeight2,
            trace::pStackStackItemHeight3,
            trace::pStackStackItemHeight4);

    final List<Function<BigInteger, Trace.TraceBuilder>> stampTracers =
        List.of(
            trace::pStackStackItemStamp1,
            trace::pStackStackItemStamp2,
            trace::pStackStackItemStamp3,
            trace::pStackStackItemStamp4);

    final int alpha = this.stack.getCurrentOpcodeData().stackSettings().alpha();
    final int delta = this.stack.getCurrentOpcodeData().stackSettings().delta();

    var heightUnder = stack.getHeight() - delta;
    var heightOver = 0;

    if (!stack.isUnderflow()) {
      if (!(alpha == 1 && delta == 0 && stack.getHeight() == Stack.MAX_STACK_SIZE)) {
        var overflow = stack.isOverflow() ? 1 : 0;
        heightOver = (2 * overflow - 1) * (heightUnder + alpha - Stack.MAX_STACK_SIZE) - overflow;
      }
    } else {
      heightUnder = -heightUnder - 1;
    }

    var it = stackOps.listIterator();
    while (it.hasNext()) {
      var i = it.nextIndex();
      var op = it.next();

      heightTracers.get(i).apply(BigInteger.valueOf(op.height()));
      valLoTracers.get(i).apply(op.value().loBigInt());
      valHiTracers.get(i).apply(op.value().hiBigInt());
      popTracers.get(i).apply(op.action() == Action.POP);
      stampTracers.get(i).apply(BigInteger.valueOf(op.stackStamp()));
    }

    return trace
        .peekAtStack(true)
        // Stack height
        .pStackHeight(BigInteger.valueOf(stack.getHeight()))
        .pStackHeightNew(BigInteger.valueOf(stack.getHeightNew()))
        .pStackHeightUnder(BigInteger.valueOf(heightUnder))
        .pStackHeightOver(BigInteger.valueOf(heightOver))
        // Instruction details
        .pStackInst(BigInteger.valueOf(this.stack.getCurrentOpcodeData().value()))
        .pStackStaticGas(BigInteger.valueOf(staticGas))
        .pStackDecodedFlag1(this.stack.getCurrentOpcodeData().stackSettings().flag1())
        .pStackDecodedFlag2(this.stack.getCurrentOpcodeData().stackSettings().flag2())
        .pStackDecodedFlag3(this.stack.getCurrentOpcodeData().stackSettings().flag3())
        .pStackDecodedFlag4(this.stack.getCurrentOpcodeData().stackSettings().flag4())
        // Exception flag
        .pStackOpcx(exceptions.invalidOpcode())
        .pStackSux(exceptions.stackUnderflow())
        .pStackSox(exceptions.stackOverflow())
        .pStackOogx(exceptions.outOfGas())
        .pStackMxpx(exceptions.outOfMemoryExpansion())
        .pStackRdcx(exceptions.returnDataCopyFault())
        .pStackJumpx(exceptions.jumpFault())
        .pStackStaticx(exceptions.staticViolation())
        .pStackSstorex(exceptions.outOfSStore())
        .pStackInvprex(contextExceptions.invalidCodePrefix())
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
        .pStackTrmFlag(
            this.stack.getCurrentOpcodeData().stackSettings().addressTrimmingInstruction())
        .pStackStaticFlag(this.stack.getCurrentOpcodeData().stackSettings().forbiddenInStatic())
        .pStackOobFlag(this.stack.getCurrentOpcodeData().stackSettings().oobFlag())
        // Hash data
        .pStackHashInfoSize(BigInteger.valueOf(hashInfoSize))
        .pStackHashInfoKecHi(this.hashInfoKeccak.hiBigInt())
        .pStackHashInfoKecLo(this.hashInfoKeccak.loBigInt())
        .pStackHashInfoFlag(this.hashInfoFlag);
  }
}
