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

import kotlin.random.Random
import kotlin.random.nextULong
import maru.core.BeaconBlockHeader
import maru.core.HashUtil
import maru.core.Validator
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class BeaconBlockHeaderSerializerTest {
  private val serializer =
    BeaconBlockHeaderSerializer(
      validatorSerializer = ValidatorSerializer(),
      hasher = KeccakHasher,
      headerHashFunction = HashUtil::headerHash,
    )

  @Test
  fun `can serialize and deserialize same value`() {
    val testValue =
      BeaconBlockHeader(
        number = Random.nextULong(),
        round = Random.nextULong(),
        timestamp = Random.nextULong(),
        proposer = Validator(Random.nextBytes(128)),
        parentRoot = Random.nextBytes(32),
        stateRoot = Random.nextBytes(32),
        bodyRoot = Random.nextBytes(32),
        headerHashFunction = HashUtil.headerHash(serializer, KeccakHasher),
      )
    val serializedData = serializer.serialize(testValue)
    val deserializedValue = serializer.deserialize(serializedData)

    assertThat(deserializedValue).isEqualTo(testValue)
  }
}
