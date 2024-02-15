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

package net.consensys.linea.config;

import lombok.Builder;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;

/** The Linea L1 L2 bridge configuration. */
@Builder(toBuilder = true)
public record LineaL1L2BridgeConfiguration(Address contract, Bytes topic) {
  public static final LineaL1L2BridgeConfiguration EMPTY =
      LineaL1L2BridgeConfiguration.builder().contract(Address.ZERO).topic(Bytes.EMPTY).build();

  public boolean isEmpty() {
    return this.contract.equals(Address.ZERO) || this.topic.isEmpty();
  }
}
