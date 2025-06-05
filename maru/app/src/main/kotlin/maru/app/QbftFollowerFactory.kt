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
