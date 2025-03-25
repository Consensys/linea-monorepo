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
package maru.consensus

import kotlin.test.Test
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat

class StaticValidatorProviderTest {
  private val validators = DataGenerators.randomValidators()
  private val staticValidatorProvider = StaticValidatorProvider(validators)

  @Test
  fun `can get validators at after block`() {
    assertThat(staticValidatorProvider.getValidatorsAfterBlock(0U).get()).isEqualTo(validators)
    assertThat(staticValidatorProvider.getValidatorsAfterBlock(1U).get()).isEqualTo(validators)
  }

  @Test
  fun `can get validators for block`() {
    assertThat(staticValidatorProvider.getValidatorsForBlock(0U).get()).isEqualTo(validators)
    assertThat(staticValidatorProvider.getValidatorsForBlock(1U).get()).isEqualTo(validators)
  }
}
