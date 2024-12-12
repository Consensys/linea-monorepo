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

package net.consensys.linea.zktracer.opcode;

import static net.consensys.linea.zktracer.opcode.InstructionFamily.*;

import java.util.Objects;

import net.consensys.linea.zktracer.opcode.gas.Billing;
import net.consensys.linea.zktracer.opcode.gas.BillingRate;
import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import net.consensys.linea.zktracer.opcode.gas.MxpType;
import net.consensys.linea.zktracer.opcode.stack.Pattern;
import net.consensys.linea.zktracer.opcode.stack.StackSettings;

/**
 * Contains the {@link OpCode} and its related metadata.
 *
 * @param mnemonic The type of the opcode represented by {@link OpCode}.
 * @param value The actual unsigned byte value of the opcode.
 * @param instructionFamily The {@link InstructionFamily} to which the opcode belongs.
 * @param stackSettings A {@link StackSettings} instance describing how the opcode alters the EVM
 *     stack.
 * @param ramSettings A {@link RamSettings} instance describing how the opcode alters the memory.
 * @param billing A {@link Billing} instance describing the billing scheme of the instruction.
 */
public record OpCodeData(
    OpCode mnemonic,
    int value,
    InstructionFamily instructionFamily,
    StackSettings stackSettings,
    RamSettings ramSettings,
    Billing billing) {

  public Billing billing() {
    return Objects.requireNonNullElse(billing, Billing.DEFAULT);
  }

  public static OpCodeData forNonOpCodes(int value) {
    return new OpCodeData(
        OpCode.INVALID,
        value,
        INVALID,
        new StackSettings(
            Pattern.ZERO_ZERO,
            0,
            0,
            GasConstants.G_ZERO,
            false,
            false,
            false,
            false,
            false,
            false,
            false,
            false),
        new RamSettings(DataLocation.NONE, DataLocation.NONE),
        new Billing(GasConstants.G_ZERO, BillingRate.NONE, MxpType.NONE));
  }

  /**
   * A method singling out <code>PUSHx</code> instructions.
   *
   * @return <code>true</code> if this opcode is a <code>PUSHx</code>
   */
  public boolean isPush() {
    return (0x60 <= value) && (value < 0x80);
  }

  public boolean isJumpDest() {
    return value == 0x5b;
  }

  public boolean isJump() {
    return instructionFamily == JUMP;
  }

  /**
   * Returns whether this instruction belong to the HALT family.
   *
   * @return true if {@link InstructionFamily} is HALT
   */
  public boolean isHalt() {
    return instructionFamily == HALT;
  }

  public boolean isCall() {
    return instructionFamily == CALL;
  }

  public boolean isCreate() {
    return instructionFamily == CREATE;
  }

  public boolean isLog() {
    return instructionFamily == LOG;
  }

  /**
   * Returns whether this instruction belong to the INVALID family.
   *
   * @return true if {@link InstructionFamily} is INVALID
   */
  public boolean isInvalid() {
    return instructionFamily == INVALID;
  }

  public boolean isMxp() {
    return this.billing().type() != MxpType.NONE;
  }

  public int numberOfStackRows() {
    return stackSettings.twoLineInstruction() ? 2 : 1;
  }
}
