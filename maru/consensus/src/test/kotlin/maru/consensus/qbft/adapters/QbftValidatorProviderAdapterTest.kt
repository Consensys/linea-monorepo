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

import kotlin.test.Test
import maru.consensus.ValidatorProvider
import maru.consensus.qbft.toAddress
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.mockito.Mockito
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture.completedFuture

class QbftValidatorProviderAdapterTest {
  @Test
  fun `can get validators after block`() {
    val validatorProvider = Mockito.mock(ValidatorProvider::class.java)
    val validators1 = DataGenerators.randomValidators()
    val validators2 = DataGenerators.randomValidators()
    val header1 = QbftBlockHeaderAdapter(DataGenerators.randomBeaconBlockHeader(10U))
    val header2 = QbftBlockHeaderAdapter(DataGenerators.randomBeaconBlockHeader(11U))
    whenever(
      validatorProvider.getValidatorsAfterBlock(header1.beaconBlockHeader.number),
    ).thenReturn(completedFuture(validators1))
    whenever(
      validatorProvider.getValidatorsAfterBlock(header2.beaconBlockHeader.number),
    ).thenReturn(completedFuture(validators2))

    val qbftValidatorProviderAdapter = QbftValidatorProviderAdapter(validatorProvider)

    assertThat(
      qbftValidatorProviderAdapter.getValidatorsAfterBlock(header1),
    ).containsAll(validators1.map { it.toAddress() })
    assertThat(
      qbftValidatorProviderAdapter.getValidatorsAfterBlock(header2),
    ).containsAll(validators2.map { it.toAddress() })
  }

  @Test
  fun `can get validators for block`() {
    val validatorProvider = Mockito.mock(ValidatorProvider::class.java)
    val validators1 = DataGenerators.randomValidators()
    val validators2 = DataGenerators.randomValidators()
    val header1 = QbftBlockHeaderAdapter(DataGenerators.randomBeaconBlockHeader(10U))
    val header2 = QbftBlockHeaderAdapter(DataGenerators.randomBeaconBlockHeader(11U))
    whenever(
      validatorProvider.getValidatorsForBlock(header1.beaconBlockHeader.number),
    ).thenReturn(completedFuture(validators1))
    whenever(
      validatorProvider.getValidatorsForBlock(header2.beaconBlockHeader.number),
    ).thenReturn(completedFuture(validators2))

    val qbftValidatorProviderAdapter = QbftValidatorProviderAdapter(validatorProvider)
    assertThat(
      qbftValidatorProviderAdapter.getValidatorsForBlock(header1),
    ).containsAll(validators1.map { it.toAddress() })
    assertThat(
      qbftValidatorProviderAdapter.getValidatorsForBlock(header2),
    ).containsAll(validators2.map { it.toAddress() })
  }
}
