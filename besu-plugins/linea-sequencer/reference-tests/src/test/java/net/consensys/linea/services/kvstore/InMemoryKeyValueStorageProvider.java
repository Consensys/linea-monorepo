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

package net.consensys.linea.services.kvstore;

import org.hyperledger.besu.ethereum.chain.DefaultBlockchain;
import org.hyperledger.besu.ethereum.chain.MutableBlockchain;
import org.hyperledger.besu.ethereum.chain.VariablesStorage;
import org.hyperledger.besu.ethereum.core.Block;
import org.hyperledger.besu.ethereum.core.BlockHeaderFunctions;
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions;
import org.hyperledger.besu.ethereum.storage.keyvalue.KeyValueStoragePrefixedKeyBlockchainStorage;
import org.hyperledger.besu.ethereum.storage.keyvalue.KeyValueStorageProvider;
import org.hyperledger.besu.ethereum.storage.keyvalue.VariablesKeyValueStorage;
import org.hyperledger.besu.ethereum.storage.keyvalue.WorldStateKeyValueStorage;
import org.hyperledger.besu.ethereum.storage.keyvalue.WorldStatePreimageKeyValueStorage;
import org.hyperledger.besu.ethereum.worldstate.DefaultWorldStateArchive;
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem;

public class InMemoryKeyValueStorageProvider extends KeyValueStorageProvider {

  public InMemoryKeyValueStorageProvider() {
    super(
        segmentIdentifiers -> new SegmentedInMemoryKeyValueStorage(),
        new InMemoryKeyValueStorage(),
        new NoOpMetricsSystem());
  }

  public static MutableBlockchain createInMemoryBlockchain(final Block genesisBlock) {
    return createInMemoryBlockchain(genesisBlock, createInMemoryVariablesStorage());
  }

  public static MutableBlockchain createInMemoryBlockchain(
      final Block genesisBlock, final VariablesStorage variablesStorage) {
    return createInMemoryBlockchain(
        genesisBlock, new MainnetBlockHeaderFunctions(), variablesStorage);
  }

  public static MutableBlockchain createInMemoryBlockchain(
      final Block genesisBlock,
      final BlockHeaderFunctions blockHeaderFunctions,
      final VariablesStorage variablesStorage) {
    final InMemoryKeyValueStorage keyValueStorage = new InMemoryKeyValueStorage();
    return DefaultBlockchain.createMutable(
        genesisBlock,
        new KeyValueStoragePrefixedKeyBlockchainStorage(
            keyValueStorage, variablesStorage, blockHeaderFunctions),
        new NoOpMetricsSystem(),
        0);
  }

  public static DefaultWorldStateArchive createInMemoryWorldStateArchive() {
    return new DefaultWorldStateArchive(
        new WorldStateKeyValueStorage(new InMemoryKeyValueStorage()),
        new WorldStatePreimageKeyValueStorage(new InMemoryKeyValueStorage()));
  }

  public static VariablesStorage createInMemoryVariablesStorage() {
    return new VariablesKeyValueStorage(new InMemoryKeyValueStorage());
  }
}
