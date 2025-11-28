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

package net.consensys.linea.zktracer.module.ecdata;

import static net.consensys.linea.zktracer.Fork.isPostOsaka;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_RETURN_DATA_SIZE___P256_VERIFY;
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;
import java.io.InputStream;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@ExtendWith(UnitTestWatcher.class)
public class P256VerifyTest extends TracerTestBase {

  @ParameterizedTest
  @MethodSource("p256VerifySource")
  void testP256Verify(String inputAsAsString, String expectedAsString, TestInfo testInfo) {
    testP256VerifyBody(inputAsAsString, expectedAsString, testInfo);
  }

  @Tag("nightly")
  @ParameterizedTest
  @MethodSource("p256VerifySourceNightly")
  void testP256VerifyNightly(String inputAsAsString, String expectedAsString, TestInfo testInfo) {
    testP256VerifyBody(inputAsAsString, expectedAsString, testInfo);
  }

  private void testP256VerifyBody(
      String inputAsAsString, String expectedAsString, TestInfo testInfo) {
    final Bytes input = Bytes.fromHexString(inputAsAsString);

    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);

    final Address codeOwnerAddress = Address.fromHexString("0xC0DE");
    final ToyAccount codeOwnerAccount =
        ToyAccount.builder()
            .balance(Wei.of(0))
            .nonce(1)
            .address(codeOwnerAddress)
            .code(input)
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
        .push(PRECOMPILE_RETURN_DATA_SIZE___P256_VERIFY) // retSize
        .push(input.size()) // retOffset
        .push(input.size()) // argSize
        .push(0) // argOffset
        .push(Address.P256_VERIFY) // address
        .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
        .op(OpCode.STATICCALL);
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run(List.of(codeOwnerAccount), chainConfig, testInfo);

    if (isPostOsaka(fork)) {
      assertEquals(
          Bytes.fromHexString(expectedAsString),
          bytecodeRunner.getHub().ecData().ecDataOperation().returnData());
    }
  }

  private static Stream<Arguments> p256VerifySource() throws IOException {
    List<Arguments> arguments = new ArrayList<>(p256VerifySourceNightly().toList());
    Collections.shuffle(arguments, new Random(LocalDate.now().toEpochDay()));
    return arguments.stream().limit(arguments.size() / 20); // Execute 5 % of the tests
  }

  private static Stream<Arguments> p256VerifySourceNightly() throws IOException {
    // Read json
    // Test vector comes from https://eips.ethereum.org/assets/eip-7951/test-vectors.json
    InputStream inputStream =
        P256VerifyTest.class.getResourceAsStream("/p256_verify_test_vectors.json");
    ObjectMapper mapper = new ObjectMapper();
    JsonNode root = mapper.readTree(inputStream);
    // Fill list of arguments
    List<Arguments> arguments = new ArrayList<>();
    for (JsonNode node : root) {
      String input = node.get("Input").asText();
      String expected = node.get("Expected").asText();
      arguments.add(Arguments.of(input, expected));
    }
    return arguments.stream();
  }
}
