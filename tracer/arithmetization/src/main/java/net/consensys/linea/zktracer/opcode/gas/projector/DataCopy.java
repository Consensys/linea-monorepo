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

package net.consensys.linea.zktracer.opcode.gas.projector;

import static org.hyperledger.besu.evm.internal.Words.clampedToLong;

import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

public final class DataCopy implements GasProjection {
  private final MessageFrame frame;
  private long offset = 0;
  private long len = 0;

  public DataCopy(MessageFrame frame) {
    this.frame = frame;
    if (frame.stackSize() > 2) {
      this.offset = clampedToLong(frame.getStackItem(1));
      this.len = clampedToLong(frame.getStackItem(2));
    }
  }

  @Override
  public long staticGas() {
    return gc.getVeryLowTierGasCost();
  }

  @Override
  public long memoryExpansion() {
    return gc.memoryExpansionGasCost(frame, this.offset, this.len);
  }

  @Override
  public long largestOffset() {
    return this.len == 0 ? 0 : Words.clampedAdd(this.offset, this.len);
  }

  @Override
  public long linearPerWord() {
    return linearCost(GasConstants.G_COPY.cost(), this.len, 32);
  }
}
