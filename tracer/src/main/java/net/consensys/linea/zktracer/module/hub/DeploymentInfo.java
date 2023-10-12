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

import java.util.HashMap;
import java.util.Map;

import org.hyperledger.besu.datatypes.Address;

/** Stores information relative to contract deployment. */
public class DeploymentInfo {
  private final Map<Address, Integer> deploymentNumber = new HashMap<>();
  private final Map<Address, Boolean> isDeploying = new HashMap<>();

  /**
   * Returns the deployment number of the given address; sets it to zero if it is the first
   * deployment of this address.
   *
   * @param address the address to get information for
   * @return the deployment number for this address
   */
  public final int number(Address address) {
    return this.deploymentNumber.getOrDefault(address, 0);
  }

  void deploy(Address address) {
    this.deploymentNumber.put(address, this.number(address) + 1);
  }

  public final boolean isDeploying(Address address) {
    return this.isDeploying.getOrDefault(address, false);
  }

  public final void markDeploying(Address address) {
    this.deploy(address);
    this.isDeploying.put(address, true);
  }

  public final void unmarkDeploying(Address address) {
    this.isDeploying.put(address, false);
  }
}
