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

package net.consensys.linea.zktracer.module.hub.transients;

import static net.consensys.linea.zktracer.module.UtilCalculator.allButOneSixtyFourth;
import static net.consensys.linea.zktracer.types.AddressUtils.isPrecompile;

import com.google.common.base.Preconditions;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.module.constants.GlobalConstants;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

/** This class provides facilities to access data that are opcode-lived. */
@Slf4j
@RequiredArgsConstructor
public class OperationAncillaries {
  private final Hub hub;

  private static Bytes maybeShadowReadMemory(final MemorySpan span, final MessageFrame frame) {
    // Accesses to huge offset with 0-length are valid
    if (span.isEmpty()) {
      return Bytes.EMPTY;
    }

    // Besu is limited to i32 for memory offset/length
    if (span.besuOverflow()) {
      log.warn("Overflowing memory access: {}", span);
      return Bytes.EMPTY;
    }

    return frame.shadowReadMemory(span.offset(), span.length());
  }

  /**
   * Compute the gas allowance for the child context if in a CALL, throws otherwise.
   *
   * @return the CALL gas allowance
   */
  public long gasAllowanceForCall() {
    final OpCode opCode = hub.opCode();

    if (opCode.isCall()) {
      final long gas = Words.clampedToLong(hub.messageFrame().getStackItem(0));
      EWord value = EWord.ZERO;
      if (opCode == OpCode.CALL || opCode == OpCode.CALLCODE) {
        value = EWord.of(hub.messageFrame().getStackItem(2));
      }
      final long stipend = value.isZero() ? 0 : GlobalConstants.GAS_CONST_G_CALL_STIPEND;
      final long upfrontCost = Hub.GAS_PROJECTOR.of(hub.messageFrame(), opCode).total();
      return stipend
          + Math.max(
              Words.unsignedMin(
                  allButOneSixtyFourth(hub.messageFrame().getRemainingGas() - upfrontCost), gas),
              0);
    }

    throw new IllegalStateException("not a CALL");
  }

  /**
   * Returns the RAM segment of the caller containing the calldata if the {@link MessageFrame}
   * operation is a call, throws otherwise.
   *
   * @param frame the execution context
   * @return the input data segment
   */
  public static MemorySpan callDataSegment(final MessageFrame frame) {
    switch (OpCode.of(frame.getCurrentOperation().getOpcode())) {
      case CALL, CALLCODE -> {
        long offset = Words.clampedToLong(frame.getStackItem(3));
        long length = Words.clampedToLong(frame.getStackItem(4));
        return MemorySpan.fromStartLength(offset, length);
      }
      case DELEGATECALL, STATICCALL -> {
        long offset = Words.clampedToLong(frame.getStackItem(2));
        long length = Words.clampedToLong(frame.getStackItem(3));
        return MemorySpan.fromStartLength(offset, length);
      }
      default -> throw new IllegalArgumentException(
          "callDataSegment called outside of a CALL-type instruction");
    }
  }

  public static MemorySpan initCodeSegment(final MessageFrame frame) {
    switch (OpCode.of(frame.getCurrentOperation().getOpcode())) {
      case CREATE, CREATE2 -> {
        long offset = Words.clampedToLong(frame.getStackItem(1));
        long length = Words.clampedToLong(frame.getStackItem(2));
        return MemorySpan.fromStartLength(offset, length);
      }
      default -> throw new IllegalArgumentException(
          "callDataSegment called outside of a CREATE(2)");
    }
  }

  /**
   * Returns the RAM segment of the caller containing the calldata if the current operation is a
   * call, throws otherwise.
   *
   * @return the input data segment
   */
  public MemorySpan callDataSegment() {
    return callDataSegment(hub.messageFrame());
  }

  public MemorySpan initCodeSegment() {
    return initCodeSegment(hub.messageFrame());
  }

  /**
   * Return the bytes of the calldata if the current operation is a call, throws otherwise.
   *
   * @return the calldata content
   */
  public Bytes callData() {
    final MemorySpan callDataSegment = callDataSegment();
    return maybeShadowReadMemory(callDataSegment, hub.messageFrame());
  }

  /**
   * Return the bytes of the calldata if the current operation is a call, throws otherwise.
   *
   * @param frame the execution context
   * @return the calldata content
   */
  public static Bytes callData(final MessageFrame frame) {
    final MemorySpan callDataSegment = callDataSegment(frame);
    return maybeShadowReadMemory(callDataSegment, frame);
  }

  public static Bytes initCode(final MessageFrame frame) {
    final MemorySpan initCodeSegment = initCodeSegment(frame);
    return maybeShadowReadMemory(initCodeSegment, frame);
  }

  /**
   * Returns the RAM segment offered by the caller for the return data if the current operation is a
   * call, throws otherwise.
   *
   * @param frame the execution context
   * @return the return data target
   */
  public static MemorySpan returnDataRequestedSegment(final MessageFrame frame) {
    switch (OpCode.of(frame.getCurrentOperation().getOpcode())) {
      case CALL, CALLCODE -> {
        long offset = Words.clampedToLong(frame.getStackItem(5));
        long length = Words.clampedToLong(frame.getStackItem(6));
        return MemorySpan.fromStartLength(offset, length);
      }
      case DELEGATECALL, STATICCALL -> {
        long offset = Words.clampedToLong(frame.getStackItem(4));
        long length = Words.clampedToLong(frame.getStackItem(5));
        return MemorySpan.fromStartLength(offset, length);
      }
      default -> throw new IllegalArgumentException(
          "returnDataRequestedSegment called outside of a *CALL");
    }
  }

  /**
   * Returns the RAM segment offered by the caller for the return data if the current operation is a
   * call, throws otherwise.
   *
   * @return the return data target
   */
  public MemorySpan returnDataRequestedSegment() {
    return returnDataRequestedSegment(hub.messageFrame());
  }

  /**
   * Returns the RAM segment offered by the callee for the return data if the current operation is a
   * RETURN/REVERT, throws otherwise.
   *
   * @param frame the execution context
   * @return the return data segment
   */
  public static MemorySpan outputDataSpan(final MessageFrame frame) {

    if (frame.getExceptionalHaltReason().isPresent()) {
      return MemorySpan.empty();
    }

    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    if (opCode == OpCode.RETURN && frame.getType() == MessageFrame.Type.CONTRACT_CREATION) {
      return MemorySpan.empty();
    }

    switch (opCode) {
      case RETURN, REVERT -> {
        long offset = Words.clampedToLong(frame.getStackItem(0));
        long length = Words.clampedToLong(frame.getStackItem(1));
        return MemorySpan.fromStartLength(offset, length);
      }
      case STOP, SELFDESTRUCT -> {
        return MemorySpan.empty();
      }

        // TODO: what the case below provides isn't output data, but the return data ...
        //  We cannot use this method for that purpose.
      case CALL, CALLCODE, DELEGATECALL, STATICCALL -> {
        Address target = Words.toAddress(frame.getStackItem(1));
        if (isPrecompile(target)) {
          return MemorySpan.fromStartLength(0, 0);
        }
        Preconditions.checkArgument(isPrecompile(target)); // useless (?) sanity check
        // TODO: this will not work for MODEXP as return data starts at offset
        //  512 - modulusByteSize
        return MemorySpan.fromStartLength(0, frame.getReturnData().size());
      }
      default -> throw new IllegalArgumentException(
          "returnDataRequestedSegment called outside of a RETURN/REVERT");
    }
  }

  /**
   * Returns the RAM segment offered by the caller for the return data if the current operation is a
   * call, throws otherwise.
   *
   * @return the return data target
   */
  public MemorySpan outputDataSpan() {
    return outputDataSpan(hub.messageFrame());
  }

  /**
   * Return the bytes of the return data if the current operation is a call, throws otherwise.
   *
   * @return the return data content
   */
  public Bytes outputData() {
    final MemorySpan outputDataSpan = outputDataSpan();

    // Accesses to huge offset with 0-length are valid
    if (outputDataSpan.isEmpty()) {
      return Bytes.EMPTY;
    }

    // Besu is limited to i32 for memory offset/length
    if (outputDataSpan.besuOverflow()) {
      log.warn("Overflowing memory access: {}", outputDataSpan);
      return Bytes.EMPTY;
    }

    // TODO: this WON'T work for precompiles, they don't have memory.
    return maybeShadowReadMemory(outputDataSpan, hub.messageFrame());
  }

  /**
   * Return the bytes of the returndata if the current operation is a return/revert, throws
   * otherwise.
   *
   * @param frame the execution context
   * @return the returndata content
   */
  public static Bytes outputData(final MessageFrame frame) {
    final MemorySpan returnDataSegment = outputDataSpan(frame);
    return maybeShadowReadMemory(returnDataSegment, frame);
  }

  public static MemorySpan logDataSegment(final MessageFrame frame) {
    long offset = Words.clampedToLong(frame.getStackItem(0));
    long length = Words.clampedToLong(frame.getStackItem(1));
    return MemorySpan.fromStartLength(offset, length);
  }

  public MemorySpan logDataSegment() {
    return logDataSegment(this.hub.messageFrame());
  }

  public static Bytes logData(final MessageFrame frame) {
    return maybeShadowReadMemory(logDataSegment(frame), frame);
  }

  public Bytes logData() {
    return logData(this.hub.messageFrame());
  }
}
