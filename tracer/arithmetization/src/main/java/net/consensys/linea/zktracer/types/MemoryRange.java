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

package net.consensys.linea.zktracer.types;

import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.experimental.Accessors;

@Getter
@AllArgsConstructor(access = AccessLevel.PRIVATE)
@Accessors(fluent = true)
public class MemoryRange {
  private MemoryPoint start;
  private MemoryPoint end;

  public static MemoryRange fromOffsetSize(EWord start, EWord length) {
    final EWord end = start.add(length);

    return new MemoryRange(MemoryPoint.fromAddress(start), MemoryPoint.fromAddress(end));
  }

  public static MemoryRange fromStartEnd(EWord start, EWord end) {
    return new MemoryRange(MemoryPoint.fromAddress(start), MemoryPoint.fromAddress(end));
  }

  public static MemoryRange fromStartEnd(long start, long end) {
    return new MemoryRange(MemoryPoint.fromAddress(start), MemoryPoint.fromAddress(end));
  }

  public EWord length() {
    return end.word()
        .subtract(start.word())
        .multiply(32)
        .add(end.uByte().toBigInteger().longValue())
        .subtract(start.uByte().toBigInteger().longValue());
  }

  public EWord absolute() {
    return start.toAbsolute();
  }

  public boolean isUint64() {
    return length().isUInt64();
  }

  @Override
  public String toString() {
    return "%s -- %s (%d)".formatted(start, end, length().toBigInteger().longValue());
  }
}
