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

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import maru.consensus.ProposerSelector
import maru.consensus.ValidatorProvider
import maru.consensus.validation.BlockValidator
import maru.consensus.validation.BlockValidator.Companion.error
import maru.consensus.validation.BlockValidator.Companion.ok
import maru.core.BeaconBlock
import maru.core.BeaconBlockHeader
import maru.core.BeaconState
import maru.core.HashUtil
import maru.core.Validator
import maru.core.ext.DataGenerators
import maru.serialization.rlp.bodyRoot
import maru.serialization.rlp.stateRoot
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture

class StateTransitionImplTest {
  private val currentBlock = DataGenerators.randomBeaconBlock(9u)
  private val preState =
    DataGenerators.randomBeaconState(9u).copy(
      latestBeaconBlockHeader = currentBlock.beaconBlockHeader,
      latestBeaconBlockRoot = HashUtil.bodyRoot(currentBlock.beaconBlockBody),
    )

  private val newBeaconBlockBody = DataGenerators.randomBeaconBlockBody()
  private val stateRootBlockHeader =
    DataGenerators.randomBeaconBlockHeader(10u).copy(
      parentRoot = currentBlock.beaconBlockHeader.hash,
      stateRoot = ByteArray(0),
      bodyRoot = HashUtil.bodyRoot(newBeaconBlockBody),
    )
  private val stateRootHash =
    HashUtil.stateRoot(
      preState.copy(
        latestBeaconBlockHeader = stateRootBlockHeader,
        latestBeaconBlockRoot = stateRootBlockHeader.bodyRoot,
      ),
    )
  private val newBlockHeader = stateRootBlockHeader.copy(stateRoot = stateRootHash)
  private val newBlock = BeaconBlock(newBlockHeader, newBeaconBlockBody)

  private val validators = setOf(newBlockHeader.proposer)

  private val validatorProvider =
    object : ValidatorProvider {
      override fun getValidatorsForBlock(header: BeaconBlockHeader): SafeFuture<Set<Validator>> =
        SafeFuture.completedFuture(validators)

      override fun getValidatorsForBlock(blockNumber: ULong): SafeFuture<Set<Validator>> =
        SafeFuture.completedFuture(validators)
    }

  private val proposerSelector =
    object : ProposerSelector {
      override fun getProposerForBlock(header: BeaconBlockHeader): SafeFuture<Validator> =
        SafeFuture.completedFuture(newBlockHeader.proposer)
    }

  private val postState =
    BeaconState(
      latestBeaconBlockHeader = newBlockHeader,
      latestBeaconBlockRoot = newBlockHeader.bodyRoot,
      validators = validators,
    )

  @Test
  fun `processBlock should return failure if block validation fails`() {
    val blockValidator =
      BlockValidator { _, _, _ ->
        SafeFuture.completedFuture(error("Block validation failed"))
      }
    val stateTransition =
      StateTransitionImpl(
        blockValidator = blockValidator,
        validatorProvider = validatorProvider,
        proposerSelector = proposerSelector,
      )

    val result = stateTransition.processBlock(preState, newBlock).get()
    val expectedResult =
      Err(StateTransition.StateTransitionError("State Transition failed. Reason: Block validation failed"))
    assertThat(result).isEqualTo(expectedResult)
  }

  @Test
  fun `processBlock should return ok if all checks pass`() {
    val blockValidator =
      BlockValidator { _, _, _ ->
        SafeFuture.completedFuture(ok())
      }
    val stateTransition =
      StateTransitionImpl(
        blockValidator = blockValidator,
        validatorProvider = validatorProvider,
        proposerSelector = proposerSelector,
      )

    val result = stateTransition.processBlock(preState, newBlock).get()
    val expectedResult = Ok(postState)
    assertThat(result).isEqualTo(expectedResult)
  }
}
