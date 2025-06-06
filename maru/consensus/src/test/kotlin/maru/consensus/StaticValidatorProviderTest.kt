/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
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
