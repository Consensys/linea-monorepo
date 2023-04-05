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
package net.consensys.linea.zktracer;

import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

public enum OpCode {
  // add
  ADD(0x01),
  SUB(0x03),
  // mod
  DIV(0x04),
  SDIV(0x05),
  MOD(0x06),
  SMOD(0x07),
  //wcp
  LT(0x10),
  GT(0x11),
  SLT(0x12),
  SGT(0x13),
  EQ(0x14),
  ISZERO(0x15),
  // shf
  SHL(0x1b),
  SHR(0x1c),
  SAR(0x1d);

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
