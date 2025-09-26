/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import java.time.Clock
import maru.config.QbftConfig
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.NewBlockHandlerMultiplexer
import maru.consensus.NextBlockTimestampProvider
import maru.consensus.ProtocolFactory
import maru.consensus.SealedBeaconBlockHandlerAdapter
import maru.consensus.blockimport.FollowerBeaconBlockImporter
import maru.consensus.blockimport.NewSealedBeaconBlockHandlerMultiplexer
import maru.consensus.qbft.QbftValidatorFactory
import maru.consensus.state.FinalizationProvider
import maru.core.Protocol
import maru.database.BeaconChain
import maru.p2p.P2PNetwork
import maru.p2p.SealedBeaconBlockBroadcaster
import maru.syncing.CLSyncStatus
import maru.syncing.SyncStatusProvider
import net.consensys.linea.metrics.MetricsFacade
import org.hyperledger.besu.plugin.services.MetricsSystem
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient

class QbftProtocolValidatorFactory(
  private val qbftOptions: QbftConfig,
  private val privateKeyBytes: ByteArray,
  private val validatorELNodeEngineApiWeb3JClient: Web3JClient,
  private val followerELNodeEngineApiWeb3JClients: Map<String, Web3JClient>,
  private val metricsSystem: MetricsSystem,
  private val finalizationStateProvider: FinalizationProvider,
  private val beaconChain: BeaconChain,
  private val nextTargetBlockTimestampProvider: NextBlockTimestampProvider,
  private val clock: Clock,
  private val p2pNetwork: P2PNetwork,
  private val metricsFacade: MetricsFacade,
  private val allowEmptyBlocks: Boolean,
  private val syncStatusProvider: SyncStatusProvider,
  private val forksSchedule: ForksSchedule,
) : ProtocolFactory {
  override fun create(forkSpec: ForkSpec): Protocol {
    require(forkSpec.configuration is QbftConsensusConfig) {
      "Unexpected fork specification! ${
        forkSpec
          .configuration
      } instead of ${QbftConsensusConfig::class.simpleName}"
    }
    val qbftConsensusConfig = forkSpec.configuration as QbftConsensusConfig

    val payloadValidatorExecutionLayerManager =
      Helpers.buildExecutionLayerManager(
        web3JEngineApiClient = validatorELNodeEngineApiWeb3JClient,
        elFork = qbftConsensusConfig.elFork,
        metricsFacade = metricsFacade,
      )
    val elFollowersNewBlockHandlerMap =
      followerELNodeEngineApiWeb3JClients.mapValues { (followerName, web3JClient) ->
        val elFollowerExecutionLayerManager =
          Helpers.buildExecutionLayerManager(
            web3JEngineApiClient = web3JClient,
            elFork = qbftConsensusConfig.elFork,
            metricsFacade = metricsFacade,
          )
        FollowerBeaconBlockImporter.create(
          executionLayerManager = elFollowerExecutionLayerManager,
          finalizationStateProvider = finalizationStateProvider,
          importerName = followerName,
        )
      }
    val blockImportHandlers =
      NewBlockHandlerMultiplexer(elFollowersNewBlockHandlerMap)
    val sealedBlockHandlers =
      mutableMapOf(
        "beacon block handlers" to SealedBeaconBlockHandlerAdapter(blockImportHandlers),
        "p2p broadcast sealed beacon block handler" to
          SealedBeaconBlockBroadcaster(p2pNetwork),
      )
    val sealedBlockHandlerMultiplexer = NewSealedBeaconBlockHandlerMultiplexer<Unit>(handlersMap = sealedBlockHandlers)

    val qbftValidatorFactory =
      QbftValidatorFactory(
        beaconChain = beaconChain,
        privateKeyBytes = privateKeyBytes,
        qbftOptions = qbftOptions,
        metricsSystem = metricsSystem,
        finalizationStateProvider = finalizationStateProvider,
        nextBlockTimestampProvider = nextTargetBlockTimestampProvider,
        newBlockHandler = sealedBlockHandlerMultiplexer,
        executionLayerManager = payloadValidatorExecutionLayerManager,
        clock = clock,
        p2PNetwork = p2pNetwork,
        allowEmptyBlocks = allowEmptyBlocks,
        forksSchedule = forksSchedule,
      )
    val qbftProtocol = qbftValidatorFactory.create(forkSpec)
    syncStatusProvider.onFullSyncComplete {
      qbftProtocol.start()
    }
    syncStatusProvider.onClSyncStatusUpdate {
      if (it == CLSyncStatus.SYNCING) {
        qbftProtocol.stop()
      }
    }

    return qbftProtocol
  }
}
