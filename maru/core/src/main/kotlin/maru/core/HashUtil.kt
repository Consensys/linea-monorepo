/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.core

import maru.serialization.Serializer

typealias HeaderHashFunction = (BeaconBlockHeader) -> ByteArray

fun interface Hasher {
  fun hash(serializedBytes: ByteArray): ByteArray
}

/**
 * Utility class for hashing various parts of the beacon chain
 */
object HashUtil {
  fun headerHash(
    serializer: Serializer<BeaconBlockHeader>,
    hasher: Hasher,
  ): HeaderHashFunction = { header -> rootHash(header, serializer, hasher) }

  fun <T> rootHash(
    t: T,
    serializer: Serializer<T>,
    hasher: Hasher,
  ): ByteArray {
    val serialized = serializer.serialize(t)
    return hasher.hash(serialized)
  }
}
