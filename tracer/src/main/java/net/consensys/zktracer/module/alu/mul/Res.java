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
package net.consensys.zktracer.module.alu.mul;

import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.bytes.Bytes16;
import net.consensys.zktracer.module.shf.Shifter;
import org.apache.tuweni.bytes.Bytes32;

import java.math.BigInteger;

public class Res {
  final Bytes16 resHi;
  final Bytes16 resLo;
  final boolean isZero;

  private Res(Bytes16 resHi, Bytes16 resLo, boolean isZero) {
    this.resHi = resHi;
    this.resLo = resLo;
    this.isZero = isZero;
  }

  public Bytes16 getResHi() {
    return resHi;
  }

  public Bytes16 getResLo() {
    return resLo;
  }

  public static Res create(final OpCode opCode, final Bytes32 arg1, final Bytes32 arg2) {
    final Bytes32 result = Muler.operate(opCode, arg1, arg2);

    return new Res(Bytes16.wrap(result.slice(0, 16)), Bytes16.wrap(result.slice(16)), result.isZero());
  }

  public boolean isZero() {
    return isZero;
  }

}
