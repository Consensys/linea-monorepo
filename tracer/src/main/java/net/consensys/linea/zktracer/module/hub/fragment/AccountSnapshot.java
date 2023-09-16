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

package net.consensys.linea.zktracer.module.hub.fragment;

import net.consensys.linea.zktracer.module.hub.Bytecode;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;

public record AccountSnapshot(
    Address address,
    long nonce,
    Wei balance,
    boolean warm,
    Bytecode code,
    int deploymentNumber,
    boolean deploymentStatus) {
  public static AccountSnapshot fromAccount(
      Account account, boolean warm, int deploymentNumber, boolean deploymentStatus) {
    if (account == null) {
      return new AccountSnapshot(
          Address.ZERO, 0, Wei.ZERO, warm, Bytecode.EMPTY, deploymentNumber, deploymentStatus);
    }

    return new AccountSnapshot(
        account.getAddress(),
        account.getNonce(),
        account.getBalance(),
        warm,
        new Bytecode(account.getCode()),
        deploymentNumber,
        deploymentStatus);
  }
}
