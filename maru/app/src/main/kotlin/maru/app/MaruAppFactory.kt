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
import maru.config.P2P
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ForkIdHashProvider
import maru.consensus.ForkIdHashProviderImpl
import maru.consensus.ForkIdHasher
import maru.consensus.ForksSchedule
import maru.consensus.LatestElBlockMetadataCache
import maru.consensus.StaticValidatorProvider
import maru.consensus.Web3jMetadataProvider
import maru.consensus.state.FinalizationProvider
import maru.consensus.state.InstantFinalizationProvider
import maru.crypto.Hashing
import maru.database.BeaconChain
import maru.database.kv.KvDatabaseFactory
import maru.executionlayer.client.PragueWeb3JJsonRpcExecutionLayerEngineApiClient
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.finalization.LineaFinalizationProvider
import maru.metrics.BesuMetricsCategoryAdapter
import maru.metrics.BesuMetricsSystemAdapter
import maru.metrics.MaruMetricsCategory
import maru.p2p.NoOpP2PNetwork
import maru.p2p.P2PNetwork
import maru.p2p.P2PNetworkDataProvider
import maru.p2p.P2PNetworkImpl
import maru.p2p.P2PPeersHeadBlockProvider
import maru.p2p.messages.StatusMessageFactory
import maru.serialization.ForkIdSerializers
import maru.serialization.rlp.RLPSerializers
import maru.syncing.AlwaysSyncedController
import maru.syncing.BeaconSyncControllerImpl
import maru.syncing.ELSyncService
import maru.syncing.PeerChainTracker
import maru.syncing.SyncController
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
  ): MaruApp {
    log.info("configs: {}", config)
    val privateKey = getOrGeneratePrivateKey(config.persistence.privateKeyPath)
    val vertx =
      VertxFactory.createVertx(
        jvmMetricsEnabled = config.observabilityOptions.jvmMetricsEnabled,
        prometheusMetricsEnabled = config.observabilityOptions.prometheusMetricsEnabled,
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
    val beaconChain =
      KvDatabaseFactory
        .createRocksDbDatabase(
          databasePath = config.persistence.dataPath,
          metricsSystem = besuMetricsSystemAdapter,
          metricCategory = BesuMetricsCategoryAdapter.from(MaruMetricsCategory.STORAGE),
        ).also {
          dbInitialization(beaconGenesisConfig, it)
        }

    val qbftFork = beaconGenesisConfig.getForkByConfigType(QbftConsensusConfig::class)
    val qbftForkTimestamp = qbftFork.timestampSeconds.toULong()
    val qbftConfig = qbftFork.configuration as QbftConsensusConfig
    BeaconChainInitialization(
      beaconChain = beaconChain,
      genesisTimestamp = qbftForkTimestamp,
    ).ensureDbIsInitialized(
      validatorSet = qbftConfig.validatorSet,
    )

    val forkIdHasher =
      ForkIdHasher(
        ForkIdSerializers
          .ForkIdSerializer,
        Hashing::shortShaHash,
      )
    val forkIdHashProvider =
      ForkIdHashProviderImpl(
        chainId = beaconGenesisConfig.chainId,
        beaconChain = beaconChain,
        forksSchedule = beaconGenesisConfig,
        forkIdHasher = forkIdHasher,
      )
    val ethereumJsonRpcClient =
      Helpers.createWeb3jClient(
        config.validatorElNode.ethApiEndpoint,
      )
    val asyncMetadataProvider = Web3jMetadataProvider(ethereumJsonRpcClient.eth1Web3j)
    val latestElBlockMetadataCache =
      LatestElBlockMetadataCache(asyncMetadataProvider.getLatestBlockMetadata())
    val statusMessageFactory = StatusMessageFactory(beaconChain, forkIdHashProvider)
    val engineApiWeb3jClient =
      Helpers.createWeb3jClient(
        config.validatorElNode.engineApiEndpoint,
      )
    val engineApiClient =
      PragueWeb3JJsonRpcExecutionLayerEngineApiClient(
        engineApiWeb3jClient,
        metricsFacade,
      )
    val executionLayerManager = JsonRpcExecutionLayerManager(engineApiClient)
    // Because of the circular dependency between SyncStatusProvider, P2PNetwork and P2PPeersHeadBlockProvider
    var syncControllerImpl: SyncController? = null

    val p2pNetwork =
      overridingP2PNetwork ?: setupP2PNetwork(
        p2pConfig = config.p2pConfig,
        privateKey = privateKey,
        chainId = beaconGenesisConfig.chainId,
        beaconChain = beaconChain,
        metricsFacade = metricsFacade,
        statusMessageFactory = statusMessageFactory,
        besuMetricsSystem = besuMetricsSystemAdapter,
        forkIdHashProvider = forkIdHashProvider,
        isBlockImportEnabledProvider = { syncControllerImpl!!.isBeaconChainSynced() },
      )
    val peersHeadBlockProvider = P2PPeersHeadBlockProvider(p2pNetwork.getPeerLookup())
    val finalizationProvider =
      overridingFinalizationProvider
        ?: setupFinalizationProvider(config, overridingLineaContractClient, vertx)
    syncControllerImpl =
      if (config.p2pConfig != null) {
        BeaconSyncControllerImpl.create(
          beaconChain = beaconChain,
          elManager = executionLayerManager,
          peersHeadsProvider = peersHeadBlockProvider,
          validatorProvider = StaticValidatorProvider(qbftConfig.validatorSet),
          peerLookup = p2pNetwork.getPeerLookup(),
          besuMetrics = besuMetricsSystemAdapter,
          metricsFacade = metricsFacade,
          peerChainTrackerConfig =
            PeerChainTracker.Config(
              config.syncing.peerChainHeightPollingInterval,
              config.syncing.peerChainHeightGranularity,
            ),
          elSyncServiceConfig = ELSyncService.Config(config.syncing.elSyncStatusRefreshInterval),
          finalizationProvider = finalizationProvider,
          allowEmptyBlocks = config.allowEmptyBlocks,
        )
      } else {
        AlwaysSyncedController(beaconChain)
      }

    val apiServer =
      overridingApiServer
        ?: ApiServerImpl(
          config =
            ApiServerImpl.Config(
              port = config.apiConfig.port,
            ),
          networkDataProvider = P2PNetworkDataProvider(p2pNetwork),
          versionProvider = MaruVersionProvider(),
          chainDataProvider = ChainDataProviderImpl(beaconChain),
          syncStatusProvider = syncControllerImpl,
          isElOnlineProvider = { engineApiClient.isOnline().get() },
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
        beaconChain = beaconChain,
        metricsSystem = besuMetricsSystemAdapter,
        lastElBlockMetadataCache = latestElBlockMetadataCache,
        ethereumJsonRpcClient = ethereumJsonRpcClient,
        engineApiWeb3jService = engineApiWeb3jClient,
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
              ),
            pollingUpdateInterval = lineaConfig.l1PollingInterval,
            l1HighestBlock = lineaConfig.l1HighestBlockTag,
          )
        } ?: InstantFinalizationProvider

    private fun setupP2PNetwork(
      p2pConfig: P2P?,
      privateKey: ByteArray,
      chainId: UInt,
      beaconChain: BeaconChain,
      isBlockImportEnabledProvider: () -> Boolean,
      metricsFacade: MetricsFacade,
      statusMessageFactory: StatusMessageFactory,
      besuMetricsSystem: BesuMetricsSystem,
      forkIdHashProvider: ForkIdHashProvider,
    ): P2PNetwork =
      p2pConfig?.let {
        P2PNetworkImpl(
          privateKeyBytes = privateKey,
          p2pConfig = p2pConfig,
          chainId = chainId,
          serDe = RLPSerializers.SealedBeaconBlockSerializer,
          metricsFacade = metricsFacade,
          statusMessageFactory = statusMessageFactory,
          beaconChain = beaconChain,
          isBlockImportEnabledProvider = isBlockImportEnabledProvider,
          metricsSystem = besuMetricsSystem,
          forkIdHashProvider = forkIdHashProvider,
        )
      } ?: run {
        log.info("No P2P configuration provided, using NoOpP2PNetwork")
        NoOpP2PNetwork
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
    val qbftForkTimestamp =
      qbftForkConfig
        .timestampSeconds
        .toULong()
    val beaconChainInitialization =
      BeaconChainInitialization(
        beaconChain = beaconChain,
        genesisTimestamp = qbftForkTimestamp,
      )
    val qbftConsensusConfig = qbftForkConfig.configuration as QbftConsensusConfig
    beaconChainInitialization.ensureDbIsInitialized(qbftConsensusConfig.validatorSet)
  }
}
