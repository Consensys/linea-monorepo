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
package net.consensys.linea.zktracer.module.blockdata;

import static net.consensys.linea.zktracer.MultiBlockUtils.multiBlocksTest;
import static net.consensys.linea.zktracer.module.blockdata.NextGasLimitScenario.*;
import static net.consensys.linea.zktracer.module.blockdata.Trace.GAS_LIMIT_MAXIMUM;
import static net.consensys.linea.zktracer.module.blockdata.Trace.GAS_LIMIT_MINIMUM;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.*;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@ExtendWith(UnitTestWatcher.class)
public class GasLimitTest {

  @Test
  void legalGasLimitVariationsTest() {
    Bytes p = BytecodeCompiler.newProgram().push(1).compile();

    long gasLimit = GAS_LIMIT_MINIMUM;
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_SAME)));
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_INCREMENT)));
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_MAX)));

    gasLimit = 100_000_000L;
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_SAME)));
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_INCREMENT)));
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_DECREMENT)));
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_MAX)));
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_MIN)));

    gasLimit = GAS_LIMIT_MAXIMUM;
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_SAME)));
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_DECREMENT)));
    multiBlocksTest(List.of(p, p), List.of(gasLimit, nextGasLimit(gasLimit, IN_RANGE_MIN)));
  }

  /**
   * The test below deliberately contains test vectors that are expected to fail. The test
   * is @Disabled as a consequence.
   *
   * @param gasLimit
   * @param nextGasLimitScenario
   */
  @Disabled
  @ParameterizedTest
  @MethodSource("blockDataVariableGasLimitTestSource")
  void variableGasLimitTest(long gasLimit, NextGasLimitScenario nextGasLimitScenario) {
    Bytes program = BytecodeCompiler.newProgram().op(OpCode.STOP).compile();

    multiBlocksTest(
        List.of(program, program), List.of(gasLimit, nextGasLimit(gasLimit, nextGasLimitScenario)));
  }

  private static Stream<Arguments> blockDataVariableGasLimitTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    // TODO: use LINEA_BLOCK_GAS_LIMIT_MIN, LINEA_BLOCK_GAS_LIMIT_MAX and something in between,
    //  e.g., 100M
    List<Long> gasLimits =
        List.of(
            0L,
            (long) ETHEREUM_GAS_LIMIT_MINIMUM,
            (long) LINEA_GAS_LIMIT_MINIMUM,
            100_000_000L,
            (long) LINEA_GAS_LIMIT_MAXIMUM,
            Long.MAX_VALUE);
    for (Long gasLimit : gasLimits) {
      for (NextGasLimitScenario scenario : values()) {
        arguments.add(Arguments.of(gasLimit, scenario));
      }
    }
    return arguments.stream();
  }

  private long nextGasLimit(long gasLimit, NextGasLimitScenario nextGasLimitScenario) {
    long maxDeviation = maxDeviation(gasLimit);
    return switch (nextGasLimitScenario) {
      case IN_RANGE_SAME -> gasLimit;
      case IN_RANGE_INCREMENT -> gasLimit + maxDeviation / 2;
      case IN_RANGE_DECREMENT -> gasLimit - maxDeviation / 2;
      case IN_RANGE_MAX -> gasLimit + maxDeviation - 1;
      case IN_RANGE_MIN -> gasLimit - maxDeviation + 1;
      case OUT_OF_RANGE_INCREMENT -> gasLimit + maxDeviation;
      case OUT_OF_RANGE_DECREMENT -> gasLimit - maxDeviation;
      case OUT_OF_RANGE_GENERIC -> 200_000_000L;
    };
  }

  private boolean isInRange(long gasLimit, long nextGasLimit) {
    long maxDeviation = maxDeviation(gasLimit);
    return nextGasLimit < gasLimit + maxDeviation && nextGasLimit > gasLimit - maxDeviation;
  }

  private long maxDeviation(long gasLimit) {
    return gasLimit / GAS_LIMIT_ADJUSTMENT_FACTOR;
  }
}
