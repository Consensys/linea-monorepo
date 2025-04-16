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

import java.time.Duration
import java.util.Optional
import java.util.UUID
import kotlin.io.path.Path
import kotlin.time.Duration.Companion.minutes
import kotlin.time.toJavaDuration
import maru.config.ApiEndpointConfig
import maru.consensus.qbft.FollowerBeaconBlockImporter
import maru.consensus.state.FinalizationState
import maru.executionlayer.client.MetadataProvider
import maru.executionlayer.client.PragueWeb3jJsonRpcExecutionLayerClient
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.executionlayer.manager.NoopValidator
import tech.pegasys.teku.ethereum.executionclient.auth.JwtConfig
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3jClientBuilder
import tech.pegasys.teku.infrastructure.time.SystemTimeProvider

object Helpers {
  fun createBlockImporter(
    apiEndpointConfig: ApiEndpointConfig,
    metadataProvider: MetadataProvider,
  ): BlockImportHandler {
    val web3JEngineApiClient: Web3JClient =
      Web3jClientBuilder()
        .endpoint(apiEndpointConfig.endpoint.toString())
        .timeout(Duration.ofMinutes(1))
        .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
        .executionClientEventsPublisher { }
        .jwtConfigOpt(wrapJwtPath(apiEndpointConfig.jwtSecretPath))
        .build()
    val web3jExecutionLayerClient = Web3JExecutionEngineClient(web3JEngineApiClient)
    val executionLayerClient = PragueWeb3jJsonRpcExecutionLayerClient(web3jExecutionLayerClient)
    val executionLayerManager =
      JsonRpcExecutionLayerManager
        .create(
          executionLayerClient = executionLayerClient,
          metadataProvider = metadataProvider,
          payloadValidator = NoopValidator,
        ).get()
    val latestBlockMetadata = metadataProvider.getLatestBlockMetadata().get()
    val blockImporter =
      FollowerBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        finalizationStateProvider = {
          FinalizationState(
            latestBlockMetadata.blockHash,
            latestBlockMetadata.blockHash,
          )
        },
      )
    return BlockImportHandler(executionLayerManager, blockImporter)
  }

  private fun wrapJwtPath(jwtPath: String?): Optional<JwtConfig> {
    val jwtConfigPath = Optional.ofNullable(jwtPath)
    return JwtConfig.createIfNeeded(
      /* needed = */ jwtConfigPath.isPresent,
      jwtConfigPath,
      Optional.of(UUID.randomUUID().toString()),
      Path("/dev/null"), // Teku's API limitation. Would be good to clean it
    )
  }

  fun createWeb3jClient(apiEndpointConfig: ApiEndpointConfig): Web3JClient =
    Web3jClientBuilder()
      .timeout(1.minutes.toJavaDuration())
      .endpoint(apiEndpointConfig.endpoint.toString())
      .jwtConfigOpt(wrapJwtPath(apiEndpointConfig.jwtSecretPath))
      .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
      .executionClientEventsPublisher {}
      .build()
}
