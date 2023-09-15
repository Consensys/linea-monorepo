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
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;

public record Create2(
    GasCalculator gc, MessageFrame frame, long initCodeOffset, long initCodeLength)
    implements GasProjection {
  @Override
  public long staticGas() {
    return GasConstants.G_CREATE.cost();
  }

  @Override
  public long memoryExpansion() {
    return gc.memoryExpansionGasCost(frame, initCodeOffset, initCodeLength);
  }

  @Override
  public long linearPerWord() {
    return linearCost(GasConstants.G_KECCAK_256_WORD.cost(), initCodeLength, 32);
  }

  @Override
  public long rawStipend() {
    long currentGas = frame.getRemainingGas();
    long gasCost = this.staticGas() + this.memoryExpansion() + this.linearPerWord();

    if (gasCost > currentGas) {
      return 0;
    } else {
      return currentGas - currentGas / 64;
    }
  }
}
