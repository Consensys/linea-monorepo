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

import com.google.common.base.Preconditions;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.limits.CountingOnlyModule;

@RequiredArgsConstructor
public final class EcRecoverEffectiveCall extends CountingOnlyModule {

  @Override
  public String moduleKey() {
    return "PRECOMPILE_ECRECOVER_EFFECTIVE_CALLS";
  }

  @Override
  public void addPrecompileLimit(final int numberEffectiveCall) {
    Preconditions.checkArgument(
        numberEffectiveCall == 1, "can't add more than one effective precompile call at a time");
    counts.add(numberEffectiveCall);
  }
}
