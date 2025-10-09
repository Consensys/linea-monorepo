/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package testutils.maru

import io.libp2p.core.PeerId
import io.libp2p.core.crypto.KeyType
import io.libp2p.core.crypto.generateKeyPair
import io.libp2p.core.crypto.marshalPrivateKey
import io.libp2p.core.crypto.unmarshalPrivateKey
import java.net.URI
import java.nio.file.Files
import java.nio.file.Path
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import linea.contract.l1.LineaRollupSmartContractClientReadOnly
import linea.kotlin.decodeHex
import linea.kotlin.encodeHex
import maru.api.ApiServer
import maru.app.MaruApp
import maru.app.MaruAppFactory
import maru.config.ApiConfig
import maru.config.ApiEndpointConfig
import maru.config.FollowersConfig
import maru.config.LineaConfig
import maru.config.MaruConfig
import maru.config.ObservabilityConfig
import maru.config.P2PConfig
import maru.config.Persistence
import maru.config.QbftConfig
import maru.config.SyncingConfig
import maru.config.SyncingConfig.SyncTargetSelection
import maru.config.ValidatorElNode
import maru.consensus.ChainFork
import maru.consensus.ClFork
import maru.consensus.DifficultyAwareQbftConfig
import maru.consensus.ElFork
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.QbftConsensusConfig
import maru.consensus.state.FinalizationProvider
import maru.core.SealedBeaconBlock
import maru.core.Validator
import maru.crypto.Crypto
import maru.database.BeaconChain
import maru.database.P2PState
import maru.extensions.fromHexToByteArray
import maru.p2p.NoOpP2PNetwork
import maru.p2p.P2PNetwork
import maru.p2p.P2PNetworkImpl
import maru.p2p.fork.ForkPeeringManager
import maru.p2p.messages.StatusManager
import maru.serialization.SerDe
import maru.syncing.SyncStatusProvider
import net.consensys.linea.metrics.MetricsFacade
import org.hyperledger.besu.plugin.services.MetricsSystem as BesuMetricsSystem

/**
 * The same MaruFactory should be used per network. Otherwise, validators won't match between Maru instances
 */
class MaruFactory(
  validatorPrivateKey: ByteArray = generatePrivateKey(),
  shanghaiTimestamp: ULong? = null,
  cancunTimestamp: ULong? = null,
  pragueTimestamp: ULong? = null,
  ttd: ULong? = null,
) {
  init {
    // If one of pragueTimestamp, cancunTimestamp, shanghaiTimestamp is defined and some other is not, throw
    require(
      (pragueTimestamp == null && cancunTimestamp == null && shanghaiTimestamp == null) ||
        (pragueTimestamp != null && cancunTimestamp != null && shanghaiTimestamp != null),
    ) {
      "pragueTimestamp, cancunTimestamp and shanghaiTimestamp should be defined or all be absent!"
    }
  }

  companion object {
    val defaultReconnectDelay = 500.milliseconds
    val defaultSyncingConfig =
      SyncingConfig(
        peerChainHeightPollingInterval = 1.seconds,
        syncTargetSelection = SyncTargetSelection.Highest,
        elSyncStatusRefreshInterval = 500.milliseconds,
        desyncTolerance = 0UL,
        download = SyncingConfig.Download(),
      )

    fun enumeratingSyncingConfigs(): List<SyncingConfig> {
      val syncTargetSelectionForMostFrequent =
        SyncTargetSelection.MostFrequent(
          peerChainHeightGranularity = 10U,
        )
      return listOf(
        defaultSyncingConfig,
        defaultSyncingConfig.copy(
          syncTargetSelection = syncTargetSelectionForMostFrequent,
          download = SyncingConfig.Download(useUnconditionalRandomDownloadPeer = true),
        ),
      )
    }

    fun generatePrivateKey(): ByteArray = marshalPrivateKey(generateKeyPair(KeyType.SECP256K1).component1())
  }

  private val validatorPrivateKeyWithPrefix = unmarshalPrivateKey(validatorPrivateKey)
  private val validatorPrivateKeyWithPrefixString = marshalPrivateKey(validatorPrivateKeyWithPrefix).encodeHex()
  private val validatorNodeId = PeerId.fromPubKey(validatorPrivateKeyWithPrefix.publicKey())
  val qbftValidator =
    Crypto.privateKeyToValidator(Crypto.privateKeyBytesWithoutPrefix(validatorPrivateKeyWithPrefix.bytes()))
  val validatorAddress = qbftValidator.address.encodeHex()

  private val validatorQbftOptions =
    QbftConfig(
      feeRecipient = qbftValidator.address.reversedArray(),
      minBlockBuildTime = 200.milliseconds,
    )

  private val beaconGenesisConfig: ForksSchedule =
    if (pragueTimestamp != null && cancunTimestamp != null && shanghaiTimestamp != null) {
      ForksSchedule(
        1337u,
        setOf(
          ForkSpec(
            timestampSeconds = 0UL,
            blockTimeSeconds = 1u,
            configuration =
              DifficultyAwareQbftConfig(
                QbftConsensusConfig(
                  validatorSet = setOf(Validator(validatorAddress.fromHexToByteArray())),
                  fork = ChainFork(ClFork.QBFT_PHASE0, ElFork.Paris),
                ),
                terminalTotalDifficulty = ttd!!,
              ),
          ),
          ForkSpec(
            timestampSeconds = shanghaiTimestamp,
            blockTimeSeconds = 1u,
            configuration =
              QbftConsensusConfig(
                validatorSet = setOf(Validator(validatorAddress.fromHexToByteArray())),
                fork = ChainFork(ClFork.QBFT_PHASE0, ElFork.Shanghai),
              ),
          ),
          ForkSpec(
            timestampSeconds = cancunTimestamp,
            blockTimeSeconds = 1u,
            configuration =
              QbftConsensusConfig(
                validatorSet = setOf(Validator(validatorAddress.fromHexToByteArray())),
                fork = ChainFork(ClFork.QBFT_PHASE0, ElFork.Cancun),
              ),
          ),
          ForkSpec(
            timestampSeconds = pragueTimestamp,
            blockTimeSeconds = 1u,
            configuration =
              QbftConsensusConfig(
                validatorSet = setOf(Validator(validatorAddress.fromHexToByteArray())),
                fork = ChainFork(ClFork.QBFT_PHASE0, ElFork.Prague),
              ),
          ),
        ),
      )
    } else {
      ForksSchedule(
        1337u,
        setOf(
          ForkSpec(
            timestampSeconds = 0UL,
            blockTimeSeconds = 1u,
            configuration =
              QbftConsensusConfig(
                validatorSet = setOf(Validator(validatorAddress.fromHexToByteArray())),
                fork = ChainFork(ClFork.QBFT_PHASE0, ElFork.Prague),
              ),
          ),
        ),
      )
    }

  private fun buildMaruConfig(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    p2pConfig: P2PConfig? = null,
    followers: FollowersConfig = FollowersConfig(emptyMap()),
    qbftOptions: QbftConfig? = null,
    observabilityOptions: ObservabilityConfig =
      ObservabilityConfig(port = 0u, prometheusMetricsEnabled = true, jvmMetricsEnabled = true),
    overridingLineaContractClient: LineaRollupSmartContractClientReadOnly? = null,
    apiConfig: ApiConfig = ApiConfig(port = 0u),
    syncingConfig: SyncingConfig = defaultSyncingConfig,
    allowEmptyBlocks: Boolean = false,
    enablePayloadValidation: Boolean = true,
  ): MaruConfig {
    val lineaConfig =
      overridingLineaContractClient?.let {
        LineaConfig(
          contractAddress = overridingLineaContractClient.getAddress().decodeHex(),
          l1EthApi = ApiEndpointConfig(URI.create(ethereumJsonRpcUrl).toURL()),
          l1PollingInterval = 100.milliseconds,
        )
      }

    return MaruConfig(
      allowEmptyBlocks = allowEmptyBlocks,
      persistence = Persistence(dataPath = dataDir),
      qbft = qbftOptions,
      validatorElNode =
        ValidatorElNode(
          ethApiEndpoint = ApiEndpointConfig(URI.create(ethereumJsonRpcUrl).toURL()),
          engineApiEndpoint = ApiEndpointConfig(URI.create(engineApiRpc).toURL()),
          payloadValidationEnabled = enablePayloadValidation,
        ),
      p2p = p2pConfig,
      followers = followers,
      observability = observabilityOptions,
      linea = lineaConfig,
      api = apiConfig,
      syncing = syncingConfig,
    )
  }

  fun buildTestMaruFollowerWithConsensusSwitch(
    dataDir: Path,
    engineApiConfig: ApiEndpointConfig,
    ethereumApiConfig: ApiEndpointConfig,
    validatorPortForStaticPeering: UInt? = null,
    overridingP2PNetwork: P2PNetwork? = null,
    desyncTolerance: ULong = 10UL,
  ): MaruApp {
    val p2pConfig = buildP2pConfig(validatorPortForStaticPeering = validatorPortForStaticPeering)
    val beaconGenesisConfig = beaconGenesisConfig

    val syncingConfig =
      SyncingConfig(
        peerChainHeightPollingInterval = 1.seconds,
        syncTargetSelection = SyncTargetSelection.Highest,
        elSyncStatusRefreshInterval = 500.milliseconds,
        desyncTolerance = desyncTolerance,
        download = SyncingConfig.Download(),
      )
    val config =
      buildMaruConfig(
        engineApiEndpointConfig = engineApiConfig,
        ethereumApiEndpointConfig = ethereumApiConfig,
        dataDir = dataDir,
        p2pConfig = p2pConfig,
        syncingConfig = syncingConfig,
      )
    return buildApp(
      config = config,
      beaconGenesisConfig = beaconGenesisConfig,
      overridingP2PNetwork = overridingP2PNetwork,
    )
  }

  private fun buildMaruConfig(
    engineApiEndpointConfig: ApiEndpointConfig,
    ethereumApiEndpointConfig: ApiEndpointConfig,
    dataDir: Path,
    p2pConfig: P2PConfig? = null,
    followers: FollowersConfig = FollowersConfig(emptyMap()),
    qbftOptions: QbftConfig? = null,
    observabilityOptions: ObservabilityConfig =
      ObservabilityConfig(port = 0u, prometheusMetricsEnabled = true, jvmMetricsEnabled = true),
    overridingLineaContractClient: LineaRollupSmartContractClientReadOnly? = null,
    apiConfig: ApiConfig = ApiConfig(port = 0u),
    syncingConfig: SyncingConfig = defaultSyncingConfig,
    allowEmptyBlocks: Boolean = false,
    enablePayloadValidation: Boolean = true,
  ): MaruConfig {
    val lineaConfig =
      overridingLineaContractClient?.let {
        LineaConfig(
          contractAddress = overridingLineaContractClient.getAddress().decodeHex(),
          l1EthApi = ethereumApiEndpointConfig,
          l1PollingInterval = 100.milliseconds,
        )
      }

    return MaruConfig(
      allowEmptyBlocks = allowEmptyBlocks,
      persistence = Persistence(dataPath = dataDir),
      qbft = qbftOptions,
      validatorElNode =
        ValidatorElNode(
          ethApiEndpoint = ethereumApiEndpointConfig,
          engineApiEndpoint = engineApiEndpointConfig,
          payloadValidationEnabled = enablePayloadValidation,
        ),
      p2p = p2pConfig,
      followers = followers,
      observability = observabilityOptions,
      linea = lineaConfig,
      api = apiConfig,
      syncing = syncingConfig,
    )
  }

  private fun writeValidatorPrivateKey(config: MaruConfig) {
    Files.writeString(config.persistence.privateKeyPath, validatorPrivateKeyWithPrefixString)
  }

  private fun buildApp(
    config: MaruConfig,
    beaconGenesisConfig: ForksSchedule = this.beaconGenesisConfig,
    overridingP2PNetwork: P2PNetwork? = null,
    overridingFinalizationProvider: FinalizationProvider? = null,
    overridingLineaContractClient: LineaRollupSmartContractClientReadOnly? = null,
    p2pNetworkFactory: (
      ByteArray,
      P2PConfig,
      UInt,
      SerDe<SealedBeaconBlock>,
      MetricsFacade,
      BesuMetricsSystem,
      StatusManager,
      BeaconChain,
      ForkPeeringManager,
      () -> Boolean,
      P2PState,
      () -> SyncStatusProvider,
    ) -> P2PNetworkImpl = ::P2PNetworkImpl,
  ): MaruApp =
    MaruAppFactory().create(
      config = config,
      beaconGenesisConfig = beaconGenesisConfig,
      overridingP2PNetwork = overridingP2PNetwork,
      overridingFinalizationProvider = overridingFinalizationProvider,
      overridingLineaContractClient = overridingLineaContractClient,
      overridingApiServer =
        object : ApiServer {
          override fun start() {}

          override fun stop() {}

          override fun port(): Int = 0
        },
      p2pNetworkFactory = p2pNetworkFactory,
    )

  private fun buildP2pConfig(
    p2pPort: UInt = 0u,
    validatorPortForStaticPeering: UInt? = null,
    discoveryPort: UInt? = null,
    bootnode: String? = null,
    banPeriod: Duration = 2000.milliseconds,
    cooldownPeriod: Duration = 1000.milliseconds,
  ): P2PConfig {
    val ip = "127.0.0.1"
    val staticPeers =
      if (validatorPortForStaticPeering != null) {
        val validatorPeer = "/ip4/$ip/tcp/$validatorPortForStaticPeering/p2p/$validatorNodeId"
        listOf(validatorPeer)
      } else {
        emptyList()
      }
    val discovery =
      if (discoveryPort != null) {
        P2PConfig.Discovery(refreshInterval = 1.seconds, port = discoveryPort, bootnodes = listOfNotNull(bootnode))
      } else {
        null
      }
    return P2PConfig(
      ip,
      port = p2pPort,
      staticPeers = staticPeers,
      reconnectDelay = defaultReconnectDelay,
      statusUpdate = P2PConfig.StatusUpdate(refreshInterval = 1.seconds), // For faster syncing in the tests
      reputation = P2PConfig.Reputation(banPeriod = banPeriod, cooldownPeriod = cooldownPeriod),
      discovery = discovery,
    )
  }

  fun buildTestMaruValidatorWithoutP2pPeering(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    overridingP2PNetwork: P2PNetwork? = null,
    allowEmptyBlocks: Boolean = false,
    syncingConfig: SyncingConfig = defaultSyncingConfig,
  ): MaruApp {
    val config =
      buildMaruConfig(
        ethereumJsonRpcUrl = ethereumJsonRpcUrl,
        engineApiRpc = engineApiRpc,
        dataDir = dataDir,
        qbftOptions = validatorQbftOptions,
        allowEmptyBlocks = allowEmptyBlocks,
        syncingConfig = syncingConfig,
      )
    writeValidatorPrivateKey(config)
    return buildApp(config, overridingP2PNetwork = overridingP2PNetwork)
  }

  fun buildTestMaruValidatorWithP2pPeering(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    overridingP2PNetwork: P2PNetwork? = null,
    overridingFinalizationProvider: FinalizationProvider? = null,
    overridingLineaContractClient: LineaRollupSmartContractClientReadOnly? = null,
    p2pPort: UInt = 0u,
    allowEmptyBlocks: Boolean = false,
    syncingConfig: SyncingConfig = defaultSyncingConfig,
    p2pNetworkFactory: (
      ByteArray,
      P2PConfig,
      UInt,
      SerDe<SealedBeaconBlock>,
      MetricsFacade,
      BesuMetricsSystem,
      StatusManager,
      BeaconChain,
      ForkPeeringManager,
      () -> Boolean,
      P2PState,
      () -> SyncStatusProvider,
    ) -> P2PNetworkImpl = ::P2PNetworkImpl,
  ): MaruApp {
    val p2pConfig = buildP2pConfig(p2pPort = p2pPort)
    val config =
      buildMaruConfig(
        ethereumJsonRpcUrl = ethereumJsonRpcUrl,
        engineApiRpc = engineApiRpc,
        dataDir = dataDir,
        p2pConfig = p2pConfig,
        followers = FollowersConfig(emptyMap()),
        qbftOptions = validatorQbftOptions,
        overridingLineaContractClient = overridingLineaContractClient,
        allowEmptyBlocks = allowEmptyBlocks,
        syncingConfig = syncingConfig,
      )
    writeValidatorPrivateKey(config)

    return buildApp(
      config = config,
      overridingP2PNetwork = overridingP2PNetwork,
      overridingFinalizationProvider = overridingFinalizationProvider,
      overridingLineaContractClient = overridingLineaContractClient,
      p2pNetworkFactory = p2pNetworkFactory,
    )
  }

  fun buildTestMaruValidatorWithDiscovery(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    overridingP2PNetwork: P2PNetwork? = null,
    overridingFinalizationProvider: FinalizationProvider? = null,
    overridingLineaContractClient: LineaRollupSmartContractClientReadOnly? = null,
    p2pPort: UInt = 0u,
    discoveryPort: UInt = 0u,
    bootnode: String? = null,
    banPeriod: Duration = 10.seconds,
    cooldownPeriod: Duration = 10.seconds,
    blockRangeRequestTimeout: Duration = 5.seconds,
    allowEmptyBlocks: Boolean = false,
    syncingConfig: SyncingConfig =
      SyncingConfig(
        peerChainHeightPollingInterval = 1.seconds,
        syncTargetSelection = SyncTargetSelection.Highest,
        elSyncStatusRefreshInterval = 500.milliseconds,
        desyncTolerance = 0UL,
        download = SyncingConfig.Download(blockRangeRequestTimeout = blockRangeRequestTimeout),
      ),
    p2pNetworkFactory: (
      ByteArray,
      P2PConfig,
      UInt,
      SerDe<SealedBeaconBlock>,
      MetricsFacade,
      BesuMetricsSystem,
      StatusManager,
      BeaconChain,
      ForkPeeringManager,
      () -> Boolean,
      P2PState,
      () -> SyncStatusProvider,
    ) -> P2PNetworkImpl = ::P2PNetworkImpl,
  ): MaruApp {
    val p2pConfig =
      buildP2pConfig(
        p2pPort = p2pPort,
        discoveryPort = discoveryPort,
        bootnode = bootnode,
        banPeriod = banPeriod,
        cooldownPeriod = cooldownPeriod,
      )
    val config =
      buildMaruConfig(
        ethereumJsonRpcUrl = ethereumJsonRpcUrl,
        engineApiRpc = engineApiRpc,
        dataDir = dataDir,
        p2pConfig = p2pConfig,
        followers = FollowersConfig(emptyMap()),
        qbftOptions = validatorQbftOptions,
        overridingLineaContractClient = overridingLineaContractClient,
        allowEmptyBlocks = allowEmptyBlocks,
        syncingConfig = syncingConfig,
      )
    writeValidatorPrivateKey(config)

    return buildApp(
      config = config,
      overridingP2PNetwork = overridingP2PNetwork,
      overridingFinalizationProvider = overridingFinalizationProvider,
      overridingLineaContractClient = overridingLineaContractClient,
      p2pNetworkFactory = p2pNetworkFactory,
    )
  }

  fun buildTestMaruFollowerWithDiscovery(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    overridingP2PNetwork: P2PNetwork? = null,
    overridingFinalizationProvider: FinalizationProvider? = null,
    overridingLineaContractClient: LineaRollupSmartContractClientReadOnly? = null,
    p2pPort: UInt = 0u,
    discoveryPort: UInt = 0u,
    bootnode: String? = null,
    banPeriod: Duration = 10.seconds,
    cooldownPeriod: Duration = 10.seconds,
    allowEmptyBlocks: Boolean = false,
    blockRangeRequestTimeout: Duration = 5.seconds,
    syncingConfig: SyncingConfig =
      SyncingConfig(
        peerChainHeightPollingInterval = 1.seconds,
        syncTargetSelection = SyncTargetSelection.Highest,
        elSyncStatusRefreshInterval = 500.milliseconds,
        desyncTolerance = 0UL,
        download = SyncingConfig.Download(blockRangeRequestTimeout = blockRangeRequestTimeout),
      ),
    p2pNetworkFactory: (
      ByteArray,
      P2PConfig,
      UInt,
      SerDe<SealedBeaconBlock>,
      MetricsFacade,
      BesuMetricsSystem,
      StatusManager,
      BeaconChain,
      ForkPeeringManager,
      () -> Boolean,
      P2PState,
      () -> SyncStatusProvider,
    ) -> P2PNetworkImpl = ::P2PNetworkImpl,
  ): MaruApp {
    val p2pConfig =
      buildP2pConfig(
        p2pPort = p2pPort,
        discoveryPort = discoveryPort,
        bootnode = bootnode,
        banPeriod = banPeriod,
        cooldownPeriod = cooldownPeriod,
      )
    val config =
      buildMaruConfig(
        ethereumJsonRpcUrl = ethereumJsonRpcUrl,
        engineApiRpc = engineApiRpc,
        dataDir = dataDir,
        p2pConfig = p2pConfig,
        overridingLineaContractClient = overridingLineaContractClient,
        allowEmptyBlocks = allowEmptyBlocks,
        syncingConfig = syncingConfig,
      )

    return buildApp(
      config = config,
      overridingP2PNetwork = overridingP2PNetwork,
      overridingFinalizationProvider = overridingFinalizationProvider,
      overridingLineaContractClient = overridingLineaContractClient,
      p2pNetworkFactory = p2pNetworkFactory,
    )
  }

  fun buildTestMaruFollowerWithP2pPeering(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    validatorPortForStaticPeering: UInt?,
    followers: FollowersConfig = FollowersConfig(emptyMap()),
    overridingFinalizationProvider: FinalizationProvider? = null,
    overridingLineaContractClient: LineaRollupSmartContractClientReadOnly? = null,
    allowEmptyBlocks: Boolean = false,
    syncingConfig: SyncingConfig = defaultSyncingConfig,
    enablePayloadValidation: Boolean = true,
  ): MaruApp {
    val p2pConfig = buildP2pConfig(validatorPortForStaticPeering = validatorPortForStaticPeering)
    val config =
      buildMaruConfig(
        allowEmptyBlocks = allowEmptyBlocks,
        ethereumJsonRpcUrl = ethereumJsonRpcUrl,
        engineApiRpc = engineApiRpc,
        dataDir = dataDir,
        p2pConfig = p2pConfig,
        followers = followers,
        overridingLineaContractClient = overridingLineaContractClient,
        syncingConfig = syncingConfig,
        enablePayloadValidation = enablePayloadValidation,
      )
    return buildApp(
      config,
      overridingFinalizationProvider = overridingFinalizationProvider,
      overridingLineaContractClient = overridingLineaContractClient,
    )
  }

  fun buildTestMaruFollowerWithoutP2pPeering(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    p2pNetwork: P2PNetwork = NoOpP2PNetwork,
    syncingConfig: SyncingConfig = defaultSyncingConfig,
  ): MaruApp {
    val config =
      buildMaruConfig(
        ethereumJsonRpcUrl = ethereumJsonRpcUrl,
        engineApiRpc = engineApiRpc,
        dataDir = dataDir,
        syncingConfig = syncingConfig,
      )
    return buildApp(config, overridingP2PNetwork = p2pNetwork)
  }

  fun buildSwitchableTestMaruValidatorWithP2pPeering(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    overridingP2PNetwork: P2PNetwork? = null,
    overridingFinalizationProvider: FinalizationProvider? = null,
    overridingLineaContractClient: LineaRollupSmartContractClientReadOnly? = null,
    p2pPort: UInt = 0u,
    allowEmptyBlocks: Boolean = false,
    followers: FollowersConfig = FollowersConfig(emptyMap()),
    syncingConfig: SyncingConfig = defaultSyncingConfig,
  ): MaruApp {
    val beaconGenesisConfig = beaconGenesisConfig
    val p2pConfig = buildP2pConfig(p2pPort = p2pPort)
    val config =
      buildMaruConfig(
        ethereumJsonRpcUrl = ethereumJsonRpcUrl,
        engineApiRpc = engineApiRpc,
        dataDir = dataDir,
        p2pConfig = p2pConfig,
        followers = followers,
        qbftOptions = validatorQbftOptions,
        overridingLineaContractClient = overridingLineaContractClient,
        allowEmptyBlocks = allowEmptyBlocks,
        syncingConfig = syncingConfig,
      )
    writeValidatorPrivateKey(config)

    return buildApp(
      config = config,
      beaconGenesisConfig = beaconGenesisConfig,
      overridingP2PNetwork = overridingP2PNetwork,
      overridingFinalizationProvider = overridingFinalizationProvider,
      overridingLineaContractClient = overridingLineaContractClient,
    )
  }

  fun buildTestMaruFollowerWithConsensusSwitch(
    ethereumJsonRpcUrl: String,
    engineApiRpc: String,
    dataDir: Path,
    validatorPortForStaticPeering: UInt? = null,
    overridingP2PNetwork: P2PNetwork? = null,
    desyncTolerance: ULong = 10UL,
  ): MaruApp =
    buildTestMaruFollowerWithConsensusSwitch(
      ethereumApiConfig = ApiEndpointConfig(endpoint = URI.create(ethereumJsonRpcUrl).toURL()),
      engineApiConfig = ApiEndpointConfig(endpoint = URI.create(engineApiRpc).toURL()),
      dataDir = dataDir,
      validatorPortForStaticPeering = validatorPortForStaticPeering,
      overridingP2PNetwork = overridingP2PNetwork,
      desyncTolerance = desyncTolerance,
    )
}
