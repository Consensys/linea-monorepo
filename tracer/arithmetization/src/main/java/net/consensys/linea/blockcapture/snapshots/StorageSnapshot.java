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

package net.consensys.linea.blockcapture.snapshots;

import java.util.Optional;

import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.worldstate.WorldView;

public record StorageSnapshot(String address, String key, String value) {
  public static Optional<StorageSnapshot> from(
      Address address, UInt256 key, final WorldView world) {
    return Optional.ofNullable(world.get(address))
        .map(
            account ->
                new StorageSnapshot(
                    address.toHexString(),
                    key.toHexString(),
                    account.getStorageValue(key).toHexString()));
  }
}
