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

package net.consensys.linea.zktracer.module.hub;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjector;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

/** Encode the exceptions that may be triggered byt the execution of an instruction. */
@Getter
@RequiredArgsConstructor
@Accessors(fluent = true)
public final class Exceptions {
  private static final byte EIP_3541_MARKER = (byte) 0xEF;
  public static final int MAX_CODE_SIZE = 24576;

  private final Hub hub;

  private boolean invalidOpcode;
  private boolean stackUnderflow;
  private boolean stackOverflow;
  private boolean outOfMemoryExpansion;
  private boolean outOfGas;
  private boolean returnDataCopyFault;
  private boolean jumpFault;
  private boolean staticFault;
  private boolean outOfSStore;
  private boolean invalidCodePrefix;
  private boolean codeSizeOverflow;

  /**
   * @param invalidOpcode unknown opcode
   * @param stackUnderflow stack underflow
   * @param stackOverflow stack overflow
   * @param outOfMemoryExpansion tried to use memory too far away
   * @param outOfGas not enough gas for instruction
   * @param returnDataCopyFault trying to read pas the RETURNDATA end
   * @param jumpFault jumping to an invalid destination
   * @param staticFault trying to execute a non-static instruction in a static context
   * @param outOfSStore not enough gas to execute an SSTORE
   */
  public Exceptions(
      boolean invalidOpcode,
      boolean stackUnderflow,
      boolean stackOverflow,
      boolean outOfMemoryExpansion,
      boolean outOfGas,
      boolean returnDataCopyFault,
      boolean jumpFault,
      boolean staticFault,
      boolean outOfSStore,
      boolean invalidCodePrefix,
      boolean codeSizeOverflow) {
    this.hub = null;
    this.invalidOpcode = invalidOpcode;
    this.stackUnderflow = stackUnderflow;
    this.stackOverflow = stackOverflow;
    this.outOfMemoryExpansion = outOfMemoryExpansion;
    this.outOfGas = outOfGas;
    this.returnDataCopyFault = returnDataCopyFault;
    this.jumpFault = jumpFault;
    this.staticFault = staticFault;
    this.outOfSStore = outOfSStore;
    this.invalidCodePrefix = invalidCodePrefix;
    this.codeSizeOverflow = codeSizeOverflow;
  }

  @Override
  public String toString() {
    if (this.invalidOpcode) {
      return "Invalid opcode";
    }
    if (this.stackUnderflow) {
      return "Stack underflow";
    }
    if (this.stackOverflow) {
      return "Stack overflow";
    }
    if (this.outOfMemoryExpansion) {
      return "Out of MXP";
    }
    if (this.outOfGas) {
      return "Out of gas";
    }
    if (this.returnDataCopyFault) {
      return "RDC fault";
    }
    if (this.jumpFault) {
      return "JMP fault";
    }
    if (this.staticFault) {
      return "Static fault";
    }
    if (this.outOfSStore) {
      return "Out of SSTORE";
    }
    if (this.invalidCodePrefix) {
      return "Invalid code prefix";
    }
    if (this.codeSizeOverflow) {
      return "Code size overflow";
    }
    return "No exception";
  }

  public void reset() {
    this.invalidOpcode = false;
    this.stackUnderflow = false;
    this.stackOverflow = false;
    this.outOfMemoryExpansion = false;
    this.outOfGas = false;
    this.returnDataCopyFault = false;
    this.jumpFault = false;
    this.staticFault = false;
    this.outOfSStore = false;
    this.invalidCodePrefix = false;
    this.codeSizeOverflow = false;
  }

  public boolean stackException() {
    return this.stackUnderflow() || this.stackOverflow();
  }
  /**
   * @return true if no stack exception has been raised
   */
  public boolean noStackException() {
    return !this.stackException();
  }

  /**
   * Creates a snapshot (a copy) of the given Exceptions object.
   *
   * @return a new Exceptions
   */
  public Exceptions snapshot() {
    return new Exceptions(
        invalidOpcode,
        stackUnderflow,
        stackOverflow,
        outOfMemoryExpansion,
        outOfGas,
        returnDataCopyFault,
        jumpFault,
        staticFault,
        outOfSStore,
        invalidCodePrefix,
        codeSizeOverflow);
  }

  /**
   * @return true if any exception flag has been raised
   */
  public boolean any() {
    return this.invalidOpcode
        || this.stackUnderflow
        || this.stackOverflow
        || this.outOfMemoryExpansion
        || this.outOfGas
        || this.returnDataCopyFault
        || this.jumpFault
        || this.staticFault
        || this.outOfSStore
        || this.invalidCodePrefix
        || this.codeSizeOverflow;
  }

  /**
   * @return true if no exception flag has been raised
   */
  public boolean none() {
    return !this.any();
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
    return !deployedCode.isEmpty() && (deployedCode.get(0) == EIP_3541_MARKER);
  }

  private static boolean isCodeSizeOverflow(MessageFrame frame) {
    if (frame.getType() != MessageFrame.Type.CONTRACT_CREATION) {
      return false;
    }

    // TODO: don't get it from getOutputData, but only when OPCODE == RETURN && read in memory
    final Bytes deployedCode = frame.getOutputData();
    return deployedCode.size() > MAX_CODE_SIZE;
  }

  public static Exceptions fromOutOfGas() {
    return new Exceptions(
        false, false, false, false, true, false, false, false, false, false, false);
  }

  /**
   * Compute all the first exception that may have happened in the current frame. Wlthout multiple
   * exceptions may be triggered, the one minimizing the quantity of trace lines is generated.
   *
   * @param frame the context from which to compute the putative exceptions
   */
  public void prepare(final MessageFrame frame, GasProjector gp) {
    OpCode opCode = hub.opCode();
    OpCodeData opCodeData = hub.opCodeData();

    this.reset();

    this.invalidOpcode = isInvalidOpcode(opCode);
    if (this.invalidOpcode) {
      return;
    }

    this.stackUnderflow = isStackUnderflow(frame, opCodeData);
    if (this.stackUnderflow) {
      return;
    }

    this.stackOverflow = isStackOverflow(frame, opCodeData);
    if (this.stackOverflow) {
      return;
    }

    this.staticFault = isStaticFault(frame, opCodeData);
    if (this.staticFault) {
      return;
    }

    this.codeSizeOverflow = isCodeSizeOverflow(frame);
    if (this.codeSizeOverflow) {
      return;
    }

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
        this.outOfMemoryExpansion = isMemoryExpansionFault(frame, opCode, gp);
        if (this.outOfMemoryExpansion) {
          return;
        }

        this.outOfGas = isOutOfGas(frame, opCode, gp);
        if (this.outOfGas) {
          return;
        }
      }
      case RETURNDATACOPY -> {
        this.returnDataCopyFault = isReturnDataCopyFault(frame, opCode);
        if (this.returnDataCopyFault) {
          return;
        }

        this.outOfMemoryExpansion = isMemoryExpansionFault(frame, opCode, gp);
        if (this.outOfMemoryExpansion) {
          return;
        }

        this.outOfGas = isOutOfGas(frame, opCode, gp);
        if (this.outOfGas) {
          return;
        }
      }
      case STOP -> {}
      case JUMP, JUMPI -> {
        this.outOfGas = isOutOfGas(frame, opCode, gp);
        if (this.outOfGas) {
          return;
        }

        this.jumpFault = isJumpFault(frame, opCode);
        if (this.jumpFault) {
          return;
        }
      }
      case SSTORE -> {
        this.outOfSStore = isOutOfSStore(frame, opCode);
        if (this.outOfSStore) {
          return;
        }

        this.outOfGas = isOutOfGas(frame, opCode, gp);
        if (this.outOfGas) {
          return;
        }
      }
      default -> {
        this.outOfGas = isOutOfGas(frame, opCode, gp);
        if (this.outOfGas) {
          return;
        }
      }
    }

    this.invalidCodePrefix = isInvalidCodePrefix(frame);
  }
}
