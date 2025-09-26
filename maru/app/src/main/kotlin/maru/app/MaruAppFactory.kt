/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import io.libp2p.core.PeerId
import io.libp2p.core.crypto.unmarshalPrivateKey
import io.micrometer.core.instrument.MeterRegistry
import io.vertx.core.Vertx
import io.vertx.micrometer.MicrometerMetricsOptions
import io.vertx.micrometer.backends.BackendRegistries
import java.nio.file.Files
import java.nio.file.Path
import java.time.Clock
import kotlin.io.path.exists
import linea.contract.l1.LineaRollupSmartContractClientReadOnly
import linea.contract.l1.Web3JLineaRollupSmartContractClientReadOnly
import linea.kotlin.encodeHex
import linea.web3j.createWeb3jHttpClient
import linea.web3j.ethapi.createEthApiClient
import maru.api.ApiServer
import maru.api.ApiServerImpl
import maru.api.ChainDataProviderImpl
import maru.config.MaruConfig
import maru.config.P2PConfig
import maru.config.SyncingConfig
import maru.config.consensus.ElFork
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ForkIdHashProvider
import maru.consensus.ForkIdHashProviderImpl
import maru.consensus.ForkIdHasher
import maru.consensus.ForksSchedule
import maru.consensus.StaticValidatorProvider
import maru.consensus.state.FinalizationProvider
import maru.consensus.state.InstantFinalizationProvider
import maru.core.SealedBeaconBlock
import maru.crypto.Hashing
import maru.database.BeaconChain
import maru.database.P2PState
import maru.database.kv.KvDatabaseFactory
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.finalization.LineaFinalizationProvider
import maru.metrics.BesuMetricsCategoryAdapter
import maru.metrics.BesuMetricsSystemAdapter
import maru.metrics.MaruMetricsCategory
import maru.p2p.NetworkHelper
import maru.p2p.NoOpP2PNetwork
import maru.p2p.P2PNetwork
import maru.p2p.P2PNetworkDataProvider
import maru.p2p.P2PNetworkImpl
import maru.p2p.P2PPeersHeadBlockProvider
import maru.p2p.messages.StatusMessageFactory
import maru.serialization.ForkIdSerializer
import maru.serialization.SerDe
import maru.serialization.rlp.RLPSerializers
import maru.syncing.AlwaysSyncedController
import maru.syncing.BeaconSyncControllerImpl
import maru.syncing.ELSyncService
import maru.syncing.HighestHeadTargetSelector
import maru.syncing.MostFrequentHeadTargetSelector
import maru.syncing.PeerChainTracker
import maru.syncing.SyncController
import maru.syncing.SyncTargetSelector
import maru.syncing.beaconchain.pipeline.BeaconChainDownloadPipelineFactory
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.metrics.Tag
import net.consensys.linea.metrics.micrometer.MicrometerMetricsFacade
import net.consensys.linea.vertx.VertxFactory
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.networking.p2p.network.config.GeneratingFilePrivateKeySource
import org.hyperledger.besu.plugin.services.MetricsSystem as BesuMetricsSystem

class MaruAppFactory {
  private val log = LogManager.getLogger(this.javaClass)

  fun create(
    config: MaruConfig,
    beaconGenesisConfig: ForksSchedule,
    clock: Clock = Clock.systemUTC(),
    overridingP2PNetwork: P2PNetwork? = null,
    overridingFinalizationProvider: FinalizationProvider? = null,
    overridingLineaContractClient: LineaRollupSmartContractClientReadOnly? = null,
    overridingApiServer: ApiServer? = null,
    p2pNetworkFactory: (
      ByteArray,
      P2PConfig,
      UInt,
      SerDe<SealedBeaconBlock>,
      MetricsFacade,
      BesuMetricsSystem,
      StatusMessageFactory,
      BeaconChain,
      ForkIdHashProvider,
      ForkIdHasher,
      () -> Boolean,
      P2PState,
    ) -> P2PNetworkImpl = ::P2PNetworkImpl,
  ): MaruApp {
    log.info("configs={}", config)
    log.info("beaconGenesisConfig={}", beaconGenesisConfig)
    val privateKey = getOrGeneratePrivateKey(config.persistence.privateKeyPath)
    val vertx =
      VertxFactory.createVertx(
        jvmMetricsEnabled = config.observability.jvmMetricsEnabled,
        prometheusMetricsEnabled = config.observability.prometheusMetricsEnabled,
      )

    val nodeId = PeerId.fromPubKey(unmarshalPrivateKey(privateKey).publicKey())
    val metricsFacade =
      MicrometerMetricsFacade(
        registry = getMetricsRegistry(),
        metricsPrefix = "maru",
        allMetricsCommonTags = listOf(Tag("nodeid", nodeId.toBase58())),
      )
    val besuMetricsSystemAdapter =
      BesuMetricsSystemAdapter(
        metricsFacade = metricsFacade,
        vertx = vertx,
      )

    ensureDirectoryExists(config.persistence.dataPath)
    val kvDatabase =
      KvDatabaseFactory
        .createRocksDbDatabase(
          databasePath = config.persistence.dataPath,
          metricsSystem = besuMetricsSystemAdapter,
          metricCategory = BesuMetricsCategoryAdapter.from(MaruMetricsCategory.STORAGE),
        ).also {
          dbInitialization(beaconGenesisConfig, it)
        }

    val qbftFork = beaconGenesisConfig.getForkByConfigType(QbftConsensusConfig::class)
    val qbftConfig = qbftFork.configuration as QbftConsensusConfig

    val forkIdHasher =
      ForkIdHasher(
        ForkIdSerializer,
        Hashing::shortShaHash,
      )
    val forkIdHashProvider =
      ForkIdHashProviderImpl(
        chainId = beaconGenesisConfig.chainId,
        beaconChain = kvDatabase,
        forksSchedule = beaconGenesisConfig,
        forkIdHasher = forkIdHasher,
      )
    val ethereumJsonRpcClient =
      Helpers.createWeb3jClient(
        apiEndpointConfig = config.validatorElNode.ethApiEndpoint,
        log = LogManager.getLogger("maru.clients.el.ethapi"),
      )
    val statusMessageFactory = StatusMessageFactory(kvDatabase, forkIdHashProvider)

    val engineApiWeb3jClient =
      Helpers.createWeb3jClient(
        apiEndpointConfig = config.validatorElNode.engineApiEndpoint,
        log = LogManager.getLogger("maru.clients.el.engineapi"),
      )

    val elManagerMap =
      ElFork.entries.associateWith {
        val engineApiClient =
          Helpers.buildExecutionEngineClient(
            web3JEngineApiClient = engineApiWeb3jClient,
            elFork = it,
            metricsFacade = metricsFacade,
          )
        JsonRpcExecutionLayerManager(engineApiClient)
      }

    // Because of the circular dependency between SyncStatusProvider, P2PNetwork and P2PPeersHeadBlockProvider
    var syncControllerImpl: SyncController? = null

    val p2pNetwork =
      overridingP2PNetwork ?: setupP2PNetwork(
        p2pConfig = config.p2p,
        privateKey = privateKey,
        chainId = beaconGenesisConfig.chainId,
        beaconChain = kvDatabase,
        metricsFacade = metricsFacade,
        statusMessageFactory = statusMessageFactory,
        besuMetricsSystem = besuMetricsSystemAdapter,
        forkIdHashProvider = forkIdHashProvider,
        isBlockImportEnabledProvider = { syncControllerImpl!!.isNodeFullInSync() },
        forkIdHasher = forkIdHasher,
        p2PState = kvDatabase,
        p2pNetworkFactory = p2pNetworkFactory,
      )
    val peersHeadBlockProvider = P2PPeersHeadBlockProvider(p2pNetwork.getPeerLookup())
    val finalizationProvider =
      overridingFinalizationProvider
        ?: setupFinalizationProvider(config, overridingLineaContractClient, vertx)
    syncControllerImpl =
      if (config.p2p != null) {
        BeaconSyncControllerImpl.create(
          beaconChain = kvDatabase,
          forksSchedule = beaconGenesisConfig,
          elManagerMap = elManagerMap,
          peersHeadsProvider = peersHeadBlockProvider,
          targetChainHeadCalculator = createSyncTargetSelector(config.syncing.syncTargetSelection),
          validatorProvider = StaticValidatorProvider(qbftConfig.validatorSet),
          peerLookup = p2pNetwork.getPeerLookup(),
          besuMetrics = besuMetricsSystemAdapter,
          metricsFacade = metricsFacade,
          peerChainTrackerConfig =
            PeerChainTracker.Config(
              config.syncing.peerChainHeightPollingInterval,
            ),
          elSyncServiceConfig = ELSyncService.Config(config.syncing.elSyncStatusRefreshInterval),
          finalizationProvider = finalizationProvider,
          desyncTolerance = config.syncing.desyncTolerance,
          pipelineConfig =
            BeaconChainDownloadPipelineFactory.Config(
              blockRangeRequestTimeout = config.syncing.download.blockRangeRequestTimeout,
              backoffDelay = config.syncing.download.backoffDelay,
              blocksBatchSize = config.syncing.download.blocksBatchSize,
              blocksParallelism = config.syncing.download.blocksParallelism,
              maxRetries = config.syncing.download.maxRetries,
              useUnconditionalRandomDownloadPeer = config.syncing.download.useUnconditionalRandomDownloadPeer,
            ),
          allowEmptyBlocks = config.allowEmptyBlocks,
        )
      } else {
        AlwaysSyncedController(kvDatabase)
      }

    val apiServer =
      overridingApiServer
        ?: ApiServerImpl(
          config =
            ApiServerImpl.Config(
              port = config.api.port,
            ),
          networkDataProvider = P2PNetworkDataProvider(p2pNetwork),
          versionProvider = MaruVersionProvider(),
          chainDataProvider = ChainDataProviderImpl(kvDatabase),
          syncStatusProvider = syncControllerImpl,
          isElOnlineProvider = { elManagerMap[ElFork.Prague]!!.isOnline().get() },
        )

    val maru =
      MaruApp(
        config = config,
        beaconGenesisConfig = beaconGenesisConfig,
        clock = clock,
        p2pNetwork = p2pNetwork,
        privateKeyProvider = { privateKey },
        finalizationProvider = finalizationProvider,
        metricsFacade = metricsFacade,
        vertx = vertx,
        beaconChain = kvDatabase,
        metricsSystem = besuMetricsSystemAdapter,
        validatorELNodeEthJsonRpcClient = ethereumJsonRpcClient,
        validatorELNodeEngineApiWeb3JClient = engineApiWeb3jClient,
        apiServer = apiServer,
        syncControllerManager = syncControllerImpl,
        syncStatusProvider = syncControllerImpl,
      )

    return maru
  }

  companion object {
    private val log = LogManager.getLogger(MaruAppFactory::class.java)

    // the second case can happen in case of concurrent access i.e. in the tests
    private fun getMetricsRegistry(): MeterRegistry =
      BackendRegistries.getDefaultNow() ?: BackendRegistries
        .setupBackend(
          MicrometerMetricsOptions(),
          null,
        ).let { BackendRegistries.getDefaultNow() }

    private fun setupFinalizationProvider(
      config: MaruConfig,
      overridingLineaContractClient: LineaRollupSmartContractClientReadOnly?,
      vertx: Vertx,
    ): FinalizationProvider =
      config.linea
        ?.let { lineaConfig ->
          val contractClient =
            overridingLineaContractClient
              ?: Web3JLineaRollupSmartContractClientReadOnly(
                web3j =
                  createWeb3jHttpClient(
                    rpcUrl = lineaConfig.l1EthApi.endpoint.toString(),
                    log = LogManager.getLogger("clients.l1.linea"),
                  ),
                contractAddress = lineaConfig.contractAddress.encodeHex(),
                log = LogManager.getLogger("clients.l1.linea"),
              )
          LineaFinalizationProvider(
            lineaContract = contractClient,
            l2EthApi =
              createEthApiClient(
                rpcUrl =
                  config.validatorElNode.ethApiEndpoint.endpoint
                    .toString(),
                log = LogManager.getLogger("clients.l2.eth.el"),
                requestRetryConfig = config.validatorElNode.ethApiEndpoint.requestRetries,
                vertx = vertx,
                stopRetriesOnErrorPredicate = { true },
              ),
            pollingUpdateInterval = lineaConfig.l1PollingInterval,
            l1HighestBlock = lineaConfig.l1HighestBlockTag,
          )
        } ?: InstantFinalizationProvider

    private fun setupP2PNetwork(
      p2pConfig: P2PConfig?,
      privateKey: ByteArray,
      chainId: UInt,
      beaconChain: BeaconChain,
      isBlockImportEnabledProvider: () -> Boolean,
      metricsFacade: MetricsFacade,
      statusMessageFactory: StatusMessageFactory,
      besuMetricsSystem: BesuMetricsSystem,
      forkIdHashProvider: ForkIdHashProvider,
      forkIdHasher: ForkIdHasher,
      p2PState: P2PState,
      p2pNetworkFactory: (
        ByteArray,
        P2PConfig,
        UInt,
        SerDe<SealedBeaconBlock>,
        MetricsFacade,
        BesuMetricsSystem,
        StatusMessageFactory,
        BeaconChain,
        ForkIdHashProvider,
        ForkIdHasher,
        () -> Boolean,
        P2PState,
      ) -> P2PNetworkImpl = ::P2PNetworkImpl,
    ): P2PNetwork =
      p2pConfig?.let {
        p2pNetworkFactory(
          privateKey,
          NetworkHelper
            .selectIpV4ForP2P(
              targetIpV4 = p2pConfig.ipAddress,
              excludeLoopback = true,
            ).also { log.info("using p2p ip={}", it) }
            .let { p2pConfig.copy(ipAddress = it) },
          chainId,
          RLPSerializers.SealedBeaconBlockCompressorSerializer,
          metricsFacade,
          besuMetricsSystem,
          statusMessageFactory,
          beaconChain,
          forkIdHashProvider,
          forkIdHasher,
          isBlockImportEnabledProvider,
          p2PState,
        )
      } ?: run {
        log.info("No P2P configuration provided, using NoOpP2PNetwork")
        NoOpP2PNetwork
      }

    private fun createSyncTargetSelector(config: SyncingConfig.SyncTargetSelection): SyncTargetSelector =
      when (config) {
        is SyncingConfig.SyncTargetSelection.Highest ->
          HighestHeadTargetSelector()

        is SyncingConfig.SyncTargetSelection.MostFrequent ->
          MostFrequentHeadTargetSelector(config.peerChainHeightGranularity)
      }

    private fun getOrGeneratePrivateKey(privateKeyPath: Path): ByteArray {
      if (!privateKeyPath
          .toFile()
          .exists()
      ) {
        log.info(
          "Private key file {} does not exist. A new private key will be generated and stored in that location.",
          privateKeyPath.toString(),
        )
      } else {
        log.info("Maru is using private key defined in file={}", privateKeyPath.toString())
      }

      return GeneratingFilePrivateKeySource(privateKeyPath.toString()).privateKeyBytes.toArray()
    }
  }

  private fun ensureDirectoryExists(path: Path) {
    if (!path.exists()) Files.createDirectories(path)
  }

  private fun dbInitialization(
    beaconGenesisConfig: ForksSchedule,
    beaconChain: BeaconChain,
  ) {
    val qbftForkConfig = beaconGenesisConfig.getForkByConfigType(QbftConsensusConfig::class)
    val beaconChainInitialization =
      BeaconChainInitialization(
        beaconChain = beaconChain,
      )
    val qbftConsensusConfig = qbftForkConfig.configuration as QbftConsensusConfig
    beaconChainInitialization.ensureDbIsInitialized(qbftConsensusConfig.validatorSet)
  }
}
