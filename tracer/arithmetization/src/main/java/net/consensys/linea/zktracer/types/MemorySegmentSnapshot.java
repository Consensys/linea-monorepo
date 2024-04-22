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

package net.consensys.linea.zktracer.types;

import java.util.Arrays;

import lombok.Getter;
import lombok.experimental.Accessors;

@Accessors(fluent = true)
public class MemorySegmentSnapshot {
  @Getter private UnsignedByte[] memory;
  private boolean clean;

  public MemorySegmentSnapshot(UnsignedByte[] memory) {
    this(memory, true);
  }

  private MemorySegmentSnapshot(UnsignedByte[] memory, boolean clean) {
    this.memory = memory;
    this.clean = clean;
  }

  public UnsignedByte[] limbAtIndex(final int index) {
    UnsignedByte[] limb = UnsignedByte.EMPTY_BYTES16;
    int limbIndex = Math.min(16 * (index + 1), memory.length) - 16 * index;

    if (limbIndex >= 0) {
      System.arraycopy(memory, 16 * index, limb, 0, limbIndex);
    }

    return limb;
  }

  public void updateLimb(int limbIndex, UnsignedByte[] valNew) {
    if (clean) {
      clean = false;
      this.memory = Arrays.copyOf(memory, memory.length);
    }

    int potNewLen = (limbIndex + 1) * 16;
    expand(potNewLen);

    int copyLen = potNewLen - 1 - (limbIndex * 16);
    System.arraycopy(valNew, 0, memory, limbIndex * 16, copyLen);
  }

  /**
   * Should be called once per RAM macro-instruction.
   *
   * @param potNewLen expanded memory length
   */
  private void expand(int potNewLen) {
    if (potNewLen <= memory.length) {
      return;
    }

    this.memory = Arrays.copyOf(memory, memory.length + potNewLen);
  }
}
