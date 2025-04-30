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
package maru.consensus.qbft.adapters

import maru.consensus.qbft.ProposerSelector
import maru.consensus.state.StateTransition
import maru.consensus.validation.BlockNumberValidator
import maru.consensus.validation.BodyRootValidator
import maru.consensus.validation.CompositeBlockValidator
import maru.consensus.validation.EmptyBlockValidator
import maru.consensus.validation.ExecutionPayloadValidator
import maru.consensus.validation.ParentRootValidator
import maru.consensus.validation.ProposerValidator
import maru.consensus.validation.StateRootValidator
import maru.consensus.validation.TimestampValidator
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockHeader
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockImporter
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockValidator
import org.hyperledger.besu.consensus.qbft.core.types.QbftProtocolSchedule

class QbftProtocolScheduleAdapter(
  private val blockImporter: QbftBlockImporter,
  private val beaconChain: BeaconChain,
  proposerSelector: ProposerSelector,
  stateTransition: StateTransition,
  executionLayerManager: ExecutionLayerManager,
) : QbftProtocolSchedule {
  private val stateRootValidator = StateRootValidator(stateTransition)
  private val bodyRootValidator = BodyRootValidator()
  private val executionPayloadValidator = ExecutionPayloadValidator(executionLayerManager)
  private val proposerValidator = ProposerValidator(proposerSelector, beaconChain)

  override fun getBlockImporter(blockHeader: QbftBlockHeader): QbftBlockImporter = blockImporter

  override fun getBlockValidator(blockHeader: QbftBlockHeader): QbftBlockValidator {
    val beaconBlockHeader = blockHeader.toBeaconBlockHeader()
    val parentHeader =
      beaconChain.getSealedBeaconBlock(beaconBlockHeader.number - 1UL)!!.beaconBlock.beaconBlockHeader
    val compositeValidator =
      CompositeBlockValidator(
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
    return QbftBlockValidatorAdapter(compositeValidator)
  }
}
