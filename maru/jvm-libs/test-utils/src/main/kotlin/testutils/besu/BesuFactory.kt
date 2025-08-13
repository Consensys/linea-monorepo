/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package testutils.besu

import java.util.Optional
import org.hyperledger.besu.crypto.KeyPairUtil
import org.hyperledger.besu.ethereum.core.AddressHelpers
import org.hyperledger.besu.ethereum.core.ImmutableMiningConfiguration
import org.hyperledger.besu.ethereum.core.ImmutableMiningConfiguration.MutableInitValues
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
  private const val PRAGUE_GENESIS = "/el_prague.json"
  const val MIN_BLOCK_TIME = 1L
  const val BLOCK_REBUILD_TIME = 100L

  fun buildTestBesu(
    genesisFile: String = GenesisConfigurationFactory.readGenesisFile(PRAGUE_GENESIS),
    validator: Boolean = true,
  ): BesuNode =
    BesuNodeFactory().createNode("miner") { builder: BesuNodeConfigurationBuilder ->
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
        .engineRpcEnabled(true)
        .jsonRpcEnabled()
        .webSocketEnabled()

      if (validator) {
        val defaultSigner = KeyPairUtil.loadKeyPairFromResource("default-signer-key")
        val miningConfiguration =
          ImmutableMiningConfiguration
            .builder()
            .mutableInitValues(
              MutableInitValues
                .builder()
                .coinbase(AddressHelpers.ofValue(1))
                .isMiningEnabled(true)
                .build(),
            ).unstable(
              ImmutableMiningConfiguration.Unstable
                .builder()
                .posBlockCreationRepetitionMinDuration(BLOCK_REBUILD_TIME)
                .build(),
            ).build()
        builder
          .miningConfiguration(miningConfiguration)
          .keyPair(defaultSigner)
      } else {
        builder
      }
    }

  fun buildSwitchableBesu(
    switchTimestamp: Long = 0,
    pragueTimestamp: Long = 0,
    expectedBlocksInClique: Int = 0,
    validator: Boolean,
  ): BesuNode {
    val genesisContent =
      BesuFactory::class.java
        .getResourceAsStream(PRAGUE_GENESIS)
        ?.bufferedReader()
        ?.use { it.readText() }
        ?: throw IllegalStateException("Could not read genesis file: $PRAGUE_GENESIS")

    val ttd = expectedBlocksInClique * 2
    val genesisFile =
      genesisContent
        .replace("\"shanghaiTime\": 0", "\"shanghaiTime\": $switchTimestamp")
        .replace("\"cancunTime\": 0", "\"cancunTime\": $pragueTimestamp")
        .replace("\"pragueTime\": 0", "\"pragueTime\": $pragueTimestamp")
        .replace("\"terminalTotalDifficulty\": 0", "\"terminalTotalDifficulty\": $ttd")
    return buildTestBesu(genesisFile = genesisFile, validator = validator)
  }
}
