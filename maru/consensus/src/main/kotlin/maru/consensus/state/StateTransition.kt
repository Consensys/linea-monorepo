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
package maru.consensus.state

import maru.consensus.ValidatorProvider
import maru.core.BeaconBlock
import maru.core.BeaconState
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface StateTransition {
  fun processBlock(block: BeaconBlock): SafeFuture<BeaconState>
}

class StateTransitionImpl(
  private val validatorProvider: ValidatorProvider,
) : StateTransition {
  override fun processBlock(block: BeaconBlock): SafeFuture<BeaconState> {
    val validatorsForBlockFuture = validatorProvider.getValidatorsForBlock(block.beaconBlockHeader.number)
    return validatorsForBlockFuture.thenApply { validatorsForBlock ->
      BeaconState(
        latestBeaconBlockHeader = block.beaconBlockHeader,
        validators = validatorsForBlock,
      )
    }
  }
}
