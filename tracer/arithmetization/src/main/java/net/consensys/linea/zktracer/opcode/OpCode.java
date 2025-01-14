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

import static com.google.common.base.Preconditions.*;

import net.consensys.linea.zktracer.opcode.gas.MxpType;
import net.consensys.linea.zktracer.types.UnsignedByte;

/** Represents the entire set of opcodes that are required by the arithmetization process. */
public enum OpCode {
  STOP,
  ADD,
  MUL,
  SUB,
  DIV,
  SDIV,
  MOD,
  SMOD,
  ADDMOD,
  MULMOD,
  EXP,
  SIGNEXTEND,
  LT,
  GT,
  SLT,
  SGT,
  EQ,
  ISZERO,
  AND,
  OR,
  XOR,
  NOT,
  BYTE,
  SHL,
  SHR,
  SAR,
  SHA3,
  ADDRESS,
  BALANCE,
  ORIGIN,
  CALLER,
  CALLVALUE,
  CALLDATALOAD,
  CALLDATASIZE,
  CALLDATACOPY,
  CODESIZE,
  CODECOPY,
  GASPRICE,
  EXTCODESIZE,
  EXTCODECOPY,
  RETURNDATASIZE,
  RETURNDATACOPY,
  EXTCODEHASH,
  BLOCKHASH,
  COINBASE,
  TIMESTAMP,
  NUMBER,
  DIFFICULTY,
  GASLIMIT,
  CHAINID,
  SELFBALANCE,
  BASEFEE,
  POP,
  MLOAD,
  MSTORE,
  MSTORE8,
  SLOAD,
  SSTORE,
  JUMP,
  JUMPI,
  PC,
  MSIZE,
  GAS,
  JUMPDEST,
  // PUSH0,
  PUSH1,
  PUSH2,
  PUSH3,
  PUSH4,
  PUSH5,
  PUSH6,
  PUSH7,
  PUSH8,
  PUSH9,
  PUSH10,
  PUSH11,
  PUSH12,
  PUSH13,
  PUSH14,
  PUSH15,
  PUSH16,
  PUSH17,
  PUSH18,
  PUSH19,
  PUSH20,
  PUSH21,
  PUSH22,
  PUSH23,
  PUSH24,
  PUSH25,
  PUSH26,
  PUSH27,
  PUSH28,
  PUSH29,
  PUSH30,
  PUSH31,
  PUSH32,
  DUP1,
  DUP2,
  DUP3,
  DUP4,
  DUP5,
  DUP6,
  DUP7,
  DUP8,
  DUP9,
  DUP10,
  DUP11,
  DUP12,
  DUP13,
  DUP14,
  DUP15,
  DUP16,
  SWAP1,
  SWAP2,
  SWAP3,
  SWAP4,
  SWAP5,
  SWAP6,
  SWAP7,
  SWAP8,
  SWAP9,
  SWAP10,
  SWAP11,
  SWAP12,
  SWAP13,
  SWAP14,
  SWAP15,
  SWAP16,
  LOG0,
  LOG1,
  LOG2,
  LOG3,
  LOG4,
  CREATE,
  CALL,
  CALLCODE,
  RETURN,
  DELEGATECALL,
  CREATE2,
  STATICCALL,
  REVERT,
  INVALID,
  SELFDESTRUCT;

  /**
   * Convert a mnemonic in any case into the matching {@link OpCode}.
   *
   * @param mnemonic the opcode menmonic
   * @return the corresponding OpCode
   */
  public static OpCode fromMnemonic(final String mnemonic) {
    return OpCode.valueOf(mnemonic.toUpperCase());
  }

  /**
   * Retrieves {@link OpCode} metadata of type {@link OpCodeData}.
   *
   * @return the current {@link OpCode}'s {@link OpCodeData}
   */
  public OpCodeData getData() {
    return OpCodes.of(this);
  }

  /**
   * Retrieves the {@link OpCode} corresponding to a given value.
   *
   * @return the {@link OpCode}
   */
  public static OpCode of(final int value) {
    return OpCodes.of(value).mnemonic();
  }

  /**
   * Returns the {@link OpCode}'s long value as a byte type.
   *
   * @return the {@link OpCode}'s value as a byte
   */
  public byte byteValue() {
    return (byte) this.getData().value();
  }

  /**
   * Returns the {@link OpCode}'s long value as an {@link UnsignedByte} type.
   *
   * @return the {@link OpCode}'s value as an {@link UnsignedByte}
   */
  public UnsignedByte unsignedByteValue() {
    return UnsignedByte.of(byteValue());
  }

  /** Returns true for PUSH-type instructions */
  public boolean isPush() {
    return getData().isPush();
  }

  /** Returns true for JUMP-type instructions */
  public boolean isJump() {
    return getData().isJump();
  }

  public boolean isLog() {
    return getData().isLog();
  }

  /** Returns whether the {@link OpCode} entails a contract creation. */
  public boolean isCreate() {
    return getData().isCreate();
  }

  /** Returns whether the {@link OpCode} is one of the CALL opcodes */
  public boolean isCall() {
    return getData().isCall();
  }

  public boolean isHalt() {
    return getData().isHalt();
  }

  public boolean isCallOrCreate() {
    return isCall() || isCreate();
  }

  public boolean callHasNoValueArgument() {
    checkArgument(isCall());
    return this == OpCode.DELEGATECALL || this == OpCode.STATICCALL;
  }

  public boolean callHasValueArgument() {
    checkArgument(isCall());
    return this == OpCode.CALL || this == OpCode.CALLCODE;
  }

  /**
   * Matches if the current {@link OpCode} is contained within a list of {@link OpCode}s.
   *
   * @param opCodes list of {@link OpCode}s to match against.
   * @return if the current {@link OpCode} is contained within the list.
   */
  public boolean isAnyOf(OpCode... opCodes) {
    for (OpCode opCode : opCodes) {
      if (opCode.equals(this)) {
        return true;
      }
    }

    return false;
  }

  public short numberOfStackRows() {
    return (short) (this.getData().numberOfStackRows());
  }

  public boolean mayTriggerStackUnderflow() {
    return this.getData().stackSettings().delta() > 0;
  }

  public boolean mayTriggerStackOverflow() {
    return this.getData().stackSettings().alpha() > 0;
  }

  public boolean mayTriggerStaticException() {
    return this.getData().stackSettings().forbiddenInStatic();
  }

  public boolean mayTriggerMemoryExpansionException() {
    return this != MSIZE && this.getData().billing().type() != MxpType.NONE;
  }
}
