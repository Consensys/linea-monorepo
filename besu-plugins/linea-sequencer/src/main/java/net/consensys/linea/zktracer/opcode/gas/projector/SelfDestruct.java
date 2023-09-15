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
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.internal.Words;

public record SelfDestruct(GasCalculator gc, MessageFrame frame) implements GasProjection {
  @Override
  public long staticGas() {
    return GasConstants.G_SELF_DESTRUCT.cost();
  }

  @Override
  public long accountAccess() {
    final Address beneficiaryAddress = Words.toAddress(frame.getStackItem(0));
    if (frame.isAddressWarm(beneficiaryAddress)) {
      return 0L;
    } else {
      return GasConstants.G_COLD_ACCOUNT_ACCESS.cost();
    }
  }

  @Override
  public long accountCreation() {
    final Address beneficiaryAddress = Words.toAddress(frame.getStackItem(0));
    final Account beneficiaryAccount = frame.getWorldUpdater().get(beneficiaryAddress);
    final Address me = frame.getRecipientAddress();
    final Wei balance = frame.getWorldUpdater().get(me).getBalance();

    if ((beneficiaryAccount == null || beneficiaryAccount.isEmpty()) && !balance.isZero()) {
      return GasConstants.G_NEW_ACCOUNT.cost();
    } else {
      return 0L;
    }
  }

  @Override
  public long refund() {
    return GasConstants.R_SELF_DESTRUCT.cost();
  }
}
