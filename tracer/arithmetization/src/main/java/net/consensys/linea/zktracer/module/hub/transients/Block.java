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

package net.consensys.linea.zktracer.module.hub.transients;

import lombok.Getter;
import lombok.experimental.Accessors;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

/** Stores block-specific information. */
@Accessors(fluent = true)
@Getter
public class Block {
  private int blockNumber = 0;
  private Address coinbaseAddress;
  private Wei baseFee;

  /**
   * Update block-specific information on new block entrance.
   *
   * @param processableBlockHeader the processable block header
   */
  public void update(
      final ProcessableBlockHeader processableBlockHeader, final Address miningBeneficiary) {
    this.blockNumber++;
    this.coinbaseAddress = miningBeneficiary;
    this.baseFee = Wei.fromQuantity(processableBlockHeader.getBaseFee().orElseThrow());
  }
}
