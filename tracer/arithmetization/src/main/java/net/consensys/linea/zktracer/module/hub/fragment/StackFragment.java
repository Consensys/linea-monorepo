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
import net.consensys.linea.zktracer.module.hub.DeploymentExceptions;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.State;
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
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.evm.internal.Words;

@Accessors(fluent = true)
public final class StackFragment implements TraceFragment {
  private final Stack stack;
  @Getter private final List<StackOperation> stackOps;
  private final short exceptions;
  @Setter private DeploymentExceptions contextExceptions;
  private final long staticGas;
  @Setter public boolean hashInfoFlag;
  private EWord hashInfoKeccak = EWord.ZERO;
  @Setter public Bytes hash;
  @Getter private final OpCode opCode;
  @Setter private boolean jumpDestinationVettingRequired;
  @Setter private boolean validJumpDestination;
  private final boolean willRevert;
  private final State.TxState.Stamps stamps;

  private StackFragment(
      final Hub hub,
      Stack stack,
      List<StackOperation> stackOps,
      short exceptions,
      AbortingConditions aborts,
      DeploymentExceptions contextExceptions,
      GasProjection gp,
      boolean isDeploying,
      boolean willRevert) {
    this.stack = stack;
    this.stackOps = stackOps;
    this.exceptions = exceptions;
    this.contextExceptions = contextExceptions;
    this.opCode = stack.getCurrentOpcodeData().mnemonic();
    this.hashInfoFlag =
        switch (this.opCode) {
          case SHA3 -> Exceptions.none(exceptions) && gp.messageSize() > 0;
          case RETURN -> Exceptions.none(exceptions) && gp.messageSize() > 0 && isDeploying;
          case CREATE2 -> Exceptions.none(exceptions) && aborts.none() && gp.messageSize() > 0;
          default -> false;
        };
    if (this.hashInfoFlag) {
      Bytes memorySegmentToHash;
      switch (this.opCode) {
        case SHA3, RETURN -> {
          final long offset = Words.clampedToLong(hub.currentFrame().frame().getStackItem(0));
          final long size = Words.clampedToLong(hub.currentFrame().frame().getStackItem(1));
          memorySegmentToHash = hub.messageFrame().shadowReadMemory(offset, size);
        }
        case CREATE2 -> {
          final long offset = Words.clampedToLong(hub.currentFrame().frame().getStackItem(1));
          final long size = Words.clampedToLong(hub.currentFrame().frame().getStackItem(2));
          memorySegmentToHash = hub.messageFrame().shadowReadMemory(offset, size);
        }
        default -> throw new UnsupportedOperationException(
            "Hash was attempted by the following opcode: " + this.opCode().toString());
      }
      this.hashInfoKeccak = EWord.of(Hash.hash(memorySegmentToHash));
    }

    this.staticGas = gp.staticGas();

    if (opCode.isJump() && !Exceptions.stackException(exceptions)) {
      final BigInteger prospectivePcNew =
          hub.currentFrame().frame().getStackItem(0).toUnsignedBigInteger();
      final BigInteger codeSize = BigInteger.valueOf(hub.currentFrame().code().getSize());

      boolean prospectivePcNewIsInBounds = codeSize.compareTo(prospectivePcNew) > 0;

      if (opCode.equals(OpCode.JUMPI)) {
        boolean nonzeroJumpCondition =
            !hub.currentFrame()
                .frame()
                .getStackItem(1)
                .toUnsignedBigInteger()
                .equals(BigInteger.ZERO);
        prospectivePcNewIsInBounds = prospectivePcNewIsInBounds && nonzeroJumpCondition;
      }

      jumpDestinationVettingRequired = prospectivePcNewIsInBounds;
    } else {
      jumpDestinationVettingRequired = false;
    }

    this.willRevert = willRevert;
    this.stamps = hub.state().stamps();
  }

  public static StackFragment prepare(
      final Hub hub,
      final Stack stack,
      final List<StackOperation> stackOperations,
      final short exceptions,
      final AbortingConditions aborts,
      final GasProjection gp,
      boolean isDeploying,
      boolean willRevert) {
    return new StackFragment(
        hub,
        stack,
        stackOperations,
        exceptions,
        aborts,
        DeploymentExceptions.empty(),
        gp,
        isDeploying,
        willRevert);
  }

  private boolean traceLog() {
    return this.opCode.isLog()
        && Exceptions.none(
            this.exceptions) // TODO: should be redundant (exceptions trigger reverts) --- this
        // could be asserted
        && !this.willRevert;
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

    final InstructionFamily currentInstFamily =
        this.stack.getCurrentOpcodeData().instructionFamily();

    return trace
        .peekAtStack(true)
        // Instruction details
        .pStackAlpha(UnsignedByte.of(this.stack.getCurrentOpcodeData().stackSettings().alpha()))
        .pStackDelta(UnsignedByte.of(this.stack.getCurrentOpcodeData().stackSettings().delta()))
        .pStackNbAdded(UnsignedByte.of(this.stack.getCurrentOpcodeData().stackSettings().nbAdded()))
        .pStackNbRemoved(
            UnsignedByte.of(this.stack.getCurrentOpcodeData().stackSettings().nbRemoved()))
        .pStackInstruction(Bytes.of(this.stack.getCurrentOpcodeData().value()))
        .pStackStaticGas(staticGas)
        // Opcode families
        .pStackAccFlag(currentInstFamily == InstructionFamily.ACCOUNT)
        .pStackAddFlag(currentInstFamily == InstructionFamily.ADD)
        .pStackBinFlag(currentInstFamily == InstructionFamily.BIN)
        .pStackBtcFlag(currentInstFamily == InstructionFamily.BATCH)
        .pStackCallFlag(currentInstFamily == InstructionFamily.CALL)
        .pStackConFlag(currentInstFamily == InstructionFamily.CONTEXT)
        .pStackCopyFlag(currentInstFamily == InstructionFamily.COPY)
        .pStackCreateFlag(currentInstFamily == InstructionFamily.CREATE)
        .pStackDupFlag(currentInstFamily == InstructionFamily.DUP)
        .pStackExtFlag(currentInstFamily == InstructionFamily.EXT)
        .pStackHaltFlag(currentInstFamily == InstructionFamily.HALT)
        .pStackInvalidFlag(currentInstFamily == InstructionFamily.INVALID)
        .pStackJumpFlag(currentInstFamily == InstructionFamily.JUMP)
        .pStackKecFlag(currentInstFamily == InstructionFamily.KEC)
        .pStackLogFlag(currentInstFamily == InstructionFamily.LOG)
        .pStackMachineStateFlag(currentInstFamily == InstructionFamily.MACHINE_STATE)
        .pStackModFlag(currentInstFamily == InstructionFamily.MOD)
        .pStackMulFlag(currentInstFamily == InstructionFamily.MUL)
        .pStackPushpopFlag(currentInstFamily == InstructionFamily.PUSH_POP)
        .pStackShfFlag(currentInstFamily == InstructionFamily.SHF)
        .pStackStackramFlag(currentInstFamily == InstructionFamily.STACK_RAM)
        .pStackStoFlag(currentInstFamily == InstructionFamily.STORAGE)
        .pStackSwapFlag(currentInstFamily == InstructionFamily.SWAP)
        .pStackTxnFlag(currentInstFamily == InstructionFamily.TRANSACTION)
        .pStackWcpFlag(currentInstFamily == InstructionFamily.WCP)
        .pStackDecFlag1(this.stack.getCurrentOpcodeData().stackSettings().flag1())
        .pStackDecFlag2(this.stack.getCurrentOpcodeData().stackSettings().flag2())
        .pStackDecFlag3(this.stack.getCurrentOpcodeData().stackSettings().flag3())
        .pStackDecFlag4(this.stack.getCurrentOpcodeData().stackSettings().flag4())
        .pStackMxpFlag(
            Optional.ofNullable(this.stack.getCurrentOpcodeData().billing())
                .map(b -> b.type() != MxpType.NONE)
                .orElse(false))
        .pStackStaticFlag(this.stack.getCurrentOpcodeData().stackSettings().forbiddenInStatic())
        .pStackPushValueHi(pushValue.hi())
        .pStackPushValueLo(pushValue.lo())
        .pStackJumpDestinationVettingRequired(
            this.jumpDestinationVettingRequired) // TODO: confirm this
        // Exception flag
        .pStackOpcx(Exceptions.invalidCodePrefix(exceptions))
        .pStackSux(Exceptions.stackUnderflow(exceptions))
        .pStackSox(Exceptions.stackOverflow(exceptions))
        .pStackMxpx(Exceptions.memoryExpansionException(exceptions))
        .pStackOogx(Exceptions.outOfGasException(exceptions))
        .pStackRdcx(Exceptions.returnDataCopyFault(exceptions))
        .pStackJumpx(Exceptions.jumpFault(exceptions))
        .pStackStaticx(Exceptions.staticFault(exceptions))
        .pStackSstorex(Exceptions.outOfSStore(exceptions))
        .pStackIcpx(contextExceptions.invalidCodePrefix())
        .pStackMaxcsx(contextExceptions.codeSizeOverflow())
        // Hash data
        .pStackHashInfoFlag(this.hashInfoFlag)
        .pStackHashInfoKeccakHi(this.hashInfoKeccak.hi())
        .pStackHashInfoKeccakLo(this.hashInfoKeccak.lo())
        .pStackLogInfoFlag(this.traceLog()) // TODO: confirm this
    ;
  }
}
