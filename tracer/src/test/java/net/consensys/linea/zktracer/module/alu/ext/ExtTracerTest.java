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
package net.consensys.linea.zktracer.module.alu.ext;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.module.AbstractModuleTracerTest;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.provider.Arguments;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class ExtTracerTest extends AbstractModuleTracerTest {
  static final Random rand = new Random();

  @Test
  public void testExactValue() {
    Bytes32 arg1 = fromBigInteger(BigInteger.valueOf(6));
    Bytes32 arg2 = fromBigInteger(BigInteger.valueOf(7));
    Bytes32 arg3 = fromBigInteger(BigInteger.valueOf(13));
    runTest(OpCode.MULMOD, arg1, arg2, arg3);
  }

  @Test
  public void testMAXExactValue() {
    Bytes32 arg1 = UInt256.MAX_VALUE;
    Bytes32 arg2 = UInt256.valueOf(2);
    Bytes32 arg3 = UInt256.valueOf(10);
    runTest(OpCode.MULMOD, arg1, arg2, arg3);
  }

  @Override
  public Stream<Arguments> provideNonRandomArguments() {
    List<Arguments> arguments = new ArrayList<>();
    for (OpCode opCode : getModuleTracer().supportedOpCodes()) {
      for (int k = 1; k <= 4; k++) {
        for (int i = 1; i <= 4; i++) {
          arguments.add(
              Arguments.of(opCode, UInt256.valueOf(i), UInt256.valueOf(k), UInt256.valueOf(k)));
        }
      }
    }
    return arguments.stream();
  }

  @Override
  protected ModuleTracer getModuleTracer() {
    return new ExtTracer();
  }

  private Bytes32 fromBigInteger(BigInteger bigInteger) {
    return Bytes32.leftPad(Bytes.wrap(bigInteger.toByteArray()));
  }
}
