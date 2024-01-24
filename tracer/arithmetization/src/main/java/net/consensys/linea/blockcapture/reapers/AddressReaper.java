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
import java.util.Collection;
import java.util.HashSet;
import java.util.Set;
import java.util.stream.Collectors;

import org.hyperledger.besu.datatypes.Address;

public class AddressReaper {
  private final ArrayDeque<Set<Address>> reaped = new ArrayDeque<>();

  public AddressReaper() {
    // “Bedrock” address set for block-level gathering.
    this.reaped.addLast(new HashSet<>());
  }

  public void enterTransaction() {
    this.reaped.addLast(new HashSet<>());
  }

  public void exitTransaction(boolean success) {
    if (!success) {
      this.reaped.removeLast();
    }
  }

  public void touch(final Address... addresses) {
    for (Address address : addresses) {
      this.reaped.peekLast().add(address);
    }
  }

  public Set<Address> collapse() {
    return this.reaped.stream().flatMap(Collection::stream).collect(Collectors.toSet());
  }
}
