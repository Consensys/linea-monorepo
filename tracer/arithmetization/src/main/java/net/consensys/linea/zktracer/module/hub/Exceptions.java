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

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjector;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

/**
 * Encode the exceptions that mey be triggered byt the execution of an instruction.
 *
 * @param invalidOpcode unknown opcode
 * @param stackUnderflow stack underflow
 * @param stackOverflow stack overflow
 * @param outOfMemoryExpansion tried to use memory too far away
 * @param outOfGas not enough gas for instruction
 * @param returnDataCopyFault
 * @param jumpFault jumping to an invalid destination
 * @param staticViolation trying to execute a non-static instruction in a static context
 * @param outOfSStore
 */
public record Exceptions(
    boolean invalidOpcode,
    boolean stackUnderflow,
    boolean stackOverflow,
    boolean outOfMemoryExpansion,
    boolean outOfGas,
    boolean returnDataCopyFault,
    boolean jumpFault,
    boolean staticViolation,
    boolean outOfSStore) {
  /**
   * @return true if no stack exception has been raised
   */
  public boolean noStackException() {
    return !this.stackOverflow() && !this.stackUnderflow();
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
        staticViolation,
        outOfSStore);
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
        || this.staticViolation
        || this.outOfSStore;
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

  private static boolean isReturnDataCopyFault(final MessageFrame frame) {
    if (OpCode.of(frame.getCurrentOperation().getOpcode()) == OpCode.RETURNDATACOPY) {
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

  private static boolean isStaticFault(final MessageFrame frame) {
    final OpCodeData opCode = OpCode.of(frame.getCurrentOperation().getOpcode()).getData();
    return frame.isStatic() && opCode.stackSettings().forbiddenInStatic();
  }

  private static boolean isOutOfSStore(MessageFrame frame, OpCode opCode) {
    return opCode == OpCode.SSTORE && frame.getRemainingGas() <= GasConstants.G_CALL_STIPEND.cost();
  }

  /**
   * Compute all the exceptions that may have happened in the current frame and package them in an
   * {@link Exceptions} record.
   *
   * @param frame the context from which to compute the putative exceptions
   * @return all {@link Exceptions} relative to the given frame
   */
  public static Exceptions forFrame(final MessageFrame frame, GasProjector gp) {
    OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    OpCodeData opCodeData = opCode.getData();

    final boolean invalidOpcode = isInvalidOpcode(opCode);
    final boolean stackUnderflow = invalidOpcode ? false : isStackUnderflow(frame, opCodeData);
    final boolean stackOverflow = invalidOpcode ? false : isStackOverflow(frame, opCodeData);
    final boolean outOfMxp =
        (stackUnderflow || stackOverflow) ? false : isMemoryExpansionFault(frame, opCode, gp);
    final boolean oufOfGas =
        (stackUnderflow || stackOverflow) ? false : isOutOfGas(frame, opCode, gp);
    final boolean returnDataCopyFault =
        (stackUnderflow || stackOverflow) ? false : isReturnDataCopyFault(frame);
    final boolean jumpFault =
        (stackUnderflow || stackOverflow) ? false : isJumpFault(frame, opCode);

    return new Exceptions(
        invalidOpcode,
        stackUnderflow,
        stackOverflow,
        outOfMxp,
        oufOfGas,
        returnDataCopyFault,
        jumpFault,
        isStaticFault(frame),
        isOutOfSStore(frame, opCode));
  }

  public static Exceptions empty() {
    return new Exceptions(false, false, false, false, false, false, false, false, false);
  }

  /**
   * Generate the exceptions for a transaction whose execution was skipped from the beginning.
   * Should map to an OoG? TODO: cf. @Olivier
   *
   * @return an Exceptions encoding an out of gas
   */
  public static Exceptions fromOutOfGas() {
    return new Exceptions(false, false, false, false, true, false, false, false, false);
  }
}
