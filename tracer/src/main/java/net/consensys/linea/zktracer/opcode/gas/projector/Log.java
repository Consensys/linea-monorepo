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

import static org.hyperledger.besu.evm.internal.Words.clampedToLong;

import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

public final class Log implements GasProjection {
  private final MessageFrame frame;
  private long offset = 0;
  private long size = 0;
  private final int numTopics;

  public Log(MessageFrame frame, int numTopics) {
    this.frame = frame;
    this.numTopics = numTopics;
    if (frame.stackSize() > 1) {
      this.offset = clampedToLong(frame.getStackItem(0));
      this.size = clampedToLong(frame.getStackItem(1));
    }
  }

  @Override
  public long staticGas() {
    switch (numTopics) {
      case 0 -> {
        return GasConstants.G_LOG_0.cost();
      }
      case 1 -> {
        return GasConstants.G_LOG_1.cost();
      }
      case 2 -> {
        return GasConstants.G_LOG_2.cost();
      }
      case 3 -> {
        return GasConstants.G_LOG_3.cost();
      }
      case 4 -> {
        return GasConstants.G_LOG_4.cost();
      }
      default -> throw new IllegalStateException("Unexpected value: " + numTopics);
    }
  }

  @Override
  public long memoryExpansion() {
    return gc.memoryExpansionGasCost(frame, offset, size);
  }

  @Override
  public long largestOffset() {
    return size == 0 ? 0 : Words.clampedAdd(offset, size);
  }

  @Override
  public long linearPerByte() {
    return linearCost(GasConstants.G_LOG_DATA.cost(), size, 1);
  }
}
