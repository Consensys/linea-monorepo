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
import java.util.Map;
import java.util.Optional;
import java.util.SequencedMap;
import java.util.UUID;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonValue;
import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.rpc.methods.LineaSendBundle;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.parameters.UnsignedLongParameter;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.rlp.BytesValueRLPOutput;

/** TransactionBundle class representing a collection of pending transactions with metadata. */
@Accessors(fluent = true)
@Getter
@EqualsAndHashCode
public class TransactionBundle {
  private final Hash bundleIdentifier;

  private final List<PendingBundleTx> pendingTransactions;
  private final Long blockNumber;
  private final Optional<Long> minTimestamp;
  private final Optional<Long> maxTimestamp;
  private final Optional<List<Hash>> revertingTxHashes;
  private final Optional<UUID> replacementUUID;

  public TransactionBundle(
      final Hash bundleIdentifier,
      final List<Transaction> transactions,
      final Long blockNumber,
      final Optional<Long> minTimestamp,
      final Optional<Long> maxTimestamp,
      final Optional<List<Hash>> revertingTxHashes,
      final Optional<UUID> replacementUUID) {
    this.bundleIdentifier = bundleIdentifier;
    this.pendingTransactions = transactions.stream().map(PendingBundleTx::new).toList();
    this.blockNumber = blockNumber;
    this.minTimestamp = minTimestamp;
    this.maxTimestamp = maxTimestamp;
    this.revertingTxHashes = revertingTxHashes;
    this.replacementUUID = replacementUUID;
  }

  public LineaSendBundle.BundleParameter toBundleParameter() {
    return new LineaSendBundle.BundleParameter(
        pendingTransactions.stream().map(PendingBundleTx::toBase64String).toList(),
        new UnsignedLongParameter(blockNumber),
        minTimestamp,
        maxTimestamp,
        revertingTxHashes,
        replacementUUID.map(UUID::toString),
        Optional.empty());
  }

  @JsonValue
  public Map<Hash, LineaSendBundle.BundleParameter> serialize() {
    return Map.of(bundleIdentifier, toBundleParameter());
  }

  @JsonCreator
  public static TransactionBundle deserialize(
      final SequencedMap<Hash, LineaSendBundle.BundleParameter> serialized) {
    final var entry = serialized.firstEntry();
    final var hash = entry.getKey();
    final var parameters = entry.getValue();

    return new TransactionBundle(
        hash,
        parameters.txs().stream().map(Bytes::fromBase64String).map(Transaction::readFrom).toList(),
        parameters.blockNumber(),
        parameters.minTimestamp(),
        parameters.maxTimestamp(),
        parameters.revertingTxHashes(),
        parameters.replacementUUID().map(UUID::fromString));
  }

  /** A pending transaction contained in a bundle. */
  public class PendingBundleTx
      extends org.hyperledger.besu.ethereum.eth.transactions.PendingTransaction.Local {

    public PendingBundleTx(final Transaction transaction) {
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

    String toBase64String() {
      final var rlpOutput = new BytesValueRLPOutput();
      getTransaction().writeTo(rlpOutput);
      return rlpOutput.encoded().toBase64String();
    }
  }
}
