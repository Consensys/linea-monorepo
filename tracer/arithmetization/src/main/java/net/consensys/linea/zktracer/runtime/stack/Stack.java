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

package net.consensys.linea.zktracer.runtime.stack;

import lombok.Getter;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class Stack {
  public static final int MAX_STACK_SIZE = 1024;

  @Getter int height;
  @Getter int heightNew;
  @Getter OpCodeData currentOpcodeData;
  Status status;
  int stamp;

  public Stack() {
    this.height = 0;
    this.heightNew = 0;
    this.status = Status.NORMAL;
  }

  public Stack snapshot() {
    var r = new Stack();
    r.height = this.height;
    r.heightNew = this.heightNew;
    r.currentOpcodeData = this.currentOpcodeData;
    r.status = this.status;

    return r;
  }

  private int stackStampWithOffset(int offset) {
    return this.stamp + offset;
  }

  private Bytes getStack(MessageFrame frame, int i) {
    return frame.getStackItem(i);
  }

  private void oneZero(MessageFrame frame, StackContext pending) {
    Bytes val = getStack(frame, 0);
    pending.addLine(
        new IndexedStackOperation(
            1, StackOperation.pop(this.height, val, stackStampWithOffset(0))));
  }

  private void twoZero(MessageFrame frame, StackContext pending) {
    Bytes val1 = getStack(frame, 0);
    Bytes val2 = getStack(frame, 1);

    pending.addLine(
        new IndexedStackOperation(
            1, StackOperation.pop(this.height, val1, stackStampWithOffset(0))),
        new IndexedStackOperation(
            2, StackOperation.pop(this.height, val2, stackStampWithOffset(1))));
  }

  private void zeroOne(MessageFrame ignoredFrame, StackContext pending) {
    pending.addArmingLine(
        new IndexedStackOperation(
            4, StackOperation.push(this.height + 1, stackStampWithOffset(0))));
  }

  private void oneOne(MessageFrame frame, StackContext pending) {
    Bytes val = getStack(frame, 0);

    pending.addArmingLine(
        new IndexedStackOperation(1, StackOperation.pop(this.height, val, stackStampWithOffset(0))),
        new IndexedStackOperation(4, StackOperation.push(this.height, stackStampWithOffset(1))));
  }

  private void twoOne(MessageFrame frame, StackContext pending) {
    Bytes val1 = getStack(frame, 0);
    Bytes val2 = getStack(frame, 1);

    pending.addArmingLine(
        new IndexedStackOperation(
            1, StackOperation.pop(this.height, val1, stackStampWithOffset(0))),
        new IndexedStackOperation(
            2, StackOperation.pop(this.height - 1, val2, stackStampWithOffset(1))),
        new IndexedStackOperation(
            4, StackOperation.push(this.height - 1, stackStampWithOffset(2))));
  }

  private void threeOne(MessageFrame frame, StackContext pending) {
    Bytes val1 = getStack(frame, 0);
    Bytes val2 = getStack(frame, 1);
    Bytes val3 = getStack(frame, 2);

    pending.addArmingLine(
        new IndexedStackOperation(
            1, StackOperation.pop(this.height, val1, stackStampWithOffset(0))),
        new IndexedStackOperation(
            2, StackOperation.pop(this.height - 1, val2, stackStampWithOffset(1))),
        new IndexedStackOperation(
            3, StackOperation.pop(this.height - 2, val3, stackStampWithOffset(2))),
        new IndexedStackOperation(
            4, StackOperation.push(this.height - 2, stackStampWithOffset(3))));
  }

  private void loadStore(MessageFrame frame, StackContext pending) {
    if (this.currentOpcodeData.mnemonic().isAnyOf(OpCode.MSTORE, OpCode.MSTORE8, OpCode.SSTORE)) {
      Bytes val1 = getStack(frame, 0);
      Bytes val2 = getStack(frame, 1);

      pending.addLine(
          new IndexedStackOperation(
              1, StackOperation.pop(this.height, val1, stackStampWithOffset(0))),
          new IndexedStackOperation(
              4, StackOperation.pop(this.height - 1, val2, stackStampWithOffset(1))));
    } else {
      Bytes val = getStack(frame, 0);

      pending.addArmingLine(
          new IndexedStackOperation(
              1, StackOperation.pop(this.height, val, stackStampWithOffset(0))),
          new IndexedStackOperation(4, StackOperation.push(this.height, stackStampWithOffset(1))));
    }
  }

  private void dup(MessageFrame frame, StackContext pending) {
    int depth = this.currentOpcodeData.stackSettings().delta() - 1;
    Bytes val = getStack(frame, depth);

    pending.addLine(
        new IndexedStackOperation(
            1, StackOperation.pop(this.height - depth, val, stackStampWithOffset(0))),
        new IndexedStackOperation(
            2, StackOperation.pushImmediate(this.height - depth, val, stackStampWithOffset(1))),
        new IndexedStackOperation(
            4, StackOperation.pushImmediate(this.height + 1, val, stackStampWithOffset(2))));
  }

  private void swap(MessageFrame frame, StackContext pending) {
    int depth = this.currentOpcodeData.stackSettings().delta() - 1;
    Bytes val1 = getStack(frame, 0);
    Bytes val2 = getStack(frame, depth);

    pending.addLine(
        new IndexedStackOperation(
            1, StackOperation.pop(this.height - depth, val1, stackStampWithOffset(0))),
        new IndexedStackOperation(
            2, StackOperation.pop(this.height, val2, stackStampWithOffset(1))),
        new IndexedStackOperation(
            3, StackOperation.pushImmediate(this.height - depth, val2, stackStampWithOffset(2))),
        new IndexedStackOperation(
            4, StackOperation.pushImmediate(this.height, val1, stackStampWithOffset(3))));
  }

  private void log(MessageFrame frame, StackContext pending) {
    Bytes offset = getStack(frame, 0);
    Bytes size = getStack(frame, 1);

    // Stack line 1
    pending.addLine(
        new IndexedStackOperation(
            1, StackOperation.pop(this.height, offset, stackStampWithOffset(0))),
        new IndexedStackOperation(
            2, StackOperation.pop(this.height - 1, size, stackStampWithOffset(1))));

    // Stack line 2
    IndexedStackOperation[] line2 = new IndexedStackOperation[] {};
    switch (this.currentOpcodeData.mnemonic()) {
      case LOG0 -> {}
      case LOG1 -> {
        Bytes topic1 = getStack(frame, 2);

        line2 =
            new IndexedStackOperation[] {
              new IndexedStackOperation(
                  1, StackOperation.pop(this.height - 2, topic1, stackStampWithOffset(0))),
            };
      }
      case LOG2 -> {
        Bytes topic1 = getStack(frame, 2);
        Bytes topic2 = getStack(frame, 3);

        line2 =
            new IndexedStackOperation[] {
              new IndexedStackOperation(
                  1, StackOperation.pop(this.height - 2, topic1, stackStampWithOffset(2))),
              new IndexedStackOperation(
                  2, StackOperation.pop(this.height - 3, topic2, stackStampWithOffset(3))),
            };
      }
      case LOG3 -> {
        Bytes topic1 = getStack(frame, 2);
        Bytes topic2 = getStack(frame, 3);
        Bytes topic3 = getStack(frame, 4);

        line2 =
            new IndexedStackOperation[] {
              new IndexedStackOperation(
                  1, StackOperation.pop(this.height - 2, topic1, stackStampWithOffset(2))),
              new IndexedStackOperation(
                  2, StackOperation.pop(this.height - 3, topic2, stackStampWithOffset(3))),
              new IndexedStackOperation(
                  3, StackOperation.pop(this.height - 4, topic3, stackStampWithOffset(4))),
            };
      }
      case LOG4 -> {
        Bytes topic1 = getStack(frame, 2);
        Bytes topic2 = getStack(frame, 3);
        Bytes topic3 = getStack(frame, 4);
        Bytes topic4 = getStack(frame, 5);

        line2 =
            new IndexedStackOperation[] {
              new IndexedStackOperation(
                  1, StackOperation.pop(this.height - 2, topic1, stackStampWithOffset(2))),
              new IndexedStackOperation(
                  2, StackOperation.pop(this.height - 3, topic2, stackStampWithOffset(3))),
              new IndexedStackOperation(
                  3, StackOperation.pop(this.height - 4, topic3, stackStampWithOffset(4))),
              new IndexedStackOperation(
                  4, StackOperation.pop(this.height - 5, topic4, stackStampWithOffset(5))),
            };
      }
      default -> throw new RuntimeException("not a LOGx");
    }
    pending.addLine(line2);
  }

  private void copy(MessageFrame frame, StackContext pending) {
    if (this.currentOpcodeData.stackSettings().addressTrimmingInstruction()) {
      Bytes val0 = getStack(frame, 0);
      Bytes val1 = getStack(frame, 1);
      Bytes val2 = getStack(frame, 2);
      Bytes val3 = getStack(frame, 3);

      pending.addLine(
          new IndexedStackOperation(
              1, StackOperation.pop(this.height - 1, val1, stackStampWithOffset(1))),
          new IndexedStackOperation(
              2, StackOperation.pop(this.height - 3, val3, stackStampWithOffset(2))),
          new IndexedStackOperation(
              3, StackOperation.pop(this.height - 2, val2, stackStampWithOffset(3))),
          new IndexedStackOperation(4, StackOperation.pop(this.height, val0, this.stamp)));
    } else {
      Bytes val1 = getStack(frame, 0);
      Bytes val2 = getStack(frame, 2);
      Bytes val3 = getStack(frame, 1);

      pending.addLine(
          new IndexedStackOperation(
              1, StackOperation.pop(this.height, val1, stackStampWithOffset(1))),
          new IndexedStackOperation(
              2, StackOperation.pop(this.height - 2, val2, stackStampWithOffset(2))),
          new IndexedStackOperation(
              3, StackOperation.pop(this.height - 1, val3, stackStampWithOffset(3))));
    }
  }

  private void call(MessageFrame frame, StackContext pending) {
    Bytes val1 = getStack(frame, 0);
    Bytes val2 = getStack(frame, 1);
    Bytes val3 = getStack(frame, 2);
    Bytes val4 = getStack(frame, 3);
    Bytes val5 = getStack(frame, 4);
    Bytes val6 = getStack(frame, 5);

    boolean callCanTransferValue = this.currentOpcodeData.mnemonic().callCanTransferValue();

    if (callCanTransferValue) {
      Bytes val7 = getStack(frame, 6);

      pending.addLine(
          new IndexedStackOperation(
              1, StackOperation.pop(this.height - 3, val4, stackStampWithOffset(3))),
          new IndexedStackOperation(
              2, StackOperation.pop(this.height - 4, val5, stackStampWithOffset(4))),
          new IndexedStackOperation(
              3, StackOperation.pop(this.height - 5, val6, stackStampWithOffset(5))),
          new IndexedStackOperation(
              4, StackOperation.pop(this.height - 6, val7, stackStampWithOffset(6))));
      pending.addArmingLine(
          new IndexedStackOperation(
              1, StackOperation.pop(this.height, val1, stackStampWithOffset(0))),
          new IndexedStackOperation(
              2, StackOperation.pop(this.height - 1, val2, stackStampWithOffset(1))),
          new IndexedStackOperation(
              3, StackOperation.pop(this.height - 2, val3, stackStampWithOffset(2))),
          new IndexedStackOperation(
              4, StackOperation.push(this.height - 6, stackStampWithOffset(7))));
    } else {

      pending.addLine(
          new IndexedStackOperation(
              1, StackOperation.pop(this.height - 2, val3, stackStampWithOffset(3))),
          new IndexedStackOperation(
              2, StackOperation.pop(this.height - 3, val4, stackStampWithOffset(4))),
          new IndexedStackOperation(
              3, StackOperation.pop(this.height - 4, val5, stackStampWithOffset(5))),
          new IndexedStackOperation(
              4, StackOperation.pop(this.height - 5, val6, stackStampWithOffset(6))));

      pending.addArmingLine(
          new IndexedStackOperation(
              1, StackOperation.pop(this.height, val1, stackStampWithOffset(0))),
          new IndexedStackOperation(
              2, StackOperation.pop(this.height - 1, val2, stackStampWithOffset(1))),
          new IndexedStackOperation(
              4, StackOperation.push(this.height - 5, stackStampWithOffset(7))));
    }
  }

  private void create(MessageFrame frame, StackContext pending) {
    final Bytes val1 = getStack(frame, 1);
    final Bytes val2 = getStack(frame, 2);

    pending.addLine(
        new IndexedStackOperation(
            1, StackOperation.pop(this.height - 1, val1, stackStampWithOffset(1))),
        new IndexedStackOperation(
            2, StackOperation.pop(this.height - 2, val2, stackStampWithOffset(2))));
    // case CREATE2
    if (this.currentOpcodeData.stackSettings().flag2()) {
      final Bytes val3 = getStack(frame, 3);
      final Bytes val4 = getStack(frame, 0);

      pending.addArmingLine(
          new IndexedStackOperation(
              2, StackOperation.pop(this.height - 3, val3, stackStampWithOffset(3))),
          new IndexedStackOperation(
              3, StackOperation.pop(this.height, val4, stackStampWithOffset(0))),
          new IndexedStackOperation(
              4, StackOperation.push(this.height - 3, stackStampWithOffset(4))));
    } else
    // case CREATE
    {
      final Bytes val4 = getStack(frame, 0);

      pending.addArmingLine(
          new IndexedStackOperation(
              3, StackOperation.pop(this.height, val4, stackStampWithOffset(0))),
          new IndexedStackOperation(
              4, StackOperation.push(this.height - 2, stackStampWithOffset(4))));
    }
  }

  /**
   * @return true if no stack exception has been raised
   */
  public boolean isOk() {
    return this.status == Status.NORMAL;
  }

  /**
   * @return true if a stack underflow exception has been raised
   */
  public boolean isUnderflow() {
    return this.status == Status.UNDERFLOW;
  }

  /**
   * @return true if a stack underflow exception has been raised
   */
  public boolean isOverflow() {
    return this.status == Status.OVERFLOW;
  }

  public void processInstruction(final Hub hub, MessageFrame frame, int stackStamp) {
    final CallFrame callFrame = hub.currentFrame();
    this.stamp = stackStamp;
    this.height = this.heightNew;
    this.currentOpcodeData = hub.opCodeData();
    callFrame.pending(new StackContext(this.currentOpcodeData.mnemonic()));

    final int alpha = this.currentOpcodeData.stackSettings().alpha();
    final int delta = this.currentOpcodeData.stackSettings().delta();

    this.heightNew += this.currentOpcodeData.stackSettings().nbAdded();
    this.heightNew -= this.currentOpcodeData.stackSettings().nbRemoved();

    if (frame.stackSize() < delta) { // Testing for underflow
      this.status = Status.UNDERFLOW;
    } else if (this.heightNew > MAX_STACK_SIZE) { // Testing for overflow
      this.status = Status.OVERFLOW;
    }

    // CALL WCP for the SUX/SOX lookup
    hub.wcp().callLT(this.height, delta);
    if (!this.isUnderflow()) {
      hub.wcp().callGT(this.height - delta + alpha, MAX_STACK_SIZE);
    }

    if (this.status.isFailure()) {
      this.heightNew = 0;

      if (this.currentOpcodeData.stackSettings().twoLineInstruction()) {
        this.stamp += callFrame.pending().addEmptyLines(2);
      } else {
        this.stamp += callFrame.pending().addEmptyLines(1);
      }

      return;
    }

    switch (this.currentOpcodeData.stackSettings().pattern()) {
      case ZERO_ZERO -> this.stamp += callFrame.pending().addEmptyLines(1);
      case ONE_ZERO -> this.oneZero(frame, callFrame.pending());
      case TWO_ZERO -> this.twoZero(frame, callFrame.pending());
      case ZERO_ONE -> this.zeroOne(frame, callFrame.pending());
      case ONE_ONE -> this.oneOne(frame, callFrame.pending());
      case TWO_ONE -> this.twoOne(frame, callFrame.pending());
      case THREE_ONE -> this.threeOne(frame, callFrame.pending());
      case LOAD_STORE -> this.loadStore(frame, callFrame.pending());
      case DUP -> this.dup(frame, callFrame.pending());
      case SWAP -> this.swap(frame, callFrame.pending());
      case LOG -> this.log(frame, callFrame.pending());
      case COPY -> this.copy(frame, callFrame.pending());
      case CALL -> this.call(frame, callFrame.pending());
      case CREATE -> this.create(frame, callFrame.pending());
    }
  }
}
