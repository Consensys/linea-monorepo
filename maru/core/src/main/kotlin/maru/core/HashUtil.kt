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
package maru.core

import maru.serialization.Serializer

typealias HeaderHashFunction = (BeaconBlockHeader) -> ByteArray

interface Hasher {
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
