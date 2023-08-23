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

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.opcode.OpCode;

enum Action {
  NONE,
  PUSH,
  POP
}

/**
 * An atomic operation (read/pop or write/push) on the stack, indexed within a {@link
 * IndexedStackOperation}.
 *
 * <p>Alterations of the stack by an EVM instruction are decomposed in one or two chunks, or {@link
 * StackLine}, made of {@link IndexedStackOperation}, which are then stored, with the associated
 * metadata, in a {@link StackContext}.
 */
final class StackOperation {
  /**
   * the relative height of the element with regard to the stack height just before executing the
   * linked EVM instruction
   */
  private final int height;
  /** the value having been popped from/pushed on the stack */
  private EWord value;
  /** whether this action is a push or a pop */
  private final Action action;
  /**
   * the stamp of this operation relative to the stack stamp before executing the linked EVM
   * instruction
   */
  private final int stackStamp;

  StackOperation() {
    this.height = 0;
    this.value = EWord.ZERO;
    this.action = Action.NONE;
    this.stackStamp = 0;
  }

  StackOperation(int height, EWord value, Action action, int stackStamp) {
    this.height = height;
    this.value = value;
    this.action = action;
    this.stackStamp = stackStamp;
  }

  public static StackOperation pop(int height, EWord value, int stackStamp) {
    return new StackOperation(height, value, Action.POP, stackStamp);
  }

  public static StackOperation push(int height, int stackStamp) {
    return new StackOperation(
        height,
        EWord.of(0xDEADBEEFL) /* marker value, erased on unlatching */,
        Action.PUSH,
        stackStamp);
  }

  public static StackOperation pushImmediate(int height, EWord val, int stackStamp) {
    return new StackOperation(height, val.copy(), Action.PUSH, stackStamp);
  }

  public void setValue(EWord x) {
    this.value = x;
  }

  public int height() {
    return height;
  }

  public EWord value() {
    return value;
  }

  public Action action() {
    return action;
  }

  public int stackStamp() {
    return stackStamp;
  }

  @Override
  public String toString() {
    return "StackOperation["
        + "height="
        + height
        + ", "
        + "value="
        + value
        + ", "
        + "action="
        + action
        + ", "
        + "stackStamp="
        + stackStamp
        + ']';
  }
}

/**
 * An operation within a {@link StackLine}. This structure is useful because stack lines may be
 * sparse. TODO: replace with a map[int->StackOperation] within StackLine?
 *
 * @param i the index of the stack item within a stack line -- within [[1, 4]
 * @param it the details of the {@link StackOperation} to apply to a column of a stack line
 */
record IndexedStackOperation(int i, StackOperation it) {
  /**
   * For the sake of homogeneity with the zkEVM spec, {@param i} is set with 1-based indices.
   * However, these indices are used to index 0-based array; hence this sneaky conversion.
   *
   * @return the actual stack operation index
   */
  public int i() {
    return this.i - 1;
  }
}

/**
 * As the zkEVM spec can only handle up to four stack operations per trace line of the {@link Hub},
 * operations on the stack must be decomposed in “lines” (mapping 1-to-1 with a trace line from the
 * hub, hence the name) of zero to four atomic {@link StackOperation}.
 *
 * @param items zero to four stack operations contained within this line
 * @param ct the index of this line within its parent {@link StackContext}
 * @param resultColumn if positive, in which item to store the expected retroactive result
 */
record StackLine(
    List<IndexedStackOperation> items,
    int ct, // TODO: could probably be inferred at trace-time
    int resultColumn) {

  /** The default constructor, an empty stack line */
  StackLine() {
    this(new ArrayList<>(), 0, -1);
  }

  /** The default constructor, an empty stack line at a given counter */
  StackLine(int ct) {
    this(new ArrayList<>(), ct, -1);
  }

  /**
   * Build a stack line from a set of {@link StackOperation}
   *
   * @param ct the index of this line within the parent {@link StackContext}
   * @param items the {@link IndexedStackOperation} to include in this line
   */
  StackLine(int ct, IndexedStackOperation... items) {
    this(Arrays.stream(items).toList(), ct, -1);
  }

  /**
   * @return a consolidated 4-elements array of the {@link StackOperation} – or no-ops
   */
  List<StackOperation> asStackOperations() {
    StackOperation[] r =
        new StackOperation[] {
          new StackOperation(), new StackOperation(), new StackOperation(), new StackOperation()
        };
    for (IndexedStackOperation item : this.items) {
      r[item.i()] = item.it();
    }
    return Arrays.asList(r);
  }

  /**
   * Sets the value of a stack item in the line. Used to retroactively set the value of push {@link
   * Action} during the unlatching process.
   *
   * @param i the 1-based stack item to alter
   * @param value the {@link EWord} to use
   */
  public void setResult(int i, EWord value) {
    for (var item : this.items) {
      if (item.i() == i - 1) {
        item.it().setValue(value);
        return;
      }
    }

    throw new RuntimeException(String.format("Item #%s not found in stack line", i));
  }

  /**
   * Sets the value of stack item <code>resultColumn</code>. Used to retroactively set the value of
   * push {@link Action} during the unlatching process.
   *
   * @param value the {@link EWord} to use
   */
  public void setResult(EWord value) {
    if (this.resultColumn == -1) {
      throw new RuntimeException("Stack line has no result column");
    }
    this.setResult(this.resultColumn, value);
  }

  /**
   * @return whether an item in this stack line requires a retroactively set value.
   */
  boolean needsResult() {
    return this.resultColumn >= 0;
  }
}

/**
 * A StackContext encode the stack-related information pertaining to the execution of an opcode
 * within a {@link CallFrame}. These cached information are used by the {@link Hub} to generate its
 * traces in the stack perspective.
 */
final class StackContext {
  /** The opcode that triggered the stack operations. */
  OpCode opCode;
  /** One or two lines to be traced, representing the stack operations performed by the opcode. */
  List<StackLine> lines;
  /** At which line in the {@link Hub} trace this stack operation is to be found. */
  int startInTrace;

  /**
   * The default constructor for a valid, albeit empty line.
   *
   * @param opCode the {@link OpCode} triggering the lines creation
   */
  StackContext(OpCode opCode) {
    this.opCode = opCode;
    this.lines = new ArrayList<>();
    this.startInTrace = 0;
  }

  /**
   * Generate a given number of empty stack lines; typically used as valid padding in the case of
   * stack exception.
   *
   * @param k the number of empty lines to generate
   * @return the number of empty lines generated
   */
  int addEmptyLines(int k) {
    for (int i = 0; i < k; i++) {
      this.lines.add(new StackLine());
    }
    return k;
  }

  /**
   * Creates a new stack lint that will not require unlatching, either because no value are pushed
   * or because they are already known.
   *
   * @param items the stack operations to execute
   */
  void addLine(IndexedStackOperation... items) {
    int newPos = this.lines.size();
    this.lines.add(new StackLine(newPos, items));
  }

  /**
   * Creates a new stack line that will require unlatching.
   *
   * @param posResult in which stack item the result shall be unlatched
   * @param items the stack operations to execute
   */
  void addArmingLine(int posResult, IndexedStackOperation... items) {
    int newPos = this.lines.size();
    this.lines.add(new StackLine(Arrays.stream(items).toList(), newPos, posResult));
  }

  /**
   * As virtually all latched stack operations write to item #4, this provides a shortcut for it.
   *
   * @param items the stack operations to execute
   */
  void addArmingLine(IndexedStackOperation... items) {
    this.addArmingLine(4, items);
  }
}
