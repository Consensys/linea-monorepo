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

package net.consensys.linea.zktracer.runtime.callstack;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.types.Bytecode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.Code;

/**
 * This class represents the call hierarchy of a transaction.
 *
 * <p>Although it is accessible in a stack-like manner, it is actually a tree, the stack access
 * representing the path from the latest leaf to the root context.
 */
@Accessors(fluent = true)
public final class CallStack {
  /** the maximal depth of the call stack (as defined by Ethereum) */
  static final int MAX_CALLSTACK_SIZE = 1024;

  /** a never-pruned-tree of the {@link CallFrame} executed by the {@link Hub} */
  @Getter
  private final List<CallFrame> callFrames =
      new ArrayList<>(
          50) { // TODO: PERF as the List of TraceSection, we should have an estimate based on
        // gasLimit on the nb of CallFrame a tx might have
        {
          add(CallFrame.EMPTY);
        }
      };

  /** the current depth of the call stack. */
  @Getter private int depth;

  /** a "pointer" to the current {@link CallFrame} in <code>frames</code>. */
  private int current;

  public void newPrecompileResult(
      final int hubStamp,
      final Bytes precompileResult,
      final int returnDataOffset,
      final Address precompileAddress) {

    final CallFrame newFrame =
        new CallFrame(
            -1,
            -1,
            false,
            this.callFrames.size(),
            hubStamp,
            precompileAddress,
            precompileAddress,
            precompileAddress,
            Bytecode.EMPTY,
            CallFrameType.PRECOMPILE_RETURN_DATA,
            this.current,
            Wei.ZERO,
            0,
            precompileResult,
            returnDataOffset,
            precompileResult.size(),
            -1,
            this.depth);

    this.callFrames.add(newFrame);
  }

  public void newBedrock(
      int hubStamp,
      Address from,
      Address to,
      CallFrameType type,
      Bytecode toCode,
      Wei value,
      long gas,
      Bytes callData,
      int accountDeploymentNumber,
      int codeDeploymentNumber,
      boolean codeDeploymentStatus) {
    this.depth = -1;
    this.enter(
        hubStamp,
        to,
        to,
        from,
        toCode == null ? Bytecode.EMPTY : toCode,
        type,
        value,
        gas,
        callData,
        0,
        callData.size(),
        hubStamp,
        accountDeploymentNumber,
        codeDeploymentNumber,
        codeDeploymentStatus);
    this.current = this.callFrames.size() - 1;
  }

  /**
   * A “mantle” {@link CallFrame} holds the call data for a message call with a non-empty call data
   *
   * @param hubStamp
   * @param from
   * @param to
   * @param type
   * @param toCode
   * @param value
   * @param gas
   * @param callData
   * @param accountDeploymentNumber
   * @param codeDeploymentNumber
   * @param codeDeploymentStatus
   */
  public void newMantleAndBedrock(
      int hubStamp,
      Address from,
      Address to,
      CallFrameType type,
      Bytecode toCode,
      Wei value,
      long gas,
      Bytes callData,
      int accountDeploymentNumber,
      int codeDeploymentNumber,
      boolean codeDeploymentStatus) {
    this.depth = -1;
    this.callFrames.add(new CallFrame(from, callData, hubStamp));
    this.enter(
        hubStamp,
        to,
        to,
        from,
        toCode == null ? Bytecode.EMPTY : toCode,
        CallFrameType.ROOT,
        value,
        gas,
        callData,
        0,
        callData.size(),
        hubStamp,
        accountDeploymentNumber,
        codeDeploymentNumber,
        codeDeploymentStatus);
    this.current = this.callFrames.size() - 1;
  }

  /**
   * @return the currently executing {@link CallFrame}
   */
  public CallFrame current() {
    return this.callFrames.get(this.current);
  }

  public boolean isEmpty() {
    return this.callFrames.isEmpty();
  }

  public int futureId() {
    return this.callFrames.size();
  }

  /**
   * @return the parent {@link CallFrame} of the current frame
   */
  public CallFrame parent() {
    if (this.current().parentFrameId() != -1) {
      return this.callFrames.get(this.current().parentFrameId());
    } else {
      return CallFrame.EMPTY;
    }
  }

  public Optional<CallFrame> maybeCurrent() {
    return this.callFrames.isEmpty() ? Optional.empty() : Optional.of(this.current());
  }

  /**
   * Creates a new call frame.
   *
   * @param hubStamp the hub stamp at the time of entry in the new frame
   * @param accountAddress the {@link Address} of the bytecode being executed
   * @param code the {@link Code} being executed
   * @param type the execution type of call frame
   * @param value the value given to this call frame
   * @param gas the gas provided to this call frame
   * @param input the call data sent to this call frame
   * @param accountDeploymentNumber
   * @param codeDeploymentNumber
   * @param isDeployment
   */
  public void enter(
      int hubStamp,
      Address accountAddress,
      Address byteCodeAddress,
      Address callerAddress,
      Bytecode code,
      CallFrameType type,
      Wei value,
      long gas,
      Bytes input,
      long callDataOffset,
      long callDataSize,
      long callDataContextNumber,
      int accountDeploymentNumber,
      int codeDeploymentNumber,
      boolean isDeployment) {
    final int caller = this.depth == -1 ? -1 : this.current;
    final int newTop = this.callFrames.size();
    this.depth += 1;

    Bytes callData = Bytes.EMPTY;
    if (type != CallFrameType.INIT_CODE) {
      callData = input;
    }
    final CallFrame newFrame =
        new CallFrame(
            accountDeploymentNumber,
            codeDeploymentNumber,
            isDeployment,
            newTop,
            hubStamp,
            accountAddress,
            callerAddress,
            byteCodeAddress,
            code,
            type,
            caller,
            value,
            gas,
            callData,
            callDataOffset,
            callDataSize,
            callDataContextNumber,
            this.depth);

    this.callFrames.add(newFrame);
    this.current = newTop;
    if (caller != -1) {
      this.callFrames.get(caller).returnData(Bytes.EMPTY);
      this.callFrames.get(caller).childFramesId().add(newTop);
    }
  }

  /**
   * Exit the current context, sets it return data for the caller to read, and marks its last
   * position in the hub traces.
   */
  public void exit() {
    this.depth -= 1;
    Preconditions.checkState(this.depth >= 0);
    this.current = this.current().parentFrameId();
  }

  /**
   * @return whether the call stack is in an overflow state
   */
  public boolean isOverflow() {
    return this.depth > MAX_CALLSTACK_SIZE;
  }

  /**
   * @return whether the call stack is at its maximum capacity and a new frame would overflow it
   */
  public boolean wouldOverflow() {
    return this.depth >= MAX_CALLSTACK_SIZE;
  }

  /**
   * @return whether the current frame is a static context
   */
  public boolean isStatic() {
    return this.current().type() == CallFrameType.STATIC;
  }

  /**
   * Get the {@link CallFrame} representing the caller of the current frame
   *
   * @return the caller of the current frame
   */
  public CallFrame caller() {
    return this.callFrames.get(this.current().parentFrameId());
  }

  /**
   * Returns the ith {@link CallFrame} in this call stack.
   *
   * @param i ID of the call frame to fetch
   * @return the ith call frame
   * @throws IndexOutOfBoundsException if the index is out of range
   */
  public CallFrame getById(int i) {
    if (i < 0 || this.callFrames.isEmpty()) {
      return CallFrame.EMPTY;
    }
    return this.callFrames.get(i);
  }

  /**
   * Returns the {@link CallFrame} in this call stack of the given context number.
   *
   * @param i context number of the call frame to fetch
   * @return the call frame with the specifies
   * @throws IndexOutOfBoundsException if the index is out of range
   */
  public CallFrame getByContextNumber(final long i) {
    for (CallFrame f : this.callFrames) {
      if (f.contextNumber() == i) {
        return f;
      }
    }

    throw new IllegalArgumentException(String.format("call frame CN %s not found", i));
  }

  /**
   * Returns the parent of the ith {@link CallFrame} in this call stack.
   *
   * @param id ID of the call frame whose parent to fetch
   * @return the ith call frame parent
   * @throws IndexOutOfBoundsException if the index is out of range
   */
  public CallFrame getParentCallFrameById(int id) {
    if (this.callFrames.isEmpty()) {
      return CallFrame.EMPTY;
    }

    return this.getById(this.callFrames.get(id).parentFrameId());
  }

  /**
   * Retrieves the context number of the parent {@link CallFrame} for a given call frame ID.
   *
   * @param id the ID of the call frame whose parent's context number is to be retrieved.
   * @return the context number of the parent call frame. If the call frame has no parent, or if the
   *     specified ID does not correspond to a valid call frame, this method returns the context
   *     number of the {@link CallFrame#EMPTY} which is typically 0.
   */
  public int getParentContextNumberById(int id) {
    return this.getParentCallFrameById(id).contextNumber();
  }

  public void revert(int stamp) {
    this.current().revert(this, stamp);
  }
}
