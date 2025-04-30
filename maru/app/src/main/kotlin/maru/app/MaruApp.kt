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
import kotlin.system.exitProcess
import maru.config.FollowersConfig
import maru.config.MaruConfig
import maru.consensus.BlockMetadata
import maru.consensus.ElFork
import maru.consensus.ForksSchedule
import maru.consensus.LatestBlockMetadataCache
import maru.consensus.NewBlockHandler
import maru.consensus.NewBlockHandlerMultiplexer
import maru.consensus.NextBlockTimestampProviderImpl
import maru.consensus.OmniProtocolFactory
import maru.consensus.ProtocolStarter
import maru.consensus.ProtocolStarterBlockHandler
import maru.consensus.Web3jMetadataProvider
import maru.consensus.blockimport.FollowerBeaconBlockImporter
import maru.consensus.delegated.ElDelegatedConsensusFactory
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlockBody
import maru.database.kv.KvDatabaseFactory
import maru.executionlayer.manager.ForkChoiceUpdatedResult
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem
import tech.pegasys.teku.infrastructure.async.SafeFuture

class MaruApp(
  config: MaruConfig,
  beaconGenesisConfig: ForksSchedule,
  clock: Clock = Clock.systemUTC(),
) : AutoCloseable {
  private val log: Logger = LogManager.getLogger(this::class.java)

  init {
    if (config.p2pConfig == null) {
      log.warn("P2P is disabled!")
    }
    if (config.validator == null) {
      log.info("Validator is not defined. Maru is running in follower-only node")
      log.error("Follower-only mode is not supported yet! Exiting application.")
      exitProcess(1)
    }
    log.info(config.toString())
  }

  private val ethereumJsonRpcClient =
    Helpers.createWeb3jClient(
      config.sotNode,
    )

  private val asyncMetadataProvider = Web3jMetadataProvider(ethereumJsonRpcClient.eth1Web3j)
  private val lastBlockMetadataCache: LatestBlockMetadataCache =
    LatestBlockMetadataCache(asyncMetadataProvider.getLatestBlockMetadata())
  private val metadataProviderCacheUpdater =
    NewBlockHandler { beaconBlock ->
      val blockMetadata = BlockMetadata.fromBeaconBlock(beaconBlock)
      lastBlockMetadataCache.updateLatestBlockMetadata(blockMetadata)
      SafeFuture.completedFuture(Unit)
    }

  private val nextTargetBlockTimestampProvider =
    NextBlockTimestampProviderImpl(
      clock = clock,
      forksSchedule = beaconGenesisConfig,
    )

  private val metricsSystem = NoOpMetricsSystem()
  private val finalizationStateProviderStub = { it: BeaconBlockBody ->
    LogManager.getLogger("FinalizationStateProvider").debug("fetching the latest finalized state")
    FinalizationState(it.executionPayload.blockHash, it.executionPayload.blockHash)
  }

  private val beaconChain =
    KvDatabaseFactory
      .createRocksDbDatabase(
        databasePath = config.persistence.dataPath,
        metricsSystem = metricsSystem,
        metricCategory = MaruMetricsCategory.STORAGE,
      )
  private val protocolStarter =
    let {
      val metadataCacheUpdaterHandlerEntry = "latest block metadata updater" to metadataProviderCacheUpdater
      val delegatedConsensusNewBlockHandler =
        NewBlockHandlerMultiplexer(
          mapOf(metadataCacheUpdaterHandlerEntry),
        )
      val qbftConsensusNewBlockHandler =
        NewBlockHandlerMultiplexer(createFollowerHandlers(config.followers) + metadataCacheUpdaterHandlerEntry)
      ProtocolStarter(
        forksSchedule = beaconGenesisConfig,
        protocolFactory =
          OmniProtocolFactory(
            elDelegatedConsensusFactory =
              ElDelegatedConsensusFactory(
                ethereumJsonRpcClient = ethereumJsonRpcClient.eth1Web3j,
                newBlockHandler = delegatedConsensusNewBlockHandler,
              ),
            qbftConsensusFactory =
              QbftProtocolFactoryWithBeaconChainInitialization(
                maruConfig = config,
                metricsSystem = metricsSystem,
                finalizationStateProvider = finalizationStateProviderStub,
                executionLayerClient = ethereumJsonRpcClient.eth1Web3j,
                nextTargetBlockTimestampProvider = nextTargetBlockTimestampProvider,
                newBlockHandler = qbftConsensusNewBlockHandler,
                beaconChain = beaconChain,
                clock = clock,
              ),
          ),
        metadataProvider = lastBlockMetadataCache,
        nextBlockTimestampProvider = nextTargetBlockTimestampProvider,
      ).also {
        val protocolStarterBlockHandlerEntry = "protocol starter" to ProtocolStarterBlockHandler(it)
        delegatedConsensusNewBlockHandler.addHandler(
          protocolStarterBlockHandlerEntry.first,
          protocolStarterBlockHandlerEntry.second,
        )
        qbftConsensusNewBlockHandler.addHandler(
          protocolStarterBlockHandlerEntry.first,
          protocolStarterBlockHandlerEntry.second,
        )
      }
    }

  private fun createFollowerHandlers(
    followers: FollowersConfig,
  ): Map<String, NewBlockHandler<ForkChoiceUpdatedResult>> =
    followers.followers
      .mapValues {
        val engineApiClient = Helpers.buildExecutionEngineClient(it.value, ElFork.Prague)
        FollowerBeaconBlockImporter.create(engineApiClient)
      }

  fun start() {
    protocolStarter.start()
    log.info("Maru is up")
  }

  fun stop() {
    protocolStarter.stop()
    log.info("Maru is down")
  }

  override fun close() {
    beaconChain.close()
  }
}
