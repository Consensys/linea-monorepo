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

import maru.consensus.ValidatorProvider
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlockBody
import maru.core.Validator
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
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
) : QbftBlockCreatorFactory {
  private val log: Logger = LogManager.getLogger(this.javaClass)
  private var hasCreatedFirstBlockCreator = false

  override fun create(round: Int): BesuQbftBlockCreator {
    val delayedQbftBlockCreator =
      DelayedQbftBlockCreator(
        manager = manager,
        proposerSelector = proposerSelector,
        validatorProvider = validatorProvider,
        beaconChain = beaconChain,
        round = round,
      )
    val blockNumber = beaconChain.getLatestBeaconState().latestBeaconBlockHeader.number + 1u
    val blockCreator = createBlockCreator(round, blockNumber, delayedQbftBlockCreator)
    hasCreatedFirstBlockCreator = true
    return blockCreator
  }

  private fun createBlockCreator(
    round: Int,
    blockNumber: ULong,
    delayedQbftBlockCreator: DelayedQbftBlockCreator,
  ): BesuQbftBlockCreator =
    if (round == 0 && hasCreatedFirstBlockCreator) {
      log.debug("Using delayed block creator number={}, round={}", blockNumber, round)
      delayedQbftBlockCreator
    } else {
      log.debug("Using eager block creator number={}, round={} ", blockNumber, round)
      EagerQbftBlockCreator(
        manager = manager,
        delegate = delayedQbftBlockCreator,
        finalizationStateProvider = finalizationStateProvider,
        blockBuilderIdentity = blockBuilderIdentity,
        beaconChain = beaconChain,
        config = eagerQbftBlockCreatorConfig,
      )
    }
}
