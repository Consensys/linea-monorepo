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

import static java.nio.charset.StandardCharsets.US_ASCII;
import static net.consensys.linea.rpc.services.LineaLimitedBundlePool.BUNDLE_SAVE_FILENAME;
import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.Mockito.RETURNS_DEEP_STUBS;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECPPrivateKey;
import org.hyperledger.besu.crypto.SECPPublicKey;
import org.hyperledger.besu.crypto.SignatureAlgorithm;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.core.TransactionTestFixture;
import org.hyperledger.besu.ethereum.eth.transactions.PendingTransaction;
import org.hyperledger.besu.plugin.data.AddedBlockContext;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.BesuEvents;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

class LineaLimitedBundlePoolTest {
  private static final KeyPair KEY_PAIR_1 =
      new KeyPair(
          SECPPrivateKey.create(BigInteger.valueOf(Long.MAX_VALUE), SignatureAlgorithm.ALGORITHM),
          SECPPublicKey.create(BigInteger.valueOf(Long.MIN_VALUE), SignatureAlgorithm.ALGORITHM));

  private static final Transaction TX1 =
      new TransactionTestFixture().nonce(0).gasLimit(21000).createTransaction(KEY_PAIR_1);
  private static final Transaction TX2 =
      new TransactionTestFixture().nonce(1).gasLimit(21000).createTransaction(KEY_PAIR_1);
  private static final Transaction TX3 =
      new TransactionTestFixture().nonce(2).gasLimit(21000).createTransaction(KEY_PAIR_1);

  @TempDir Path dataDir;
  private LineaLimitedBundlePool pool;
  private BesuEvents eventService;
  private AddedBlockContext addedBlockContext;
  private BlockHeader blockHeader;
  private BlockchainService blockchainService;

  @BeforeEach
  void setUp() {
    eventService = mock(BesuEvents.class);
    addedBlockContext = mock(AddedBlockContext.class);
    blockchainService = mock(BlockchainService.class, RETURNS_DEEP_STUBS);
    when(blockchainService.getChainHeadHeader().getNumber()).thenReturn(10L);
    pool =
        new LineaLimitedBundlePool(
            dataDir, 10_000L, eventService, blockchainService); // Max 100 entries, 10 KB size
    blockHeader = mock(BlockHeader.class);
  }

  @Test
  void smokeTestPutAndGetByHash() {
    Hash hash = Hash.fromHexStringLenient("0x1234");
    TransactionBundle bundle = createBundle(hash, 1);

    pool.putOrReplace(hash, bundle);
    TransactionBundle retrieved = pool.get(hash);

    assertNotNull(retrieved, "Bundle should be retrieved by hash");
    assertEquals(hash, retrieved.bundleIdentifier(), "Retrieved bundle hash should match");
  }

  @Test
  void smokeTestGetBundlesByBlockNumber() {
    Hash hash1 = Hash.fromHexStringLenient("0x1234");
    Hash hash2 = Hash.fromHexStringLenient("0x5678");
    TransactionBundle bundle1 = createBundle(hash1, 1);
    TransactionBundle bundle2 = createBundle(hash2, 1);

    pool.putOrReplace(hash1, bundle1);
    pool.putOrReplace(hash2, bundle2);

    List<TransactionBundle> bundles = pool.getBundlesByBlockNumber(1);

    assertEquals(2, bundles.size(), "There should be two bundles for block 1");
    assertTrue(bundles.contains(bundle1), "Bundles should contain bundle1");
    assertTrue(bundles.contains(bundle2), "Bundles should contain bundle2");
  }

  @Test
  void testPutAndGetByUUID() {
    UUID uuid = UUID.randomUUID();
    TransactionBundle bundle = createBundle(Hash.ZERO, 1L);

    pool.putOrReplace(uuid, bundle);

    assertNotNull(pool.get(uuid));
    assertEquals(bundle, pool.get(uuid));
  }

  @Test
  void testPutAndGetByHash() {
    Hash hash = Hash.hash(Bytes.fromHexStringLenient("0x1234"));
    TransactionBundle bundle = createBundle(hash, 1L);

    pool.putOrReplace(hash, bundle);

    assertNotNull(pool.get(hash));
    assertEquals(bundle, pool.get(hash));
  }

  @Test
  void testPutAndGet_ThrowsException_WhenFrozen() {
    // saving to disk freeze the pool
    pool.saveToDisk();
    assertTrue(pool.isFrozen());

    Hash hash = Hash.hash(Bytes.fromHexStringLenient("0x1234"));
    TransactionBundle bundle = createBundle(hash, 1L);

    assertThatThrownBy(() -> pool.putOrReplace(hash, bundle))
        .isInstanceOf(IllegalStateException.class)
        .hasMessage("Bundle pool is not accepting modifications");
  }

  @Test
  void testRemoveByUUID() {
    UUID uuid = UUID.randomUUID();
    TransactionBundle bundle = createBundle(Hash.ZERO, 1L);

    pool.putOrReplace(uuid, bundle);

    assertTrue(pool.remove(uuid));
    assertNull(pool.get(uuid));
  }

  @Test
  void testRemoveByHash() {
    Hash hash = Hash.hash(Bytes.fromHexStringLenient("0x5678"));
    TransactionBundle bundle = createBundle(hash, 1L);

    pool.putOrReplace(hash, bundle);

    assertTrue(pool.remove(hash));
    assertNull(pool.get(hash));
  }

  @Test
  void testGetByUUID_NotFound() {
    UUID uuid = UUID.randomUUID();
    assertNull(pool.get(uuid));
  }

  @Test
  void testGetByHash_NotFound() {
    Hash hash = Hash.hash(Bytes.fromHexStringLenient("0x9876"));
    assertNull(pool.get(hash));
  }

  @Test
  void testRemoveByUUID_NotFound() {
    UUID uuid = UUID.randomUUID();
    assertFalse(pool.remove(uuid));
  }

  @Test
  void testRemoveByHash_NotFound() {
    Hash hash = Hash.hash(Bytes.fromHexStringLenient("0xabcd"));
    assertFalse(pool.remove(hash));
  }

  @Test
  void testRemove_ThrowsException_WhenFrozen() {
    // saving to disk freeze the pool
    pool.saveToDisk();
    assertTrue(pool.isFrozen());

    Hash hash = Hash.hash(Bytes.fromHexStringLenient("0xabcd"));
    assertThatThrownBy(() -> pool.remove(hash))
        .isInstanceOf(IllegalStateException.class)
        .hasMessage("Bundle pool is not accepting modifications");
  }

  @Test
  void testOnBlockAdded_RemovesOldBundles() {
    // Prepare old block number
    long oldBlockNumber = 10L;
    long newBlockNumber = 15L;
    Hash mockOldHash = Hash.ZERO;

    // Mock block header behavior
    when(addedBlockContext.getBlockHeader()).thenReturn(blockHeader);
    when(blockHeader.getNumber()).thenReturn(newBlockNumber);

    // Create a fake transaction bundle
    TransactionBundle oldBundle = createBundle(mockOldHash, oldBlockNumber);

    // Manually insert old bundle into the block index
    pool.putOrReplace(mockOldHash, oldBundle);

    // Ensure bundle exists before adding new block
    assert !pool.getBundlesByBlockNumber(oldBlockNumber).isEmpty();

    // Call the method under test
    pool.onBlockAdded(addedBlockContext);

    // Verify that the old bundle is removed
    assert pool.getBundlesByBlockNumber(oldBlockNumber).isEmpty();
  }

  @Test
  void testOnBlockAdded_DoesNothing_WhenFrozen() {
    // Prepare old block number
    long oldBlockNumber = 10L;
    long newBlockNumber = 15L;
    Hash mockOldHash = Hash.ZERO;

    // Mock block header behavior
    when(addedBlockContext.getBlockHeader()).thenReturn(blockHeader);
    when(blockHeader.getNumber()).thenReturn(newBlockNumber);

    // Create a fake transaction bundle
    TransactionBundle oldBundle = createBundle(mockOldHash, oldBlockNumber);

    // Manually insert old bundle into the block index
    pool.putOrReplace(mockOldHash, oldBundle);

    // Ensure bundle exists before adding new block
    assert !pool.getBundlesByBlockNumber(oldBlockNumber).isEmpty();

    // saving to disk freeze the pool
    pool.saveToDisk();
    assertTrue(pool.isFrozen());

    // Call the method under test
    pool.onBlockAdded(addedBlockContext);

    // Verify that the old bundle is still there
    assertThat(pool.getBundlesByBlockNumber(oldBlockNumber)).containsExactly(oldBundle);
  }

  @Test
  void saveToDisk() throws IOException {

    Hash hash1 = Hash.fromHexStringLenient("0x1234");
    TransactionBundle bundle1 = createBundle(hash1, 1, List.of(TX1, TX2));
    pool.putOrReplace(hash1, bundle1);

    Hash hash2 = Hash.fromHexStringLenient("0x5678");
    TransactionBundle bundle2 = createBundle(hash2, 2, List.of(TX3));
    pool.putOrReplace(hash2, bundle2);

    pool.saveToDisk();

    final var saved = Files.readString(dataDir.resolve(BUNDLE_SAVE_FILENAME), US_ASCII);
    assertThat(saved)
        .isEqualTo(
            """
            {"version":1}
            {"0x0000000000000000000000000000000000000000000000000000000000001234":{"blockNumber":1,"txs":["+E+AghOIglIIgASAggqWoHNvbkX5jC5D+Q0GW88l7bP45W+b8oubebJsfXgE+lRzoAVzHPSnS/zQmUxq3Hg9UHQ3p51KWM6dyYuqKVM7HYz7","+E8BghOIglIIgASAggqVoGgwjcqbkx9qWzUse4MmYxq5fGYo617lp3j9YAj74GDhoFrjtX1uTIbDgflVrS1EPJv2jmbGV2NbxukBL0sNVpBf"]}}
            {"0x0000000000000000000000000000000000000000000000000000000000005678":{"blockNumber":2,"txs":["+E8CghOIglIIgASAggqVoMmdnUf+4fBBE+l/IAxacTZhj5elWnFdplP+s4jg92yyoHUWAGDUZ5Vo6dg3q7e9+PyBAkwlk4Fprh1UFmyQhhjx"]}}""");
  }

  @Test
  void loadFromDisk() throws IOException {
    Files.writeString(
        dataDir.resolve(BUNDLE_SAVE_FILENAME),
        """
             {"version":1}
             {"0x0000000000000000000000000000000000000000000000000000000000001234":{"blockNumber":11,"txs":["+E+AghOIglIIgASAggqWoHNvbkX5jC5D+Q0GW88l7bP45W+b8oubebJsfXgE+lRzoAVzHPSnS/zQmUxq3Hg9UHQ3p51KWM6dyYuqKVM7HYz7","+E8BghOIglIIgASAggqVoGgwjcqbkx9qWzUse4MmYxq5fGYo617lp3j9YAj74GDhoFrjtX1uTIbDgflVrS1EPJv2jmbGV2NbxukBL0sNVpBf"]}}
             {"0x0000000000000000000000000000000000000000000000000000000000005678":{"blockNumber":12,"txs":["+E8CghOIglIIgASAggqVoMmdnUf+4fBBE+l/IAxacTZhj5elWnFdplP+s4jg92yyoHUWAGDUZ5Vo6dg3q7e9+PyBAkwlk4Fprh1UFmyQhhjx"]}}""",
        US_ASCII);

    pool.loadFromDisk();

    Hash hash1 = Hash.fromHexStringLenient("0x1234");
    TransactionBundle bundle1 = pool.get(hash1);

    assertThat(bundle1.blockNumber()).isEqualTo(11);
    assertThat(bundle1.bundleIdentifier()).isEqualTo(hash1);
    assertThat(bundle1.pendingTransactions())
        .map(PendingTransaction::getTransaction)
        .map(Transaction::getHash)
        .containsExactly(TX1.getHash(), TX2.getHash());

    Hash hash2 = Hash.fromHexStringLenient("0x5678");

    TransactionBundle bundle2 = pool.get(hash2);

    assertThat(bundle2.blockNumber()).isEqualTo(12);
    assertThat(bundle2.bundleIdentifier()).isEqualTo(hash2);
    assertThat(bundle2.pendingTransactions())
        .map(PendingTransaction::getTransaction)
        .map(Transaction::getHash)
        .containsExactly(TX3.getHash());
  }

  @Test
  void loadFromDisk_UnsupportedVersion() throws IOException {
    Files.writeString(
        dataDir.resolve(BUNDLE_SAVE_FILENAME),
        """
             {"version":0}
             {"0x0000000000000000000000000000000000000000000000000000000000001234":{"blockNumber":11,"txs":["+E+AghOIglIIgASAggqWoHNvbkX5jC5D+Q0GW88l7bP45W+b8oubebJsfXgE+lRzoAVzHPSnS/zQmUxq3Hg9UHQ3p51KWM6dyYuqKVM7HYz7","+E8BghOIglIIgASAggqVoGgwjcqbkx9qWzUse4MmYxq5fGYo617lp3j9YAj74GDhoFrjtX1uTIbDgflVrS1EPJv2jmbGV2NbxukBL0sNVpBf"]}}""",
        US_ASCII);

    pool.loadFromDisk();

    // no bundle should be restored
    assertThat(pool.size()).isEqualTo(0);
  }

  @Test
  void partialLoadFromDisk_DueToInvalidLine() throws IOException {
    Files.writeString(
        dataDir.resolve(BUNDLE_SAVE_FILENAME),
        """
             {"version":1}
             {"0x0000000000000000000000000000000000000000000000000000000000001234":{"blockNumber":11,"txs":["+E+AghOIglIIgASAggqWoHNvbkX5jC5D+Q0GW88l7bP45W+b8oubebJsfXgE+lRzoAVzHPSnS/zQmUxq3Hg9UHQ3p51KWM6dyYuqKVM7HYz7","+E8BghOIglIIgASAggqVoGgwjcqbkx9qWzUse4MmYxq5fGYo617lp3j9YAj74GDhoFrjtX1uTIbDgflVrS1EPJv2jmbGV2NbxukBL0sNVpBf"]}}
             {"0x0000000000000000000000000000000000000000000000000000000000005678":{"blockNumber":"not a number","txs":["+E8CghOIglIIgASAggqVoMmdnUf+4fBBE+l/IAxacTZhj5elWnFdplP+s4jg92yyoHUWAGDUZ5Vo6dg3q7e9+PyBAkwlk4Fprh1UFmyQhhjx"]}}""",
        US_ASCII);

    pool.loadFromDisk();

    assertThat(pool.size()).isEqualTo(1);

    Hash hash1 = Hash.fromHexStringLenient("0x1234");
    TransactionBundle bundle1 = pool.get(hash1);

    assertThat(bundle1.blockNumber()).isEqualTo(11);
    assertThat(bundle1.bundleIdentifier()).isEqualTo(hash1);
    assertThat(bundle1.pendingTransactions())
        .map(PendingTransaction::getTransaction)
        .map(Transaction::getHash)
        .containsExactly(TX1.getHash(), TX2.getHash());

    Hash hash2 = Hash.fromHexStringLenient("0x5678");

    assertThat(pool.get(hash2)).isNull();
  }

  @Test
  void partialLoadFromDisk_DueOldBlockNumber() throws IOException {
    Files.writeString(
        dataDir.resolve(BUNDLE_SAVE_FILENAME),
        """
            {"version":1}
            {"0x0000000000000000000000000000000000000000000000000000000000005678":{"blockNumber":10,"txs":["+E8CghOIglIIgASAggqVoMmdnUf+4fBBE+l/IAxacTZhj5elWnFdplP+s4jg92yyoHUWAGDUZ5Vo6dg3q7e9+PyBAkwlk4Fprh1UFmyQhhjx"]}}
            {"0x0000000000000000000000000000000000000000000000000000000000001234":{"blockNumber":11,"txs":["+E+AghOIglIIgASAggqWoHNvbkX5jC5D+Q0GW88l7bP45W+b8oubebJsfXgE+lRzoAVzHPSnS/zQmUxq3Hg9UHQ3p51KWM6dyYuqKVM7HYz7","+E8BghOIglIIgASAggqVoGgwjcqbkx9qWzUse4MmYxq5fGYo617lp3j9YAj74GDhoFrjtX1uTIbDgflVrS1EPJv2jmbGV2NbxukBL0sNVpBf"]}}""",
        US_ASCII);

    pool.loadFromDisk();

    assertThat(pool.size()).isEqualTo(1);

    Hash hash1 = Hash.fromHexStringLenient("0x1234");
    TransactionBundle bundle1 = pool.get(hash1);

    assertThat(bundle1.blockNumber()).isEqualTo(11);
    assertThat(bundle1.bundleIdentifier()).isEqualTo(hash1);
    assertThat(bundle1.pendingTransactions())
        .map(PendingTransaction::getTransaction)
        .map(Transaction::getHash)
        .containsExactly(TX1.getHash(), TX2.getHash());

    Hash hash2 = Hash.fromHexStringLenient("0x5678");

    assertThat(pool.get(hash2)).isNull();
  }

  private TransactionBundle createBundle(Hash hash, long blockNumber) {
    return createBundle(hash, blockNumber, Collections.emptyList());
  }

  private TransactionBundle createBundle(Hash hash, long blockNumber, List<Transaction> maybeTxs) {
    return new TransactionBundle(
        hash,
        maybeTxs,
        blockNumber,
        Optional.empty(),
        Optional.empty(),
        Optional.empty(),
        Optional.empty());
  }
}
