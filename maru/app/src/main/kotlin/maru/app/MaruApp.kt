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
package maru.app

import java.time.Clock
import java.time.Duration
import maru.config.MaruConfig
import maru.consensus.ForksSchedule
import maru.consensus.MetadataOnlyHandlerAdapter
import maru.consensus.NewBlockHandlerMultiplexer
import maru.consensus.OmniProtocolFactory
import maru.consensus.ProtocolStarter
import maru.executionlayer.client.ExecutionLayerClient
import maru.executionlayer.client.Web3jJsonRpcExecutionLayerClient
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3jClientBuilder
import tech.pegasys.teku.infrastructure.time.SystemTimeProvider

class MaruApp(
  config: MaruConfig,
  beaconGenesisConfig: ForksSchedule,
  clock: Clock = Clock.systemUTC(),
) {
  private val log: Logger = LogManager.getLogger(this::class.java)

  init {
    if (config.p2pConfig == null) {
      log.warn("P2P is disabled!")
    }
    if (config.validator == null) {
      log.info("Maru is running in follower-only node")
    }
  }

  private val ethereumJsonRpcClient =
    buildJsonRpcClient(
      config.executionClientConfig.ethereumJsonRpcEndpoint
        .toString(),
    )
  private val executionLayerClient =
    buildExecutionEngineClient(
      config.executionClientConfig.engineApiJsonRpcEndpoint
        .toString(),
      ethereumJsonRpcClient,
    )

  private val newBlockHandlerMultiplexer = NewBlockHandlerMultiplexer(emptyMap())

  private val protocolStarter =
    ProtocolStarter(
      forksSchedule = beaconGenesisConfig,
      protocolFactory =
        OmniProtocolFactory(
          forksSchedule = beaconGenesisConfig,
          clock = clock,
          config = config,
          executionLayerClient = executionLayerClient,
          ethereumJsonRpcClient = ethereumJsonRpcClient.eth1Web3j,
          newBlockHandler = newBlockHandlerMultiplexer,
        ),
      executionLayerClient = executionLayerClient,
    ).also {
      newBlockHandlerMultiplexer.addHandler("protocol starter", MetadataOnlyHandlerAdapter(it))
    }

  private fun buildExecutionEngineClient(
    endpoint: String,
    web3JEthereumApiClient: Web3JClient,
  ): ExecutionLayerClient {
    val web3JEngineApiClient: Web3JClient =
      Web3jClientBuilder()
        .endpoint(endpoint)
        .timeout(Duration.ofMinutes(1))
        .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
        .executionClientEventsPublisher { }
        .build()
    val web3jExecutionLayerClient = Web3JExecutionEngineClient(web3JEngineApiClient)
    return Web3jJsonRpcExecutionLayerClient(web3jExecutionLayerClient, web3JEthereumApiClient)
  }

  private fun buildJsonRpcClient(endpoint: String): Web3JClient =
    Web3jClientBuilder()
      .endpoint(endpoint)
      .timeout(Duration.ofMinutes(1))
      .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
      .executionClientEventsPublisher { }
      .build()

  fun start() {
    protocolStarter.start()
    log.info("Maru is up")
  }

  fun stop() {
    protocolStarter.stop()
  }
}
