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

import maru.core.BeaconBlockHeader
import maru.core.Validator
import tech.pegasys.teku.infrastructure.async.SafeFuture

/**
 * Provides access to the set of validators for a given block.
 */
interface ValidatorProvider {
  fun getValidatorsForBlock(header: BeaconBlockHeader): SafeFuture<Set<Validator>>

  fun getValidatorsAfterBlock(blockNumber: ULong): SafeFuture<Set<Validator>> = getValidatorsForBlock(blockNumber + 1u)

  fun getValidatorsForBlock(blockNumber: ULong): SafeFuture<Set<Validator>>
}

/**
 * A [ValidatorProvider] that always returns the same [Validator] instance. This is useful for the single validator case.
 */
class StaticValidatorProvider(
  private val validators: Set<Validator>,
) : ValidatorProvider {
  // TODO: will be removed in the future
  override fun getValidatorsForBlock(header: BeaconBlockHeader): SafeFuture<Set<Validator>> =
    SafeFuture.completedFuture<Set<Validator>>(validators)

  override fun getValidatorsForBlock(blockNumber: ULong): SafeFuture<Set<Validator>> =
    SafeFuture.completedFuture<Set<Validator>>(validators)
}
