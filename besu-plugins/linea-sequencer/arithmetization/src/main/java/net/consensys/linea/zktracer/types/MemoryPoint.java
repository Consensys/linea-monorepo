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

@AllArgsConstructor(access = AccessLevel.PRIVATE)
@Accessors(fluent = true)
public class MemoryPoint {
  @Getter private EWord word;
  @Getter private UnsignedByte uByte;

  public static MemoryPoint fromAddress(long memAddr) {
    EWord word = EWord.of(memAddr).divide(EWord.THIRTY_TWO);
    UnsignedByte uByte = UnsignedByte.of(memAddr % 32);

    return new MemoryPoint(word, uByte);
  }

  public static MemoryPoint fromAddress(EWord memAddr) {
    EWord word = memAddr.divide(EWord.THIRTY_TWO);
    EWord remainder = memAddr.mod(32);
    UnsignedByte uByte = UnsignedByte.of(remainder.toLong());

    return new MemoryPoint(word, uByte);
  }

  public EWord toAbsolute() {
    // out = 32×Word + Byte
    EWord out = EWord.THIRTY_TWO.multiply(word);
    return out.add(EWord.of(uByte.toBigInteger().longValue()));
  }

  @Override
  public String toString() {
    return "[%s]%02d".formatted(word.toString(), uByte.toBigInteger().longValue());
  }
}
