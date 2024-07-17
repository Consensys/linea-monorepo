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

import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;

@RequiredArgsConstructor
public class EcPairingG2MembershipCalls implements Module {
  private final EcPairingFinalExponentiations ecPairingFinalExponentiations;

  @Override
  public String moduleKey() {
    return "PRECOMPILE_ECPAIRING_G2_MEMBERSHIP_CALLS";
  }

  @Override
  public void enterTransaction() {}

  @Override
  public void popTransaction() {}

  @Override
  public int lineCount() {
    long g2MembershipTests = 0;

    for (EcPairingTallier count : this.ecPairingFinalExponentiations.counts()) {
      g2MembershipTests += count.numberOfG2MembershipTests();
    }

    if (g2MembershipTests > Integer.MAX_VALUE) {
      throw new RuntimeException("Ludicrous amount of " + moduleKey());
    }

    return (int) g2MembershipTests;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    throw new UnsupportedOperationException("should never be called");
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    throw new UnsupportedOperationException("should never be called");
  }
}
