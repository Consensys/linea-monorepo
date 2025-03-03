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

package net.consensys.linea.zktracer.module.limits.precompiles;

import static net.consensys.linea.zktracer.CurveOperations.*;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.container.module.CountingOnlyModule;
import net.consensys.linea.zktracer.container.stacked.CountOnlyOperation;

@Slf4j
@RequiredArgsConstructor
@Accessors(fluent = true)
@Getter
public final class EcPairingFinalExponentiations implements CountingOnlyModule {
  private final CountOnlyOperation counts = new CountOnlyOperation();
  private static final int PRECOMPILE_BASE_GAS_FEE = 45_000; // cf EIP-1108
  private static final int PRECOMPILE_MILLER_LOOP_GAS_FEE = 34_000; // cf EIP-1108
  private static final int ECPAIRING_NB_BYTES_PER_MILLER_LOOP = 192;
  private static final int ECPAIRING_NB_BYTES_PER_SMALL_POINT = 64;
  private static final int ECPAIRING_NB_BYTES_PER_LARGE_POINT = 128;

  @Override
  public String moduleKey() {
    return "PRECOMPILE_ECPAIRING_FINAL_EXPONENTIATIONS";
  }

  @Override
  public void updateTally(final int numberEffectiveCall) {
    Preconditions.checkArgument(
        numberEffectiveCall <= 1, "can't add more than one effective precompile call at a time");
    counts.add(numberEffectiveCall);
  }
}
