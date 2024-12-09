package net.consensys.linea.zktracer.module.mul;

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

import static net.consensys.linea.zktracer.module.HexStringUtils.and;
import static net.consensys.linea.zktracer.module.HexStringUtils.rightShift;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.IntStream;
import java.util.stream.Stream;

import lombok.experimental.Accessors;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.module.HexStringUtils;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@Accessors(fluent = true)
@Tag("weekly")
@ExtendWith(UnitTestWatcher.class)
public class ExpExtensiveTest {
  // Test vectors
  static final String P_1 = "f076b857fa9947c1f9ec558262c72704099ca8cd325566f73fb99238102ed171";
  static final String P_2 = "c809c9170ca6faec82d43ee6754dbad01d198ecae0823bac23ca30c7f0c9657d";

  static final String[] EVEN_BASES =
      IntStream.rangeClosed(1, 256).mapToObj(HexStringUtils::even).toArray(String[]::new);

  static final String[] SIMPLE_ODD_BASES =
      IntStream.rangeClosed(0, 31)
          .map(i -> i * 8) // Generates 0, 8, 16, ..., 248
          .mapToObj(HexStringUtils::odd)
          .toArray(String[]::new);

  static final String[] OTHER_ODD_BASES =
      Stream.of(SIMPLE_ODD_BASES).map(mask -> and(P_1, mask)).toArray(String[]::new);

  static final String[] COMPLEX_EXPONENTS =
      IntStream.rangeClosed(0, 256).mapToObj(n -> rightShift(P_2, n)).toArray(String[]::new);

  static final String[] TINY_EXPONENTS =
      Stream.concat(
              IntStream.rangeClosed(0, 257) // Generates numbers 0 to 257
                  .mapToObj(BigInteger::valueOf),
              Stream.of(
                  BigInteger.valueOf(65535), BigInteger.valueOf(65536), BigInteger.valueOf(65537)))
          .map(bigInteger -> bigInteger.toString(16))
          .toArray(String[]::new);

  static final String[] THRESHOLD_EXPONENTS =
      Stream.of(
              BigInteger.ONE.shiftLeft(128).subtract(BigInteger.ONE), // (1 << 128) - 1
              BigInteger.ONE.shiftLeft(128), // (1 << 128)
              BigInteger.ONE.shiftLeft(128).add(BigInteger.ONE), // (1 << 128) + 1
              BigInteger.ONE.shiftLeft(256).subtract(BigInteger.ONE) // (1 << 256) - 1
              )
          .map(bigInteger -> bigInteger.toString(16))
          .toArray(String[]::new);

  static final String[] SPECIAL_EXPONENTS =
      Stream.of(TINY_EXPONENTS, THRESHOLD_EXPONENTS).flatMap(Stream::of).toArray(String[]::new);

  static final String[] BASES =
      Stream.of(EVEN_BASES, SIMPLE_ODD_BASES, OTHER_ODD_BASES)
          .flatMap(Stream::of)
          .toArray(String[]::new);

  static final String[] EXPONENTS =
      Stream.of(COMPLEX_EXPONENTS, SPECIAL_EXPONENTS).flatMap(Stream::of).toArray(String[]::new);

  static final List<Arguments> EXPONENTS_LIST = Stream.of(EXPONENTS).map(Arguments::of).toList();

  // Note that flatMap(Stream::of) converts stream of String[] to stream of String

  // This is not an actual test, but just a utility to generate test cases
  // @Disabled
  @Test
  public void generateTestCases() {
    for (int i = 0; i < BASES.length; i++) {
      System.out.println(
          "@Test\n"
              + "void expTestForBase_"
              + BASES[i]
              + "() {\n"
              + "    expProgramOf(BASES["
              + i
              + "]).run();\n"
              + "}\n");
    }
  }

  // Tests
  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe() {
    expProgramOf(BASES[0]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc() {
    expProgramOf(BASES[1]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8() {
    expProgramOf(BASES[2]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0() {
    expProgramOf(BASES[3]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0() {
    expProgramOf(BASES[4]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc0() {
    expProgramOf(BASES[5]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80() {
    expProgramOf(BASES[6]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00() {
    expProgramOf(BASES[7]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe00() {
    expProgramOf(BASES[8]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc00() {
    expProgramOf(BASES[9]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff800() {
    expProgramOf(BASES[10]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000() {
    expProgramOf(BASES[11]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe000() {
    expProgramOf(BASES[12]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc000() {
    expProgramOf(BASES[13]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000() {
    expProgramOf(BASES[14]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000() {
    expProgramOf(BASES[15]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0000() {
    expProgramOf(BASES[16]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc0000() {
    expProgramOf(BASES[17]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80000() {
    expProgramOf(BASES[18]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000() {
    expProgramOf(BASES[19]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe00000() {
    expProgramOf(BASES[20]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc00000() {
    expProgramOf(BASES[21]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffff800000() {
    expProgramOf(BASES[22]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000() {
    expProgramOf(BASES[23]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffe000000() {
    expProgramOf(BASES[24]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffffc000000() {
    expProgramOf(BASES[25]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000000() {
    expProgramOf(BASES[26]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000() {
    expProgramOf(BASES[27]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0000000() {
    expProgramOf(BASES[28]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffc0000000() {
    expProgramOf(BASES[29]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffff80000000() {
    expProgramOf(BASES[30]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000() {
    expProgramOf(BASES[31]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffe00000000() {
    expProgramOf(BASES[32]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffffc00000000() {
    expProgramOf(BASES[33]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffff800000000() {
    expProgramOf(BASES[34]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffff000000000() {
    expProgramOf(BASES[35]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffe000000000() {
    expProgramOf(BASES[36]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffc000000000() {
    expProgramOf(BASES[37]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffff8000000000() {
    expProgramOf(BASES[38]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000000() {
    expProgramOf(BASES[39]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffe0000000000() {
    expProgramOf(BASES[40]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffffc0000000000() {
    expProgramOf(BASES[41]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffff80000000000() {
    expProgramOf(BASES[42]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000() {
    expProgramOf(BASES[43]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffe00000000000() {
    expProgramOf(BASES[44]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffc00000000000() {
    expProgramOf(BASES[45]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffff800000000000() {
    expProgramOf(BASES[46]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffff000000000000() {
    expProgramOf(BASES[47]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffe000000000000() {
    expProgramOf(BASES[48]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffffc000000000000() {
    expProgramOf(BASES[49]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffff8000000000000() {
    expProgramOf(BASES[50]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffff0000000000000() {
    expProgramOf(BASES[51]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffe0000000000000() {
    expProgramOf(BASES[52]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffc0000000000000() {
    expProgramOf(BASES[53]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffff80000000000000() {
    expProgramOf(BASES[54]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000() {
    expProgramOf(BASES[55]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffe00000000000000() {
    expProgramOf(BASES[56]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffffc00000000000000() {
    expProgramOf(BASES[57]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffff800000000000000() {
    expProgramOf(BASES[58]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffff000000000000000() {
    expProgramOf(BASES[59]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffe000000000000000() {
    expProgramOf(BASES[60]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffc000000000000000() {
    expProgramOf(BASES[61]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffff8000000000000000() {
    expProgramOf(BASES[62]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffff0000000000000000() {
    expProgramOf(BASES[63]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffe0000000000000000() {
    expProgramOf(BASES[64]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffffc0000000000000000() {
    expProgramOf(BASES[65]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffff80000000000000000() {
    expProgramOf(BASES[66]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffff00000000000000000() {
    expProgramOf(BASES[67]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffe00000000000000000() {
    expProgramOf(BASES[68]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffc00000000000000000() {
    expProgramOf(BASES[69]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffff800000000000000000() {
    expProgramOf(BASES[70]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffff000000000000000000() {
    expProgramOf(BASES[71]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffe000000000000000000() {
    expProgramOf(BASES[72]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffffc000000000000000000() {
    expProgramOf(BASES[73]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffff8000000000000000000() {
    expProgramOf(BASES[74]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffff0000000000000000000() {
    expProgramOf(BASES[75]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffe0000000000000000000() {
    expProgramOf(BASES[76]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffc0000000000000000000() {
    expProgramOf(BASES[77]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffff80000000000000000000() {
    expProgramOf(BASES[78]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffff00000000000000000000() {
    expProgramOf(BASES[79]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffe00000000000000000000() {
    expProgramOf(BASES[80]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffffc00000000000000000000() {
    expProgramOf(BASES[81]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffff800000000000000000000() {
    expProgramOf(BASES[82]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffff000000000000000000000() {
    expProgramOf(BASES[83]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffe000000000000000000000() {
    expProgramOf(BASES[84]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffc000000000000000000000() {
    expProgramOf(BASES[85]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffff8000000000000000000000() {
    expProgramOf(BASES[86]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffff0000000000000000000000() {
    expProgramOf(BASES[87]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffe0000000000000000000000() {
    expProgramOf(BASES[88]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffffc0000000000000000000000() {
    expProgramOf(BASES[89]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffff80000000000000000000000() {
    expProgramOf(BASES[90]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffff00000000000000000000000() {
    expProgramOf(BASES[91]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffe00000000000000000000000() {
    expProgramOf(BASES[92]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffc00000000000000000000000() {
    expProgramOf(BASES[93]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffff800000000000000000000000() {
    expProgramOf(BASES[94]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffff000000000000000000000000() {
    expProgramOf(BASES[95]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffe000000000000000000000000() {
    expProgramOf(BASES[96]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffffc000000000000000000000000() {
    expProgramOf(BASES[97]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffff8000000000000000000000000() {
    expProgramOf(BASES[98]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffff0000000000000000000000000() {
    expProgramOf(BASES[99]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffe0000000000000000000000000() {
    expProgramOf(BASES[100]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffc0000000000000000000000000() {
    expProgramOf(BASES[101]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffff80000000000000000000000000() {
    expProgramOf(BASES[102]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffff00000000000000000000000000() {
    expProgramOf(BASES[103]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffe00000000000000000000000000() {
    expProgramOf(BASES[104]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffffc00000000000000000000000000() {
    expProgramOf(BASES[105]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffff800000000000000000000000000() {
    expProgramOf(BASES[106]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffff000000000000000000000000000() {
    expProgramOf(BASES[107]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffe000000000000000000000000000() {
    expProgramOf(BASES[108]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffc000000000000000000000000000() {
    expProgramOf(BASES[109]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffff8000000000000000000000000000() {
    expProgramOf(BASES[110]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffff0000000000000000000000000000() {
    expProgramOf(BASES[111]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffe0000000000000000000000000000() {
    expProgramOf(BASES[112]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffffc0000000000000000000000000000() {
    expProgramOf(BASES[113]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffff80000000000000000000000000000() {
    expProgramOf(BASES[114]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffff00000000000000000000000000000() {
    expProgramOf(BASES[115]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffe00000000000000000000000000000() {
    expProgramOf(BASES[116]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffc00000000000000000000000000000() {
    expProgramOf(BASES[117]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffff800000000000000000000000000000() {
    expProgramOf(BASES[118]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffff000000000000000000000000000000() {
    expProgramOf(BASES[119]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffe000000000000000000000000000000() {
    expProgramOf(BASES[120]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffffc000000000000000000000000000000() {
    expProgramOf(BASES[121]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffff8000000000000000000000000000000() {
    expProgramOf(BASES[122]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffff0000000000000000000000000000000() {
    expProgramOf(BASES[123]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffe0000000000000000000000000000000() {
    expProgramOf(BASES[124]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffc0000000000000000000000000000000() {
    expProgramOf(BASES[125]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffff80000000000000000000000000000000() {
    expProgramOf(BASES[126]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffff00000000000000000000000000000000() {
    expProgramOf(BASES[127]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffe00000000000000000000000000000000() {
    expProgramOf(BASES[128]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffffc00000000000000000000000000000000() {
    expProgramOf(BASES[129]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffff800000000000000000000000000000000() {
    expProgramOf(BASES[130]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffff000000000000000000000000000000000() {
    expProgramOf(BASES[131]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffe000000000000000000000000000000000() {
    expProgramOf(BASES[132]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffc000000000000000000000000000000000() {
    expProgramOf(BASES[133]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffff8000000000000000000000000000000000() {
    expProgramOf(BASES[134]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffff0000000000000000000000000000000000() {
    expProgramOf(BASES[135]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffe0000000000000000000000000000000000() {
    expProgramOf(BASES[136]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffffc0000000000000000000000000000000000() {
    expProgramOf(BASES[137]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffff80000000000000000000000000000000000() {
    expProgramOf(BASES[138]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffff00000000000000000000000000000000000() {
    expProgramOf(BASES[139]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffe00000000000000000000000000000000000() {
    expProgramOf(BASES[140]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffc00000000000000000000000000000000000() {
    expProgramOf(BASES[141]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffff800000000000000000000000000000000000() {
    expProgramOf(BASES[142]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffff000000000000000000000000000000000000() {
    expProgramOf(BASES[143]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffe000000000000000000000000000000000000() {
    expProgramOf(BASES[144]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffffc000000000000000000000000000000000000() {
    expProgramOf(BASES[145]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffff8000000000000000000000000000000000000() {
    expProgramOf(BASES[146]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffff0000000000000000000000000000000000000() {
    expProgramOf(BASES[147]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffe0000000000000000000000000000000000000() {
    expProgramOf(BASES[148]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffc0000000000000000000000000000000000000() {
    expProgramOf(BASES[149]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffff80000000000000000000000000000000000000() {
    expProgramOf(BASES[150]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffff00000000000000000000000000000000000000() {
    expProgramOf(BASES[151]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffe00000000000000000000000000000000000000() {
    expProgramOf(BASES[152]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffffc00000000000000000000000000000000000000() {
    expProgramOf(BASES[153]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffff800000000000000000000000000000000000000() {
    expProgramOf(BASES[154]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffff000000000000000000000000000000000000000() {
    expProgramOf(BASES[155]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffe000000000000000000000000000000000000000() {
    expProgramOf(BASES[156]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffc000000000000000000000000000000000000000() {
    expProgramOf(BASES[157]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffff8000000000000000000000000000000000000000() {
    expProgramOf(BASES[158]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffff0000000000000000000000000000000000000000() {
    expProgramOf(BASES[159]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffe0000000000000000000000000000000000000000() {
    expProgramOf(BASES[160]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffffc0000000000000000000000000000000000000000() {
    expProgramOf(BASES[161]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffff80000000000000000000000000000000000000000() {
    expProgramOf(BASES[162]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffff00000000000000000000000000000000000000000() {
    expProgramOf(BASES[163]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffe00000000000000000000000000000000000000000() {
    expProgramOf(BASES[164]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffc00000000000000000000000000000000000000000() {
    expProgramOf(BASES[165]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffff800000000000000000000000000000000000000000() {
    expProgramOf(BASES[166]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffff000000000000000000000000000000000000000000() {
    expProgramOf(BASES[167]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffe000000000000000000000000000000000000000000() {
    expProgramOf(BASES[168]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffffc000000000000000000000000000000000000000000() {
    expProgramOf(BASES[169]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffff8000000000000000000000000000000000000000000() {
    expProgramOf(BASES[170]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffff0000000000000000000000000000000000000000000() {
    expProgramOf(BASES[171]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffe0000000000000000000000000000000000000000000() {
    expProgramOf(BASES[172]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffc0000000000000000000000000000000000000000000() {
    expProgramOf(BASES[173]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffff80000000000000000000000000000000000000000000() {
    expProgramOf(BASES[174]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffff00000000000000000000000000000000000000000000() {
    expProgramOf(BASES[175]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffe00000000000000000000000000000000000000000000() {
    expProgramOf(BASES[176]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffffc00000000000000000000000000000000000000000000() {
    expProgramOf(BASES[177]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffff800000000000000000000000000000000000000000000() {
    expProgramOf(BASES[178]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffff000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[179]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffe000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[180]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffc000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[181]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffff8000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[182]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffff0000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[183]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffe0000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[184]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffffc0000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[185]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffff80000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[186]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffff00000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[187]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffe00000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[188]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffc00000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[189]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffff800000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[190]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffff000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[191]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffe000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[192]).run();
  }

  @Test
  void expTestForBase_fffffffffffffffc000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[193]).run();
  }

  @Test
  void expTestForBase_fffffffffffffff8000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[194]).run();
  }

  @Test
  void expTestForBase_fffffffffffffff0000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[195]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffe0000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[196]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffc0000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[197]).run();
  }

  @Test
  void expTestForBase_ffffffffffffff80000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[198]).run();
  }

  @Test
  void expTestForBase_ffffffffffffff00000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[199]).run();
  }

  @Test
  void expTestForBase_fffffffffffffe00000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[200]).run();
  }

  @Test
  void expTestForBase_fffffffffffffc00000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[201]).run();
  }

  @Test
  void expTestForBase_fffffffffffff800000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[202]).run();
  }

  @Test
  void expTestForBase_fffffffffffff000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[203]).run();
  }

  @Test
  void expTestForBase_ffffffffffffe000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[204]).run();
  }

  @Test
  void expTestForBase_ffffffffffffc000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[205]).run();
  }

  @Test
  void expTestForBase_ffffffffffff8000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[206]).run();
  }

  @Test
  void expTestForBase_ffffffffffff0000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[207]).run();
  }

  @Test
  void expTestForBase_fffffffffffe0000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[208]).run();
  }

  @Test
  void expTestForBase_fffffffffffc0000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[209]).run();
  }

  @Test
  void expTestForBase_fffffffffff80000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[210]).run();
  }

  @Test
  void expTestForBase_fffffffffff00000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[211]).run();
  }

  @Test
  void expTestForBase_ffffffffffe00000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[212]).run();
  }

  @Test
  void expTestForBase_ffffffffffc00000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[213]).run();
  }

  @Test
  void expTestForBase_ffffffffff800000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[214]).run();
  }

  @Test
  void expTestForBase_ffffffffff000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[215]).run();
  }

  @Test
  void expTestForBase_fffffffffe000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[216]).run();
  }

  @Test
  void expTestForBase_fffffffffc000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[217]).run();
  }

  @Test
  void expTestForBase_fffffffff8000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[218]).run();
  }

  @Test
  void expTestForBase_fffffffff0000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[219]).run();
  }

  @Test
  void expTestForBase_ffffffffe0000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[220]).run();
  }

  @Test
  void expTestForBase_ffffffffc0000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[221]).run();
  }

  @Test
  void expTestForBase_ffffffff80000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[222]).run();
  }

  @Test
  void expTestForBase_ffffffff00000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[223]).run();
  }

  @Test
  void expTestForBase_fffffffe00000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[224]).run();
  }

  @Test
  void expTestForBase_fffffffc00000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[225]).run();
  }

  @Test
  void expTestForBase_fffffff800000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[226]).run();
  }

  @Test
  void expTestForBase_fffffff000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[227]).run();
  }

  @Test
  void expTestForBase_ffffffe000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[228]).run();
  }

  @Test
  void expTestForBase_ffffffc000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[229]).run();
  }

  @Test
  void expTestForBase_ffffff8000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[230]).run();
  }

  @Test
  void expTestForBase_ffffff0000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[231]).run();
  }

  @Test
  void expTestForBase_fffffe0000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[232]).run();
  }

  @Test
  void expTestForBase_fffffc0000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[233]).run();
  }

  @Test
  void expTestForBase_fffff80000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[234]).run();
  }

  @Test
  void expTestForBase_fffff00000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[235]).run();
  }

  @Test
  void expTestForBase_ffffe00000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[236]).run();
  }

  @Test
  void expTestForBase_ffffc00000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[237]).run();
  }

  @Test
  void expTestForBase_ffff800000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[238]).run();
  }

  @Test
  void expTestForBase_ffff000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[239]).run();
  }

  @Test
  void expTestForBase_fffe000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[240]).run();
  }

  @Test
  void expTestForBase_fffc000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[241]).run();
  }

  @Test
  void expTestForBase_fff8000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[242]).run();
  }

  @Test
  void expTestForBase_fff0000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[243]).run();
  }

  @Test
  void expTestForBase_ffe0000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[244]).run();
  }

  @Test
  void expTestForBase_ffc0000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[245]).run();
  }

  @Test
  void expTestForBase_ff80000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[246]).run();
  }

  @Test
  void expTestForBase_ff00000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[247]).run();
  }

  @Test
  void expTestForBase_fe00000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[248]).run();
  }

  @Test
  void expTestForBase_fc00000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[249]).run();
  }

  @Test
  void expTestForBase_f800000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[250]).run();
  }

  @Test
  void expTestForBase_f000000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[251]).run();
  }

  @Test
  void expTestForBase_e000000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[252]).run();
  }

  @Test
  void expTestForBase_c000000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[253]).run();
  }

  @Test
  void expTestForBase_8000000000000000000000000000000000000000000000000000000000000000() {
    expProgramOf(BASES[254]).run();
  }

  @Test
  void expTestForBase_0() {
    expProgramOf(BASES[255]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[256]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[257]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[258]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[259]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[260]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[261]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[262]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[263]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[264]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[265]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[266]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[267]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[268]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[269]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[270]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[271]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[272]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffffff() {
    expProgramOf(BASES[273]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffffff() {
    expProgramOf(BASES[274]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffffff() {
    expProgramOf(BASES[275]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffffff() {
    expProgramOf(BASES[276]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffffff() {
    expProgramOf(BASES[277]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffffff() {
    expProgramOf(BASES[278]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffffff() {
    expProgramOf(BASES[279]).run();
  }

  @Test
  void expTestForBase_ffffffffffffffff() {
    expProgramOf(BASES[280]).run();
  }

  @Test
  void expTestForBase_ffffffffffffff() {
    expProgramOf(BASES[281]).run();
  }

  @Test
  void expTestForBase_ffffffffffff() {
    expProgramOf(BASES[282]).run();
  }

  @Test
  void expTestForBase_ffffffffff() {
    expProgramOf(BASES[283]).run();
  }

  @Test
  void expTestForBase_ffffffff() {
    expProgramOf(BASES[284]).run();
  }

  @Test
  void expTestForBase_ffffff() {
    expProgramOf(BASES[285]).run();
  }

  @Test
  void expTestForBase_ffff() {
    expProgramOf(BASES[286]).run();
  }

  @Test
  void expTestForBase_ff() {
    expProgramOf(BASES[287]).run();
  }

  @Test
  void expTestForBase_f076b857fa9947c1f9ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[288]).run();
  }

  @Test
  void expTestForBase_76b857fa9947c1f9ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[289]).run();
  }

  @Test
  void expTestForBase_b857fa9947c1f9ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[290]).run();
  }

  @Test
  void expTestForBase_57fa9947c1f9ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[291]).run();
  }

  @Test
  void expTestForBase_fa9947c1f9ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[292]).run();
  }

  @Test
  void expTestForBase_9947c1f9ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[293]).run();
  }

  @Test
  void expTestForBase_47c1f9ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[294]).run();
  }

  @Test
  void expTestForBase_c1f9ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[295]).run();
  }

  @Test
  void expTestForBase_f9ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[296]).run();
  }

  @Test
  void expTestForBase_ec558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[297]).run();
  }

  @Test
  void expTestForBase_558262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[298]).run();
  }

  @Test
  void expTestForBase_8262c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[299]).run();
  }

  @Test
  void expTestForBase_62c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[300]).run();
  }

  @Test
  void expTestForBase_c72704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[301]).run();
  }

  @Test
  void expTestForBase_2704099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[302]).run();
  }

  @Test
  void expTestForBase_4099ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[303]).run();
  }

  @Test
  void expTestForBase_99ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[304]).run();
  }

  @Test
  void expTestForBase_9ca8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[305]).run();
  }

  @Test
  void expTestForBase_a8cd325566f73fb99238102ed171() {
    expProgramOf(BASES[306]).run();
  }

  @Test
  void expTestForBase_cd325566f73fb99238102ed171() {
    expProgramOf(BASES[307]).run();
  }

  @Test
  void expTestForBase_325566f73fb99238102ed171() {
    expProgramOf(BASES[308]).run();
  }

  @Test
  void expTestForBase_5566f73fb99238102ed171() {
    expProgramOf(BASES[309]).run();
  }

  @Test
  void expTestForBase_66f73fb99238102ed171() {
    expProgramOf(BASES[310]).run();
  }

  @Test
  void expTestForBase_f73fb99238102ed171() {
    expProgramOf(BASES[311]).run();
  }

  @Test
  void expTestForBase_3fb99238102ed171() {
    expProgramOf(BASES[312]).run();
  }

  @Test
  void expTestForBase_b99238102ed171() {
    expProgramOf(BASES[313]).run();
  }

  @Test
  void expTestForBase_9238102ed171() {
    expProgramOf(BASES[314]).run();
  }

  @Test
  void expTestForBase_38102ed171() {
    expProgramOf(BASES[315]).run();
  }

  @Test
  void expTestForBase_102ed171() {
    expProgramOf(BASES[316]).run();
  }

  @Test
  void expTestForBase_2ed171() {
    expProgramOf(BASES[317]).run();
  }

  @Test
  void expTestForBase_d171() {
    expProgramOf(BASES[318]).run();
  }

  @Test
  void expTestForBase_71() {
    expProgramOf(BASES[319]).run();
  }

  // Disabled tests due to length of time to run
  @Disabled
  @ParameterizedTest
  @MethodSource("expWithEvenBaseAndComplexExponentTestSource")
  public void expWithEvenBaseAndComplexExponentTest(String base, String exponent) {
    expProgramOf(base, exponent).run();
  }

  static Stream<Arguments> expWithEvenBaseAndComplexExponentTestSource() {
    return generateSource(EVEN_BASES, COMPLEX_EXPONENTS);
  }

  @Disabled
  @ParameterizedTest
  @MethodSource("expWithEvenBaseAndSpecialExponentTestSource")
  public void expWithEvenBaseAndSpecialExponentTest(String base, String exponent) {
    expProgramOf(base, exponent).run();
  }

  static Stream<Arguments> expWithEvenBaseAndSpecialExponentTestSource() {
    return generateSource(EVEN_BASES, SPECIAL_EXPONENTS);
  }

  @Disabled
  @ParameterizedTest
  @MethodSource("expWithSimpleOddBaseAndComplexExponentTestSource")
  public void expWithSimpleOddBaseAndComplexExponentTest(String base, String exponent) {
    expProgramOf(base, exponent).run();
  }

  static Stream<Arguments> expWithSimpleOddBaseAndComplexExponentTestSource() {
    return generateSource(SIMPLE_ODD_BASES, COMPLEX_EXPONENTS);
  }

  @Disabled
  @ParameterizedTest
  @MethodSource("expWithSimpleOddBaseAndSpecialExponentTestSource")
  public void expWithSimpleOddBaseAndSpecialExponentTest(String base, String exponent) {
    expProgramOf(base, exponent).run();
  }

  static Stream<Arguments> expWithSimpleOddBaseAndSpecialExponentTestSource() {
    return generateSource(SIMPLE_ODD_BASES, SPECIAL_EXPONENTS);
  }

  @Disabled
  @ParameterizedTest
  @MethodSource("expWithOtherOddBaseAndComplexExponentTestSource")
  public void expWithOtherOddBaseAndComplexExponentTest(String base, String exponent) {
    expProgramOf(base, exponent).run();
  }

  static Stream<Arguments> expWithOtherOddBaseAndComplexExponentTestSource() {
    return generateSource(OTHER_ODD_BASES, COMPLEX_EXPONENTS);
  }

  @Disabled
  @ParameterizedTest
  @MethodSource("expWithOtherOddBaseAndSpecialExponentTestSource")
  public void expWithOtherOddBaseAndSpecialExponentTest(String base, String exponent) {
    expProgramOf(base, exponent).run();
  }

  static Stream<Arguments> expWithOtherOddBaseAndSpecialExponentTestSource() {
    return generateSource(OTHER_ODD_BASES, SPECIAL_EXPONENTS);
  }

  // Support methods
  private BytecodeRunner expProgramOf(String base, String exponent) {
    return BytecodeRunner.of(
        BytecodeCompiler.newProgram().push(exponent).push(base).op(OpCode.EXP).compile());
  }

  private BytecodeRunner expProgramOf(String base) {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    for (String exponent : EXPONENTS) {
      program.push(exponent).push(base).op(OpCode.EXP);
    }
    return BytecodeRunner.of(program.compile());
  }

  static Stream<Arguments> generateSource(String[] bases, String[] exponents) {
    List<Arguments> arguments = new ArrayList<>();
    for (String base : bases) {
      for (String exponent : exponents) {
        arguments.add(Arguments.of(base, exponent));
      }
    }
    return arguments.stream();
  }
}
