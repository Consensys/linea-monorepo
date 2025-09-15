/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package testutils.besu

import java.util.Collections.singletonList
import java.util.Optional
import java.util.UUID
import kotlin.jvm.optionals.getOrDefault
import org.hyperledger.besu.crypto.KeyPairUtil
import org.hyperledger.besu.ethereum.api.jsonrpc.JsonRpcConfiguration
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
    engineRpcPort: Optional<Int> = Optional.empty(),
    jsonRpcPort: Optional<Int> = Optional.empty(),
  ): BesuNode =
    BesuNodeFactory().createNode("miner-${UUID.randomUUID()}") { builder: BesuNodeConfigurationBuilder ->
      val persistentStorageFactory: KeyValueStorageFactory =
        RocksDBKeyValueStorageFactory(
          RocksDBCLIOptions.create()::toDomainObject,
          KeyValueSegmentIdentifier.entries,
          RocksDBMetricsFactory.PUBLIC_ROCKS_DB_METRICS,
        )

      val engineRpcConfig = JsonRpcConfiguration.createEngineDefault()
      engineRpcConfig.setEnabled(true)
      engineRpcConfig.setPort(engineRpcPort.getOrDefault(0))
      engineRpcConfig.host = "127.0.0.1"
      engineRpcConfig.setHostsAllowlist(singletonList("*"))
      engineRpcConfig.setAuthenticationEnabled(false)

      val jsonRpcConfig = JsonRpcConfiguration.createDefault()
      jsonRpcConfig.setEnabled(true)
      jsonRpcConfig.setPort(jsonRpcPort.getOrDefault(0))
      jsonRpcConfig.setHostsAllowlist(singletonList("*"))

      builder
        .storageImplementation(persistentStorageFactory)
        .genesisConfigProvider {
          Optional.of(
            genesisFile,
          )
        }.devMode(false)
        .engineJsonRpcConfiguration(engineRpcConfig)
        .jsonRpcConfiguration(jsonRpcConfig)
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
    pragueTimestamp: ULong = 0UL,
    cancunTimestamp: ULong = pragueTimestamp,
    shanghaiTimestamp: ULong = cancunTimestamp,
    ttd: ULong = 0UL,
    validator: Boolean,
  ): BesuNode {
    val genesisContent =
      BesuFactory::class.java
        .getResourceAsStream(PRAGUE_GENESIS)
        ?.bufferedReader()
        ?.use { it.readText() }
        ?: throw IllegalStateException("Could not read genesis file: $PRAGUE_GENESIS")

    val genesisFile =
      genesisContent
        .replace("\"shanghaiTime\": 0", "\"shanghaiTime\": $shanghaiTimestamp")
        .replace("\"cancunTime\": 0", "\"cancunTime\": $cancunTimestamp")
        .replace("\"pragueTime\": 0", "\"pragueTime\": $pragueTimestamp")
        .replace("\"terminalTotalDifficulty\": 0", "\"terminalTotalDifficulty\": $ttd")
    return buildTestBesu(genesisFile = genesisFile, validator = validator)
  }
}
