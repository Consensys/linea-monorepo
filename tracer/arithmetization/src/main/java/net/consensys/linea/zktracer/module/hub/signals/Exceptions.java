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

package net.consensys.linea.zktracer.module.hub.signals;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EIP_3541_MARKER;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MAX_CODE_SIZE;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjector;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

/** Encode the exceptions that may be triggered by the execution of an instruction. */
public class Exceptions {
  private Exceptions() {}

  public static final short NONE = 0; // no exceptions occurred
  private static final short INVALID_OPCODE = 1; // unknown opcode
  private static final short STACK_UNDERFLOW = 2; // stack underflow
  private static final short STACK_OVERFLOW = 4; // stack overflow
  private static final short OUT_OF_MEMORY_EXPANSION = 8; // tried to use memory too far away
  private static final short OUT_OF_GAS = 16; // not enough gas for instruction
  private static final short RETURN_DATA_COPY_FAULT = 32; // trying to read past the RETURNDATA end
  private static final short JUMP_FAULT = 64; // jumping to an invalid destination
  private static final short STATIC_FAULT =
      128; // trying to execute non-static instruction in a static context
  private static final short OUT_OF_SSTORE = 256; // not enough gas to execute an SSTORE
  private static final short INVALID_CODE_PREFIX = 512;
  private static final short CODE_SIZE_OVERFLOW = 2048;

  public static boolean stackException(final short bitmask) {
    return stackOverflow(bitmask) || stackUnderflow(bitmask);
  }

  /**
   * @return true if no stack exception has been raised
   */
  public static boolean noStackException(final short bitmask) {
    return !stackException(bitmask);
  }

  /**
   * @return true if any exception flag has been raised
   */
  public static boolean any(final short bitmask) {
    return !none(bitmask);
  }

  /**
   * @return true if no exception flag has been raised
   */
  public static boolean none(final short bitmask) {
    return bitmask == NONE;
  }

  public static boolean invalidOpcode(final short bitmask) {
    return (bitmask & INVALID_OPCODE) != 0;
  }

  public static boolean stackUnderflow(final short bitmask) {
    return (bitmask & STACK_UNDERFLOW) != 0;
  }

  public static boolean stackOverflow(final short bitmask) {
    return (bitmask & STACK_OVERFLOW) != 0;
  }

  public static boolean outOfMemoryExpansion(final short bitmask) {
    return (bitmask & OUT_OF_MEMORY_EXPANSION) != 0;
  }

  public static boolean outOfGas(final short bitmask) {
    return (bitmask & OUT_OF_GAS) != 0;
  }

  public static boolean returnDataCopyFault(final short bitmask) {
    return (bitmask & RETURN_DATA_COPY_FAULT) != 0;
  }

  public static boolean jumpFault(final short bitmask) {
    return (bitmask & JUMP_FAULT) != 0;
  }

  public static boolean staticFault(final short bitmask) {
    return (bitmask & STATIC_FAULT) != 0;
  }

  public static boolean outOfSStore(final short bitmask) {
    return (bitmask & OUT_OF_SSTORE) != 0;
  }

  public static boolean invalidCodePrefix(final short bitmask) {
    return (bitmask & INVALID_CODE_PREFIX) != 0;
  }

  public static boolean codeSizeOverflow(final short bitmask) {
    return (bitmask & CODE_SIZE_OVERFLOW) != 0;
  }

  private static boolean isInvalidOpcode(final OpCode opCode) {
    return opCode == OpCode.INVALID;
  }

  private static boolean isStackUnderflow(final MessageFrame frame, OpCodeData opCodeData) {
    return frame.stackSize() < opCodeData.stackSettings().nbRemoved();
  }

  private static boolean isStackOverflow(final MessageFrame frame, OpCodeData opCodeData) {
    return frame.stackSize()
            + opCodeData.stackSettings().nbAdded()
            - opCodeData.stackSettings().nbRemoved()
        > 1024;
  }

  private static boolean isMemoryExpansionFault(
      MessageFrame frame, OpCode opCode, GasProjector gp) {
    return gp.of(frame, opCode).largestOffset() > 0xffffffffL;
  }

  private static boolean isOutOfGas(MessageFrame frame, OpCode opCode, GasProjector gp) {
    final long required = gp.of(frame, opCode).total();
    return required > frame.getRemainingGas();
  }

  private static boolean isReturnDataCopyFault(final MessageFrame frame, final OpCode opCode) {
    if (opCode == OpCode.RETURNDATACOPY) {
      long returnDataSize = frame.getReturnData().size();
      long askedOffset = Words.clampedToLong(frame.getStackItem(1));
      long askedSize = Words.clampedToLong(frame.getStackItem(2));

      return Words.clampedAdd(askedOffset, askedSize) > returnDataSize;
    }

    return false;
  }

  private static boolean isJumpFault(final MessageFrame frame, OpCode opCode) {
    if (opCode == OpCode.JUMP || opCode == OpCode.JUMPI) {
      final long target = Words.clampedToLong(frame.getStackItem(0));
      final boolean invalidDestination = frame.getCode().isJumpDestInvalid((int) target);

      switch (opCode) {
        case JUMP -> {
          return invalidDestination;
        }
        case JUMPI -> {
          long condition = Words.clampedToLong(frame.getStackItem(1));
          return (condition != 0) && invalidDestination;
        }
        default -> {
          return false;
        }
      }
    }

    return false;
  }

  private static boolean isStaticFault(final MessageFrame frame, OpCodeData opCodeData) {
    if (frame.isStatic() && opCodeData.mnemonic() == OpCode.CALL && frame.stackSize() > 2) {
      final long value = Words.clampedToLong(frame.getStackItem(2));
      if (value > 0) {
        return true;
      }
    }

    return frame.isStatic() && opCodeData.stackSettings().forbiddenInStatic();
  }

  private static boolean isOutOfSStore(MessageFrame frame, OpCode opCode) {
    return opCode == OpCode.SSTORE && frame.getRemainingGas() <= GasConstants.G_CALL_STIPEND.cost();
  }

  private static boolean isInvalidCodePrefix(MessageFrame frame) {
    if (frame.getType() != MessageFrame.Type.CONTRACT_CREATION) {
      return false;
    }

    final Bytes deployedCode = frame.getOutputData();
    return !deployedCode.isEmpty() && (deployedCode.get(0) == (byte) EIP_3541_MARKER);
  }

  private static boolean isCodeSizeOverflow(MessageFrame frame) {
    if (frame.getType() != MessageFrame.Type.CONTRACT_CREATION) {
      return false;
    }

    // TODO: don't get it from getOutputData, but only when OPCODE == RETURN && read in memory
    final Bytes deployedCode = frame.getOutputData();
    return deployedCode.size() > MAX_CODE_SIZE;
  }

  /**
   * Return the first exception that may have happened in the current frame. Although multiple
   * exceptions may be triggered, the one minimizing the quantity of trace lines is generated.
   *
   * @param frame the context from which to compute the putative exceptions
   */
  public static short fromFrame(final Hub hub, final MessageFrame frame) {
    OpCode opCode = hub.opCode();
    OpCodeData opCodeData = hub.opCodeData();

    if (isInvalidOpcode(opCode)) {
      return INVALID_OPCODE;
    }
    if (isStackUnderflow(frame, opCodeData)) {
      return STACK_UNDERFLOW;
    }
    if (isStackOverflow(frame, opCodeData)) {
      return STACK_OVERFLOW;
    }
    if (isStaticFault(frame, opCodeData)) {
      return STATIC_FAULT;
    }
    if (isCodeSizeOverflow(frame)) {
      return CODE_SIZE_OVERFLOW;
    }

    final GasProjector gp = Hub.GAS_PROJECTOR;
    switch (opCode) {
      case CALLDATACOPY,
          CODECOPY,
          EXTCODECOPY,
          LOG0,
          LOG1,
          LOG2,
          LOG3,
          LOG4,
          SHA3,
          CREATE,
          CREATE2,
          CALL,
          DELEGATECALL,
          STATICCALL,
          CALLCODE,
          RETURN,
          REVERT,
          CALLDATALOAD,
          MLOAD,
          MSTORE,
          MSTORE8 -> {
        if (isMemoryExpansionFault(frame, opCode, gp)) {
          return OUT_OF_MEMORY_EXPANSION;
        }
        if (isOutOfGas(frame, opCode, gp)) {
          return OUT_OF_GAS;
        }
      }

      case RETURNDATACOPY -> {
        if (isReturnDataCopyFault(frame, opCode)) {
          return RETURN_DATA_COPY_FAULT;
        }
        if (isMemoryExpansionFault(frame, opCode, gp)) {
          return OUT_OF_MEMORY_EXPANSION;
        }
        if (isOutOfGas(frame, opCode, gp)) {
          return OUT_OF_GAS;
        }
      }

      case STOP -> {}

      case JUMP, JUMPI -> {
        if (isOutOfGas(frame, opCode, gp)) {
          return OUT_OF_GAS;
        }
        if (isJumpFault(frame, opCode)) {
          return JUMP_FAULT;
        }
      }

      case SSTORE -> {
        if (isOutOfSStore(frame, opCode)) {
          return OUT_OF_SSTORE;
        }
        if (isOutOfGas(frame, opCode, gp)) {
          return OUT_OF_GAS;
        }
      }

      default -> {
        if (isOutOfGas(frame, opCode, gp)) {
          return OUT_OF_GAS;
        }
      }
    }

    if (isInvalidCodePrefix(frame)) {
      return INVALID_CODE_PREFIX;
    }
    return NONE;
  }
}
