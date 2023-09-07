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

package net.consensys.linea.zktracer.module.ext;

import java.util.stream.Stream;

import com.google.common.collect.ArrayListMultimap;
import com.google.common.collect.Multimap;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.DynamicTests;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.TestFactory;

class ExtTracerTest {
  private static final Module MODULE = new Ext();

  private static final DynamicTests DYN_TESTS = DynamicTests.forModule(MODULE);

  @TestFactory
  Stream<DynamicTest> runDynamicTests() {
    return DYN_TESTS
        .testCase("non random arguments test", provideNonRandomArguments())
        .testCase("zero value test", provideZeroValueTest())
        //      .testCase("modulus zero value arguments test", provideModulusZeroValueArguments())
        .testCase("tiny value arguments test", provideTinyValueArguments())
        .testCase("max value arguments test", provideMaxValueArguments())
        .run();
  }

  private Multimap<OpCode, Bytes32> provideNonRandomArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (OpCode opCode : MODULE.supportedOpCodes()) {
      for (int k = 1; k <= 4; k++) {
        for (int i = 1; i <= 4; i++) {
          arguments.put(opCode, UInt256.valueOf(i));
          arguments.put(opCode, UInt256.valueOf(k));
          arguments.put(opCode, UInt256.valueOf(k));
        }
      }
    }

    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideZeroValueTest() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (OpCode opCode : MODULE.supportedOpCodes()) {
      arguments.put(opCode, UInt256.valueOf(6));
      arguments.put(opCode, UInt256.valueOf(12));
      arguments.put(opCode, UInt256.valueOf(0));
    }

    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideModulusZeroValueArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (OpCode opCode : MODULE.supportedOpCodes()) {
      arguments.put(opCode, UInt256.valueOf(0));
      arguments.put(opCode, UInt256.valueOf(1));
      arguments.put(opCode, UInt256.valueOf(1));
    }

    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideTinyValueArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (OpCode opCode : MODULE.supportedOpCodes()) {
      arguments.put(opCode, UInt256.valueOf(6));
      arguments.put(opCode, UInt256.valueOf(7));
      arguments.put(opCode, UInt256.valueOf(13));
    }

    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideMaxValueArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (OpCode opCode : MODULE.supportedOpCodes()) {
      arguments.put(opCode, UInt256.MAX_VALUE);
      arguments.put(opCode, UInt256.MAX_VALUE);
      arguments.put(opCode, UInt256.MAX_VALUE);
    }

    return arguments;
  }
}
