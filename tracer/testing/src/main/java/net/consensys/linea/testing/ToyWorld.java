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

package net.consensys.linea.testing;

import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import lombok.Builder;
import lombok.Getter;
import lombok.Singular;
import net.consensys.linea.blockcapture.snapshots.AccountSnapshot;
import net.consensys.linea.blockcapture.snapshots.ConflationSnapshot;
import net.consensys.linea.blockcapture.snapshots.StorageSnapshot;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.account.MutableAccount;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.worldstate.AuthorizedCodeService;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;

public class ToyWorld implements WorldUpdater {
  private final ToyWorld parent;
  private final AuthorizedCodeService authorizedCodeService;
  @Getter private Map<Address, ToyAccount> addressAccountMap;

  private ToyWorld() {
    this(null);
  }

  private ToyWorld(final ToyWorld parent) {
    this(parent, Collections.emptyList());
  }

  @Builder
  private ToyWorld(final ToyWorld parent, @Singular final List<ToyAccount> accounts) {
    this.parent = parent;
    this.addressAccountMap = new HashMap<>();
    this.authorizedCodeService = new AuthorizedCodeService();
    // Initialise the account map
    for (ToyAccount account : accounts) {
      addressAccountMap.put(account.getAddress(), account);
    }
  }

  public static ToyWorld empty() {
    return builder().build();
  }

  public static ToyWorld of(final ConflationSnapshot conflation) {
    ToyWorld worldUpdater = new ToyWorld();
    initWorldUpdater(worldUpdater, conflation);
    return worldUpdater;
  }

  @Override
  public WorldUpdater updater() {
    return new ToyWorld(this);
  }

  @Override
  public Account get(final Address address) {
    if (addressAccountMap.containsKey(address)) {
      return authorizedCodeService.processAccount(this, addressAccountMap.get(address), address);
    } else if (parent != null) {
      return authorizedCodeService.processAccount(this, parent.get(address), address);
    } else {
      return authorizedCodeService.processAccount(this, null, address);
    }
  }

  @Override
  public MutableAccount createAccount(final Address address, final long nonce, final Wei balance) {
    return createAccount(null, address, nonce, balance, Bytes.EMPTY);
  }

  public MutableAccount createAccount(
      final Account parentAccount,
      final Address address,
      final long nonce,
      final Wei balance,
      final Bytes code) {

    ToyAccount account =
        ToyAccount.builder()
            .parent(parentAccount)
            .code(code)
            .address(address)
            .nonce(nonce)
            .balance(balance)
            .build();

    addressAccountMap.put(address, account);

    return authorizedCodeService.processMutableAccount(this, account, address);
  }

  @Override
  public MutableAccount getAccount(final Address address) {
    if (addressAccountMap.containsKey(address)) {
      return authorizedCodeService.processMutableAccount(
          this, addressAccountMap.get(address), address);
    } else if (parent != null) {
      Account parentAccount = parent.getAccount(address);
      if (parentAccount == null) {
        return authorizedCodeService.processMutableAccount(this, null, address);
      } else {
        return createAccount(
            parentAccount,
            parentAccount.getAddress(),
            parentAccount.getNonce(),
            parentAccount.getBalance(),
            parentAccount.getCode());
      }
    } else {
      return authorizedCodeService.processMutableAccount(this, null, address);
    }
  }

  @Override
  public void deleteAccount(final Address address) {
    addressAccountMap.put(address, null);
  }

  @Override
  public Collection<? extends Account> getTouchedAccounts() {
    return addressAccountMap.values();
  }

  @Override
  public Collection<Address> getDeletedAccountAddresses() {
    return addressAccountMap.entrySet().stream()
        .filter(e -> e.getValue() == null)
        .map(Map.Entry::getKey)
        .toList();
  }

  @Override
  public void revert() {
    addressAccountMap = new HashMap<>();
  }

  @Override
  public void commit() {
    addressAccountMap.forEach(
        (address, account) -> {
          if (!account.updateParent()) {
            parent.addressAccountMap.put(address, account);
          }
        });
  }

  @Override
  public Optional<WorldUpdater> parentUpdater() {
    return Optional.ofNullable(parent);
  }

  /**
   * Initialise a world updater given a conflation. Observe this can be applied to any WorldUpdater,
   * such as SimpleWorld.
   *
   * @param world The world to be initialised.
   * @param conflation The conflation from which to initialise.
   */
  private static void initWorldUpdater(WorldUpdater world, final ConflationSnapshot conflation) {
    for (AccountSnapshot account : conflation.accounts()) {
      // Construct contract address
      Address addr = Address.fromHexString(account.address());
      // Create account
      MutableAccount acc =
          world.createAccount(
              Words.toAddress(addr), account.nonce(), Wei.fromHexString(account.balance()));
      // Update code
      acc.setCode(Bytes.fromHexString(account.code()));
    }
    //
    for (StorageSnapshot s : conflation.storage()) {
      world
          .getAccount(Words.toAddress(Bytes.fromHexString(s.address())))
          .setStorageValue(UInt256.fromHexString(s.key()), UInt256.fromHexString(s.value()));
    }
  }
}
