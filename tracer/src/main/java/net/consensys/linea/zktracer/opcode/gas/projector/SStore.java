/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.opcode.gas.projector;

import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;

public record SStore(
    GasCalculator gc,
    MessageFrame frame,
    UInt256 key,
    UInt256 originalValue,
    UInt256 currentValue,
    UInt256 newValue)
    implements GasProjection {
  @Override
  public long storageWarmth() {
    if (frame.isStorageWarm(frame.getRecipientAddress(), key)) {
      return 0L;
    } else {
      return GasConstants.G_COLD_S_LOAD.cost();
    }
  }

  @Override
  public long sStoreValue() {
    if (newValue.equals(currentValue) || !originalValue.equals(currentValue)) {
      return GasConstants.G_WARM_ACCESS.cost();
    } else {
      return originalValue.isZero() ? GasConstants.G_S_SET.cost() : GasConstants.G_S_RESET.cost();
    }
  }

  @Override
  public long refund() {
    long rDirtyClear = 0;
    if (!originalValue.isZero() && currentValue.isZero()) {
      rDirtyClear = -GasConstants.R_S_CLEAR.cost();
    }
    if (!originalValue.isZero() && newValue.isZero()) {
      rDirtyClear = GasConstants.R_S_CLEAR.cost();
    }

    long rDirtyReset = 0;
    if (originalValue.equals(newValue) && originalValue.isZero()) {
      rDirtyReset = GasConstants.G_S_SET.cost() - GasConstants.G_WARM_ACCESS.cost();
    }
    if (originalValue.equals(newValue) && !originalValue.isZero()) {
      rDirtyReset = GasConstants.G_S_RESET.cost() - GasConstants.G_WARM_ACCESS.cost();
    }

    long r = 0;
    if (!currentValue.equals(newValue) && currentValue.equals(originalValue) && newValue.isZero()) {
      r = GasConstants.R_S_CLEAR.cost();
    }
    if (!currentValue.equals(newValue) && !currentValue.equals(originalValue)) {
      r = rDirtyClear + rDirtyReset;
    }

    return r;
  }
}
