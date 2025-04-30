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
package maru.consensus.qbft

import java.time.Clock
import maru.consensus.NextBlockTimestampProvider
import maru.consensus.ValidatorProvider
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlockBody
import maru.core.Validator
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import org.hyperledger.besu.consensus.common.bft.blockcreation.ProposerSelector
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockCreatorFactory
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockCreator as BesuQbftBlockCreator

/**
 * Maru's QbftBlockCreator factory
 */
class QbftBlockCreatorFactory(
  private val manager: ExecutionLayerManager,
  private val proposerSelector: ProposerSelector,
  private val validatorProvider: ValidatorProvider,
  private val beaconChain: BeaconChain,
  private val finalizationStateProvider: (BeaconBlockBody) -> FinalizationState,
  private val blockBuilderIdentity: Validator,
  private val eagerQbftBlockCreatorConfig: EagerQbftBlockCreator.Config,
  private val nextBlockTimestampProvider: NextBlockTimestampProvider,
  private val clock: Clock,
) : QbftBlockCreatorFactory {
  override fun create(round: Int): BesuQbftBlockCreator {
    val delayedQbftBlockCreator =
      DelayedQbftBlockCreator(
        manager = manager,
        proposerSelector = proposerSelector,
        validatorProvider = validatorProvider,
        beaconChain = beaconChain,
        round = round,
      )
    return EagerQbftBlockCreator(
      manager = manager,
      delegate = delayedQbftBlockCreator,
      finalizationStateProvider = finalizationStateProvider,
      blockBuilderIdentity = blockBuilderIdentity,
      beaconChain = beaconChain,
      nextBlockTimestampProvider = nextBlockTimestampProvider,
      config = eagerQbftBlockCreatorConfig,
      clock = clock,
    )
  }
}
