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
import kotlin.time.Duration
import maru.config.DummyConsensusOptions
import maru.consensus.EngineApiBlockCreator
import maru.consensus.ForksSchedule
import maru.consensus.NewBlockHandler
import maru.executionlayer.client.ExecutionLayerClient
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions

object DummyConsensusProtocolBuilder {
  fun build(
    forksSchedule: ForksSchedule,
    clock: Clock,
    minTimeTillNextBlock: Duration,
    dummyConsensusOptions: DummyConsensusOptions,
    executionLayerClient: ExecutionLayerClient,
    onNewBlockHandler: NewBlockHandler,
  ): TimeDrivenEventProducer {
    val jsonRpcExecutionLayerManager =
      JsonRpcExecutionLayerManager
        .create(
          executionLayerClient = executionLayerClient,
          feeRecipientProvider = {
            (forksSchedule.getForkByNumber(it) as DummyConsensusConfig).feeRecipient
          },
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
        minTimeTillNextBlock = minTimeTillNextBlock,
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
        onNewBlock = onNewBlockHandler,
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
