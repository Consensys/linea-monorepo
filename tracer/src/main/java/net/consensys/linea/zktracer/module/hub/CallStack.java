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

package net.consensys.linea.zktracer.module.hub;

import java.util.List;

import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.Code;

class CallStack {
  /** the maximal depth of the call stack (as defined by Ethereum) */
  static final int CALLSTACK_SIZE = 1024;
  /** a never-pruned-tree of the {@link CallFrame} executed by the {@link Hub} */
  private List<CallFrame> frames;
  /** the current depth of the call stack */
  private int depth;
  /** a "pointer" to the current {@link CallFrame} in <code>frames</code> */
  private int current;

  CallStack() {
    this.frames = List.of(new CallFrame());
    this.current = 0;
    this.depth = 0;
  }

  CallFrame top() {
    return this.frames.get(this.current);
  }

  CallFrame enter(
      Address address,
      Code code,
      CallFrameType type,
      Wei value,
      long gas,
      int currentLine,
      Bytes input,
      int maxContextNumber,
      int deploymentNumber) {
    final int caller = this.current;
    final int newTop = this.frames.size();
    Bytes callData;
    if (type != CallFrameType.InitCode) {
      callData = input;
    } else {
      callData = Bytes.EMPTY;
    }

    CallFrame newFrame =
        new CallFrame(
            maxContextNumber,
            deploymentNumber,
            newTop,
            address,
            code,
            type,
            caller,
            this.frames.get(caller).address,
            value,
            gas,
            currentLine,
            callData);

    this.frames.add(newFrame);
    this.current = newTop;
    this.depth += 1;
    this.frames.get(caller).childFrames.add(newTop);

    return this.top();
  }

  /**
   * Exit the current context, sets it return data for the caller to read, and marks its last
   * position in the hub traces
   *
   * @param currentLine the current line in the hub trace
   * @param returnData the return data of the current frame
   */
  void exit(int currentLine, Bytes returnData) {
    this.depth -= 1;
    assert this.depth >= 0;

    this.top().close(currentLine);
    final int parent = this.top().parentFrame;
    this.frames.get(parent).lastCalled = this.current;
    this.frames.get(parent).returnData = returnData;
    this.current = parent;
  }

  /**
   * @return whether the call stack is in an overflow state
   */
  boolean isOverflow() {
    return this.depth > CALLSTACK_SIZE;
  }

  /**
   * @return whether the current frame is a static context
   */
  boolean isStatic() {
    return this.top().type == CallFrameType.Static;
  }

  /**
   * Get the caller of the current frame
   *
   * @return the caller of the current frame
   */
  CallFrame caller() {
    return this.frames.get(this.top().parentFrame);
  }

  /**
   * Get the last frame called by the current frame
   *
   * @return the latest child
   */
  CallFrame callee() {
    return this.frames.get(this.top().lastCalled);
  }
}
