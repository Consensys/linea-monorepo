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

import static org.hyperledger.besu.evm.internal.Words.clampedAdd;
import static org.hyperledger.besu.evm.internal.Words.clampedToLong;

import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;

public record Sha3(GasCalculator gc, MessageFrame frame) implements GasProjection {
  @Override
  public long staticGas() {
    return GasConstants.G_KECCAK_256.cost();
  }

  @Override
  public long memoryExpansion() {
    long offset = clampedToLong(frame.getStackItem(0));
    long length = clampedToLong(frame.getStackItem(1));
    return gc.memoryExpansionGasCost(frame, offset, length);
  }

  @Override
  public long largestOffset() {
    long offset = clampedToLong(frame.getStackItem(0));
    long length = clampedToLong(frame.getStackItem(1));
    return clampedAdd(offset, length);
  }

  @Override
  public long linearPerWord() {
    return linearCost(GasConstants.G_KECCAK_256_WORD.cost(), frame.getStackItem(1).bitLength(), 32);
  }
}
