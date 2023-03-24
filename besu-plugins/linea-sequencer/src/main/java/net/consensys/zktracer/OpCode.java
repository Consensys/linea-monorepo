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
package net.consensys.zktracer;

import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

public enum OpCode {
  // shf
  SAR(0x1d),
  SHL(0x1b),
  SHR(0x1c),
  // add
  ADD(0x01),
  SUB(0x03),
  ADDMOD(0x08);

  public final long value;

  private static final Map<Long, OpCode> BY_VALUE = new HashMap<>(values().length);

  static {
    for (OpCode o : values()) {
      BY_VALUE.put(o.value, o);
    }
  }

  OpCode(final int value) {
    this.value = value;
  }

  public static OpCode of(final long value) {
    if (!BY_VALUE.containsKey(value)) {
      throw new AssertionError("No OpCode with value " + value + " is defined.");
    }

    return BY_VALUE.get(value);
  }

  public boolean isEqual(final long opCode) {
    return this.value == opCode;
  }

  public boolean isElementOf(OpCode... opCodeSet) {
    return Arrays.asList(opCodeSet).contains(this);
  }
}
