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
  private final List<CallFrame> frames = new ArrayList<>();

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
            this.frames.size(),
            hubStamp,
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

    this.frames.add(newFrame);
  }

  public void newBedrock(
      int hubStamp,
      //      Address from,
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
    this.current = this.frames.size() - 1;
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
    this.frames.add(new CallFrame(callData, hubStamp));
    this.enter(
        hubStamp,
        to,
        to,
        toCode == null ? Bytecode.EMPTY : toCode,
        CallFrameType.BEDROCK,
        value,
        gas,
        callData,
        0,
        callData.size(),
        hubStamp,
        accountDeploymentNumber,
        codeDeploymentNumber,
        codeDeploymentStatus);
    this.current = this.frames.size() - 1;
  }

  /**
   * @return the currently executing {@link CallFrame}
   */
  public CallFrame current() {
    return this.frames.get(this.current);
  }

  public boolean isEmpty() {
    return this.frames.isEmpty();
  }

  public int futureId() {
    return this.frames.size();
  }

  /**
   * @return the parent {@link CallFrame} of the current frame
   */
  public CallFrame parent() {
    if (this.current().parentFrame() != -1) {
      return this.frames.get(this.current().parentFrame());
    } else {
      return CallFrame.EMPTY;
    }
  }

  public Optional<CallFrame> maybeCurrent() {
    return this.frames.isEmpty() ? Optional.empty() : Optional.of(this.current());
  }

  /**
   * Creates a new call frame.
   *
   * @param hubStamp the hub stamp at the time of entry in the new frame
   * @param address the {@link Address} of the bytecode being executed
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
      Address address,
      Address codeAddress,
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
    final int newTop = this.frames.size();
    this.depth += 1;

    Bytes callData = Bytes.EMPTY;
    if (type != CallFrameType.INIT_CODE) {
      callData = input;
    }

    CallFrame newFrame =
        new CallFrame(
            accountDeploymentNumber,
            codeDeploymentNumber,
            isDeployment,
            newTop,
            hubStamp,
            address,
            codeAddress,
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

    this.frames.add(newFrame);
    this.current = newTop;
    if (caller != -1) {
      this.frames.get(caller).latestReturnData(Bytes.EMPTY);
      this.frames.get(caller).childFrames().add(newTop);
    }
  }

  /**
   * Exit the current context, sets it return data for the caller to read, and marks its last
   * position in the hub traces.
   */
  public void exit() {
    this.depth -= 1;
    Preconditions.checkState(this.depth >= 0);
    this.current = this.current().parentFrame();
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
    return this.frames.get(this.current().parentFrame());
  }

  /**
   * Returns the ith {@link CallFrame} in this call stack.
   *
   * @param i ID of the call frame to fetch
   * @return the ith call frame
   * @throws IndexOutOfBoundsException if the index is out of range
   */
  public CallFrame getById(int i) {
    if (i < 0 || this.frames.isEmpty()) {
      return CallFrame.EMPTY;
    }
    return this.frames.get(i);
  }

  /**
   * Returns the {@link CallFrame} in this call stack of the given context number.
   *
   * @param i context number of the call frame to fetch
   * @return the call frame with the specifies
   * @throws IndexOutOfBoundsException if the index is out of range
   */
  public CallFrame getByContextNumber(final long i) {
    for (CallFrame f : this.frames) {
      if (f.contextNumber() == i) {
        return f;
      }
    }

    throw new IllegalArgumentException(String.format("call frame CN %s not found", i));
  }

  /**
   * Returns the parent of the ith {@link CallFrame} in this call stack.
   *
   * @param i ID of the call frame whose parent to fetch
   * @return the ith call frame parent
   * @throws IndexOutOfBoundsException if the index is out of range
   */
  public CallFrame getParentOf(int i) {
    if (this.frames.isEmpty()) {
      return CallFrame.EMPTY;
    }

    return this.getById(this.frames.get(i).parentFrame());
  }

  public void revert(int stamp) {
    this.current().revert(this, stamp);
  }

  public String pretty() {
    StringBuilder r = new StringBuilder(2000);
    for (CallFrame c : this.frames) {
      final CallFrame parent = this.getParentOf(c.id());
      r.append(" ".repeat(c.depth()));
      r.append(
          "%d/%d (<- %d/%d): %s"
              .formatted(c.id(), c.contextNumber(), parent.id(), parent.contextNumber(), c.type()));
    }
    return r.toString();
  }
}
