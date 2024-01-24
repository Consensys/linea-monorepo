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

package net.consensys.linea.blockcapture.reapers;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Set;

import net.consensys.linea.blockcapture.snapshots.AccountSnapshot;
import net.consensys.linea.blockcapture.snapshots.BlockSnapshot;
import net.consensys.linea.blockcapture.snapshots.ConflationSnapshot;
import net.consensys.linea.blockcapture.snapshots.StorageSnapshot;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

/**
 * The Reaper collect all the information from the state that will be accessed during the execution
 * of a conflation.
 *
 * <p>This data can than be collapsed into a “replay” ({@link ConflationSnapshot}), i.e. the minimal
 * required information to replay a conflation as if it were executed on the blockchain.
 */
public class Reaper {
  /** Collect the reads from the state */
  private final StorageReaper storage = new StorageReaper();
  /** Collect the addresses read from the state */
  private final AddressReaper addresses = new AddressReaper();
  /** Collect the blocks within a conflation */
  private final List<BlockSnapshot> blocks = new ArrayList<>();

  public void enterBlock(final BlockHeader header, final BlockBody body) {
    this.blocks.add(
        BlockSnapshot.of((org.hyperledger.besu.ethereum.core.BlockHeader) header, body));
    this.addresses.touch(header.getCoinbase());
  }

  public void enterTransaction(Transaction tx) {
    this.storage.enterTransaction();
    this.addresses.enterTransaction();

    this.touchAddress(tx.getSender());
    tx.getTo().ifPresent(this::touchAddress);
  }

  public void exitTransaction(boolean success) {
    this.storage.exitTransaction(success);
    this.addresses.exitTransaction(success);
  }

  public void touchAddress(final Address address) {
    this.addresses.touch(address);
  }

  public void touchStorage(final Address address, final UInt256 key) {
    this.storage.touch(address, key);
  }

  /**
   * Uniquify and solidify the accumulated data, then return a {@link ConflationSnapshot}, which
   * contains the smallest dataset required to exactly replay the conflation within a test framework
   * without requiring access to the whole state.
   *
   * @param world the state before the conflation execution
   * @return a minimal set of information required to replay the conflation within a test framework
   */
  public ConflationSnapshot collapse(final WorldUpdater world) {
    final List<AccountSnapshot> initialAccounts =
        this.addresses.collapse().stream()
            .flatMap(a -> AccountSnapshot.from(a, world).stream())
            .toList();

    final List<StorageSnapshot> initialStorage = new ArrayList<>();
    for (Map.Entry<Address, Set<UInt256>> e : this.storage.collapse().entrySet()) {
      final Address address = e.getKey();

      e.getValue().stream()
          .flatMap(key -> StorageSnapshot.from(address, key, world).stream())
          .forEach(initialStorage::add);
    }

    return new ConflationSnapshot(this.blocks, initialAccounts, initialStorage);
  }
}
