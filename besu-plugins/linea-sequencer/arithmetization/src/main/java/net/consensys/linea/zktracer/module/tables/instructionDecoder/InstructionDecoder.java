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

package net.consensys.linea.zktracer.module.tables.instructionDecoder;

import java.math.BigInteger;

import net.consensys.linea.zktracer.module.ModuleTrace;
import net.consensys.linea.zktracer.opcode.DataLocation;
import net.consensys.linea.zktracer.opcode.InstructionFamily;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.gas.BillingRate;
import net.consensys.linea.zktracer.opcode.gas.MxpType;
import net.consensys.linea.zktracer.opcode.stack.Pattern;
import net.consensys.linea.zktracer.types.UnsignedByte;

public final class InstructionDecoder {
  private static void traceFamily(OpCodeData op, Trace.TraceBuilder trace) {
    trace
        .familyAdd(op.instructionFamily() == InstructionFamily.ADD)
        .familyMod(op.instructionFamily() == InstructionFamily.MOD)
        .familyMul(op.instructionFamily() == InstructionFamily.MUL)
        .familyExt(op.instructionFamily() == InstructionFamily.EXT)
        .familyWcp(op.instructionFamily() == InstructionFamily.WCP)
        .familyBin(op.instructionFamily() == InstructionFamily.BIN)
        .familyShf(op.instructionFamily() == InstructionFamily.SHF)
        .familyKec(op.instructionFamily() == InstructionFamily.KEC)
        .familyContext(op.instructionFamily() == InstructionFamily.CONTEXT)
        .familyAccount(op.instructionFamily() == InstructionFamily.ACCOUNT)
        .familyCopy(op.instructionFamily() == InstructionFamily.COPY)
        .familyTransaction(op.instructionFamily() == InstructionFamily.TRANSACTION)
        .familyBatch(op.instructionFamily() == InstructionFamily.BATCH)
        .familyStackRam(op.instructionFamily() == InstructionFamily.STACK_RAM)
        .familyStorage(op.instructionFamily() == InstructionFamily.STORAGE)
        .familyJump(op.instructionFamily() == InstructionFamily.JUMP)
        .familyMachineState(op.instructionFamily() == InstructionFamily.JUMP)
        .familyPushPop(op.instructionFamily() == InstructionFamily.PUSH_POP)
        .familyDup(op.instructionFamily() == InstructionFamily.DUP)
        .familySwap(op.instructionFamily() == InstructionFamily.SWAP)
        .familyLog(op.instructionFamily() == InstructionFamily.LOG)
        .familyCreate(op.instructionFamily() == InstructionFamily.CREATE)
        .familyCall(op.instructionFamily() == InstructionFamily.CALL)
        .familyHalt(op.instructionFamily() == InstructionFamily.HALT)
        .familyInvalid(op.instructionFamily() == InstructionFamily.INVALID);
  }

  private static void traceStackSettings(OpCodeData op, Trace.TraceBuilder trace) {
    trace
        .patternZeroZero(op.stackSettings().pattern() == Pattern.ZERO_ZERO)
        .patternOneZero(op.stackSettings().pattern() == Pattern.ONE_ZERO)
        .patternTwoZero(op.stackSettings().pattern() == Pattern.TWO_ZERO)
        .patternZeroOne(op.stackSettings().pattern() == Pattern.ZERO_ONE)
        .patternOneOne(op.stackSettings().pattern() == Pattern.ONE_ONE)
        .patternTwoOne(op.stackSettings().pattern() == Pattern.TWO_ONE)
        .patternThreeOne(op.stackSettings().pattern() == Pattern.THREE_ONE)
        .patternLoadStore(op.stackSettings().pattern() == Pattern.LOAD_STORE)
        .patternDup(op.stackSettings().pattern() == Pattern.DUP)
        .patternSwap(op.stackSettings().pattern() == Pattern.SWAP)
        .patternLog(op.stackSettings().pattern() == Pattern.LOG)
        .patternCopy(op.stackSettings().pattern() == Pattern.COPY)
        .patternCall(op.stackSettings().pattern() == Pattern.CALL)
        .patternCreate(op.stackSettings().pattern() == Pattern.CREATE)
        .alpha(UnsignedByte.of(op.stackSettings().alpha()))
        .delta(UnsignedByte.of(op.stackSettings().delta()))
        .nbAdded(UnsignedByte.of(op.stackSettings().nbAdded()))
        .nbRemoved(UnsignedByte.of(op.stackSettings().nbRemoved()))
        .staticGas(BigInteger.valueOf(op.stackSettings().staticGas().cost()))
        .twoLinesInstruction(op.stackSettings().twoLinesInstruction())
        .forbiddenInStatic(op.stackSettings().forbiddenInStatic())
        .addressTrimmingInstruction(op.stackSettings().addressTrimmingInstruction())
        .flag1(op.stackSettings().flag1())
        .flag2(op.stackSettings().flag2())
        .flag3(op.stackSettings().flag3())
        .flag4(op.stackSettings().flag4());
  }

  private static void traceRamSettings(OpCodeData op, Trace.TraceBuilder trace) {
    trace
        .ramEnabled(op.ramSettings().enabled())
        // Source
        .ramSourceRom(op.ramSettings().source() == DataLocation.ROM)
        .ramSourceRam(op.ramSettings().source() == DataLocation.RAM)
        .ramSourceTxnData(op.ramSettings().source() == DataLocation.TXN_CALL_DATA)
        .ramSourceStack(op.ramSettings().source() == DataLocation.STACK)
        .ramSourceEcData(op.ramSettings().source() == DataLocation.EC_DATA)
        .ramSourceEcInfo(op.ramSettings().source() == DataLocation.EC_INFO)
        .ramSourceModexpData(op.ramSettings().source() == DataLocation.MOD_EXP_DATA)
        .ramSourceHashData(op.ramSettings().source() == DataLocation.HASH_DATA)
        .ramSourceHashInfo(op.ramSettings().source() == DataLocation.HASH_INFO)
        .ramSourceBlakeData(op.ramSettings().source() == DataLocation.BLAKE_DATA)
        .ramSourceLogData(op.ramSettings().source() == DataLocation.LOG_DATA)
        // Target
        .ramTargetRom(op.ramSettings().target() == DataLocation.ROM)
        .ramTargetRam(op.ramSettings().target() == DataLocation.RAM)
        .ramTargetTxnData(op.ramSettings().target() == DataLocation.TXN_CALL_DATA)
        .ramTargetStack(op.ramSettings().target() == DataLocation.STACK)
        .ramTargetEcData(op.ramSettings().target() == DataLocation.EC_DATA)
        .ramTargetEcInfo(op.ramSettings().target() == DataLocation.EC_INFO)
        .ramTargetModexpData(op.ramSettings().target() == DataLocation.MOD_EXP_DATA)
        .ramTargetHashData(op.ramSettings().target() == DataLocation.HASH_DATA)
        .ramTargetHashInfo(op.ramSettings().target() == DataLocation.HASH_INFO)
        .ramTargetBlakeData(op.ramSettings().target() == DataLocation.BLAKE_DATA)
        .ramTargetLogData(op.ramSettings().target() == DataLocation.LOG_DATA);
  }

  private static void traceBillingSettings(OpCodeData op, Trace.TraceBuilder trace) {
    trace
        .billingPerWord(
            BigInteger.valueOf(
                op.billing().billingRate() == BillingRate.BY_WORD
                    ? op.billing().perUnit().cost()
                    : 0))
        .billingPerByte(
            BigInteger.valueOf(
                op.billing().billingRate() == BillingRate.BY_BYTE
                    ? op.billing().perUnit().cost()
                    : 0))
        .mxpType1(op.billing().type() == MxpType.TYPE_1)
        .mxpType2(op.billing().type() == MxpType.TYPE_2)
        .mxpType3(op.billing().type() == MxpType.TYPE_3)
        .mxpType4(op.billing().type() == MxpType.TYPE_4)
        .mxpType5(op.billing().type() == MxpType.TYPE_5);
  }

  public static ModuleTrace generate() {
    Trace.TraceBuilder trace = new Trace.TraceBuilder(256);

    for (int i = 0; i < 256; i++) {
      final OpCodeData op = OpCode.of(i).getData();

      traceFamily(op, trace);
      traceStackSettings(op, trace);
      traceRamSettings(op, trace);
      traceBillingSettings(op, trace);
      trace
          .opcode(BigInteger.valueOf(i))
          .isPush(op.pushFlag())
          .isJumpdest(op.jumpFlag())
          .validateRow();
    }

    return new InstructionDecoderTrace(trace.build());
  }
}
