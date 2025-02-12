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

import static net.consensys.linea.zktracer.module.tables.instructionDecoder.Trace.headers;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.opcode.InstructionFamily;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.gas.BillingRate;
import net.consensys.linea.zktracer.opcode.gas.MxpType;
import net.consensys.linea.zktracer.types.UnsignedByte;

public final class InstructionDecoder implements Module {
  private static void traceFamily(OpCodeData op, Trace trace) {
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
        .familyMachineState(op.instructionFamily() == InstructionFamily.MACHINE_STATE)
        .familyPushPop(op.instructionFamily() == InstructionFamily.PUSH_POP)
        .familyDup(op.instructionFamily() == InstructionFamily.DUP)
        .familySwap(op.instructionFamily() == InstructionFamily.SWAP)
        .familyLog(op.instructionFamily() == InstructionFamily.LOG)
        .familyCreate(op.instructionFamily() == InstructionFamily.CREATE)
        .familyCall(op.instructionFamily() == InstructionFamily.CALL)
        .familyHalt(op.instructionFamily() == InstructionFamily.HALT)
        .familyInvalid(op.instructionFamily() == InstructionFamily.INVALID);
  }

  private static void traceStackSettings(OpCodeData op, Trace trace) {
    trace
        .alpha(UnsignedByte.of(op.stackSettings().alpha()))
        .delta(UnsignedByte.of(op.stackSettings().delta()))
        .staticGas(op.stackSettings().staticGas().cost())
        .twoLineInstruction(op.stackSettings().twoLineInstruction())
        .staticFlag(op.stackSettings().forbiddenInStatic())
        .flag1(op.stackSettings().flag1())
        .flag2(op.stackSettings().flag2())
        .flag3(op.stackSettings().flag3())
        .flag4(op.stackSettings().flag4());
  }

  private static void traceBillingSettings(OpCodeData op, Trace trace) {
    trace
        .billingPerWord(
            UnsignedByte.of(
                op.billing().billingRate() == BillingRate.BY_WORD
                    ? op.billing().perUnit().cost()
                    : 0))
        .billingPerByte(
            UnsignedByte.of(
                op.billing().billingRate() == BillingRate.BY_BYTE
                    ? op.billing().perUnit().cost()
                    : 0))
        .mxpType1(op.billing().type() == MxpType.TYPE_1)
        .mxpType2(op.billing().type() == MxpType.TYPE_2)
        .mxpType3(op.billing().type() == MxpType.TYPE_3)
        .mxpType4(op.billing().type() == MxpType.TYPE_4)
        .mxpType5(op.billing().type() == MxpType.TYPE_5)
        .mxpFlag(op.isMxp());
  }

  @Override
  public String moduleKey() {
    return "INSTRUCTION_DECODER";
  }

  @Override
  public void commitTransactionBundle() {}

  @Override
  public void popTransactionBundle() {}

  @Override
  public int lineCount() {
    return 256;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    Trace trace = new Trace(buffers);

    for (int i = 0; i < 256; i++) {
      final OpCodeData op = OpCode.of(i).getData();

      traceFamily(op, trace);
      traceStackSettings(op, trace);
      traceBillingSettings(op, trace);
      trace
          .opcode(UnsignedByte.of(i))
          .isPush(op.isPush())
          .isJumpdest(op.isJumpDest())
          .validateRow();
    }
  }
}
