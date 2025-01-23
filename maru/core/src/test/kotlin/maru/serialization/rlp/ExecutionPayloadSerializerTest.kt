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
package maru.serialization.rlp

import java.math.BigInteger
import kotlin.random.Random
import kotlin.random.nextULong
import maru.core.ExecutionPayload
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class ExecutionPayloadSerializerTest {
  private val serializer = ExecutionPayloadSerializer()

  @Test
  fun `can serialize and deserialize same value`() {
    val testValue =
      ExecutionPayload(
        parentHash = Random.nextBytes(32),
        stateRoot = Random.nextBytes(32),
        receiptsRoot = Random.nextBytes(32),
        logsBloom = Random.nextBytes(32),
        prevRandao = Random.nextBytes(32),
        blockNumber = Random.nextULong(),
        gasLimit = Random.nextULong(),
        gasUsed = Random.nextULong(),
        timestamp = Random.nextULong(),
        extraData = Random.nextBytes(32),
        baseFeePerGas = BigInteger.valueOf(Random.nextLong()),
        blockHash = Random.nextBytes(32),
        transactions = buildList(3) { Random.nextBytes(100) },
      )
    val serializedData = serializer.serialize(testValue)
    val deserializedValue = serializer.deserialize(serializedData)

    assertThat(deserializedValue).isEqualTo(testValue)
  }

  @Test
  fun `can serialize and deserialize execution payload with zero transactions`() {
    val testValue =
      ExecutionPayload(
        parentHash = Random.nextBytes(32),
        stateRoot = Random.nextBytes(32),
        receiptsRoot = Random.nextBytes(32),
        logsBloom = Random.nextBytes(32),
        prevRandao = Random.nextBytes(32),
        blockNumber = Random.nextULong(),
        gasLimit = Random.nextULong(),
        gasUsed = Random.nextULong(),
        timestamp = Random.nextULong(),
        extraData = Random.nextBytes(32),
        baseFeePerGas = BigInteger.valueOf(Random.nextLong()),
        blockHash = Random.nextBytes(32),
        transactions = emptyList(),
      )
    val serializedData = serializer.serialize(testValue)
    val deserializedValue = serializer.deserialize(serializedData)

    assertThat(deserializedValue).isEqualTo(testValue)
  }
}
