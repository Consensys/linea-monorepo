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

import static net.consensys.linea.zktracer.Trace.LLARGE;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes32;

@RequiredArgsConstructor
public class Shb {

  @Getter private final UnsignedByte[][] shbHi;
  @Getter private final UnsignedByte[][] shbLo;

  public static Shb create(final OpCode opCode, final Bytes32 arg2, final UnsignedByte lsb) {
    final UnsignedByte[][] shbHi = new UnsignedByte[5][LLARGE];
    final UnsignedByte[][] shbLo = new UnsignedByte[5][LLARGE];

    for (int i = 3; i < 8; i++) {
      final UnsignedByte shiftAmount = (lsb.shiftLeft(8 - i)).shiftRight(8 - i);

      Bytes32 shiftedBytes = Shifter.shift(opCode, arg2, shiftAmount.toInteger());

      final byte[] shbHiByteArray = shiftedBytes.slice(0, 16).toArray();
      for (int j = 0; j < shbHiByteArray.length; j++) {
        shbHi[i - 3][j] = UnsignedByte.of(shbHiByteArray[j]);
      }

      final byte[] shbLoByteArray = shiftedBytes.slice(16).toArray();
      for (int j = 0; j < shbLoByteArray.length; j++) {
        shbLo[i - 3][j] = UnsignedByte.of(shbLoByteArray[j]);
      }
    }

    return new Shb(shbHi, shbLo);
  }
}
