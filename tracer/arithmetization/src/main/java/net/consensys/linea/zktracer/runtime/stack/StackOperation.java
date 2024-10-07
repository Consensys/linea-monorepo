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
import lombok.Setter;
import lombok.experimental.Accessors;
import org.apache.tuweni.bytes.Bytes;

/**
 * An atomic operation (read/pop or write/push) on the stack, indexed within a {@link
 * IndexedStackOperation}.
 *
 * <p>Alterations of the stack by an EVM instruction are decomposed in one or two chunks, or {@link
 * StackLine}, made of {@link IndexedStackOperation}, which are then stored, with the associated
 * metadata, in a {@link StackContext}.
 */
@Accessors(fluent = true)
public final class StackOperation {
  private static final Bytes MARKER = Bytes.fromHexString("0xDEADBEEF");

  /**
   * The relative height of the element with regard to the stack height just before executing the
   * linked EVM instruction.
   */
  @Getter private final int height;

  /** The value having been popped from/pushed on the stack. */
  @Getter @Setter private Bytes value;

  /** whether this action is a push or a pop. */
  @Getter private final byte action;

  /**
   * The stamp of this operation relative to the stack stamp before executing the linked EVM
   * instruction.
   */
  @Getter private final int stackStamp;

  StackOperation() {
    this.height = 0;
    this.value = Bytes.EMPTY;
    this.action = Stack.NONE;
    this.stackStamp = 0;
  }

  StackOperation(int height, Bytes value, byte action, int stackStamp) {
    this.height = height;
    this.value = value;
    this.action = action;
    this.stackStamp = stackStamp;
  }

  public static StackOperation pop(int height, Bytes value, int stackStamp) {
    return new StackOperation(height, value, Stack.POP, stackStamp);
  }

  public static StackOperation push(int height, int stackStamp) {
    return new StackOperation(
        height, MARKER /* marker value, erased on unlatching */, Stack.PUSH, stackStamp);
  }

  public static StackOperation pushImmediate(int height, Bytes val, int stackStamp) {
    return new StackOperation(height, val.copy(), Stack.PUSH, stackStamp);
  }
}
