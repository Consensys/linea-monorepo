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
package maru.testutils

import java.io.File
import java.nio.file.Files
import java.nio.file.Path
import maru.app.MaruApp
import maru.app.MaruAppCli.Companion.loadConfig
import maru.config.MaruConfigDtoToml
import maru.config.Utils
import maru.consensus.ElFork
import maru.consensus.config.JsonFriendlyForksSchedule
import maru.consensus.qbft.network.NoopGossiper
import maru.consensus.qbft.network.NoopValidatorMulticaster
import org.hyperledger.besu.consensus.common.bft.Gossiper
import org.hyperledger.besu.consensus.common.bft.network.ValidatorMulticaster

object MaruFactory {
  private val consensusConfigDir = "/e2e/config"
  private val pragueConsensusConfig = "$consensusConfigDir/qbft-prague.json"
  const val VALIDATOR_PRIVATE_KEY = "1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"
  const val VALIDATOR_PRIVATE_KEY_WITH_PREFIX = "0x08021220" + VALIDATOR_PRIVATE_KEY

  private fun buildMaruConfigString(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataPath: String,
  ): String =
    """
    [persistence]
    data-path="$dataPath"
    private-key-path="$dataPath/private-key"

    [sot-eth-endpoint]
    endpoint = "$ethereumJsonRpcUrl"

    [qbft-options]
    communication-margin=200m

    [p2p-config]
    port = 3322
    ip-address = "127.0.0.1"
    static-peers = []

    [validator]
    el-client-engine-api-endpoint = "$engineApiRpc"
    """.trimIndent()

  private fun pickConsensusConfig(elFork: ElFork): String =
    when (elFork) {
      ElFork.Prague -> pragueConsensusConfig
    }

  fun buildTestMaru(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    elFork: ElFork,
    dataDir: Path,
    gossiper: Gossiper = NoopGossiper,
    validatorMulticaster: ValidatorMulticaster = NoopValidatorMulticaster,
  ): MaruApp {
    val appConfig =
      Utils.parseTomlConfig<MaruConfigDtoToml>(
        buildMaruConfigString(
          ethereumJsonRpcUrl = ethereumJsonRpcUrl,
          engineApiRpc = engineApiRpc,
          dataPath = dataDir.toString(),
        ),
      )
    val consensusGenesisResource = this::class.java.getResource(pickConsensusConfig(elFork))
    val beaconGenesisConfig = loadConfig<JsonFriendlyForksSchedule>(listOf(File(consensusGenesisResource!!.path)))

    Files.writeString(appConfig.domainFriendly().persistence.privateKeyPath, VALIDATOR_PRIVATE_KEY_WITH_PREFIX)
    return MaruApp(
      config = appConfig.domainFriendly(),
      beaconGenesisConfig = beaconGenesisConfig.getUnsafe().domainFriendly(),
      gossiper = gossiper,
      validatorMulticaster = validatorMulticaster,
    )
  }

  fun buildTestMaruWithConsensusSwitch(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    switchTimestamp: Long,
    gossiper: Gossiper = NoopGossiper,
    validatorMulticaster: ValidatorMulticaster = NoopValidatorMulticaster,
  ): MaruApp {
    val appConfig =
      Utils.parseTomlConfig<MaruConfigDtoToml>(
        buildMaruConfigString(
          ethereumJsonRpcUrl = ethereumJsonRpcUrl,
          engineApiRpc = engineApiRpc,
          dataPath = dataDir.toString(),
        ),
      )

    // Build the genesis file string directly
    val genesisContent =
      """
      {
        "config": {
          "0": {
            "type": "delegated",
            "blockTimeSeconds": 1
          },
          "$switchTimestamp": {
            "type": "qbft",
            "blockTimeSeconds": 1,
            "feeRecipient": "0x0000000000000000000000000000000000000000",
            "elFork": "Prague"
          }
        }
      }
      """.trimIndent()

    val tempFile = Files.createTempFile("clique-to-qbft", ".json").toFile()
    tempFile.deleteOnExit()
    tempFile.writeText(genesisContent)

    val beaconGenesisConfig = loadConfig<JsonFriendlyForksSchedule>(listOf(tempFile))

    return MaruApp(
      config = appConfig.domainFriendly(),
      beaconGenesisConfig = beaconGenesisConfig.getUnsafe().domainFriendly(),
      gossiper = gossiper,
      validatorMulticaster = validatorMulticaster,
    )
  }
}
