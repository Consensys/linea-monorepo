/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.messages

import maru.core.ext.DataGenerators.randomStatus
import maru.p2p.Message
import maru.p2p.RpcMessageType
import maru.p2p.Version
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class StatusMessageSerDeTest {
  private val serDe = StatusMessageSerDe(StatusSerDe())

  @Test
  fun `can serialize and deserialize same value`() {
    val testValue =
      Message(
        RpcMessageType.STATUS,
        Version.V1,
        randomStatus(100U),
      )

    val serializedData = serDe.serialize(testValue)
    val deserializedValue = serDe.deserialize(serializedData)

    assertThat(deserializedValue).isEqualTo(testValue)
  }
}
