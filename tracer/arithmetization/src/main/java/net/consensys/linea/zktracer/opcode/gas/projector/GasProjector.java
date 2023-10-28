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

package net.consensys.linea.zktracer.opcode.gas.projector;

import static org.hyperledger.besu.evm.internal.Words.clampedToLong;

import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

public class GasProjector {
  public GasProjection of(MessageFrame frame, OpCode opCode) {
    return switch (opCode) {
      case STOP -> new Zero();
      case ADD,
          SUB,
          NOT,
          LT,
          GT,
          SLT,
          SGT,
          EQ,
          ISZERO,
          AND,
          OR,
          XOR,
          BYTE,
          SHL,
          SHR,
          SAR,
          CALLDATALOAD,
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
          SWAP16 -> new VeryLow();
      case MUL, DIV, SDIV, MOD, SMOD, SIGNEXTEND, SELFBALANCE -> new Low();
      case ADDMOD, MULMOD, JUMP -> new Mid();
      case EXP -> new Exp(frame);
      case SHA3 -> new Sha3(frame);
      case ADDRESS,
          ORIGIN,
          CALLER,
          CALLVALUE,
          CALLDATASIZE,
          CODESIZE,
          GASPRICE,
          COINBASE,
          TIMESTAMP,
          NUMBER,
          DIFFICULTY,
          GASLIMIT,
          CHAINID,
          RETURNDATASIZE,
          POP,
          PC,
          MSIZE,
          GAS,
          BASEFEE -> new Base();
      case BALANCE, EXTCODESIZE, EXTCODEHASH -> new AccountAccess(frame);
      case CALLDATACOPY, CODECOPY, RETURNDATACOPY -> new DataCopy(frame);
      case EXTCODECOPY -> new ExtCodeCopy(frame);
      case BLOCKHASH -> new BlockHash();
      case MLOAD, MSTORE -> new MLoadStore(frame);
      case MSTORE8 -> new MStore8(frame);
      case SLOAD -> new SLoad(frame);
      case SSTORE -> new SStore(frame);
      case JUMPI -> new High();
      case JUMPDEST -> new JumpDest();
      case LOG0 -> new Log(frame, 0);
      case LOG1 -> new Log(frame, 1);
      case LOG2 -> new Log(frame, 2);
      case LOG3 -> new Log(frame, 3);
      case LOG4 -> new Log(frame, 4);
      case CREATE -> new Create(frame);
      case CREATE2 -> new Create2(frame);
      case CALL -> {
        if (frame.stackSize() > 6) {
          final long stipend = clampedToLong(frame.getStackItem(0));
          final Address to = Words.toAddress(frame.getStackItem(1));
          final Account recipient = frame.getWorldUpdater().get(to);
          final Wei value = Wei.wrap(frame.getStackItem(2));
          final long inputDataOffset = clampedToLong(frame.getStackItem(3));
          final long inputDataLength = clampedToLong(frame.getStackItem(4));
          final long returnDataOffset = clampedToLong(frame.getStackItem(5));
          final long returnDataLength = clampedToLong(frame.getStackItem(6));
          yield new Call(
              frame,
              stipend,
              inputDataOffset,
              inputDataLength,
              returnDataOffset,
              returnDataLength,
              value,
              recipient,
              to);
        } else {
          yield Call.invalid();
        }
      }
      case CALLCODE -> {
        if (frame.stackSize() > 6) {
          final long stipend = clampedToLong(frame.getStackItem(0));
          final Account recipient = frame.getWorldUpdater().get(frame.getRecipientAddress());
          final Address to = Words.toAddress(frame.getStackItem(1));
          final Wei value = Wei.wrap(frame.getStackItem(2));
          final long inputDataOffset = clampedToLong(frame.getStackItem(3));
          final long inputDataLength = clampedToLong(frame.getStackItem(4));
          final long returnDataOffset = clampedToLong(frame.getStackItem(5));
          final long returnDataLength = clampedToLong(frame.getStackItem(6));
          yield new Call(
              frame,
              stipend,
              inputDataOffset,
              inputDataLength,
              returnDataOffset,
              returnDataLength,
              value,
              recipient,
              to);
        } else {
          yield Call.invalid();
        }
      }
      case DELEGATECALL -> {
        if (frame.stackSize() > 5) {
          final long stipend = clampedToLong(frame.getStackItem(0));
          final Account recipient = frame.getWorldUpdater().get(frame.getRecipientAddress());
          final Address to = Words.toAddress(frame.getStackItem(1));
          final long inputDataOffset = clampedToLong(frame.getStackItem(2));
          final long inputDataLength = clampedToLong(frame.getStackItem(3));
          final long returnDataOffset = clampedToLong(frame.getStackItem(4));
          final long returnDataLength = clampedToLong(frame.getStackItem(5));
          yield new Call(
              frame,
              stipend,
              inputDataOffset,
              inputDataLength,
              returnDataOffset,
              returnDataLength,
              Wei.ZERO,
              recipient,
              to);
        } else {
          yield Call.invalid();
        }
      }
      case STATICCALL -> {
        if (frame.stackSize() > 5) {
          final long stipend = clampedToLong(frame.getStackItem(0));
          final Address to = Words.toAddress(frame.getStackItem(1));
          final Account recipient = frame.getWorldUpdater().get(to);
          final long inputDataOffset = clampedToLong(frame.getStackItem(2));
          final long inputDataLength = clampedToLong(frame.getStackItem(3));
          final long returnDataOffset = clampedToLong(frame.getStackItem(4));
          final long returnDataLength = clampedToLong(frame.getStackItem(5));
          yield new Call(
              frame,
              stipend,
              inputDataOffset,
              inputDataLength,
              returnDataOffset,
              returnDataLength,
              Wei.ZERO,
              recipient,
              to);
        } else {
          yield Call.invalid();
        }
      }
      case RETURN -> new Return(frame);
      case REVERT -> new Revert(frame);
      case INVALID -> new GasProjection() {};
      case SELFDESTRUCT -> new SelfDestruct(frame);
      default -> throw new IllegalStateException("Unexpected value: " + opCode);
    };
  }
}
