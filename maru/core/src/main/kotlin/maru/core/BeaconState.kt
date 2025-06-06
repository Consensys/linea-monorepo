/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.core

/**
 * After every BeaconBlock there is a transition in the BeaconState by applying the operations from
 * the BeaconBlock These operations could be a new execution payload, adding/removing validators
 * etc.
 */
data class BeaconState(
  val latestBeaconBlockHeader: BeaconBlockHeader,
  val validators: Set<Validator>,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BeaconState

    if (latestBeaconBlockHeader != other.latestBeaconBlockHeader) return false
    if (validators != other.validators) return false

    return true
  }

  override fun hashCode(): Int {
    var result = latestBeaconBlockHeader.hashCode()
    result = 31 * result + validators.hashCode()
    return result
  }
}
