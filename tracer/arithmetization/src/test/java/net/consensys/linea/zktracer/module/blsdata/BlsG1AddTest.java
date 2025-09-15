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

package net.consensys.linea.zktracer.module.blsdata;

import static com.google.common.base.Preconditions.checkArgument;

import java.util.ArrayList;
import java.util.List;
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
public class BlsG1AddTest extends TracerTestBase {

  // Valid G1 point - 128 bytes (256 hex chars)
  private static final String VALID_G1_POINT =
      "0000000000000000000000000000000012196c5a43d69224d8713389285f26b98f86ee910ab3dd668e413738282003cc5b7357af9a7af54bb713d62255e80f56"
          + "0000000000000000000000000000000006ba8102bfbeea4416b710c73e8cce3032c31c6269c44906f8ac4f7874ce99fb17559992486528963884ce429a992fee";

  // Invalid G1 point (not on curve)
  private static final String INVALID_G1_POINT_NOT_ON_CURVE =
      "00000000000000000000000000000000177b39d2b8d31753ee35033df55a1f891be9196aec9cd8f512e9069d21a8bdbf693bd2e826e792cd12cb554287adf4ca"
          + "0000000000000000000000000000000003c0f5770509862f754fc474cb163c41790d844f52939e2dec87b97c2a707831a4043ab47014d501f67862e95842ba5a";

  // G1 point that's on curve but not in subgroup
  private static final String G1_POINT_NOT_IN_SUBGROUP =
      "00000000000000000000000000000000054a4326bbddbdbbca126659e6686984046d2fa49270742e5b6d9017734acf2801f370eebe7af29dfc8d50483609dc00"
          + "000000000000000000000000000000001713e9ef64254fe96d874d16e33636f186e30d7e476db9f49a16698b771f10e0f8f08e5d8dba621b887c0d257cbd8eac";

  // Invalid padding (non-zero leading bytes)
  private static final String INVALID_G1_PADDING =
      "0100000000000000000000000000000012196c5a43d69224d8713389285f26b98f86ee910ab3dd668e413738282003cc5b7357af9a7af54bb713d62255e80f56"
          + "0000000000000000000000000000000006ba8102bfbeea4416b710c73e8cce3032c31c6269c44906f8ac4f7874ce99fb17559992486528963884ce429a992fee";

  // Invalid length input
  private static final String INVALID_LENGTH_INPUT = "0001020304";

  @ParameterizedTest
  @MethodSource("blsG1AddSource")
  void testBlsG1Add(String a, String b, TestInfo testInfo) {
    checkArgument(a.length() == 256, "G1 point 'a' must be 256 hex chars");
    checkArgument(b.length() == 256, "G1 point 'b' must be 256 hex chars");

    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);

    // TODO: extract method for that
    final Address codeOwnerAddress = Address.fromHexString("0xC0DE");
    final ToyAccount codeOwnerAccount =
        ToyAccount.builder()
            .balance(Wei.of(0))
            .nonce(1)
            .address(codeOwnerAddress)
            .code(Bytes.concatenate(Bytes.fromHexString(a), Bytes.fromHexString(b)))
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
        .push(0x80) // retSize
        .push(0x100) // retOffset
        .push(0x100) // argSize
        .push(0) // argOffset
        .push(11) // address
        .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
        .op(OpCode.STATICCALL);
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run(List.of(codeOwnerAccount), chainConfig, testInfo);
  }

  private static Stream<Arguments> blsG1AddSource() {
    List<Arguments> arguments = new ArrayList<>();
    arguments.add(Arguments.of(VALID_G1_POINT, VALID_G1_POINT));
    return arguments.stream();
  }
}
