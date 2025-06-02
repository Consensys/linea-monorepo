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
package maru.consensus.validation

import maru.consensus.qbft.ProposerSelector
import maru.consensus.state.StateTransition
import maru.core.BeaconBlockHeader
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager

fun interface BeaconBlockValidatorFactory {
  fun createValidatorForBlock(beaconBlockHeader: BeaconBlockHeader): BlockValidator
}

class BeaconBlockValidatorFactoryImpl(
  val beaconChain: BeaconChain,
  proposerSelector: ProposerSelector,
  stateTransition: StateTransition,
  executionLayerManager: ExecutionLayerManager,
) : BeaconBlockValidatorFactory {
  private val stateRootValidator = StateRootValidator(stateTransition)
  private val bodyRootValidator = BodyRootValidator()
  private val executionPayloadValidator = ExecutionPayloadValidator(executionLayerManager)
  private val proposerValidator = ProposerValidator(proposerSelector, beaconChain)

  override fun createValidatorForBlock(beaconBlockHeader: BeaconBlockHeader): BlockValidator {
    val parentBeaconBlockNumber = beaconBlockHeader.number - 1UL
    val parentBlock = beaconChain.getSealedBeaconBlock(parentBeaconBlockNumber)
    if (parentBlock == null) {
      throw IllegalArgumentException("Expected block for header number=$parentBeaconBlockNumber isn't found!")
    } else {
      val parentHeader = parentBlock.beaconBlock.beaconBlockHeader
      return CompositeBlockValidator(
        blockValidators =
          listOf(
            stateRootValidator,
            BlockNumberValidator(parentHeader),
            TimestampValidator(parentHeader),
            proposerValidator,
            ParentRootValidator(parentHeader),
            bodyRootValidator,
            executionPayloadValidator,
            EmptyBlockValidator,
          ),
      )
    }
  }
}
