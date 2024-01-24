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

import java.util.ArrayDeque;
import java.util.Deque;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;

/**
 * This object gathers all non-reversed accesses to storage values during the execution of a
 * conflation, then collapse them into a single mapping of the initial values in these slots.
 */
public class StorageReaper {
  private final Deque<HashMap<Address, Set<UInt256>>> transientStates = new ArrayDeque<>();

  public void enterTransaction() {
    this.transientStates.addLast(new HashMap<>());
  }

  public void exitTransaction(boolean success) {
    if (!success) {
      this.transientStates.removeLast();
    }
  }

  public void touch(final Address address, final UInt256 key) {
    this.transientStates.peekLast().computeIfAbsent(address, k -> new HashSet<>()).add(key);
  }

  public Map<Address, Set<UInt256>> collapse() {
    final Map<Address, Set<UInt256>> r = new HashMap<>();

    for (var txEntry : this.transientStates) {
      for (Map.Entry<Address, Set<UInt256>> addressKeys : txEntry.entrySet()) {
        final Address address = addressKeys.getKey();

        // Use computeIfAbsent instead of put, as we only want to capture the **first** read.
        r.computeIfAbsent(address, k -> new HashSet<>()).addAll(addressKeys.getValue());
      }
    }

    return r;
  }
}
