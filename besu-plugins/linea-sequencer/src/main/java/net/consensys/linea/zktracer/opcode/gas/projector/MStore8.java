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

import org.hyperledger.besu.evm.frame.MessageFrame;

public final class MStore8 implements GasProjection {
  private final MessageFrame frame;
  private long offset = 0;

  public MStore8(MessageFrame frame) {
    this.frame = frame;
    if (frame.stackSize() > 0) {
      this.offset = clampedToLong(frame.getStackItem(0));
    }
  }

  @Override
  public long staticGas() {
    return gc.getVeryLowTierGasCost();
  }

  @Override
  public long memoryExpansion() {
    return gc.memoryExpansionGasCost(this.frame, this.offset, 1);
  }

  @Override
  public long largestOffset() {
    return this.offset;
  }
}
