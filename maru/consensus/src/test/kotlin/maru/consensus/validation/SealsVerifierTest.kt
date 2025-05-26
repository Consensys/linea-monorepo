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

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import maru.consensus.ValidatorProvider
import maru.core.BeaconBlockHeader
import maru.core.Seal
import maru.core.Validator
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture

class SealsVerifierTest {
  private val validators =
    listOf(
      Validator(ByteArray(20) { 1 }),
      Validator(ByteArray(20) { 2 }),
      Validator(ByteArray(20) { 3 }),
    )
  private val validatorProvider =
    object : ValidatorProvider {
      override fun getValidatorsForBlock(blockNumber: ULong) = SafeFuture.completedFuture(validators.toSet())
    }
  private val validSeal1 = Seal(ByteArray(32) { 10 })
  private val validSeal2 = Seal(ByteArray(32) { 11 })

  private val beaconBlockHeader = DataGenerators.randomBeaconBlockHeader(1u)

  @Test
  fun `test quorum threshold met`() {
    val sealVerifier =
      object : SealVerifier {
        override fun extractValidator(
          seal: Seal,
          beaconBlockHeader: BeaconBlockHeader,
        ) = when (seal) {
          validSeal1 -> Ok(validators[0])
          validSeal2 -> Ok(validators[1])
          else -> Err(SealVerifier.SealValidationError("Invalid seal"))
        }
      }
    val sealsVerifier = QuorumOfSealsVerifier(validatorProvider, sealVerifier)
    val result =
      sealsVerifier
        .verifySeals(
          setOf(validSeal1, validSeal2),
          beaconBlockHeader,
        ).get()
    assertThat(result).isEqualTo(Ok(Unit))
  }

  @Test
  fun `test quorum threshold not met`() {
    val sealVerifier =
      object : SealVerifier {
        override fun extractValidator(
          seal: Seal,
          beaconBlockHeader: BeaconBlockHeader,
        ) = Ok(validators[0])
      }
    val sealsVerifier = QuorumOfSealsVerifier(validatorProvider, sealVerifier)
    val result = sealsVerifier.verifySeals(setOf(validSeal1), beaconBlockHeader).get()
    assertThat(result).isInstanceOf(Err::class.java)
    val error = (result as Err).error
    assertThat(error).isEqualTo("Quorum threshold not met. sealers=1 validators=3 quorumCount=2")
  }

  @Test
  fun `test seal not from validator set`() {
    val nonValidator = Validator(ByteArray(20) { 9 })
    val sealVerifier =
      object : SealVerifier {
        override fun extractValidator(
          seal: Seal,
          beaconBlockHeader: BeaconBlockHeader,
        ) = Ok(nonValidator)
      }
    val sealsVerifier = QuorumOfSealsVerifier(validatorProvider, sealVerifier)
    val result = sealsVerifier.verifySeals(setOf(validSeal1), beaconBlockHeader).get()
    assertThat(result).isInstanceOf(Err::class.java)
    val error = (result as Err).error
    assertThat(error).isEqualTo("validator=$nonValidator isn't in the expectedValidatorSet=$validators")
  }

  @Test
  fun `test invalid seal extraction`() {
    val sealVerifier =
      object : SealVerifier {
        override fun extractValidator(
          seal: Seal,
          beaconBlockHeader: BeaconBlockHeader,
        ) = Err(SealVerifier.SealValidationError("Invalid seal"))
      }
    val invalidSeal = Seal(ByteArray(32) { 12 })
    val sealsVerifier = QuorumOfSealsVerifier(validatorProvider, sealVerifier)
    val result = sealsVerifier.verifySeals(setOf(invalidSeal), beaconBlockHeader).get()
    assertThat(result).isInstanceOf(Err::class.java)
    val error = (result as Err).error
    assertThat(error).isEqualTo("Invalid seal")
  }
}
