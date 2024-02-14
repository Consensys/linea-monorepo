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

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

/** This class provides facilities to access data that are opcode-lived. */
@RequiredArgsConstructor
public class Operation {
  private final Hub hub;

  /**
   * Compute the gas allowance for the child context if in a CALL, throws otherwise.
   *
   * @return the CALL gas allowance
   */
  public long gasAllowanceForCall() {
    final OpCode opCode = hub.opCode();

    switch (opCode) {
      case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
        final long gas = Words.clampedToLong(hub.messageFrame().getStackItem(0));
        EWord value = EWord.ZERO;
        if (opCode == OpCode.CALL || opCode == OpCode.CALLCODE) {
          value = EWord.of(hub.messageFrame().getStackItem(2));
        }
        final long stipend = value.isZero() ? 0 : GasConstants.G_CALL_STIPEND.cost();
        final long upfrontCost = Hub.gp.of(hub.messageFrame(), opCode).total();
        return stipend
            + Math.max(
                Words.unsignedMin(
                    allButOneSixtyFourth(hub.messageFrame().getRemainingGas() - upfrontCost), gas),
                0);
      }
      default -> throw new IllegalStateException("not a CALL");
    }
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
      case CREATE, CREATE2 -> {
        long offset = Words.clampedToLong(frame.getStackItem(1));
        long length = Words.clampedToLong(frame.getStackItem(2));
        return MemorySpan.fromStartLength(offset, length);
      }
      default -> throw new IllegalArgumentException("callDataSegment called outside of a *CALL");
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

  /**
   * Return the bytes of the calldata if the current operation is a call, throws otherwise.
   *
   * @return the calldata content
   */
  public Bytes callData() {
    final MemorySpan callDataSegment = callDataSegment();
    return hub.messageFrame().shadowReadMemory(callDataSegment.offset(), callDataSegment.length());
  }

  /**
   * Return the bytes of the calldata if the current operation is a call, throws otherwise.
   *
   * @param frame the execution context
   * @return the calldata content
   */
  public static Bytes callData(final MessageFrame frame) {
    final MemorySpan callDataSegment = callDataSegment(frame);
    return frame.shadowReadMemory(callDataSegment.offset(), callDataSegment.length());
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
  public static MemorySpan returnDataSegment(final MessageFrame frame) {
    switch (OpCode.of(frame.getCurrentOperation().getOpcode())) {
      case RETURN, REVERT -> {
        long offset = Words.clampedToLong(frame.getStackItem(0));
        long length = Words.clampedToLong(frame.getStackItem(1));
        return MemorySpan.fromStartLength(offset, length);
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
  public MemorySpan returnDataSegment() {
    return returnDataSegment(hub.messageFrame());
  }

  /**
   * Return the bytes of the calldata if the current operation is a call, throws otherwise.
   *
   * @return the calldata content
   */
  public Bytes returnData() {
    final MemorySpan returnDataSegment = returnDataSegment();
    return hub.messageFrame()
        .shadowReadMemory(returnDataSegment.offset(), returnDataSegment.length());
  }

  /**
   * Return the bytes of the returndata if the current operation is a return/revert, throws
   * otherwise.
   *
   * @param frame the execution context
   * @return the returndata content
   */
  public static Bytes returnData(final MessageFrame frame) {
    final MemorySpan returnDataSegment = returnDataSegment(frame);
    return frame.shadowReadMemory(returnDataSegment.offset(), returnDataSegment.length());
  }
}
