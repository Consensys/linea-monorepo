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
  const val VALIDATOR_PRIVATE_KEY = "0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"

  private fun buildMaruConfigString(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataPath: String,
  ): String =
    """
    [persistence]
    data-path="$dataPath"

    [sot-eth-endpoint]
    endpoint = "$ethereumJsonRpcUrl"

    [qbft-options]
    communication-margin=200m

    [p2p-config]
    port = 3322

    [validator]
    private-key = "$VALIDATOR_PRIVATE_KEY"
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

    return MaruApp(
      config = appConfig.domainFriendly(),
      beaconGenesisConfig = beaconGenesisConfig.getUnsafe().domainFriendly(),
      gossiper = gossiper,
      validatorMulticaster = validatorMulticaster,
    )
  }
}
