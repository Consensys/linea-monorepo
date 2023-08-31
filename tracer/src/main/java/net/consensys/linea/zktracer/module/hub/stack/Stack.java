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

package net.consensys.linea.zktracer.module.hub.stack;

import lombok.Getter;
import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.hub.callstack.CallFrame;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
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

  private EWord getStack(MessageFrame frame, int i) {
    return EWord.of(frame.getStackItem(i));
  }

  private void oneZero(MessageFrame frame, StackContext pending) {
    EWord val = getStack(frame, 0);
    pending.addLine(new IndexedStackOperation(1, StackOperation.pop(this.height, val, this.stamp)));
  }

  private void twoZero(MessageFrame frame, StackContext pending) {
    EWord val1 = getStack(frame, 0);
    EWord val2 = getStack(frame, 1);

    pending.addLine(
        new IndexedStackOperation(1, StackOperation.pop(this.height, val1, this.stamp)),
        new IndexedStackOperation(2, StackOperation.pop(this.height, val2, this.stamp)));
  }

  private void zeroOne(MessageFrame ignoredFrame, StackContext pending) {
    pending.addArmingLine(
        new IndexedStackOperation(4, StackOperation.push(this.height + 1, this.stamp)));
  }

  private void oneOne(MessageFrame frame, StackContext pending) {
    EWord val = getStack(frame, 0);

    pending.addArmingLine(
        new IndexedStackOperation(1, StackOperation.pop(this.height, val, this.stamp)),
        new IndexedStackOperation(4, StackOperation.push(this.height, this.stamp + 1)));
  }

  private void twoOne(MessageFrame frame, StackContext pending) {
    EWord val1 = getStack(frame, 0);
    EWord val2 = getStack(frame, 1);

    pending.addArmingLine(
        new IndexedStackOperation(1, StackOperation.pop(this.height, val1, this.stamp)),
        new IndexedStackOperation(2, StackOperation.pop(this.height - 1, val2, this.stamp + 1)),
        new IndexedStackOperation(4, StackOperation.push(this.height - 1, this.stamp + 2)));
  }

  private void threeOne(MessageFrame frame, StackContext pending) {
    EWord val1 = getStack(frame, 0);
    EWord val2 = getStack(frame, 1);
    EWord val3 = getStack(frame, 2);

    pending.addArmingLine(
        new IndexedStackOperation(1, StackOperation.pop(this.height, val1, this.stamp)),
        new IndexedStackOperation(2, StackOperation.pop(this.height - 1, val2, this.stamp + 1)),
        new IndexedStackOperation(3, StackOperation.pop(this.height - 2, val3, this.stamp + 2)),
        new IndexedStackOperation(4, StackOperation.push(this.height - 2, this.stamp + 3)));
  }

  private void loadStore(MessageFrame frame, StackContext pending) {
    if (this.currentOpcodeData.stackSettings().flag1()) {
      EWord val1 = getStack(frame, 0);
      EWord val2 = getStack(frame, 1);

      pending.addLine(
          new IndexedStackOperation(1, StackOperation.pop(this.height, val1, this.stamp)),
          new IndexedStackOperation(4, StackOperation.pop(this.height - 1, val2, this.stamp + 1)));
    } else {
      EWord val = getStack(frame, 0);

      pending.addArmingLine(
          new IndexedStackOperation(1, StackOperation.pop(this.height, val, this.stamp)),
          new IndexedStackOperation(4, StackOperation.push(this.height, this.stamp + 1)));
    }
  }

  private void dup(MessageFrame frame, StackContext pending) {
    int depth = this.currentOpcodeData.stackSettings().delta() - 1;
    EWord val = getStack(frame, depth);

    pending.addLine(
        new IndexedStackOperation(1, StackOperation.pop(this.height - depth, val, this.stamp)),
        new IndexedStackOperation(
            2, StackOperation.pushImmediate(this.height - depth, val, this.stamp + 1)),
        new IndexedStackOperation(
            4, StackOperation.pushImmediate(this.height + 1, val, this.stamp + 2)));
  }

  private void swap(MessageFrame frame, StackContext pending) {
    int depth = this.currentOpcodeData.stackSettings().delta() - 1;
    EWord val1 = getStack(frame, 0);
    EWord val2 = getStack(frame, depth);

    pending.addLine(
        new IndexedStackOperation(1, StackOperation.pop(this.height - depth, val1, this.stamp)),
        new IndexedStackOperation(2, StackOperation.pop(this.height, val2, this.stamp + 1)),
        new IndexedStackOperation(
            3, StackOperation.pushImmediate(this.height - depth, val2, this.stamp + 2)),
        new IndexedStackOperation(
            4, StackOperation.pushImmediate(this.height, val1, this.stamp + 3)));
  }

  private void log(MessageFrame frame, StackContext pending) {
    EWord offset = getStack(frame, 0);
    EWord size = getStack(frame, 1);

    // Stack line 1
    pending.addLine(
        new IndexedStackOperation(1, StackOperation.pop(this.height, offset, this.stamp)),
        new IndexedStackOperation(2, StackOperation.pop(this.height - 1, size, this.stamp + 1)));

    // Stack line 2
    IndexedStackOperation[] line2 = new IndexedStackOperation[] {};
    switch (this.currentOpcodeData.mnemonic()) {
      case LOG0 -> {}
      case LOG1 -> {
        EWord topic1 = getStack(frame, 2);

        line2 =
            new IndexedStackOperation[] {
              new IndexedStackOperation(
                  1, StackOperation.pop(this.height - 2, topic1, this.stamp + 2)),
            };
      }
      case LOG2 -> {
        EWord topic1 = getStack(frame, 2);
        EWord topic2 = getStack(frame, 3);

        line2 =
            new IndexedStackOperation[] {
              new IndexedStackOperation(
                  1, StackOperation.pop(this.height - 2, topic1, this.stamp + 2)),
              new IndexedStackOperation(
                  2, StackOperation.pop(this.height - 3, topic2, this.stamp + 3)),
            };
      }
      case LOG3 -> {
        EWord topic1 = getStack(frame, 2);
        EWord topic2 = getStack(frame, 3);
        EWord topic3 = getStack(frame, 4);

        line2 =
            new IndexedStackOperation[] {
              new IndexedStackOperation(
                  1, StackOperation.pop(this.height - 2, topic1, this.stamp + 2)),
              new IndexedStackOperation(
                  2, StackOperation.pop(this.height - 3, topic2, this.stamp + 3)),
              new IndexedStackOperation(
                  3, StackOperation.pop(this.height - 4, topic3, this.stamp + 4)),
            };
      }
      case LOG4 -> {
        EWord topic1 = getStack(frame, 2);
        EWord topic2 = getStack(frame, 3);
        EWord topic3 = getStack(frame, 4);
        EWord topic4 = getStack(frame, 5);

        line2 =
            new IndexedStackOperation[] {
              new IndexedStackOperation(
                  1, StackOperation.pop(this.height - 2, topic1, this.stamp + 2)),
              new IndexedStackOperation(
                  2, StackOperation.pop(this.height - 3, topic2, this.stamp + 3)),
              new IndexedStackOperation(
                  3, StackOperation.pop(this.height - 4, topic3, this.stamp + 4)),
              new IndexedStackOperation(
                  4, StackOperation.pop(this.height - 5, topic4, this.stamp + 5)),
            };
      }
      default -> throw new RuntimeException("not a LOGx");
    }
    pending.addLine(line2);
  }

  private void copy(MessageFrame frame, StackContext pending) {
    if (this.currentOpcodeData.stackSettings().addressTrimmingInstruction()) {
      EWord val0 = getStack(frame, 0);
      EWord val1 = getStack(frame, 1);
      EWord val2 = getStack(frame, 2);
      EWord val3 = getStack(frame, 3);

      pending.addLine(
          new IndexedStackOperation(1, StackOperation.pop(this.height - 1, val1, this.stamp + 1)),
          new IndexedStackOperation(2, StackOperation.pop(this.height - 3, val3, this.stamp + 2)),
          new IndexedStackOperation(3, StackOperation.pop(this.height - 2, val2, this.stamp + 3)),
          new IndexedStackOperation(4, StackOperation.pop(this.height, val0, this.stamp)));
    } else {
      EWord val1 = getStack(frame, 0);
      EWord val2 = getStack(frame, 2);
      EWord val3 = getStack(frame, 1);

      pending.addLine(
          new IndexedStackOperation(1, StackOperation.pop(this.height, val1, this.stamp + 1)),
          new IndexedStackOperation(2, StackOperation.pop(this.height - 2, val2, this.stamp + 2)),
          new IndexedStackOperation(3, StackOperation.pop(this.height - 1, val3, this.stamp + 3)));
    }
  }

  private void call(MessageFrame frame, StackContext pending) {
    EWord val1 = getStack(frame, 0);
    EWord val2 = getStack(frame, 1);
    EWord val3 = getStack(frame, 2);
    EWord val4 = getStack(frame, 3);
    EWord val5 = getStack(frame, 4);
    EWord val6 = getStack(frame, 5);

    boolean sevenItems =
        this.currentOpcodeData.stackSettings().flag1()
            || this.currentOpcodeData.stackSettings().flag2();
    if (sevenItems) {
      EWord val7 = getStack(frame, 6);

      pending.addLine(
          new IndexedStackOperation(1, StackOperation.pop(this.height - 3, val4, this.stamp + 3)),
          new IndexedStackOperation(2, StackOperation.pop(this.height - 4, val5, this.stamp + 4)),
          new IndexedStackOperation(3, StackOperation.pop(this.height - 5, val6, this.stamp + 5)),
          new IndexedStackOperation(4, StackOperation.pop(this.height - 6, val7, this.stamp + 6)));
      pending.addArmingLine(
          new IndexedStackOperation(1, StackOperation.pop(this.height, val1, this.stamp)),
          new IndexedStackOperation(2, StackOperation.pop(this.height - 1, val2, this.stamp + 1)),
          new IndexedStackOperation(3, StackOperation.pop(this.height - 2, val3, this.stamp + 2)),
          new IndexedStackOperation(4, StackOperation.push(this.height - 6, this.stamp + 6)));
    } else {

      pending.addLine(
          new IndexedStackOperation(1, StackOperation.pop(this.height - 2, val3, this.stamp + 3)),
          new IndexedStackOperation(2, StackOperation.pop(this.height - 3, val4, this.stamp + 4)),
          new IndexedStackOperation(3, StackOperation.pop(this.height - 4, val5, this.stamp + 5)),
          new IndexedStackOperation(4, StackOperation.pop(this.height - 5, val6, this.stamp + 6)));

      pending.addArmingLine(
          new IndexedStackOperation(1, StackOperation.pop(this.height, val1, this.stamp)),
          new IndexedStackOperation(2, StackOperation.pop(this.height - 1, val2, this.stamp + 1)),
          new IndexedStackOperation(4, StackOperation.push(this.height - 5, this.stamp + 7)));
    }
  }

  private void create(MessageFrame frame, StackContext pending) {
    EWord val1 = getStack(frame, 1);
    EWord val2 = getStack(frame, 2);

    pending.addLine(
        new IndexedStackOperation(1, StackOperation.pop(this.height - 1, val1, this.stamp + 1)),
        new IndexedStackOperation(2, StackOperation.pop(this.height - 2, val2, this.stamp + 2)));
    if (this.currentOpcodeData.stackSettings().flag1()) {
      EWord val3 = getStack(frame, 3);
      EWord val4 = getStack(frame, 0);

      pending.addArmingLine(
          new IndexedStackOperation(2, StackOperation.pop(this.height - 3, val3, this.stamp + 3)),
          new IndexedStackOperation(3, StackOperation.pop(this.height, val4, this.stamp)),
          new IndexedStackOperation(4, StackOperation.push(this.height - 3, this.stamp + 4)));
    } else {
      EWord val4 = getStack(frame, 0);

      pending.addArmingLine(
          new IndexedStackOperation(3, StackOperation.pop(this.height, val4, this.stamp)),
          new IndexedStackOperation(4, StackOperation.push(this.height - 2, this.stamp + 4)));
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

  public boolean processInstruction(MessageFrame frame, CallFrame callFrame, int stackStamp) {
    this.stamp = stackStamp;
    this.height = this.heightNew;
    this.currentOpcodeData = OpCode.of(frame.getCurrentOperation().getOpcode()).getData();
    callFrame.setPending(new StackContext(this.currentOpcodeData.mnemonic()));

    assert this.height == frame.stackSize();
    this.heightNew += this.currentOpcodeData.stackSettings().nbAdded();
    this.heightNew -= this.currentOpcodeData.stackSettings().nbRemoved();

    if (frame.stackSize()
        < this.currentOpcodeData.stackSettings().delta()) { // Testing for underflow
      this.status = Status.UNDERFLOW;
    } else if (this.heightNew > MAX_STACK_SIZE) { // Testing for overflow
      this.status = Status.OVERFLOW;
    }

    if (this.status.isFailure()) {
      this.heightNew = 0;

      if (this.currentOpcodeData.stackSettings().twoLinesInstruction()) {
        this.stamp += callFrame.getPending().addEmptyLines(2);
      } else {
        this.stamp += callFrame.getPending().addEmptyLines(1);
      }

      return false;
    }

    switch (this.currentOpcodeData.stackSettings().pattern()) {
      case ZERO_ZERO -> this.stamp += callFrame.getPending().addEmptyLines(1);
      case ONE_ZERO -> this.oneZero(frame, callFrame.getPending());
      case TWO_ZERO -> this.twoZero(frame, callFrame.getPending());
      case ZERO_ONE -> this.zeroOne(frame, callFrame.getPending());
      case ONE_ONE -> this.oneOne(frame, callFrame.getPending());
      case TWO_ONE -> this.twoOne(frame, callFrame.getPending());
      case THREE_ONE -> this.threeOne(frame, callFrame.getPending());
      case LOAD_STORE -> this.loadStore(frame, callFrame.getPending());
      case DUP -> this.dup(frame, callFrame.getPending());
      case SWAP -> this.swap(frame, callFrame.getPending());
      case LOG -> this.log(frame, callFrame.getPending());
      case COPY -> this.copy(frame, callFrame.getPending());
      case CALL -> this.call(frame, callFrame.getPending());
      case CREATE -> this.create(frame, callFrame.getPending());
    }

    return this.status == Status.NORMAL;
  }
}
