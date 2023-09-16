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

package net.consensys.linea.zktracer.module.hub;

import lombok.Getter;
import lombok.experimental.Accessors;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.BlockHeader;

/** Stores block-specific information. */
@Accessors(fluent = true)
@Getter
public class BlockInfo {
  int blockNumber = 0;
  Address minerAddress;
  Wei baseFee;

  /**
   * Update block-specific information on new block entrance.
   *
   * @param blockHeader the new block header
   */
  void update(final BlockHeader blockHeader) {
    this.blockNumber++;
    this.minerAddress = blockHeader.getCoinbase();
    this.baseFee = Wei.of(blockHeader.getBaseFee().get().getAsBigInteger());
  }
}
