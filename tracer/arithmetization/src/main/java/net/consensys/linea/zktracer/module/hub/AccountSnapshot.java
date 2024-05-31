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

import com.google.common.base.Preconditions;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.types.Bytecode;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;

@AllArgsConstructor
@Getter
@Setter
@Accessors(fluent = true)
public class AccountSnapshot {
  private Address address;
  private long nonce;
  private Wei balance;
  private boolean isWarm;
  private Bytecode code;
  private int deploymentNumber;
  private boolean deploymentStatus;

  public AccountSnapshot decrementBalance(Wei quantity) {
    Preconditions.checkState(
        this.balance.greaterOrEqualThan(quantity),
        "Insufficient balance: %s".formatted(this.balance));
    this.balance = this.balance.subtract(quantity);
    return this;
  }

  public AccountSnapshot incrementBalance(Wei quantity) {
    this.balance = this.balance.add(quantity);
    return this;
  }

  public AccountSnapshot incrementNonce() {
    this.nonce++;
    return this;
  }

  public static AccountSnapshot fromAccount(
      Account account, boolean isWarm, int deploymentNumber, boolean deploymentStatus) {
    return fromAccount(Optional.ofNullable(account), isWarm, deploymentNumber, deploymentStatus);
  }

  public static AccountSnapshot empty(
      boolean isWarm, int deploymentNumber, boolean deploymentStatus) {
    return new AccountSnapshot(
        Address.ZERO, 0, Wei.ZERO, isWarm, Bytecode.EMPTY, deploymentNumber, deploymentStatus);
  }

  public static AccountSnapshot fromAccount(
      Optional<Account> account, boolean isWarm, int deploymentNumber, boolean deploymentStatus) {

    return account
        .map(
            a ->
                new AccountSnapshot(
                    a.getAddress(),
                    a.getNonce(),
                    a.getBalance().copy(),
                    isWarm,
                    new Bytecode(a.getCode().copy()),
                    deploymentNumber,
                    deploymentStatus))
        .orElseGet(() -> AccountSnapshot.empty(isWarm, deploymentNumber, deploymentStatus));
  }

  public AccountSnapshot debit(Wei quantity) {
    return new AccountSnapshot(
        this.address,
        this.nonce + 1,
        this.balance.subtract(quantity),
        this.isWarm,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }

  public AccountSnapshot debit(Wei quantity, boolean isWarm) {
    return new AccountSnapshot(
        this.address,
        this.nonce + 1,
        this.balance.subtract(quantity),
        isWarm,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }

  public AccountSnapshot deploy(Wei value) {
    return new AccountSnapshot(
        this.address,
        this.nonce + 1,
        this.balance.add(value),
        this.isWarm,
        this.code,
        this.deploymentNumber + 1,
        this.deploymentStatus);
  }

  public AccountSnapshot deploy(Wei value, Bytecode code) {
    Preconditions.checkState(
        !this.deploymentStatus, "Deployment status should be false before deploying.");
    return new AccountSnapshot(
        this.address,
        this.nonce + 1,
        this.balance.add(value),
        true,
        code,
        this.deploymentNumber + 1,
        true);
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

  public AccountSnapshot credit(Wei value, boolean isWarm) {
    return new AccountSnapshot(
        this.address,
        this.nonce,
        this.balance.add(value),
        isWarm,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }
}
