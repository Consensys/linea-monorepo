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

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import com.google.common.base.Preconditions;
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
public class BlsG2AddTest extends TracerTestBase {

  // Valid G2 point - 256 bytes (512 hex chars)
  private static final String VALID_G2_POINT =
      "00000000000000000000000000000000124aca13d9ead2e5194eb097360743fc996551a5f339d644ded3571c5588a1fedf3f26ecdca73845241e47337e8ad990"
          + "000000000000000000000000000000000299bfd77515b688335e58acb31f7e0e6416840989cb08775287f90f7e6c921438b7b476cfa387742fcdc43bcecfe45f"
          + "00000000000000000000000000000000032e78350f525d673e75a3430048a7931d21264ac1b2c8dc58aee07e77790dfc9afb530b004145f0040c48bce128135e"
          + "0000000000000000000000000000000015963bcbd8fa50808bdce4f8de40eb9706c1a41ada22f0e469ecceb3e0b0fa3404ccdcc66a286b5a9e221c4a088a9145";

  // Invalid G2 point (not on curve)
  private static final String INVALID_G2_POINT_NOT_ON_CURVE =
      "000000000000000000000000000000000b2c619263417e8f6cffa2e53261cb8cf5fbbabb9e6f4188aeaabe50d434a0489b6cccd2b65b4d1393a26911021baffa"
          + "00000000000000000000000000000000007bcd4156af7ebe5e2f6ac63db859c9f42d5f11682792a0de2ec1db76648c0c98fdd8a82cf640bdcd309901afd4f570"
          + "00000000000000000000000000000000153a9002d117a518b2c1786f9e8b95b00e936f3f15302a27a16d7f2f8fc48ca834c0cf4fce456e96d72f01f252f4d084"
          + "000000000000000000000000000000001091fc53100190db07ec2057727859e65da996f6792ac5602cb9dfbc3ed4a5a67d6b82bd82112075ef8afc4155db2621";

  // G2 point that's on curve but not in subgroup
  private static final String G2_POINT_NOT_IN_SUBGROUP =
      "000000000000000000000000000000000380f5c0d9ae49e3904c5ae7ad83043158d68fa721b06b561e714b71a2c48c2307b5258892f999a882bed3549a286b7f"
          + "0000000000000000000000000000000004886f7f17a8e9918b4bfa8ebe450b0216ed5e1fa103dfc671332dc38b04ed3105526fb0dda7e032b6fb67debf9f0bc5"
          + "0000000000000000000000000000000018146b7ed1ecf2a4f2d1f75bb6e9ddbb9796bb03576686346995566cf3b3831ec5462e61028355504fc90f877408ac17"
          + "0000000000000000000000000000000003da9de8dcd94d7793b19e45a5521b1bc42f1a6d693139d03bb26402678ee6a635a4d50eaddfd326e446ed0330fa67fb";

  // Invalid padding (non-zero leading bytes)
  private static final String INVALID_G2_PADDING =
      "01000000000000000000000000000000124aca13d9ead2e5194eb097360743fc996551a5f339d644ded3571c5588a1fedf3f26ecdca73845241e47337e8ad990"
          + "000000000000000000000000000000000299bfd77515b688335e58acb31f7e0e6416840989cb08775287f90f7e6c921438b7b476cfa387742fcdc43bcecfe45f"
          + "00000000000000000000000000000000032e78350f525d673e75a3430048a7931d21264ac1b2c8dc58aee07e77790dfc9afb530b004145f0040c48bce128135e"
          + "0000000000000000000000000000000015963bcbd8fa50808bdce4f8de40eb9706c1a41ada22f0e469ecceb3e0b0fa3404ccdcc66a286b5a9e221c4a088a9145";

  // Invalid length input
  private static final String INVALID_LENGTH_INPUT = "0001020304";

  @ParameterizedTest
  @MethodSource("blsG2AddSource")
  void testBlsG2Add(String a, String b, TestInfo testInfo) {
    Preconditions.checkArgument(a.length() == 512, "G2 point 'a' must be 512 hex chars");
    Preconditions.checkArgument(b.length() == 512, "G2 point 'b' must be 512 hex chars");

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
        .push(13) // address
        .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
        .op(OpCode.STATICCALL);
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run(chainConfig, testInfo);
  }

  private static Stream<Arguments> blsG2AddSource() {
    List<Arguments> arguments = new ArrayList<>();
    arguments.add(Arguments.of(VALID_G2_POINT, VALID_G2_POINT));
    return arguments.stream();
  }
}
