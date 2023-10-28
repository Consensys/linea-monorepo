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

package net.consensys.linea.zktracer.opcode.gas.projector;

import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

public final class AccountAccess implements GasProjection {
  private final MessageFrame frame;
  private Address target = null;

  public AccountAccess(MessageFrame frame) {
    if (frame.stackSize() > 0) {
      this.target = Words.toAddress(frame.getStackItem(0));
    }
    this.frame = frame;
  }

  boolean isInvalid() {
    return this.target == null;
  }

  @Override
  public long accountAccess() {
    if (this.isInvalid()) {
      return 0;
    }

    if (frame.isAddressWarm(this.target)) {
      return gc.getWarmStorageReadCost();
    } else {
      return gc.getColdAccountAccessCost();
    }
  }
}
