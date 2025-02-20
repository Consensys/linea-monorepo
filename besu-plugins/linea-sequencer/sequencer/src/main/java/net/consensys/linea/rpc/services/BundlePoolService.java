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
package net.consensys.linea.rpc.services;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import lombok.Getter;
import lombok.experimental.Accessors;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.services.BesuService;

public interface BundlePoolService extends BesuService {

  /** TransactionBundle record representing a collection of pending transactions with metadata. */
  @Accessors(fluent = true)
  @Getter
  class TransactionBundle {
    private final Hash bundleIdentifier;
    private final List<PendingBundleTx> pendingTransactions;
    private final Long blockNumber;
    private final Optional<Long> minTimestamp;
    private final Optional<Long> maxTimestamp;
    private final Optional<List<Hash>> revertingTxHashes;

    public TransactionBundle(
        final Hash bundleIdentifier,
        final List<Transaction> transactions,
        final Long blockNumber,
        final Optional<Long> minTimestamp,
        final Optional<Long> maxTimestamp,
        final Optional<List<Hash>> revertingTxHashes) {
      this.bundleIdentifier = bundleIdentifier;
      this.pendingTransactions = transactions.stream().map(PendingBundleTx::new).toList();
      this.blockNumber = blockNumber;
      this.minTimestamp = minTimestamp;
      this.maxTimestamp = maxTimestamp;
      this.revertingTxHashes = revertingTxHashes;
    }

    /** A pending transaction contained in a bundle. */
    public class PendingBundleTx
        extends org.hyperledger.besu.ethereum.eth.transactions.PendingTransaction.Local {

      public PendingBundleTx(final org.hyperledger.besu.ethereum.core.Transaction transaction) {
        super(transaction);
      }

      public TransactionBundle getBundle() {
        return TransactionBundle.this;
      }

      public boolean isBundleStart() {
        return getBundle().pendingTransactions().getFirst().equals(this);
      }

      @Override
      public String toTraceLog() {
        return "Bundle tx: " + super.toTraceLog();
      }
    }
  }

  /**
   * Retrieves a list of TransactionBundles associated with a block number.
   *
   * @param blockNumber The block number to look up.
   * @return A list of TransactionBundles for the given block number, or an empty list if none are
   *     found.
   */
  List<TransactionBundle> getBundlesByBlockNumber(long blockNumber);

  /**
   * Retrieves a TransactionBundle by its unique hash identifier.
   *
   * @param hash The hash identifier of the TransactionBundle.
   * @return The TransactionBundle associated with the hash, or null if not found.
   */
  TransactionBundle get(Hash hash);

  /**
   * Retrieves a TransactionBundle by its replacement UUID
   *
   * @param replacementUUID identifier of the TransactionBundle.
   * @return The TransactionBundle associated with the uuid, or null if not found.
   */
  TransactionBundle get(UUID replacementUUID);

  /**
   * Puts or replaces an existing TransactionBundle in the cache and updates the block index.
   *
   * @param hash The hash identifier of the TransactionBundle.
   * @param bundle The new TransactionBundle to replace the existing one.
   */
  void putOrReplace(Hash hash, TransactionBundle bundle);

  /**
   * Puts or replaces an existing TransactionBundle by UUIDin the cache and updates the block index.
   *
   * @param replacementUUID identifier of the TransactionBundle.
   * @param bundle The new TransactionBundle to replace the existing one.
   */
  void putOrReplace(UUID replacementUUID, TransactionBundle bundle);

  /**
   * removes an existing TransactionBundle in the cache and updates the block index.
   *
   * @param replacementUUID identifier of the TransactionBundle.
   * @return boolean indicating if bundle was found and removed
   */
  boolean remove(UUID replacementUUID);

  /**
   * removes an existing TransactionBundle in the cache and updates the block index.
   *
   * @param hash The hash identifier of the TransactionBundle.
   * @return boolean indicating if bundle was found and removed
   */
  boolean remove(Hash hash);

  /**
   * Removes all TransactionBundles associated with the given block number. First removes them from
   * the block index, then removes them from the cache.
   *
   * @param blockNumber The block number whose bundles should be removed.
   */
  void removeByBlockNumber(long blockNumber);

  /**
   * Get the number of bundles in the pool
   *
   * @return the number of bundles in the pool
   */
  long size();
}
