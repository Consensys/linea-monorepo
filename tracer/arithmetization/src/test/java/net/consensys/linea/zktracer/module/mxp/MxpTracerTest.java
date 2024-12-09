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

package net.consensys.linea.zktracer.module.mxp;

import static net.consensys.linea.zktracer.opcode.OpCode.MLOAD;
import static net.consensys.linea.zktracer.opcode.OpCode.MSTORE;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.DynamicTests;
import net.consensys.linea.testing.OpcodeCall;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.TestFactory;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(UnitTestWatcher.class)
public class MxpTracerTest {
  // private static final Random RAND = new Random();
  private static final int TEST_REPETITIONS = 2;
  private static final Module MODULE = new Mxp();
  private static final DynamicTests DYN_TESTS = DynamicTests.forModule(MODULE);

  @TestFactory
  Stream<DynamicTest> runDynamicTests() {
    return DYN_TESTS
        .testCase("non random arguments test", provideNonRandomArguments())
        .testCase("simple mload arguments test", simpleMloadArgs())
        .testCase(
            "one of each type2 and type3 instruction MLOAD, MSTORE, MSTORE8", simpleType2And3Args())
        .run();
  }

  private List<OpcodeCall> provideNonRandomArguments() {
    return DYN_TESTS.newModuleArgumentsProvider(
        (arguments, opCode) -> {
          for (int i = 0; i < TEST_REPETITIONS; i++) {
            final List<Bytes32> args = new ArrayList<>(12);
            for (int j = 0; j < (opCode.equals(OpCode.EXTCODECOPY) ? 4 : 3); j++) {
              args.add(UInt256.valueOf(j));
            }
            arguments.add(new OpcodeCall(opCode, args));
          }
        });
  }

  protected List<OpcodeCall> simpleMloadArgs() {
    Bytes32 arg1 =
        Bytes32.fromHexString("0xdcd5cf52e4daec5389587d0d0e996e6ce2d0546b63d3ea0a0dc48ad984d180a9");
    return List.of(new OpcodeCall(MLOAD, List.of(arg1)));
  }

  protected List<OpcodeCall> simpleType2And3Args() {
    // one of each type2 and type3 instruction MLOAD, MSTORE, MSTORE8
    Bytes32 arg1 =
        Bytes32.fromHexString("0xdcd5cf52e4daec5389587d0d0e996e6ce2d0546b63d3ea0a0dc48ad984d180a9");
    return List.of(
        new OpcodeCall(MLOAD, List.of(arg1)),
        new OpcodeCall(MSTORE, List.of(arg1)),
        new OpcodeCall(MSTORE, List.of(arg1)));
  }
}
