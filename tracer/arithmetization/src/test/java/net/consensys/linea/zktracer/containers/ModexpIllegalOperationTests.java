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

package net.consensys.linea.zktracer.containers;

import static java.lang.Integer.MAX_VALUE;
import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.limits.precompiles.ModexpEffectiveCall;
import org.junit.jupiter.api.Test;

public class ModexpIllegalOperationTests {
  @Test
  void legalThenTwoIllegals() {
    final ZkTracer state = new ZkTracer();
    final ModexpEffectiveCall countingOnlyModule = state.getHub().modexpEffectiveCall();

    countingOnlyModule.addPrecompileLimit(1);

    countingOnlyModule.addPrecompileLimit(MAX_VALUE);
    assertThat(countingOnlyModule.lineCount()).isEqualTo(MAX_VALUE);

    countingOnlyModule.addPrecompileLimit(MAX_VALUE);
    assertThat(countingOnlyModule.lineCount()).isEqualTo(MAX_VALUE);

    countingOnlyModule.popTransactionBundle();
    assertThat(countingOnlyModule.lineCount()).isEqualTo(0);
  }

  @Test
  void legalIllegalLegal() {
    final ZkTracer state = new ZkTracer();
    final ModexpEffectiveCall countingOnlyModule = state.getHub().modexpEffectiveCall();

    countingOnlyModule.addPrecompileLimit(1);

    countingOnlyModule.addPrecompileLimit(MAX_VALUE);
    assertThat(countingOnlyModule.lineCount()).isEqualTo(MAX_VALUE);

    countingOnlyModule.addPrecompileLimit(1);
    assertThat(countingOnlyModule.lineCount()).isEqualTo(MAX_VALUE);

    countingOnlyModule.popTransactionBundle();
    assertThat(countingOnlyModule.lineCount()).isEqualTo(0);
  }

  @Test
  void TwoIllegals() {
    final ZkTracer state = new ZkTracer();
    final ModexpEffectiveCall countingOnlyModule = state.getHub().modexpEffectiveCall();

    countingOnlyModule.addPrecompileLimit(MAX_VALUE);
    assertThat(countingOnlyModule.lineCount()).isEqualTo(MAX_VALUE);

    countingOnlyModule.addPrecompileLimit(MAX_VALUE);
    assertThat(countingOnlyModule.lineCount()).isEqualTo(MAX_VALUE);

    countingOnlyModule.popTransactionBundle();
    assertThat(countingOnlyModule.lineCount()).isEqualTo(0);
  }
}
