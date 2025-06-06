/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import maru.config.ValidatorElNode
import maru.consensus.ForkSpec
import maru.consensus.NewBlockHandler
import maru.consensus.ProtocolFactory
import maru.consensus.StaticValidatorProvider
import maru.consensus.blockimport.TransactionalSealedBeaconBlockImporter
import maru.consensus.blockimport.ValidatingSealedBeaconBlockImporter
import maru.consensus.qbft.ProposerSelectorImpl
import maru.consensus.qbft.QbftConsensusConfig
import maru.consensus.qbft.QbftConsensusFollower
import maru.consensus.state.StateTransitionImpl
import maru.consensus.validation.BeaconBlockValidatorFactoryImpl
import maru.consensus.validation.QuorumOfSealsVerifier
import maru.consensus.validation.SCEP256SealVerifier
import maru.core.Protocol
import maru.database.BeaconChain
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.p2p.P2PNetwork

class QbftFollowerFactory(
  val p2PNetwork: P2PNetwork,
  val beaconChain: BeaconChain,
  val newBlockHandler: NewBlockHandler<*>,
  val validatorElNodeConfig: ValidatorElNode,
  val beaconChainInitialization: BeaconChainInitialization,
) : ProtocolFactory {
  override fun create(forkSpec: ForkSpec): Protocol {
    val qbftConsensusConfig = (forkSpec.configuration as QbftConsensusConfig)
    val validatorProvider = StaticValidatorProvider(validators = qbftConsensusConfig.validatorSet)
    val stateTransition = StateTransitionImpl(validatorProvider)
    val transactionalSealedBeaconBlockImporter =
      TransactionalSealedBeaconBlockImporter(beaconChain, stateTransition) { _, beaconBlock ->
        newBlockHandler.handleNewBlock(beaconBlock)
      }
    val sealsVerifier = QuorumOfSealsVerifier(validatorProvider, SCEP256SealVerifier())
    val engineApiExecutionLayerClient =
      Helpers.buildExecutionEngineClient(
        validatorElNodeConfig.engineApiEndpoint,
        qbftConsensusConfig.elFork,
      )
    val executionLayerManager =
      JsonRpcExecutionLayerManager(
        executionLayerEngineApiClient = engineApiExecutionLayerClient,
      )
    val beaconBlockValidatorFactory =
      BeaconBlockValidatorFactoryImpl(
        beaconChain = beaconChain,
        proposerSelector = ProposerSelectorImpl,
        stateTransition = stateTransition,
        executionLayerManager = executionLayerManager,
      )
    val blockImporter =
      ValidatingSealedBeaconBlockImporter(
        beaconBlockImporter = transactionalSealedBeaconBlockImporter,
        sealsVerifier = sealsVerifier,
        beaconBlockValidatorFactory = beaconBlockValidatorFactory,
      )

    beaconChainInitialization.ensureDbIsInitialized(qbftConsensusConfig.validatorSet)

    return QbftConsensusFollower(p2PNetwork, blockImporter)
  }
}
