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

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import net.consensys.linea.rpc.services.BundlePoolService.TransactionBundle;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.data.AddedBlockContext;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.BesuEvents;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class LineaLimitedBundlePoolTest {

  private LineaLimitedBundlePool pool;
  private BesuEvents eventService;
  private AddedBlockContext addedBlockContext;
  private BlockHeader blockHeader;

  @BeforeEach
  void setUp() {
    eventService = mock(BesuEvents.class);
    addedBlockContext = mock(AddedBlockContext.class);
    pool = new LineaLimitedBundlePool(10_000L, eventService); // Max 100 entries, 10 KB size
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
  void smokeTestRemoveByBlockNumber() {
    Hash hash1 = Hash.fromHexStringLenient("0x1234");
    Hash hash2 = Hash.fromHexStringLenient("0x5678");
    TransactionBundle bundle1 = createBundle(hash1, 1);
    TransactionBundle bundle2 = createBundle(hash2, 1);

    pool.putOrReplace(hash1, bundle1);
    pool.putOrReplace(hash2, bundle2);

    pool.removeByBlockNumber(1);

    assertNull(pool.get(hash1), "Bundle1 should be removed from the cache");
    assertNull(pool.get(hash2), "Bundle2 should be removed from the cache");
    assertTrue(
        pool.getBundlesByBlockNumber(1).isEmpty(), "Block index for block 1 should be empty");
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

  private TransactionBundle createBundle(Hash hash, long blockNumber) {
    return createBundle(hash, blockNumber, Collections.emptyList());
  }

  private TransactionBundle createBundle(Hash hash, long blockNumber, List<Transaction> maybeTxs) {
    return new TransactionBundle(
        hash, maybeTxs, blockNumber, Optional.empty(), Optional.empty(), Optional.empty());
  }
}
