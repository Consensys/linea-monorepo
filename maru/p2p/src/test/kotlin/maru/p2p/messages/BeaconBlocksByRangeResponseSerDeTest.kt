/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.messages

import maru.core.HashUtil
import maru.core.ext.DataGenerators
import maru.serialization.rlp.RLPSerializers
import maru.serialization.rlp.bodyRoot
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class BeaconBlocksByRangeResponseSerDeTest {
  @Test
  fun `response serDe serializes and deserializes correctly`() {
    val serDe =
      BeaconBlocksByRangeResponseSerDe(
        RLPSerializers.SealedBeaconBlockSerializer,
      )

    val response =
      BeaconBlocksByRangeResponse(
        blocks = listOf(
          DataGenerators.randomSealedBeaconBlock(
            number = 5UL,
            headerHashFunction = RLPSerializers.DefaultHeaderHashFunction,
            bodyRootFunction = { body -> HashUtil.bodyRoot(body) },
          ),
        ),
      )

    val serialized = serDe.serialize(response)
    val deserialized = serDe.deserialize(serialized)

    assertThat(deserialized).isEqualTo(response)
  }
}
