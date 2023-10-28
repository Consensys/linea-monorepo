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

package net.consensys.linea.zktracer.module.hub;

import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;

/**
 * Records the aborting conditions that may happen during a CALL or a CREATE.
 *
 * @param callStackOverflow too many nested contexts
 * @param balanceTooLow trying to give more ETH than the caller has
 */
public record Aborts(boolean callStackOverflow, boolean balanceTooLow) {
  public static Aborts forFrame(Hub hub) {
    return new Aborts(
        hub.callStack().wouldOverflow(),
        switch (hub.currentFrame().opCode()) {
          case CALL, CALLCODE -> {
            if (hub.exceptions().none()) {
              final Address myAddress = hub.currentFrame().address();
              final Wei myBalance =
                  hub.messageFrame().getWorldUpdater().getAccount(myAddress).getBalance();
              final Wei value = Wei.of(UInt256.fromBytes(hub.messageFrame().getStackItem(2)));

              yield value.greaterThan(myBalance);
            } else {
              yield false;
            }
          }
          case CREATE, CREATE2 -> {
            if (hub.exceptions().none()) {
              final Address myAddress = hub.currentFrame().address();
              final Wei myBalance =
                  hub.messageFrame().getWorldUpdater().getAccount(myAddress).getBalance();
              final Wei value = Wei.of(UInt256.fromBytes(hub.messageFrame().getStackItem(0)));

              yield value.greaterThan(myBalance);
            } else {
              yield false;
            }
          }
          default -> false;
        });
  }

  public Aborts snapshot() {
    return new Aborts(this.callStackOverflow, this.balanceTooLow);
  }

  public boolean none() {
    return !this.any();
  }

  public boolean any() {
    return this.callStackOverflow || this.balanceTooLow;
  }
}
