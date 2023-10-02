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
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

public final class Sha3 implements GasProjection {
  private final MessageFrame frame;
  private long offset = 0;
  private long length = 0;
  private int bitLength = 0;

  public Sha3(MessageFrame frame) {
    this.frame = frame;
    if (frame.stackSize() >= 2) {
      Bytes biLength = frame.getStackItem(1);
      this.offset = clampedToLong(frame.getStackItem(0));
      this.length = clampedToLong(biLength);
      this.bitLength = biLength.bitLength();
    }
  }

  @Override
  public long staticGas() {
    return GasConstants.G_KECCAK_256.cost();
  }

  @Override
  public long memoryExpansion() {

    return gc.memoryExpansionGasCost(frame, offset, length);
  }

  @Override
  public long largestOffset() {
    return clampedAdd(this.offset, this.length);
  }

  @Override
  public long linearPerWord() {
    return linearCost(GasConstants.G_KECCAK_256_WORD.cost(), this.bitLength, 32);
  }
}
