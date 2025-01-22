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
import java.util.Set;

import net.consensys.linea.blockcapture.snapshots.AccountSnapshot;
import net.consensys.linea.blockcapture.snapshots.BlockHashSnapshot;
import net.consensys.linea.blockcapture.snapshots.BlockSnapshot;
import net.consensys.linea.blockcapture.snapshots.ConflationSnapshot;
import net.consensys.linea.blockcapture.snapshots.StorageSnapshot;
import net.consensys.linea.blockcapture.snapshots.TransactionResultSnapshot;
import net.consensys.linea.blockcapture.snapshots.TransactionSnapshot;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;
import org.hyperledger.besu.evm.worldstate.WorldView;
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
  /** Collect storage locations read / written by the entire conflation */
  private final StorageReaper conflationStorage = new StorageReaper();

  /** Collect the account address read / written by the entire conflation */
  private final AddressReaper conflationAddresses = new AddressReaper();

  /** Collection all block hashes read during the conflation * */
  private final BlockHashReaper conflationHashes = new BlockHashReaper();

  /** Collect the blocks within a conflation */
  private final List<BlockSnapshot> blocks = new ArrayList<>();

  /**
   * Transaction index is used to do determine the current transaction being processed within the
   * current block.
   */
  private int txIndex;

  /** Collect storage locations read / written by the current transaction */
  private StorageReaper txStorage = null;

  /** Collect the account address read / written by the current transaction */
  private AddressReaper txAddresses = null;

  public void enterBlock(
      final BlockHeader header, final BlockBody body, final Address miningBeneficiary) {
    this.blocks.add(
        BlockSnapshot.of((org.hyperledger.besu.ethereum.core.BlockHeader) header, body));
    this.conflationAddresses.touch(miningBeneficiary);
    txIndex = 0; // reset
  }

  public void prepareTransaction(Transaction tx) {
    // Configure tx-local reapers
    this.txStorage = new StorageReaper();
    this.txAddresses = new AddressReaper();
    this.touchAddress(tx.getSender());
    tx.getTo().ifPresent(this::touchAddress);
  }

  public void exitTransaction(
      WorldView world,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed,
      Set<Address> selfDestructs) {
    // Identify relevant transaction snapshot
    TransactionSnapshot txSnapshot = blocks.getLast().txs().get(txIndex);
    // Convert logs into hex strings
    List<String> logStrings = logs.stream().map(l -> l.getData().toHexString()).toList();
    // Convert destructed account addresses into into hex strings
    List<String> destructStrings = selfDestructs.stream().map(Address::toHexString).toList();
    // Collapse accounts
    final List<AccountSnapshot> accounts = this.txAddresses.collapse(world);
    // Collapse storage
    final List<StorageSnapshot> storage = txStorage.collapse(world);
    // Construct the result snapshot
    TransactionResultSnapshot resultSnapshot =
        new TransactionResultSnapshot(
            status, output.toHexString(), logStrings, gasUsed, accounts, storage, destructStrings);
    // Record the outcome
    txSnapshot.setTransactionResult(resultSnapshot);
    // Reset for next transaction
    txIndex++;
    this.txStorage = null;
    this.txAddresses = null;
  }

  public void touchAddress(final Address address) {
    this.conflationAddresses.touch(address);
    this.txAddresses.touch(address);
  }

  public void touchStorage(final Address address, final UInt256 key) {
    this.conflationStorage.touch(address, key);
    this.txStorage.touch(address, key);
  }

  public void touchBlockHash(final long blockNumber, Hash blockHash) {
    this.conflationHashes.touch(blockNumber, blockHash);
    // No need to tx local hashes, since they are a global concept.
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
    // Collapse accounts
    final List<AccountSnapshot> accounts = this.conflationAddresses.collapse(world);
    // Collapse storage
    final List<StorageSnapshot> storage = conflationStorage.collapse(world);
    // Collapse block hashes
    final List<BlockHashSnapshot> hashes = conflationHashes.collapse();
    // Done
    return new ConflationSnapshot(this.blocks, accounts, storage, hashes);
  }
}
