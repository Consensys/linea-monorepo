/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import io.vertx.core.Vertx
import java.time.Clock
import maru.api.ApiServer
import maru.config.FollowersConfig
import maru.config.MaruConfig
import maru.config.consensus.ElFork
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ElBlockMetadata
import maru.consensus.ForksSchedule
import maru.consensus.LatestElBlockMetadataCache
import maru.consensus.NewBlockHandler
import maru.consensus.NewBlockHandlerMultiplexer
import maru.consensus.NextBlockTimestampProviderImpl
import maru.consensus.OmniProtocolFactory
import maru.consensus.ProtocolStarter
import maru.consensus.ProtocolStarterBlockHandler
import maru.consensus.SealedBeaconBlockHandlerAdapter
import maru.consensus.blockimport.FollowerBeaconBlockImporter
import maru.consensus.blockimport.NewSealedBeaconBlockHandlerMultiplexer
import maru.consensus.delegated.ElDelegatedConsensusFactory
import maru.consensus.state.FinalizationProvider
import maru.core.Protocol
import maru.crypto.Crypto
import maru.database.BeaconChain
import maru.metrics.MaruMetricsCategory
import maru.p2p.P2PNetwork
import maru.p2p.SealedBeaconBlockBroadcaster
import maru.p2p.ValidationResult
import maru.services.LongRunningService
import maru.syncing.SyncStatusProvider
import net.consensys.linea.async.get
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.vertx.ObservabilityServer
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.plugin.services.MetricsSystem
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.infrastructure.async.SafeFuture

class MaruApp(
  val config: MaruConfig,
  beaconGenesisConfig: ForksSchedule,
  clock: Clock = Clock.systemUTC(),
  // This will only be used if config.p2pConfig is undefined
  private val p2pNetwork: P2PNetwork,
  private val privateKeyProvider: () -> ByteArray,
  private val finalizationProvider: FinalizationProvider,
  private val vertx: Vertx,
  private val metricsFacade: MetricsFacade,
  private val beaconChain: BeaconChain,
  private val metricsSystem: MetricsSystem,
  private val lastElBlockMetadataCache: LatestElBlockMetadataCache,
  private val ethereumJsonRpcClient: Web3JClient,
  private val apiServer: ApiServer,
  private val syncStatusProvider: SyncStatusProvider,
  private val syncControllerManager: LongRunningService,
) : AutoCloseable {
  private val log: Logger = LogManager.getLogger(this.javaClass)

  init {
    if (config.qbftOptions == null) {
      log.info("Qbft options are not defined. Maru is running in follower-only node")
    }

    metricsFacade.createGauge(
      category = MaruMetricsCategory.METADATA,
      name = "block.height",
      description = "Latest block height",
      measurementSupplier = {
        lastElBlockMetadataCache.getLatestBlockMetadata().blockNumber.toLong()
      },
    )
  }

  fun p2pPort(): UInt = p2pNetwork.port

  private val metadataProviderCacheUpdater =
    NewBlockHandler<Unit> { beaconBlock ->
      val elBlockMetadata = ElBlockMetadata.fromBeaconBlock(beaconBlock)
      lastElBlockMetadataCache.updateLatestBlockMetadata(elBlockMetadata)
      SafeFuture.completedFuture(Unit)
    }
  private val nextTargetBlockTimestampProvider =
    NextBlockTimestampProviderImpl(
      clock = clock,
      forksSchedule = beaconGenesisConfig,
    )
  private val protocolStarter = createProtocolStarter(config, beaconGenesisConfig, clock)

  @Suppress("UNCHECKED_CAST")
  private fun createFollowerHandlers(followers: FollowersConfig): Map<String, NewBlockHandler<Unit>> =
    followers.followers
      .mapValues {
        val engineApiClient = Helpers.buildExecutionEngineClient(it.value, ElFork.Prague, metricsFacade)
        FollowerBeaconBlockImporter.create(engineApiClient, finalizationProvider) as NewBlockHandler<Unit>
      }

  fun start() {
    try {
      vertx
        .deployVerticle(
          ObservabilityServer(
            ObservabilityServer.Config(applicationName = "maru", port = config.observabilityOptions.port.toInt()),
          ),
        ).get()
    } catch (th: Throwable) {
      log.error("Error while trying to start the observability server", th)
      throw th
    }
    try {
      p2pNetwork.start().get()
    } catch (th: Throwable) {
      log.error("Error while trying to start the P2P network", th)
      throw th
    }
    try {
      syncControllerManager.start()
    } catch (th: Throwable) {
      log.error("Error while trying to start the Sync Service", th)
      throw th
    }
    apiServer.start()
    log.info("Maru is up")
  }

  fun stop() {
    try {
      vertx.deploymentIDs().forEach {
        vertx.undeploy(it).get()
      }
    } catch (th: Throwable) {
      log.warn("Error while trying to stop the vertx verticles", th)
    }
    try {
      p2pNetwork.stop().get()
    } catch (th: Throwable) {
      log.warn("Error while trying to stop the P2P network", th)
    }
    try {
      syncControllerManager.stop()
    } catch (th: Throwable) {
      log.error("Error while trying to stop the Sync Service", th)
    }
    protocolStarter.stop()
    apiServer.stop()
    log.info("Maru is down")
  }

  override fun close() {
    beaconChain.close()
  }

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

    val qbftForkTimestamp =
      beaconGenesisConfig.getForkByConfigType(QbftConsensusConfig::class).timestampSeconds.toULong()
    val beaconChainInitialization =
      BeaconChainInitialization(
        beaconChain = beaconChain,
        genesisTimestamp = qbftForkTimestamp,
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
          privateKeyBytes = Crypto.privateKeyBytesWithoutPrefix(privateKeyProvider()),
          validatorElNodeConfig = config.validatorElNode,
          metricsSystem = metricsSystem,
          finalizationStateProvider = finalizationProvider,
          nextTargetBlockTimestampProvider = nextTargetBlockTimestampProvider,
          newBlockHandler = sealedBlockHandlerMultiplexer,
          beaconChain = beaconChain,
          clock = clock,
          p2pNetwork = p2pNetwork,
          beaconChainInitialization = beaconChainInitialization,
          metricsFacade = metricsFacade,
          allowEmptyBlocks = config.allowEmptyBlocks,
          syncStatusProvider = syncStatusProvider,
        )
      } else {
        QbftFollowerFactory(
          p2PNetwork = p2pNetwork,
          beaconChain = beaconChain,
          newBlockHandler = blockImportHandlers,
          validatorElNodeConfig = config.validatorElNode,
          beaconChainInitialization = beaconChainInitialization,
          metricsFacade = metricsFacade,
          allowEmptyBlocks = config.allowEmptyBlocks,
        )
      }
    val delegatedConsensusNewBlockHandler =
      NewBlockHandlerMultiplexer(
        mapOf(metadataCacheUpdaterHandlerEntry),
      )

    val protocolStarter =
      ProtocolStarter.create(
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
        elMetadataProvider = lastElBlockMetadataCache,
        nextBlockTimestampProvider = nextTargetBlockTimestampProvider,
        syncStatusProvider = syncStatusProvider,
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
