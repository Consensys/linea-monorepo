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
import maru.consensus.DifficultyAwareQbftConfig
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.NextBlockTimestampProviderImpl
import maru.consensus.OmniProtocolFactory
import maru.consensus.ProtocolStarter
import maru.consensus.QbftConsensusConfig
import maru.consensus.qbft.DifficultyAwareQbftFactory
import maru.consensus.state.FinalizationProvider
import maru.core.Protocol
import maru.core.Validator
import maru.crypto.Crypto
import maru.database.BeaconChain
import maru.extensions.encodeHex
import maru.finalization.LineaFinalizationProvider
import maru.metrics.MaruMetricsCategory
import maru.p2p.P2PNetwork
import maru.services.LongRunningService
import maru.subscription.InOrderFanoutSubscriptionManager
import maru.syncing.SyncStatusProvider
import net.consensys.linea.async.get
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.vertx.ObservabilityServer
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.plugin.services.MetricsSystem
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient

class MaruApp(
  val config: MaruConfig,
  val beaconGenesisConfig: ForksSchedule,
  clock: Clock = Clock.systemUTC(),
  // This will only be used if config.p2pConfig is undefined
  val p2pNetwork: P2PNetwork,
  private val privateKeyProvider: () -> ByteArray,
  private val finalizationProvider: FinalizationProvider,
  private val vertx: Vertx,
  private val metricsFacade: MetricsFacade,
  private val beaconChain: BeaconChain,
  private val metricsSystem: MetricsSystem,
  private val validatorELNodeEthJsonRpcClient: Web3JClient,
  private val validatorELNodeEngineApiWeb3JClient: Web3JClient,
  private val apiServer: ApiServer,
  private val syncStatusProvider: SyncStatusProvider,
  private val syncControllerManager: LongRunningService,
) : AutoCloseable {
  private val log: Logger = LogManager.getLogger(this.javaClass)

  private fun getPrivateKeyWithoutPrefix() = Crypto.privateKeyBytesWithoutPrefix(privateKeyProvider())

  init {
    if (config.qbft == null) {
      log.info("Qbft options are not defined. nodeRole=follower")
    } else {
      val localValidator = Crypto.privateKeyToValidator(getPrivateKeyWithoutPrefix())
      log.info("Qbft options are defined. nodeRole=validator with address={}", localValidator.address.encodeHex())
      // TODO: This may be not needed when we use dynamic validator set from a smart contract
      warnIfValidatorIsNotInTheGenesis(localValidator)
    }

    metricsFacade.createGauge(
      category = MaruMetricsCategory.METADATA,
      name = "cl.block.height",
      description = "Latest CL block height",
      measurementSupplier = {
        beaconChain
          .getLatestBeaconState()
          .beaconBlockHeader.number
          .toLong()
      },
    )
  }

  private val followerELNodeEngineApiWeb3JClients: Map<String, Web3JClient> =
    config.followers.followers.mapValues { (followerLabel, apiEndpointConfig) ->
      Helpers.createWeb3jClient(
        apiEndpointConfig = apiEndpointConfig,
        log = LogManager.getLogger("maru.clients.follower.$followerLabel"),
      )
    }

  fun p2pPort(): UInt = p2pNetwork.port

  private val nextTargetBlockTimestampProvider =
    NextBlockTimestampProviderImpl(
      clock = clock,
      forksSchedule = beaconGenesisConfig,
    )
  private val protocolStarter =
    createProtocolStarter(
      config = config,
      beaconGenesisConfig = beaconGenesisConfig,
      clock = clock,
      beaconChain = beaconChain,
      privateKeyWithoutPrefix = getPrivateKeyWithoutPrefix(),
    )

  fun start() {
    if (finalizationProvider is LineaFinalizationProvider) {
      start("Finalization Provider", finalizationProvider::start)
    }
    start("P2P Network") { p2pNetwork.start().get() }
    start("Sync Service", syncControllerManager::start)
    start("Beacon Api", apiServer::start)
    // observability shall be the last to start because of liveness/readiness probe
    start("Observability Server") {
      ObservabilityServer(
        ObservabilityServer.Config(applicationName = "maru", port = config.observability.port.toInt()),
      ).let { vertx.deployVerticle(it).get() }
    }
    log.info("Maru is up")
  }

  fun stop() {
    stop("Sync service", syncControllerManager::stop)
    stop("P2P Network") { p2pNetwork.stop().get() }
    if (finalizationProvider is LineaFinalizationProvider) {
      stop("Finalization Provider", finalizationProvider::stop)
    }
    stop("Beacon API", apiServer::stop)
    stop("Protocol", protocolStarter::pause)
    stop("vertx verticles") {
      vertx.deploymentIDs().forEach {
        vertx.undeploy(it).get()
      }
    }
    log.info("Maru is down")
  }

  override fun close() {
    validatorELNodeEngineApiWeb3JClient.eth1Web3j.shutdown()
    validatorELNodeEthJsonRpcClient.eth1Web3j.shutdown()
    followerELNodeEngineApiWeb3JClients.forEach { (_, web3jClient) -> web3jClient.eth1Web3j.shutdown() }
    p2pNetwork.close()
    vertx.close()
    // close db last, otherwise other components may fail trying to save data
    beaconChain.close()
    protocolStarter.close()
  }

  private fun start(
    serviceName: String,
    action: () -> Unit,
  ) {
    runCatching(action)
      .onSuccess { log.info("{} started!", serviceName) }
      .onFailure { log.error("Failed to start {}, errorMessage={}", serviceName, it.message, it) }
      .getOrThrow()
  }

  private fun stop(
    serviceName: String,
    action: () -> Unit,
  ) = runCatching(action)
    .getOrElse { log.warn("Failed to stop {}, errorMessage={}", serviceName, it.message, it) }

  fun peersConnected(): UInt =
    p2pNetwork
      .peerCount
      .toUInt()

  private fun createProtocolStarter(
    config: MaruConfig,
    beaconGenesisConfig: ForksSchedule,
    clock: Clock,
    beaconChain: BeaconChain,
    privateKeyWithoutPrefix: ByteArray,
  ): Protocol {
    val qbftFactory =
      if (config.qbft != null) {
        QbftProtocolValidatorFactory(
          qbftOptions = config.qbft!!,
          privateKeyBytes = privateKeyWithoutPrefix,
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
          forksSchedule = beaconGenesisConfig,
          payloadValidationEnabled = config.validatorElNode.payloadValidationEnabled,
        )
      } else {
        QbftFollowerFactory(
          p2pNetwork = p2pNetwork,
          beaconChain = beaconChain,
          validatorELNodeEngineApiWeb3JClient = validatorELNodeEngineApiWeb3JClient,
          followerELNodeEngineApiWeb3JClients = followerELNodeEngineApiWeb3JClients,
          metricsFacade = metricsFacade,
          allowEmptyBlocks = config.allowEmptyBlocks,
          finalizationStateProvider = finalizationProvider,
          payloadValidationEnabled = config.validatorElNode.payloadValidationEnabled,
        )
      }
    val forkTransitionSubscriptionManager = InOrderFanoutSubscriptionManager<ForkSpec>()
    forkTransitionSubscriptionManager.addSyncSubscriber(p2pNetwork::handleForkTransition)
    val difficultyAwareQbftFactory =
      DifficultyAwareQbftFactory(
        ethereumJsonRpcClient = validatorELNodeEthJsonRpcClient.eth1Web3j,
        postTtdProtocolFactory = qbftFactory,
      )
    val protocolStarter =
      ProtocolStarter.create(
        forksSchedule = beaconGenesisConfig,
        protocolFactory =
          OmniProtocolFactory(
            qbftConsensusFactory = qbftFactory,
            difficultyAwareQbftFactory = difficultyAwareQbftFactory,
          ),
        nextBlockTimestampProvider = nextTargetBlockTimestampProvider,
        syncStatusProvider = syncStatusProvider,
        forkTransitionCheckInterval = config.protocolTransitionPollingInterval,
        forkTransitionNotifier = forkTransitionSubscriptionManager,
        clock = clock,
      )

    return protocolStarter
  }

  private fun warnIfValidatorIsNotInTheGenesis(localValidator: Validator) {
    val validatorsFromAllForks: Set<Validator> =
      beaconGenesisConfig.forks
        .flatMap<ForkSpec, Validator> {
          when (val configuration = it.configuration) {
            is DifficultyAwareQbftConfig -> configuration.postTtdConfig.validatorSet
            is QbftConsensusConfig -> configuration.validatorSet
            else -> throw IllegalArgumentException("")
          }
        }.toSet()
    if (!validatorsFromAllForks.contains(localValidator)) {
      log.warn(
        "localValidator={} isn't found in any of validatorSet-s in any of the Forks in the Genesis file!",
        localValidator,
      )
    }
  }
}
