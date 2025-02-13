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
import org.hyperledger.besu.ethereum.api.jsonrpc.JsonRpcConfiguration
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.BesuNodeConfigurationBuilder
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.BesuNodeFactory
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.genesis.GenesisConfigurationFactory

object BesuFactory {
  private val cancunGenesis = "/e2e/config/el_cancun.json"

  fun buildTestBesu(): BesuNode = BesuNodeFactory().createExecutionEngineGenesisNode("test node", cancunGenesis)

  fun copyTestBesu(besuNode: BesuNode): BesuNode {
    val genesisFile = GenesisConfigurationFactory.readGenesisFile(cancunGenesis)

    val jsonRpcConfiguration = JsonRpcConfiguration.createDefault()
    jsonRpcConfiguration.isEnabled = true
    jsonRpcConfiguration.port = besuNode.configuration.jsonRpcPort.get()
    jsonRpcConfiguration.setHostsAllowlist(listOf("*"))

    val engineApiConfiguration = JsonRpcConfiguration.createEngineDefault()
    engineApiConfiguration.isEnabled = true
    engineApiConfiguration.port = besuNode.configuration.engineJsonRpcPort.get()
    engineApiConfiguration.isAuthenticationEnabled = false
    engineApiConfiguration.setHostsAllowlist(listOf("*"))

    val nodeConfiguration =
      BesuNodeConfigurationBuilder()
        .name("test node")
        .genesisConfigProvider {
          Optional.of(
            genesisFile,
          )
        }.jsonRpcConfiguration(jsonRpcConfiguration)
        .devMode(false)
        .bootnodeEligible(false)
        .miningEnabled()
        .jsonRpcTxPool()
        .engineJsonRpcConfiguration(engineApiConfiguration)
        .jsonRpcDebug()
        .dataPath(besuNode.homeDirectory())
        .build()
    return BesuNodeFactory().create(nodeConfiguration)
  }
}
