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

public final class SLoad implements GasProjection {
  private final MessageFrame frame;
  private UInt256 key = null;

  public SLoad(MessageFrame frame) {
    this.frame = frame;
    if (frame.stackSize() > 0) {
      this.key = UInt256.fromBytes(frame.getStackItem(0));
    }
  }

  @Override
  public long storageWarmth() {
    if (key == null) {
      return 0;
    } else {
      if (frame.isStorageWarm(frame.getRecipientAddress(), key)) {
        return GasConstants.G_WARM_ACCESS.cost();
      } else {
        return GasConstants.G_COLD_S_LOAD.cost();
      }
    }
  }
}
