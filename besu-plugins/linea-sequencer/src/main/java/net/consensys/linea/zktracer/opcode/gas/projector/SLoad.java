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

public record SLoad(GasCalculator gc, MessageFrame frame) implements GasProjection {
  @Override
  public long storageWarmth() {
    final UInt256 key = UInt256.fromBytes(frame.getStackItem(0));
    if (frame.isStorageWarm(frame.getRecipientAddress(), key)) {
      return GasConstants.G_WARM_ACCESS.cost();
    } else {
      return GasConstants.G_COLD_S_LOAD.cost();
    }
  }
}
