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
import maru.core.BeaconState
import maru.core.HashUtil
import maru.core.ext.DataGenerators
import maru.serialization.rlp.bodyRoot
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.ArgumentMatchers.eq
import org.mockito.Mockito.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class StateTransitionImplTest {
  @Test
  fun `processBlock should return ok`() {
    val newBlock = DataGenerators.randomBeaconBlock(10uL)
    val validators = List(3) { DataGenerators.randomValidator() }.toSet()
    val expectedPostState =
      BeaconState(
        latestBeaconBlockHeader = newBlock.beaconBlockHeader,
        latestBeaconBlockRoot = HashUtil.bodyRoot(newBlock.beaconBlockBody),
        validators = validators,
      )

    val validatorProvider = mock<ValidatorProvider>()
    whenever(validatorProvider.getValidatorsForBlock(eq(newBlock.beaconBlockHeader.number.toLong()).toULong()))
      .thenReturn(SafeFuture.completedFuture(validators))

    val stateTransition = StateTransitionImpl(validatorProvider = validatorProvider)

    val result = stateTransition.processBlock(newBlock).get()
    assertThat(result).isEqualTo(expectedPostState)
  }
}
