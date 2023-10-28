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

package net.consensys.linea.zktracer.types;

import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.experimental.Accessors;

@AllArgsConstructor(access = AccessLevel.PRIVATE)
@Accessors(fluent = true)
public class MemoryRange {
  @Getter private MemoryPoint start;
  @Getter private MemoryPoint end;

  public static MemoryRange newInstance(EWord offset, EWord length) {
    EWord endWord = offset.add(length);

    MemoryPoint start = MemoryPoint.fromAddress(offset);
    MemoryPoint end = MemoryPoint.fromAddress(endWord);

    return new MemoryRange(start, end);
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
