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

package net.consensys.linea.blockcapture.reapers;

import java.util.HashMap;
import java.util.List;

import net.consensys.linea.blockcapture.snapshots.BlockHashSnapshot;
import org.hyperledger.besu.datatypes.Hash;

public class BlockHashReaper {
  private final HashMap<Long, Hash> touchedBlockHashes = new HashMap<>();

  public void touch(final long blockNumber, Hash blockHash) {
    this.touchedBlockHashes.put(blockNumber, blockHash);
  }

  /**
   * Collapse recorded set of touched block hashes down into a list of block hash snapshots.
   *
   * @return List of block hash snapshots
   */
  public List<BlockHashSnapshot> collapse() {
    return this.touchedBlockHashes.entrySet().stream()
        .map(e -> BlockHashSnapshot.of(e.getKey(), e.getValue()))
        .toList();
  }
}
