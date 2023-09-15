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

package net.consensys.linea.zktracer.opcode.gas.projector;

import static org.hyperledger.besu.evm.internal.Words.clampedToLong;

import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.internal.Words;

public final class GasProjector {
  private final GasCalculator gc;

  public GasProjector(GasCalculator gc) {
    this.gc = gc;
  }

  public GasProjection of(MessageFrame frame, OpCode opCode) {
    switch (opCode) {
      case STOP -> new Zero(gc);
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
          SWAP16 -> new VeryLow(gc);
      case MUL, DIV, SDIV, MOD, SMOD, SIGNEXTEND, SELFBALANCE -> new Low(gc);
      case ADDMOD, MULMOD, JUMP -> new Mid(gc);
      case EXP -> new Exp(gc, frame);
      case SHA3 -> new Sha3(gc, frame);
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
          BASEFEE -> new Base(gc);
      case BALANCE, EXTCODESIZE, EXTCODEHASH -> new AccountAccess(gc, frame);
      case CALLDATACOPY, CODECOPY, RETURNDATACOPY -> new DataCopy(gc, frame);
      case EXTCODECOPY -> new ExtCodeCopy(gc, frame);
      case BLOCKHASH -> new BlockHash(gc);
      case MLOAD, MSTORE -> new MLoadStore(gc, frame);
      case MSTORE8 -> new MStore8(gc, frame);
      case SLOAD -> new SLoad(gc, frame);
      case SSTORE -> {
        final UInt256 key = UInt256.fromBytes(frame.getStackItem(0));
        final Account account = frame.getWorldUpdater().getAccount(frame.getRecipientAddress());
        final UInt256 currentValue = account.getStorageValue(key);
        final UInt256 originalValue = account.getOriginalStorageValue(key);
        final UInt256 newValue = UInt256.fromBytes(frame.getStackItem(1));

        return new SStore(gc, frame, key, originalValue, currentValue, newValue);
      }
      case JUMPI -> new High(gc);
      case JUMPDEST -> new JumpDest(gc);
      case LOG0 -> {
        long offset = clampedToLong(frame.getStackItem(0));
        long size = clampedToLong(frame.getStackItem(1));

        return new Log(gc, frame, offset, size, 0);
      }
      case LOG1 -> {
        long offset = clampedToLong(frame.getStackItem(0));
        long size = clampedToLong(frame.getStackItem(1));

        return new Log(gc, frame, offset, size, 1);
      }
      case LOG2 -> {
        long offset = clampedToLong(frame.getStackItem(0));
        long size = clampedToLong(frame.getStackItem(1));

        return new Log(gc, frame, offset, size, 2);
      }
      case LOG3 -> {
        long offset = clampedToLong(frame.getStackItem(0));
        long size = clampedToLong(frame.getStackItem(1));

        return new Log(gc, frame, offset, size, 3);
      }
      case LOG4 -> {
        long offset = clampedToLong(frame.getStackItem(0));
        long size = clampedToLong(frame.getStackItem(1));

        return new Log(gc, frame, offset, size, 4);
      }
      case CREATE -> {
        final long initCodeOffset = clampedToLong(frame.getStackItem(1));
        final long initCodeLength = clampedToLong(frame.getStackItem(2));

        return new Create(gc, frame, initCodeOffset, initCodeLength);
      }
      case CREATE2 -> {
        final long initCodeOffset = clampedToLong(frame.getStackItem(1));
        final long initCodeLength = clampedToLong(frame.getStackItem(2));

        return new Create2(gc, frame, initCodeOffset, initCodeLength);
      }
      case CALL -> {
        final long stipend = clampedToLong(frame.getStackItem(0));
        final Account recipient =
            frame.getWorldUpdater().get(Words.toAddress(frame.getStackItem(1)));
        final Address to = recipient.getAddress();
        final Wei value = Wei.wrap(frame.getStackItem(2));
        final long inputDataOffset = clampedToLong(frame.getStackItem(3));
        final long inputDataLength = clampedToLong(frame.getStackItem(4));
        final long returnDataOffset = clampedToLong(frame.getStackItem(5));
        final long returnDataLength = clampedToLong(frame.getStackItem(6));
        return new Call(
            gc,
            frame,
            stipend,
            inputDataOffset,
            inputDataLength,
            returnDataOffset,
            returnDataLength,
            value,
            recipient,
            to);
      }
      case CALLCODE -> {
        final long stipend = clampedToLong(frame.getStackItem(0));
        final Account recipient = frame.getWorldUpdater().get(frame.getRecipientAddress());
        final Address to = Words.toAddress(frame.getStackItem(1));
        final Wei value = Wei.wrap(frame.getStackItem(2));
        final long inputDataOffset = clampedToLong(frame.getStackItem(3));
        final long inputDataLength = clampedToLong(frame.getStackItem(4));
        final long returnDataOffset = clampedToLong(frame.getStackItem(5));
        final long returnDataLength = clampedToLong(frame.getStackItem(6));
        return new Call(
            gc,
            frame,
            stipend,
            inputDataOffset,
            inputDataLength,
            returnDataOffset,
            returnDataLength,
            value,
            recipient,
            to);
      }
      case DELEGATECALL -> {
        final long stipend = clampedToLong(frame.getStackItem(0));
        final Account recipient = frame.getWorldUpdater().get(frame.getRecipientAddress());
        final Address to = Words.toAddress(frame.getStackItem(1));
        final long inputDataOffset = clampedToLong(frame.getStackItem(2));
        final long inputDataLength = clampedToLong(frame.getStackItem(3));
        final long returnDataOffset = clampedToLong(frame.getStackItem(4));
        final long returnDataLength = clampedToLong(frame.getStackItem(5));
        return new Call(
            gc,
            frame,
            stipend,
            inputDataOffset,
            inputDataLength,
            returnDataOffset,
            returnDataLength,
            Wei.ZERO,
            recipient,
            to);
      }
      case STATICCALL -> {
        final long stipend = clampedToLong(frame.getStackItem(0));
        final Account recipient =
            frame.getWorldUpdater().get(Words.toAddress(frame.getStackItem(1)));
        final Address to = recipient.getAddress();
        final long inputDataOffset = clampedToLong(frame.getStackItem(2));
        final long inputDataLength = clampedToLong(frame.getStackItem(3));
        final long returnDataOffset = clampedToLong(frame.getStackItem(4));
        final long returnDataLength = clampedToLong(frame.getStackItem(5));
        return new Call(
            gc,
            frame,
            stipend,
            inputDataOffset,
            inputDataLength,
            returnDataOffset,
            returnDataLength,
            Wei.ZERO,
            recipient,
            to);
      }
      case RETURN -> new Return(gc, frame);
      case REVERT -> {
        final long offset = clampedToLong(frame.popStackItem());
        final long length = clampedToLong(frame.popStackItem());

        return new Revert(gc, frame, offset, length);
      }
      case INVALID -> new GasProjection() {};
      case SELFDESTRUCT -> new SelfDestruct(gc, frame);
      default -> throw new IllegalStateException("Unexpected value: " + opCode);
    }
    throw new IllegalStateException();
  }
}
