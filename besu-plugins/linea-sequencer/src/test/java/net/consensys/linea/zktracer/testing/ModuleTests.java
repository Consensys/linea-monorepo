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

package net.consensys.linea.zktracer.testing;

import java.util.List;

import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

/** Contains methods that execute module tests. */
public class ModuleTests {

  /**
   * Compiles and runs a predefined sequence of bytecode instructions for testing purposes.
   *
   * @param opCode opcode for which the test should be run
   * @param arguments args of the opcode for which the test should be run
   */
  public static void runTestWithOpCodeArgs(final OpCode opCode, final List<Bytes32> arguments) {
    Bytes bytecode = BytecodeCompiler.newProgram().opAnd32ByteArgs(opCode, arguments).compile();

    BytecodeExecutor.builder().byteCode(bytecode).build().run();
  }

  /**
   * Generates a JSON trace based on an opcode and associated arguments.
   *
   * @param opCode opcode for which the trace should be generated
   * @param arguments args of the opcode for which the trace should be generated
   * @return a JSON string representation of the trace
   */
  public static String generateTrace(final OpCode opCode, final List<Bytes32> arguments) {
    Bytes bytecode = BytecodeCompiler.newProgram().opAnd32ByteArgs(opCode, arguments).compile();

    return BytecodeExecutor.builder().byteCode(bytecode).build().traceCode();
  }
}
