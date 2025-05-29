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
import maru.consensus.SealedBeaconBlockHandlerAdapter
import maru.consensus.Web3jMetadataProvider
import maru.consensus.blockimport.FollowerBeaconBlockImporter
import maru.consensus.blockimport.NewSealedBeaconBlockHandlerMultiplexer
import maru.consensus.delegated.ElDelegatedConsensusFactory
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlockBody
import maru.core.Protocol
import maru.database.kv.KvDatabaseFactory
import maru.p2p.NoOpP2PNetwork
import maru.p2p.P2PNetwork
import maru.p2p.P2PNetworkImpl
import maru.p2p.SealedBeaconBlockBroadcaster
import maru.p2p.ValidationResult
import maru.serialization.rlp.RLPSerializers
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.p2p.network.config.GeneratingFilePrivateKeySource

class MaruApp(
  config: MaruConfig,
  beaconGenesisConfig: ForksSchedule,
  clock: Clock = Clock.systemUTC(),
  // This will only be used if config.p2pConfig is undefined
  private var p2pNetwork: P2PNetwork = NoOpP2PNetwork,
) : AutoCloseable {
  private val log: Logger = LogManager.getLogger(this::javaClass)

  private var privateKeyBytes: ByteArray =
    GeneratingFilePrivateKeySource(
      config.persistence.privateKeyPath.toString(),
    ).privateKeyBytes.toArray()

  init {
    if (!config.persistence.privateKeyPath
        .toFile()
        .exists()
    ) {
      log.info(
        "Private key file ${config.persistence.privateKeyPath} does not exist. A new private key will be generated and stored in that location.",
      )
    } else {
      log.info(
        "Private key file ${config.persistence.privateKeyPath} already exists. Maru will use the existing private key.",
      )
    }
    if (config.qbftOptions == null) {
      log.info("Qbft options are not defined. Maru is running in follower-only node")
    }
    if (config.p2pConfig == null) {
      log.info("P2PManager is not defined.")
    }
    log.info(config.toString())

    config.p2pConfig?.let {
      p2pNetwork =
        P2PNetworkImpl(
          privateKeyBytes = privateKeyBytes,
          p2pConfig = config.p2pConfig!!,
          chainId = beaconGenesisConfig.chainId,
          serializer = RLPSerializers.SealedBeaconBlockSerializer,
        )
    }
  }

  private val ethereumJsonRpcClient =
    Helpers.createWeb3jClient(
      config.validatorElNode.ethApiEndpoint,
    )

  private val asyncMetadataProvider = Web3jMetadataProvider(ethereumJsonRpcClient.eth1Web3j)
  private val lastBlockMetadataCache: LatestBlockMetadataCache =
    LatestBlockMetadataCache(asyncMetadataProvider.getLatestBlockMetadata())
  private val metadataProviderCacheUpdater =
    NewBlockHandler<Unit> { beaconBlock ->
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
  private val protocolStarter = createProtocolStarter(config, beaconGenesisConfig, clock)

  private fun createFollowerHandlers(followers: FollowersConfig): Map<String, NewBlockHandler<Unit>> =
    followers.followers
      .mapValues {
        val engineApiClient = Helpers.buildExecutionEngineClient(it.value, ElFork.Prague)
        FollowerBeaconBlockImporter.create(engineApiClient) as NewBlockHandler<Unit>
      }

  fun start() {
    p2pNetwork.start()
    protocolStarter.start()
    log.info("Maru is up")
  }

  fun stop() {
    p2pNetwork.stop()
    protocolStarter.stop()
    log.info("Maru is down")
  }

  override fun close() {
    beaconChain.close()
  }

  private fun privateKeyBytesWithoutPrefix() =
    privateKeyBytes
      .slice(
        privateKeyBytes.size - 32..privateKeyBytes.size - 1,
      ).toByteArray()

  private fun createProtocolStarter(
    config: MaruConfig,
    beaconGenesisConfig: ForksSchedule,
    clock: Clock,
  ): Protocol {
    val metadataCacheUpdaterHandlerEntry = "latest block metadata updater" to metadataProviderCacheUpdater

    val followerHandlersMap: Map<String, NewBlockHandler<Unit>> =
      createFollowerHandlers(config.followers)
    val followerBlockHandlers = followerHandlersMap + metadataCacheUpdaterHandlerEntry
    val blockImportHandlers =
      NewBlockHandlerMultiplexer(followerBlockHandlers)
    val adaptedBeaconBlockImporter = SealedBeaconBlockHandlerAdapter(blockImportHandlers)

    val beaconChainInitialization =
      BeaconChainInitialization(
        executionLayerClient = ethereumJsonRpcClient.eth1Web3j,
        beaconChain = beaconChain,
      )

    val qbftFactory =
      if (config.qbftOptions != null) {
        val sealedBlockHandlers =
          mutableMapOf(
            "beacon block handlers" to adaptedBeaconBlockImporter,
            "p2p broadcast sealed beacon block handler" to
              SealedBeaconBlockBroadcaster(p2pNetwork),
          )
        val sealedBlockHandlerMultiplexer =
          NewSealedBeaconBlockHandlerMultiplexer<Unit>(
            handlersMap = sealedBlockHandlers,
          )
        QbftProtocolFactoryWithBeaconChainInitialization(
          qbftOptions = config.qbftOptions!!,
          privateKeyBytes = privateKeyBytesWithoutPrefix(),
          validatorElNodeConfig = config.validatorElNode,
          metricsSystem = metricsSystem,
          finalizationStateProvider = finalizationStateProviderStub,
          nextTargetBlockTimestampProvider = nextTargetBlockTimestampProvider,
          newBlockHandler = sealedBlockHandlerMultiplexer,
          beaconChain = beaconChain,
          clock = clock,
          p2pNetwork = p2pNetwork,
          beaconChainInitialization = beaconChainInitialization,
        )
      } else {
        QbftFollowerFactory(
          p2PNetwork = p2pNetwork,
          beaconChain = beaconChain,
          newBlockHandler = blockImportHandlers,
          validatorElNodeConfig = config.validatorElNode,
          beaconChainInitialization = beaconChainInitialization,
        )
      }
    val delegatedConsensusNewBlockHandler =
      NewBlockHandlerMultiplexer(
        mapOf(metadataCacheUpdaterHandlerEntry),
      )

    val protocolStarter =
      ProtocolStarter(
        forksSchedule = beaconGenesisConfig,
        protocolFactory =
          OmniProtocolFactory(
            elDelegatedConsensusFactory =
              ElDelegatedConsensusFactory(
                ethereumJsonRpcClient = ethereumJsonRpcClient.eth1Web3j,
                newBlockHandler = delegatedConsensusNewBlockHandler,
              ),
            qbftConsensusFactory = qbftFactory,
          ),
        metadataProvider = lastBlockMetadataCache,
        nextBlockTimestampProvider = nextTargetBlockTimestampProvider,
      )

    val protocolStarterBlockHandlerEntry = "protocol starter" to ProtocolStarterBlockHandler(protocolStarter)
    delegatedConsensusNewBlockHandler.addHandler(
      protocolStarterBlockHandlerEntry.first,
    ) {
      protocolStarterBlockHandlerEntry.second.handleNewBlock(it)
    }
    blockImportHandlers.addHandler(
      protocolStarterBlockHandlerEntry.first,
    ) {
      protocolStarterBlockHandlerEntry.second
        .handleNewBlock(it)
        .thenApply { ValidationResult.Companion.Valid }
    }

    return protocolStarter
  }
}
