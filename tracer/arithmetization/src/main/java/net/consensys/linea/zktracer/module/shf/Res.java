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

package net.consensys.linea.zktracer.module.shf;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.bytes.Bytes16;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;

@RequiredArgsConstructor
public class Res {
  @Getter final Bytes16 resHi;
  @Getter final Bytes16 resLo;

  public static Res create(final OpCode opCode, final Bytes32 arg1, final Bytes32 arg2) {
    final Bytes32 result = Shifter.shift(opCode, arg2, shiftBy(arg1));

    return new Res(Bytes16.wrap(result.slice(0, 16)), Bytes16.wrap(result.slice(16)));
  }

  private static int shiftBy(final Bytes32 arg) {
    return allButLastByteZero(arg) ? arg.get(31) & 0xff : 256;
  }

  private static boolean allButLastByteZero(final Bytes32 bytes) {
    for (int i = 0; i < 31; i++) {
      if (bytes.get(i) > 0) {
        return false;
      }
    }

    return true;
  }
}
