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
import maru.app.config.DummyConsensusOptions
import maru.app.config.ExecutionClientConfig
import maru.consensus.EngineApiBlockCreator
import maru.consensus.ForksSchedule
import maru.consensus.dummy.DummyConsensusEventHandler
import maru.consensus.dummy.DummyConsensusState
import maru.consensus.dummy.EmptyBlockValidator
import maru.consensus.dummy.FinalizationState
import maru.consensus.dummy.NextBlockTimestampProviderImpl
import maru.consensus.dummy.TimeDrivenEventProducer
import maru.executionlayer.client.ExecutionLayerClient
import maru.executionlayer.client.Web3jJsonRpcExecutionLayerClient
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3jClientBuilder
import tech.pegasys.teku.infrastructure.time.SystemTimeProvider

object DummyConsensusProtocolBuilder {
  fun build(
    forksSchedule: ForksSchedule,
    clock: Clock,
    executionClientConfig: ExecutionClientConfig,
    dummyConsensusOptions: DummyConsensusOptions,
  ): TimeDrivenEventProducer {
    val web3JEngineApiClient: Web3JClient =
      Web3jClientBuilder()
        .endpoint(executionClientConfig.engineApiJsonRpcEndpoint.toString())
        .timeout(Duration.ofMinutes(1))
        .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
        .executionClientEventsPublisher { }
        .build()
    val web3jExecutionLayerClient = Web3JExecutionEngineClient(web3JEngineApiClient)

    val web3JEthereumApiClient: Web3JClient =
      Web3jClientBuilder()
        .endpoint(executionClientConfig.ethereumJsonRpcEndpoint.toString())
        .timeout(Duration.ofMinutes(1))
        .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
        .executionClientEventsPublisher({ })
        .build()
    val executionLayerClient: ExecutionLayerClient =
      Web3jJsonRpcExecutionLayerClient(web3jExecutionLayerClient, web3JEthereumApiClient)

    val jsonRpcExecutionLayerManager =
      JsonRpcExecutionLayerManager
        .create(
          executionLayerClient = executionLayerClient,
          feeRecipientProvider = { forksSchedule.getForkByNumber(it).feeRecipient },
          EmptyBlockValidator,
        ).get()
    val latestBlockMetadata = jsonRpcExecutionLayerManager.latestBlockMetadata()
    val latestBlockHash = latestBlockMetadata.blockHash

    val finalizationState = FinalizationState(latestBlockHash, latestBlockHash)
    val dummyConsensusState =
      DummyConsensusState(
        clock = clock,
        finalizationState_ = finalizationState,
        latestBlockHash_ = latestBlockHash,
      )

    val nextBlockTimestampProvider =
      NextBlockTimestampProviderImpl(
        clock = clock,
        forksSchedule = forksSchedule,
        minTimeTillNextBlock = executionClientConfig.minTimeBetweenGetPayloadAttempts,
      )
    val blockCreator =
      EngineApiBlockCreator(
        manager = jsonRpcExecutionLayerManager,
        state = dummyConsensusState,
        blockHeaderFunctions = MainnetBlockHeaderFunctions(),
        initialBlockTimestamp = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(latestBlockMetadata),
      )
    val eventHandler =
      DummyConsensusEventHandler(
        executionLayerManager = jsonRpcExecutionLayerManager,
        blockCreator = blockCreator,
        nextBlockTimestampProvider = nextBlockTimestampProvider,
        onNewBlock = {},
      )
    return TimeDrivenEventProducer(
      forksSchedule = forksSchedule,
      eventHandler = eventHandler,
      blockMetadataProvider = jsonRpcExecutionLayerManager::latestBlockMetadata,
      nextBlockTimestampProvider = nextBlockTimestampProvider,
      clock = clock,
      config = TimeDrivenEventProducer.Config(dummyConsensusOptions.communicationMargin),
    )
  }
}
