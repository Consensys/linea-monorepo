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
import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.HashUtil
import maru.core.Seal
import maru.core.SealedBeaconBlock
import maru.core.Validator
import maru.core.ext.DataGenerators.randomExecutionPayload
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class SealedBeaconBlockSerializerTest {
  private val blockHeaderSerializer =
    BeaconBlockHeaderSerializer(
      validatorSerializer = ValidatorSerializer(),
      hasher = KeccakHasher,
      headerHashFunction = HashUtil::headerHash,
    )
  private val sealSerializer = SealSerializer()
  private val blockSerializer =
    BeaconBlockSerializer(
      beaconBlockHeaderSerializer =
      blockHeaderSerializer,
      beaconBlockBodySerializer =
        BeaconBlockBodySerializer(
          sealSerializer = sealSerializer,
          executionPayloadSerializer = ExecutionPayloadSerializer(),
        ),
    )
  private val sealedBlockSerializer =
    SealedBeaconBlockSerializer(
      beaconBlockSerializer = blockSerializer,
      sealSerializer = sealSerializer,
    )

  @Test
  fun `can serialize and deserialize same value`() {
    val beaconBlockHeader =
      BeaconBlockHeader(
        number = Random.nextULong(),
        round = Random.nextULong(),
        timestamp = Random.nextULong(),
        proposer = Validator(Random.nextBytes(128)),
        parentRoot = Random.nextBytes(32),
        stateRoot = Random.nextBytes(32),
        bodyRoot = Random.nextBytes(32),
        headerHashFunction = HashUtil.headerHash(blockHeaderSerializer, KeccakHasher),
      )
    val beaconBlockBody =
      BeaconBlockBody(
        prevCommitSeals = buildList(3) { Seal(Random.nextBytes(96)) },
        executionPayload = randomExecutionPayload(),
      )
    val sealedBlock =
      SealedBeaconBlock(
        beaconBlock =
          BeaconBlock(
            beaconBlockHeader = beaconBlockHeader,
            beaconBlockBody = beaconBlockBody,
          ),
        commitSeals = buildList(3) { Seal(Random.nextBytes(96)) },
      )
    val serializedData = sealedBlockSerializer.serialize(sealedBlock)
    val deserializedValue = sealedBlockSerializer.deserialize(serializedData)

    assertThat(deserializedValue).isEqualTo(sealedBlock)
  }
}
