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

package net.consensys.linea.zktracer.module.shf;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.IntStream;
import java.util.stream.Stream;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@Accessors(fluent = true)
@Tag("weekly")
public class ShfExtensiveTest {

  private static final List<Arguments> shfTestSourceList = new ArrayList<>();
  @Getter private static Stream<Arguments> shfWithMaskTestSource;

  @BeforeAll
  public static void init() {
    List<Arguments> shfWithMaskTestSourceList = new ArrayList<>();
    for (int k = 0; k < 32; k++) {
      for (int l = 1; l <= 8; l++) {
        String value = value(k, l);
        shfTestSourceList.add(Arguments.of(value, k, l));
        for (String XY : XYs) {
          String mask = XY + "00".repeat(31);
          String maskXorValue =
              String.format("%064X", new BigInteger(mask, 16).xor(new BigInteger(value, 16)));
          shfWithMaskTestSourceList.add(Arguments.of(maskXorValue, k, l, XY));
        }
      }
    }
    shfWithMaskTestSource = shfWithMaskTestSourceList.stream();
    // shfWithMaskTestSource inputs are used only once, so it is fine to generate the corresponding
    // stream here.
    // Note that whenever a stream is used, it is also consumed,
    // that is why in the case of shfTestSourceList inputs,
    // we generate a new stream every time it is needed.
  }

  @ParameterizedTest
  @MethodSource("shfTestSource")
  void shlTest(String value, int k, int l) {
    shfProgramOf(value, OpCode.SHL).run();
  }

  @ParameterizedTest
  @MethodSource("shfTestSource")
  void shrTest(String value, int k, int l) {
    shfProgramOf(value, OpCode.SHR).run();
  }

  @ParameterizedTest
  @MethodSource("shfTestSource")
  void sarTest(String value, int k, int l) {
    shfProgramOf(value, OpCode.SAR).run();
  }

  @ParameterizedTest
  @MethodSource("shfWithMaskTestSource")
  void sarWithMaskTest(String value, int k, int l, String XY) {
    shfProgramOf(value, OpCode.SAR).run();
  }

  private static Stream<Arguments> shfTestSource() {
    return shfTestSourceList.stream();
    // A new stream is generated whenever it is necessary, starting from the same list
  }

  // Inputs and support methods

  static final String[] SHIFTS =
      Stream.concat(
              IntStream.rangeClosed(0, 257) // Generates numbers 0 to 257
                  .mapToObj(BigInteger::valueOf),
              Stream.of(
                  BigInteger.valueOf(511),
                  BigInteger.valueOf(512),
                  BigInteger.valueOf(513),
                  BigInteger.valueOf(65535),
                  BigInteger.valueOf(65536),
                  BigInteger.valueOf(65537),
                  BigInteger.ONE.shiftLeft(128).subtract(BigInteger.ONE), // (1 << 128) - 1
                  BigInteger.ONE.shiftLeft(128), // (1 << 128)
                  BigInteger.ONE.shiftLeft(128).add(BigInteger.ONE), // (1 << 128) + 1
                  BigInteger.ONE.shiftLeft(256).subtract(BigInteger.ONE) // (1 << 256) - 1
                  ))
          .map(bigInteger -> bigInteger.toString(16))
          .toArray(String[]::new);

  static final String[] P = {
    "DF", "D5", "A2", "E7", "6E", "9D", "3A", "20",
    "96", "2D", "17", "48", "19", "7F", "0D", "4C",
    "FF", "3D", "57", "A4", "A8", "87", "45", "B9",
    "C9", "34", "1A", "F3", "57", "84", "D3", "EE"
  }; // big-endian (from the least significant byte to the most significant byte)

  static final String[] XYs = new String[] {"80", "90", "A0", "B0", "C0", "D0", "E0", "F0"};

  //  Creates a program that concatenates shifts operations (with different relevant shift values)
  //  for a given value and opcode
  private BytecodeRunner shfProgramOf(String value, OpCode opCode) {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    for (String shift : SHIFTS) {
      program.push(value).push(shift).op(opCode);
    }
    return BytecodeRunner.of(program.compile());
  }

  public static String value(int k, int l) {
    String[] v = new String[32];
    // 0 to k - 1
    if (k >= 0) System.arraycopy(P, 0, v, 0, k);
    // k
    v[k] = String.format("%02X", (1 << l) - 1);
    // k + 1 to 31
    for (int i = k + 1; i < 32; i++) {
      v[i] = "00";
    }
    return String.join("", java.util.Arrays.asList(v).reversed());
  }

  // Testing support methods
  @Test
  void testValue() {
    Assertions.assertEquals(
        "0000000000000000000000003F573DFF4C0D7F1948172D96203A9D6EE7A2D5DF", value(19, 6));
  }

  // ###################################################################################################################

  // This test should be executed only occasionally since very long. Run below batched tests instead
  @Disabled
  @ParameterizedTest
  @MethodSource("shfExtensiveTestSource")
  void shfExtensiveTest(String shift, String value, OpCode opCode) {
    BytecodeRunner.of(BytecodeCompiler.newProgram().push(value).push(shift).op(opCode).compile())
        .run();
  }

  private static Stream<Arguments> shfExtensiveTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (String shift : SHIFTS) {
      for (int k = 0; k < 32; k++) {
        for (int l = 1; l <= 8; l++) {
          String value = value(k, l);
          arguments.add(Arguments.of(shift, value, OpCode.SHL));
          arguments.add(Arguments.of(shift, value, OpCode.SHR));
          arguments.add(Arguments.of(shift, value, OpCode.SAR));
          // Adding additional cases for SAR
          for (String XY : XYs) {
            String mask = XY + "00".repeat(31);
            String maskXorValue =
                String.format("%064X", new BigInteger(mask, 16).xor(new BigInteger(value, 16)));
            arguments.add(Arguments.of(shift, maskXorValue, OpCode.SAR));
          }
        }
      }
    }
    return arguments.stream();
  }
}
