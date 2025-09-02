/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.serialization

import java.nio.ByteBuffer
import maru.consensus.ForkId

object ForkIdSerializer : Serializer<ForkId> {
  override fun serialize(value: ForkId): ByteArray {
    val serializedForkSpec = ForkSpecSerializer.serialize(value.forkSpec)

    val buffer =
      ByteBuffer
        .allocate(4 + serializedForkSpec.size + 32)
        .putInt(value.chainId.toInt())
        .put(serializedForkSpec)
        .put(value.genesisRootHash)

    return buffer.array()
  }
}
