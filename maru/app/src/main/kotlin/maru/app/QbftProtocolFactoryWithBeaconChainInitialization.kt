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
import maru.config.QbftOptions
import maru.config.ValidatorElNode
import maru.consensus.ForkSpec
import maru.consensus.NextBlockTimestampProvider
import maru.consensus.ProtocolFactory
import maru.consensus.qbft.QbftConsensusConfig
import maru.consensus.qbft.QbftValidatorFactory
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlockBody
import maru.core.Protocol
import maru.database.BeaconChain
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.p2p.P2PNetwork
import maru.p2p.SealedBeaconBlockHandler
import org.hyperledger.besu.plugin.services.MetricsSystem

class QbftProtocolFactoryWithBeaconChainInitialization(
  private val qbftOptions: QbftOptions,
  private val privateKeyBytes: ByteArray,
  private val validatorElNodeConfig: ValidatorElNode,
  private val metricsSystem: MetricsSystem,
  private val finalizationStateProvider: (BeaconBlockBody) -> FinalizationState,
  private val beaconChain: BeaconChain,
  private val nextTargetBlockTimestampProvider: NextBlockTimestampProvider,
  private val newBlockHandler: SealedBeaconBlockHandler<*>,
  private val clock: Clock,
  private val p2pNetwork: P2PNetwork,
  private val beaconChainInitialization: BeaconChainInitialization,
) : ProtocolFactory {
  override fun create(forkSpec: ForkSpec): Protocol {
    require(forkSpec.configuration is QbftConsensusConfig) {
      "Unexpected fork specification! ${
        forkSpec
          .configuration
      } instead of ${QbftConsensusConfig::class.simpleName}"
    }
    val qbftConsensusConfig = forkSpec.configuration as QbftConsensusConfig

    val engineApiExecutionLayerClient =
      Helpers.buildExecutionEngineClient(
        validatorElNodeConfig.engineApiEndpoint,
        qbftConsensusConfig.elFork,
      )
    val executionLayerManager =
      JsonRpcExecutionLayerManager(
        executionLayerEngineApiClient = engineApiExecutionLayerClient,
      )

    beaconChainInitialization.ensureDbIsInitialized(validatorSet = qbftConsensusConfig.validatorSet)

    val qbftValidatorFactory =
      QbftValidatorFactory(
        beaconChain = beaconChain,
        privateKeyBytes = privateKeyBytes,
        qbftOptions = qbftOptions,
        metricsSystem = metricsSystem,
        finalizationStateProvider = finalizationStateProvider,
        nextBlockTimestampProvider = nextTargetBlockTimestampProvider,
        newBlockHandler = newBlockHandler,
        executionLayerManager = executionLayerManager,
        clock = clock,
        p2PNetwork = p2pNetwork,
      )
    return qbftValidatorFactory.create(forkSpec)
  }
}
