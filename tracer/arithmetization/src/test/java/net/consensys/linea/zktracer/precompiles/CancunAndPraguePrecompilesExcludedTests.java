/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.precompiles;

import static net.consensys.linea.zktracer.types.AddressUtils.*;
import static org.hyperledger.besu.datatypes.Address.KZG_POINT_EVAL;
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

public class CancunAndPraguePrecompilesExcludedTests extends TracerTestBase {
  // TODO: remove me when Linea supports Cancun & Prague precompiles
  @ParameterizedTest
  @MethodSource("cancunAndPraguePrecompilesExclusionTestSource")
  void cancunAndPraguePrecompilesExclusionTest(Address prc, TestInfo testInfo) {
    final Bytes bytecode =
        BytecodeCompiler.newProgram(chainConfig)
            .push(0)
            .push(0)
            .push(0)
            .push(0)
            .push(0)
            .push(prc) // address
            .push(0xffff) // gas
            .op(OpCode.CALL)
            .op(OpCode.POP)
            .compile();
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(bytecode);
    try {
      bytecodeRunner.run(chainConfig, testInfo);
    } catch (Exception e) {
      // we don't care about execution result, just the counting. Tracing is expected to fail.
    }

    // Check that the line count is made
    assertEquals(
        isKzgPrecompileCall(prc, chainConfig.fork) ? Integer.MAX_VALUE : 0,
        bytecodeRunner.getHub().pointEval().lineCount());
    assertEquals(
        isBlsPrecompileCall(prc, chainConfig.fork) ? Integer.MAX_VALUE : 0,
        bytecodeRunner.getHub().bls().lineCount());
  }

  private static Stream<Arguments> cancunAndPraguePrecompilesExclusionTestSource() {
    final List<Arguments> arguments = new ArrayList<>();
    for (Address address : BLS_PRECOMPILES) {
      arguments.add(Arguments.of(address));
    }
    arguments.add(Arguments.of(KZG_POINT_EVAL));
    return arguments.stream();
  }
}
