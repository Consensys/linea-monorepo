/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.testutils.besu

import java.util.Optional
import maru.consensus.ElFork
import org.hyperledger.besu.ethereum.storage.keyvalue.KeyValueSegmentIdentifier
import org.hyperledger.besu.plugin.services.storage.KeyValueStorageFactory
import org.hyperledger.besu.plugin.services.storage.rocksdb.RocksDBKeyValueStorageFactory
import org.hyperledger.besu.plugin.services.storage.rocksdb.RocksDBMetricsFactory
import org.hyperledger.besu.plugin.services.storage.rocksdb.configuration.RocksDBCLIOptions
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.BesuNodeConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.BesuNodeFactory
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.genesis.GenesisConfigurationFactory

object BesuFactory {
  private val elConfigsDir = "/e2e/config"
  private val pragueGenesis = "$elConfigsDir/el_prague.json"

  private fun pickGenesis(elFork: ElFork): String =
    when (elFork) {
      ElFork.Prague -> pragueGenesis
    }

  fun buildTestBesu(elFork: ElFork = ElFork.Prague): BesuNode =
    BesuNodeFactory().createMinerNode(
      "miner",
    ) { builder: BesuNodeConfigurationBuilder ->
      val genesisFilePath = pickGenesis(elFork)
      val genesisFile = GenesisConfigurationFactory.readGenesisFile(genesisFilePath)
      val persistentStorageFactory: KeyValueStorageFactory =
        RocksDBKeyValueStorageFactory(
          RocksDBCLIOptions.create()::toDomainObject,
          KeyValueSegmentIdentifier.entries,
          RocksDBMetricsFactory.PUBLIC_ROCKS_DB_METRICS,
        )
      builder
        .storageImplementation(persistentStorageFactory)
        .genesisConfigProvider {
          Optional.of(
            genesisFile,
          )
        }.devMode(false)
        .bootnodeEligible(false)
        .jsonRpcTxPool()
        .engineRpcEnabled(true)
        .jsonRpcDebug()
    }
}
