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
import maru.config.MaruConfig
import maru.consensus.ElBlockMetadata
import maru.consensus.ForksSchedule
import maru.consensus.LatestElBlockMetadataCache
import maru.consensus.NewBlockHandler
import maru.consensus.NextBlockTimestampProviderImpl
import maru.consensus.OmniProtocolFactory
import maru.consensus.ProtocolStarter
import maru.consensus.state.FinalizationProvider
import maru.core.Protocol
import maru.crypto.Crypto
import maru.database.BeaconChain
import maru.metrics.MaruMetricsCategory
import maru.p2p.P2PNetwork
import maru.p2p.PeerInfo
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
  val beaconGenesisConfig: ForksSchedule,
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
  private val validatorELNodeEthJsonRpcClient: Web3JClient,
  private val validatorELNodeEngineApiWeb3JClient: Web3JClient,
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
      name = "el.block.height",
      description = "Latest EL block height",
      measurementSupplier = {
        lastElBlockMetadataCache.getLatestBlockMetadata().blockNumber.toLong()
      },
    )

    metricsFacade.createGauge(
      category = MaruMetricsCategory.METADATA,
      name = "cl.block.height",
      description = "Latest CL block height",
      measurementSupplier = {
        beaconChain
          .getLatestBeaconState()
          .latestBeaconBlockHeader.number
          .toLong()
      },
    )
  }

  private val followerELNodeEngineApiWeb3JClients: Map<String, Web3JClient> =
    config.followers.followers.mapValues { (_, apiEndpointConfig) ->
      Helpers.createWeb3jClient(apiEndpointConfig)
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
    validatorELNodeEngineApiWeb3JClient.eth1Web3j.shutdown()
    validatorELNodeEthJsonRpcClient.eth1Web3j.shutdown()
    followerELNodeEngineApiWeb3JClients.forEach { (_, web3jClient) -> web3jClient.eth1Web3j.shutdown() }
    p2pNetwork.close()
    vertx.close()
  }

  fun peersConnected(): UInt =
    p2pNetwork
      .getPeers()
      .filter { it.status == PeerInfo.PeerStatus.CONNECTED }
      .size
      .toUInt()

  private fun createProtocolStarter(
    config: MaruConfig,
    beaconGenesisConfig: ForksSchedule,
    clock: Clock,
  ): Protocol {
    val metadataCacheUpdaterHandlerEntry = "latest block metadata updater" to metadataProviderCacheUpdater
    val qbftFactory =
      if (config.qbftOptions != null) {
        QbftProtocolValidatorFactory(
          qbftOptions = config.qbftOptions!!,
          privateKeyBytes = Crypto.privateKeyBytesWithoutPrefix(privateKeyProvider()),
          validatorELNodeEngineApiWeb3JClient = validatorELNodeEngineApiWeb3JClient,
          followerELNodeEngineApiWeb3JClients = followerELNodeEngineApiWeb3JClients,
          metricsSystem = metricsSystem,
          finalizationStateProvider = finalizationProvider,
          nextTargetBlockTimestampProvider = nextTargetBlockTimestampProvider,
          beaconChain = beaconChain,
          clock = clock,
          p2pNetwork = p2pNetwork,
          metricsFacade = metricsFacade,
          allowEmptyBlocks = config.allowEmptyBlocks,
          syncStatusProvider = syncStatusProvider,
          metadataCacheUpdaterHandlerEntry = metadataCacheUpdaterHandlerEntry,
        )
      } else {
        QbftFollowerFactory(
          p2pNetwork = p2pNetwork,
          beaconChain = beaconChain,
          validatorELNodeEngineApiWeb3JClient = validatorELNodeEngineApiWeb3JClient,
          followerELNodeEngineApiWeb3JClients = followerELNodeEngineApiWeb3JClients,
          metricsFacade = metricsFacade,
          allowEmptyBlocks = config.allowEmptyBlocks,
          metadataCacheUpdaterHandlerEntry = metadataCacheUpdaterHandlerEntry,
          finalizationStateProvider = finalizationProvider,
        )
      }

    val protocolStarter =
      ProtocolStarter.create(
        forksSchedule = beaconGenesisConfig,
        protocolFactory =
          OmniProtocolFactory(
            qbftConsensusFactory = qbftFactory,
          ),
        nextBlockTimestampProvider = nextTargetBlockTimestampProvider,
        syncStatusProvider = syncStatusProvider,
        forkTransitionCheckInterval = config.protocolTransitionPollingInterval,
      )

    return protocolStarter
  }
}
