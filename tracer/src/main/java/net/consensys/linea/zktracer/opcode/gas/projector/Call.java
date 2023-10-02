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

package net.consensys.linea.zktracer.opcode.gas.projector;

import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

public record Call(
    MessageFrame frame,
    long stipend,
    long inputDataOffset,
    long inputDataLength,
    long returnDataOffset,
    long returnDataLength,
    Wei value,
    Account recipient,
    Address to)
    implements GasProjection {
  public static Call invalid() {
    return new Call(null, 0, 0, 0, 0, 0, Wei.ZERO, null, null);
  }

  boolean isInvalid() {
    return this.frame == null;
  }

  @Override
  public long memoryExpansion() {
    if (this.isInvalid()) {
      return 0;
    }

    return Math.max(
        gc.memoryExpansionGasCost(frame, inputDataOffset, inputDataLength),
        gc.memoryExpansionGasCost(frame, returnDataOffset, returnDataLength));
  }

  @Override
  public long largestOffset() {
    if (this.isInvalid()) {
      return 0;
    }

    return Math.max(
        Words.clampedAdd(inputDataOffset, inputDataLength),
        Words.clampedAdd(returnDataOffset, returnDataLength));
  }

  @Override
  public long accountAccess() {
    if (this.isInvalid()) {
      return 0;
    }

    if (frame.isAddressWarm(to)) {
      return GasConstants.G_WARM_ACCESS.cost();
    } else {
      return GasConstants.G_COLD_ACCOUNT_ACCESS.cost();
    }
  }

  @Override
  public long accountCreation() {
    if (this.isInvalid()) {
      return 0;
    }

    if ((recipient == null || recipient.isEmpty()) && !value.isZero()) {
      return GasConstants.G_NEW_ACCOUNT.cost();
    } else {
      return 0L;
    }
  }

  @Override
  public long transferValue() {
    if (this.isInvalid()) {
      return 0;
    }

    if (value.isZero()) {
      return 0L;
    } else {
      return GasConstants.G_CALL_VALUE.cost();
    }
  }

  @Override
  public long rawStipend() {
    if (this.isInvalid()) {
      return 0;
    }

    final long cost = memoryExpansion() + accountAccess() + accountCreation() + transferValue();
    if (cost > frame.getRemainingGas()) {
      return 0L;
    } else {
      final long remaining = frame.getRemainingGas() - cost;
      final long weird = remaining - remaining / 64;

      return Math.min(weird, stipend);
    }
  }

  @Override
  public long extraStipend() {
    if (this.isInvalid()) {
      return 0;
    }

    if (value.isZero()) {
      return 0L;
    } else {
      return GasConstants.G_CALL_STIPEND.cost();
    }
  }
}
