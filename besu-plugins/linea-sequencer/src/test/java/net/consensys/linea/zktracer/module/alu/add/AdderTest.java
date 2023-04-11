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
package net.consensys.linea.zktracer.module.alu.add;

import static org.assertj.core.api.Assertions.assertThat;

import net.consensys.linea.zktracer.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.Test;

class AdderTest {

  @Test
  void zeroAddZero_isZero() {
    Bytes32 actual = Adder.addSub(OpCode.ADD, Bytes32.ZERO, Bytes32.ZERO);
    assertThat(actual).isEqualTo(Bytes32.ZERO);
  }

  @Test
  void zeroSubZero_isZero() {
    Bytes32 actual = Adder.addSub(OpCode.SUB, Bytes32.ZERO, Bytes32.ZERO);
    assertThat(actual).isEqualTo(Bytes32.ZERO);
  }

  @Test
  void xSubZero_isX() {
    Bytes32 randomBytes = Bytes32.random();
    Bytes32 actual = Adder.addSub(OpCode.SUB, randomBytes, Bytes32.ZERO);
    assertThat(actual).isEqualTo(randomBytes);
  }

  @Test
  void xAddZero_isX() {
    Bytes32 randomBytes = Bytes32.random();
    Bytes32 actual = Adder.addSub(OpCode.ADD, randomBytes, Bytes32.ZERO);
    assertThat(actual).isEqualTo(randomBytes);
  }

  @Test
  void maxSubMax_isZero() {
    byte b;
    b = 'f';
    Bytes32 max = Bytes32.repeat(b);
    Bytes32 actual = Adder.addSub(OpCode.SUB, max, max);
    assertThat(actual).isEqualTo(Bytes32.ZERO);
  }

  @Test
  void maxSubZero_isMax() {
    byte b;
    b = 'f';
    Bytes32 max = Bytes32.repeat(b);
    Bytes32 actual = Adder.addSub(OpCode.SUB, max, Bytes32.ZERO);
    assertThat(actual).isEqualTo(max);
  }

  @Test
  void overflowDoesNotError() {
    byte b;
    b = 'f';
    Bytes32 max = Bytes32.repeat(b);
    Adder.addSub(OpCode.ADD, max, max);
  }
}
