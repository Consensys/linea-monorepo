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

package net.consensys.linea.zktracer.module.shf;

import net.consensys.linea.zktracer.bytes.UnsignedByte;

public record ByteChunks(UnsignedByte ra, UnsignedByte la, UnsignedByte ones) {

  public static ByteChunks fromBytes(final UnsignedByte b, final UnsignedByte mshp) {
    if (mshp.toInteger() > 8) {
      String s =
          String.format("chunksFromByte given mshp = %d not in {0,1,...,8}", mshp.toInteger());
      throw new IllegalArgumentException(s);
    }

    final UnsignedByte mshpCmp = UnsignedByte.of(8 - mshp.toInteger());
    final UnsignedByte ra = b.shiftRight(mshp);
    final UnsignedByte la = b.shiftLeft(mshpCmp);
    final UnsignedByte ones = UnsignedByte.of(255).shiftLeft(mshpCmp);

    return new ByteChunks(ra, la, ones);
  }
}
