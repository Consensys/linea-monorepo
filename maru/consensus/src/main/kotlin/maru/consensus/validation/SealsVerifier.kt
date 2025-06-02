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
import com.github.michaelbull.result.Result
import maru.consensus.ValidatorProvider
import maru.core.BeaconBlockHeader
import maru.core.Seal
import org.hyperledger.besu.consensus.common.bft.BftHelpers
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface SealsVerifier {
  fun verifySeals(
    seals: Set<Seal>,
    beaconBlockHeader: BeaconBlockHeader,
  ): SafeFuture<Result<Unit, String>>
}

class QuorumOfSealsVerifier(
  val validatorProvider: ValidatorProvider,
  val sealVerifier: SealVerifier,
) : SealsVerifier {
  override fun verifySeals(
    seals: Set<Seal>,
    beaconBlockHeader: BeaconBlockHeader,
  ): SafeFuture<Result<Unit, String>> =
    validatorProvider
      .getValidatorsForBlock(
        beaconBlockHeader.number,
      ).thenApply { expectedValidatorSet ->
        val validSealIssuers =
          seals
            .map {
              try {
                val sealedValidator = sealVerifier.extractValidator(it, beaconBlockHeader)
                when (sealedValidator) {
                  is Err -> {
                    return@thenApply Err(sealedValidator.error.message)
                  }

                  is Ok ->
                    if (sealedValidator.value in expectedValidatorSet) {
                      sealedValidator.value
                    } else {
                      return@thenApply Err(
                        "validator=${sealedValidator.value} isn't in the expectedValidatorSet=$expectedValidatorSet",
                      )
                    }
                }
              } catch (ex: Throwable) {
                return@thenApply Err(ex.message!!)
              }
            }.toSet()

        val quorumCount = BftHelpers.calculateRequiredValidatorQuorum(expectedValidatorSet.size)
        if (quorumCount > validSealIssuers.size) {
          Err(
            "Quorum threshold not met. sealers=${seals.size} " +
              "validators=${expectedValidatorSet.size} " +
              "quorumCount=$quorumCount",
          )
        } else {
          Ok(Unit)
        }
      }
}
