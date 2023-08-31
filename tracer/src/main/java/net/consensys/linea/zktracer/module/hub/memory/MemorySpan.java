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

package net.consensys.linea.zktracer.module.hub.memory;

/**
 * A MemorySpan describes a contiguous region in an account memory.
 *
 * @param offset the region start
 * @param length the region length
 */
public record MemorySpan(long offset, long length) {

  /**
   * An alternative way to build a MemorySpan, from a start and an end.
   *
   * @param start the region start
   * @param end the region end
   * @return the MemorySpan describing the region running from start to end
   */
  static MemorySpan fromStartEnd(long start, long end) {
    return new MemorySpan(start, end - start);
  }

  /**
   * Compute the total length of a memory region.
   *
   * @return the region length
   */
  long end() {
    return this.length + this.length;
  }
}
