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
import net.consensys.linea.zktracer.types.AddressUtils;
import net.consensys.linea.zktracer.types.Bytecode;
import org.apache.tuweni.bytes.Bytes;
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

  // TODO: we require a MARKED_FOR_SELFDESTRUCT boolean
  //  The implementation will be
  //  1. is (address, deploymentNumber) ∈ effectiveSelfdestructsMap Then
  //       - get the relevant selfDestructTime
  //       - MARKED_FOR_SELFDESTRUCT     = [hubStamp > selfDestructTime]
  //       - MARKED_FOR_SELFDESTRUCT_NEW = hubStamp ≥ selfDestructTime

  // public enum AccountRowType {
  //   STANDARD,
  //   DEFERRED,
  // }

  // TODO: is there a "canonical" way to take a snapshot fo an account
  //  where getWorldUpdater().getAccount(address) return null ?
  public static AccountSnapshot canonical(Hub hub, Address address) {
    final Account account = hub.messageFrame().getWorldUpdater().getAccount(address);
    boolean isPrecompile = AddressUtils.isPrecompile(address);
    if (account != null) {
      return new AccountSnapshot(
          address,
          account.getNonce(),
          account.getBalance(),
          hub.messageFrame().isAddressWarm(address),
          new Bytecode(account.getCode()),
          hub.transients.conflation().deploymentInfo().number(address),
          hub.transients.conflation().deploymentInfo().isDeploying(address));
    } else {
      return new AccountSnapshot(
          address,
          0,
          Wei.ZERO,
          hub.messageFrame().isAddressWarm(address),
          new Bytecode(Bytes.EMPTY),
          hub.transients.conflation().deploymentInfo().number(address),
          hub.transients.conflation().deploymentInfo().isDeploying(address));
    }
  }

  public AccountSnapshot decrementBalance(Wei quantity) {
    Preconditions.checkState(
        this.balance.greaterOrEqualThan(quantity),
        "Insufficient balance\n     Address: %s\n     Balance: %s\n     Value: %s"
            .formatted(this.address, this.balance, quantity));
    this.balance = this.balance.subtract(quantity);
    return this;
  }

  public AccountSnapshot incrementBalance(Wei quantity) {
    this.balance = this.balance.add(quantity);
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

  public static AccountSnapshot fromAddress(
      Address address, boolean isWarm, int deploymentNumber, boolean deploymentStatus) {
    return new AccountSnapshot(
        address, 0, Wei.ZERO, isWarm, Bytecode.EMPTY, deploymentNumber, deploymentStatus);
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
        this.nonce,
        this.balance.subtract(quantity),
        this.isWarm,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }

  public AccountSnapshot turnOnWarmth() {
    return new AccountSnapshot(
        this.address,
        this.nonce,
        this.balance,
        true,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }

  public AccountSnapshot raiseNonce() {
    return new AccountSnapshot(
        this.address,
        this.nonce + 1,
        this.balance,
        this.isWarm,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }

  // TODO: does this update the deployment number in the deploymentInfo object ?
  public AccountSnapshot initiateDeployment(Wei value, Bytecode code, int updatedDeploymentNumber) {
    Preconditions.checkState(
        !this.deploymentStatus,
        "Deployment status should be false before initiating a deployment.");
    return new AccountSnapshot(
        this.address,
        this.nonce + 1,
        this.balance.add(value),
        true,
        code,
        updatedDeploymentNumber,
        true);
  }

  public AccountSnapshot deployByteCode(Bytecode code) {
    Preconditions.checkState(
        this.deploymentStatus, "Deployment status should be true before deploying byte code.");

    return new AccountSnapshot(
        this.address, this.nonce, this.balance, true, code, this.deploymentNumber, false);
  }

  public AccountSnapshot credit(Wei value) {
    return new AccountSnapshot(
        this.address,
        this.nonce,
        this.balance.add(value),
        this.isWarm,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }

  // TODO: confirm with Tsvetan if this creates a proper deep copy
  public AccountSnapshot deepCopy() {
    return new AccountSnapshot(
        this.address,
        this.nonce,
        this.balance,
        this.isWarm,
        this.code,
        this.deploymentNumber,
        this.deploymentStatus);
  }

  public AccountSnapshot wipe() {
    return new AccountSnapshot(
        this.address, 0, Wei.of(0), this.isWarm, Bytecode.EMPTY, this.deploymentNumber + 1, false);
  }
}
