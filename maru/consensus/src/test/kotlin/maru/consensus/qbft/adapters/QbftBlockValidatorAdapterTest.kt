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

import com.github.michaelbull.result.Result
import com.github.michaelbull.result.getError
import java.util.Optional
import maru.consensus.validation.BlockValidator
import maru.core.BeaconBlock
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockValidator
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture

class QbftBlockValidatorAdapterTest {
  private val newBlock = DataGenerators.randomBeaconBlock(10u)
  private lateinit var blockValidator: BlockValidator

  @Test
  fun `validateBlock should return false when block validation error`() {
    val blockValidatorError = BlockValidator.error("Error")
    blockValidator =
      object : BlockValidator {
        override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidator.BlockValidationError>> =
          SafeFuture.completedFuture(blockValidatorError)
      }
    val qbftBlockValidatorAdapter =
      QbftBlockValidatorAdapter(
        blockValidator = blockValidator,
      )
    val expectedResult =
      QbftBlockValidator.ValidationResult(
        false,
        Optional.of(blockValidatorError.getError().toString()),
      )
    assertThat(qbftBlockValidatorAdapter.validateBlock(QbftBlockAdapter(newBlock))).isEqualTo(expectedResult)
  }

  @Test
  fun `validateBlock should return true when valid block`() {
    blockValidator =
      object : BlockValidator {
        override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidator.BlockValidationError>> =
          SafeFuture.completedFuture(BlockValidator.ok())
      }

    val qbftBlockValidatorAdapter =
      QbftBlockValidatorAdapter(
        blockValidator = blockValidator,
      )
    val expectedResult = QbftBlockValidator.ValidationResult(true, Optional.empty())
    assertThat(qbftBlockValidatorAdapter.validateBlock(QbftBlockAdapter(newBlock))).isEqualTo(expectedResult)
  }
}
