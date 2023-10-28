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

package net.consensys.linea.zktracer.module.wcp;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.DynamicTests;
import net.consensys.linea.zktracer.testing.OpcodeCall;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.TestFactory;

class WcpTracerTest {
  private static final Module MODULE = new Wcp();

  private static final DynamicTests DYN_TESTS = DynamicTests.forModule(MODULE);

  @TestFactory
  Stream<DynamicTest> runDynamicTests() {
    return DYN_TESTS.testCase("non random arguments test", provideNonRandomArguments()).run();
  }

  private List<OpcodeCall> provideNonRandomArguments() {
    List<OpcodeCall> testCases = new ArrayList<>();

    Bytes32 arg1 =
        Bytes32.fromHexString("0xdcd5cf52e4daec5389587d0d0e996e6ce2d0546b63d3ea0a0dc48ad984d180a9");
    Bytes32 arg2 =
        Bytes32.fromHexString("0x0479484af4a59464a48818b3980174687661bafb13d06f49537995fa6c02159e");

    testCases.add(new OpcodeCall(OpCode.GT, List.of(arg1, arg2)));

    return testCases;
  }
}
