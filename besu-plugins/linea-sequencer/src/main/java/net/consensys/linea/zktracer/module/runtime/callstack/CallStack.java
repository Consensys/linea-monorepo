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

package net.consensys.linea.zktracer.module.runtime.callstack;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import com.google.common.base.Preconditions;
import lombok.Getter;
import net.consensys.linea.zktracer.module.hub.Bytecode;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.memory.MemorySpan;
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
public final class CallStack {
  /** the maximal depth of the call stack (as defined by Ethereum) */
  static final int MAX_CALLSTACK_SIZE = 1024;
  /** a never-pruned-tree of the {@link CallFrame} executed by the {@link Hub} */
  private final List<CallFrame> frames = new ArrayList<>();
  /** the current depth of the call stack. */
  @Getter private int depth;
  /** a "pointer" to the current {@link CallFrame} in <code>frames</code>. */
  private int current;

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
    this.depth = 0;
    this.frames.add(new CallFrame(from));
    this.enter(
        hubStamp,
        to,
        to,
        toCode == null ? Bytecode.EMPTY : toCode,
        type,
        value,
        gas,
        callData,
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

  public int futureId() {
    return this.frames.size();
  }

  /**
   * @return the parent {@link CallFrame} of the current frame
   */
  public CallFrame parent() {
    return this.frames.get(this.current().parentFrame());
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
      int accountDeploymentNumber,
      int codeDeploymentNumber,
      boolean isDeployment) {
    final int caller = this.current;
    final int newTop = this.frames.size();
    Bytes callData;
    if (type != CallFrameType.INIT_CODE) {
      callData = input;
    } else {
      callData = Bytes.EMPTY;
    }

    this.depth += 1;
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
            this.depth);

    this.frames.add(newFrame);
    this.current = newTop;
    this.frames.get(caller).childFrames().add(newTop);
  }

  /**
   * Exit the current context, sets it return data for the caller to read, and marks its last
   * position in the hub traces.
   *
   * @param currentLine the current line in the hub trace
   * @param returnData the return data of the current frame
   */
  public void exit(int currentLine, Bytes returnData) {
    this.depth -= 1;
    Preconditions.checkState(this.depth >= 0);
    this.current().returnDataPointer(new MemorySpan(0, 0)); // TODO: fix me Franklin
    final int parent = this.current().parentFrame();
    this.frames.get(parent).childFrames().add(this.current);
    this.frames.get(parent).returnData(returnData);
    this.current = parent;
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
  public CallFrame get(int i) {
    // The case where the CF #0 is called on an empty stack stems from a skipped transaction, where
    // no CF of interest is available to trace.
    // TODO: use an explicit -1 as marker
    if (i == 0 && this.frames.isEmpty()) {
      return CallFrame.EMPTY;
    }
    return this.frames.get(i);
  }

  /**
   * Returns the parent of the ith {@link CallFrame} in this call stack.
   *
   * @param i ID of the call frame whose parent to fetch
   * @return the ith call frame parent
   * @throws IndexOutOfBoundsException if the index is out of range
   */
  public CallFrame getParentOf(int i) {
    return this.get(this.frames.get(i).parentFrame());
  }

  public void revert(int stamp) {
    this.current().revert(this, stamp);
  }
}
