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
package maru.consensus.dummy

import java.time.Clock
import maru.config.MaruConfig
import maru.consensus.ElFork
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.NewBlockHandler
import maru.consensus.NextBlockTimestampProviderImpl
import maru.consensus.ProtocolFactory
import maru.consensus.state.FinalizationState
import maru.executionlayer.client.ExecutionLayerClient
import maru.executionlayer.client.MetadataProvider
import maru.executionlayer.client.PragueWeb3jJsonRpcExecutionLayerClient
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3jClientBuilder
import tech.pegasys.teku.infrastructure.time.SystemTimeProvider

class DummyConsensusProtocolFactory(
  private val forksSchedule: ForksSchedule,
  private val maruConfig: MaruConfig,
  private val clock: Clock,
  private val metadataProvider: MetadataProvider,
  private val newBlockHandler: NewBlockHandler,
) : ProtocolFactory {
  init {
    require(maruConfig.dummyConsensusOptions != null) {
      "Next fork is dummy consensus one, but dummyConsensusOptions are undefined!"
    }
    require(maruConfig.validator != null) {
      "Validator is required for Dummy Consensus protocol factory!"
    }
  }

  class DummyConsensusFeeRecipientProvider(
    private val forksSchedule: ForksSchedule,
  ) : FeeRecipientProvider {
    override fun getFeeRecipient(timestamp: Long): ByteArray {
      val nextExpectedFork = forksSchedule.getForkByTimestamp(timestamp)
      return (
        nextExpectedFork.configuration as DummyConsensusConfig
      ).feeRecipient
    }
  }

  override fun create(forkSpec: ForkSpec): DummyConsensus {
    require(forkSpec.configuration is DummyConsensusConfig) {
      "Unexpected fork specification! ${
        forkSpec
          .configuration
      } instead of ${DummyConsensusConfig::class.simpleName}"
    }

    val validatorConfig = maruConfig.validator!!
    val executionLayerClient =
      buildExecutionEngineClient(
        validatorConfig.client.engineApiClientConfig.endpoint
          .toString(),
        forkSpec.configuration.elFork,
      )
    val jsonRpcExecutionLayerManager =
      JsonRpcExecutionLayerManager
        .create(
          executionLayerClient = executionLayerClient,
          metadataProvider = metadataProvider,
          payloadValidator = EmptyBlockValidator,
        ).get()

    val latestBlockMetadata = metadataProvider.getLatestBlockMetadata().get()
    val blockMetadataCache = LatestBlockMetadataCache(latestBlockMetadata)
    val latestBlockHash = blockMetadataCache.getLatestBlockMetadata().blockHash

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
        minTimeTillNextBlock = validatorConfig.client.minTimeBetweenGetPayloadAttempts,
      )

    val blockCreator =
      DummyEngineApiBlockCreator(
        manager = jsonRpcExecutionLayerManager,
        state = dummyConsensusState,
        nextBlockTimestamp =
          nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(
            latestBlockMetadata.unixTimestampSeconds,
          ),
        feeRecipientProvider = DummyConsensusFeeRecipientProvider(forksSchedule),
      )
    val eventHandler =
      DummyConsensusEventHandler(
        blockCreator = blockCreator,
        nextBlockTimestampProvider = nextBlockTimestampProvider,
        onNewBlock = newBlockHandler,
        blockMetadataCache = blockMetadataCache,
      )
    return DummyConsensus(
      forksSchedule = forksSchedule,
      eventHandler = eventHandler,
      blockMetadataProvider = blockMetadataCache::getLatestBlockMetadata,
      nextBlockTimestampProvider = nextBlockTimestampProvider,
      clock = clock,
      config = DummyConsensus.Config(maruConfig.dummyConsensusOptions!!.communicationMargin),
    )
  }

  private fun buildExecutionEngineClient(
    endpoint: String,
    elFork: ElFork,
  ): ExecutionLayerClient {
    val web3JEngineApiClient: Web3JClient =
      Web3jClientBuilder()
        .endpoint(endpoint)
        .timeout(java.time.Duration.ofMinutes(1))
        .timeProvider(SystemTimeProvider.SYSTEM_TIME_PROVIDER)
        .executionClientEventsPublisher { }
        .build()
    val web3jExecutionLayerClient = Web3JExecutionEngineClient(web3JEngineApiClient)
    return when (elFork) {
      ElFork.Prague -> PragueWeb3jJsonRpcExecutionLayerClient(web3jExecutionLayerClient)
    }
  }
}
