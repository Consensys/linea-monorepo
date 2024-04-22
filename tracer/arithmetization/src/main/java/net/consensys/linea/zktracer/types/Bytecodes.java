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

import static net.consensys.linea.zktracer.module.mmu.Trace.LLARGE;

import org.apache.tuweni.bytes.Bytes;

public class Bytecodes {

  public static Bytes16 readBytes(final Bytes data, final long offset, final int sizeToRead) {
    if (offset >= data.size()) {
      return Bytes16.ZERO;
    }

    final long dataLengthToExtract = Math.min(sizeToRead, data.size() - offset);

    return Bytes16.rightPad(data.slice((int) offset, (int) dataLengthToExtract));
  }

  public static Bytes16 readLimb(final Bytes data, final long limbOffset) {
    return readBytes(data, LLARGE * limbOffset, LLARGE);
  }
}
