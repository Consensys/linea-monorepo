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

package net.consensys.linea.zktracer.module.oob;

import static java.util.Map.entry;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_SIZE___FP2_TO_G2;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_SIZE___FP_TO_G1;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_SIZE___G1_ADD;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_SIZE___G2_ADD;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_SIZE___POINT_EVALUATION;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_G1_MSM;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_G2_MSM;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_PAIRING_CHECK;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@ExtendWith(UnitTestWatcher.class)
public class BlsPrecompilesSizeTest extends TracerTestBase {

  @ParameterizedTest
  @MethodSource("blsPrecompilesSizeTestSource")
  void blsPrecompilesSizeTest(Address address, Integer size, TestInfo testInfo) {
    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);

    final Address codeOwnerAddress = Address.fromHexString("0xC0DE");
    final ToyAccount codeOwnerAccount =
        ToyAccount.builder()
            .balance(Wei.of(0))
            .nonce(1)
            .address(codeOwnerAddress)
            .code(Bytes.fromHexString("00".repeat(size)))
            .build();

    // First place the parameters in memory
    // Copy to targetOffset the code of codeOwnerAccount
    program
        .push(codeOwnerAddress)
        .op(OpCode.EXTCODESIZE) // size
        .push(0) // offset
        .push(0) // targetOffset
        .push(codeOwnerAddress) // address
        .op(OpCode.EXTCODECOPY);

    // Do the call
    program
        .push(0) // retSize
        .push(0) // retOffset, note that we are really interested in the return data
        .push(size) // argSize
        .push(0) // argOffset
        .push(address) // address
        .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
        .op(OpCode.STATICCALL);
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run(List.of(codeOwnerAccount), chainConfig, testInfo);
  }

  private static Stream<Arguments> blsPrecompilesSizeTestSource() {
    final Map<Address, Integer> FIXED_SIZE_PRECOMPILE_ADDRESS_TO_SIZE =
        Map.ofEntries(
            entry(Address.KZG_POINT_EVAL, PRECOMPILE_CALL_DATA_SIZE___POINT_EVALUATION),
            entry(Address.BLS12_G1ADD, PRECOMPILE_CALL_DATA_SIZE___G1_ADD),
            entry(Address.BLS12_G2ADD, PRECOMPILE_CALL_DATA_SIZE___G2_ADD),
            entry(Address.BLS12_MAP_FP_TO_G1, PRECOMPILE_CALL_DATA_SIZE___FP_TO_G1),
            entry(Address.BLS12_MAP_FP2_TO_G2, PRECOMPILE_CALL_DATA_SIZE___FP2_TO_G2));

    final Map<Address, Integer> VARIABLE_SIZE_PRECOMPILE_ADDRESS_TO_UNIT =
        Map.ofEntries(
            entry(Address.BLS12_G1MULTIEXP, PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_G1_MSM),
            entry(Address.BLS12_G2MULTIEXP, PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_G2_MSM),
            entry(Address.BLS12_PAIRING, PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_PAIRING_CHECK));

    List<Arguments> arguments = new ArrayList<>();

    for (Address address : FIXED_SIZE_PRECOMPILE_ADDRESS_TO_SIZE.keySet()) {
      arguments.add(Arguments.of(address, 0));
      arguments.add(Arguments.of(address, 1));
      int size = FIXED_SIZE_PRECOMPILE_ADDRESS_TO_SIZE.get(address);
      for (int cornerCase = -1; cornerCase <= 1; cornerCase++) {
        arguments.add(Arguments.of(address, size + cornerCase));
      }
    }

    for (Address address : VARIABLE_SIZE_PRECOMPILE_ADDRESS_TO_UNIT.keySet()) {
      arguments.add(Arguments.of(address, 0));
      arguments.add(Arguments.of(address, 1));
      int unit = VARIABLE_SIZE_PRECOMPILE_ADDRESS_TO_UNIT.get(address);
      for (int numberOfUnits = 1; numberOfUnits <= 128; numberOfUnits++) {
        for (int cornerCase = -1; cornerCase <= 1; cornerCase++) {
          arguments.add(Arguments.of(address, numberOfUnits * unit + cornerCase));
        }
      }
    }

    return arguments.stream();
  }
}
