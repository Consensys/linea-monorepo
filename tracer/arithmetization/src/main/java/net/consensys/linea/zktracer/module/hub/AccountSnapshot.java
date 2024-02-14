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

import java.util.Optional;

import net.consensys.linea.zktracer.types.Bytecode;
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
    return fromAccount(Optional.ofNullable(account), warm, deploymentNumber, deploymentStatus);
  }

  public static AccountSnapshot empty(
      boolean warm, int deploymentNumber, boolean deploymentStatus) {
    return new AccountSnapshot(
        Address.ZERO, 0, Wei.ZERO, warm, Bytecode.EMPTY, deploymentNumber, deploymentStatus);
  }

  public static AccountSnapshot fromAccount(
      Optional<Account> account, boolean warm, int deploymentNumber, boolean deploymentStatus) {

    return account
        .map(
            a ->
                new AccountSnapshot(
                    a.getAddress(),
                    a.getNonce(),
                    a.getBalance().copy(),
                    warm,
                    new Bytecode(a.getCode().copy()),
                    deploymentNumber,
                    deploymentStatus))
        .orElseGet(() -> AccountSnapshot.empty(warm, deploymentNumber, deploymentStatus));
  }

  public AccountSnapshot debit(Wei quantity) {
    return new AccountSnapshot(
        this.address,
        this.nonce + 1,
        this.balance.subtract(quantity),
        this.warm,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }

  public AccountSnapshot deploy(Wei value) {
    return new AccountSnapshot(
        this.address,
        this.nonce + 1,
        this.balance.add(value),
        this.warm,
        this.code,
        this.deploymentNumber + 1,
        this.deploymentStatus);
  }

  public AccountSnapshot credit(Wei value) {
    return new AccountSnapshot(
        this.address,
        this.nonce,
        this.balance.add(value),
        true,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }
}
