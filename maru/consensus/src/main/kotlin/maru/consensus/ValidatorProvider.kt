/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import maru.core.Validator
import tech.pegasys.teku.infrastructure.async.SafeFuture

/**
 * Provides access to the set of validators for a given block.
 */
interface ValidatorProvider {
  fun getValidatorsAfterBlock(blockNumber: ULong): SafeFuture<Set<Validator>> = getValidatorsForBlock(blockNumber + 1u)

  fun getValidatorsForBlock(blockNumber: ULong): SafeFuture<Set<Validator>>
}

/**
 * A [ValidatorProvider] that always returns the same [Validator] instance. This is useful for the single validator case.
 */
class StaticValidatorProvider(
  private val validators: Set<Validator>,
) : ValidatorProvider {
  override fun getValidatorsForBlock(blockNumber: ULong): SafeFuture<Set<Validator>> =
    SafeFuture.completedFuture(validators)
}
