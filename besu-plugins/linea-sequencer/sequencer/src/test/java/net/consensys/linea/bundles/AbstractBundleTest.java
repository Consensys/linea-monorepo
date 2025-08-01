/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.bundles;

import java.math.BigInteger;
import java.util.List;
import java.util.Optional;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECPPrivateKey;
import org.hyperledger.besu.crypto.SECPPublicKey;
import org.hyperledger.besu.crypto.SignatureAlgorithm;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.core.TransactionTestFixture;

abstract class AbstractBundleTest {
  protected static final KeyPair KEY_PAIR =
      new KeyPair(
          SECPPrivateKey.create(BigInteger.valueOf(Long.MAX_VALUE), SignatureAlgorithm.ALGORITHM),
          SECPPublicKey.create(BigInteger.valueOf(Long.MIN_VALUE), SignatureAlgorithm.ALGORITHM));

  protected static final Transaction TX1 =
      new TransactionTestFixture().nonce(0).gasLimit(21000).createTransaction(KEY_PAIR);
  protected static final Transaction TX2 =
      new TransactionTestFixture().nonce(1).gasLimit(21000).createTransaction(KEY_PAIR);
  protected static final Transaction TX3 =
      new TransactionTestFixture().nonce(2).gasLimit(21000).createTransaction(KEY_PAIR);

  protected TransactionBundle createBundle(
      Hash hash, long blockNumber, List<Transaction> maybeTxs) {
    return createBundle(hash, blockNumber, maybeTxs, false);
  }

  protected TransactionBundle createBundle(
      Hash hash, long blockNumber, List<Transaction> maybeTxs, boolean hasPriority) {
    return new TransactionBundle(
        hash,
        maybeTxs,
        blockNumber,
        Optional.empty(),
        Optional.empty(),
        Optional.empty(),
        Optional.empty(),
        hasPriority);
  }
}
