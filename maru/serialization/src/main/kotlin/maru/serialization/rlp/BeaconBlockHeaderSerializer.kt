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

import maru.core.BeaconBlockHeader
import maru.core.Hasher
import maru.core.HeaderHashFunction
import maru.serialization.Serializer
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

class BeaconBlockHeaderSerializer(
  private val validatorSerializer: ValidatorSerializer,
  private val hasher: Hasher,
  private val headerHashFunction: (Serializer<BeaconBlockHeader>, Hasher) -> HeaderHashFunction,
) : RLPSerializer<BeaconBlockHeader> {
  override fun writeTo(
    value: BeaconBlockHeader,
    rlpOutput: RLPOutput,
  ) {
    rlpOutput.startList()

    rlpOutput.writeLong(value.number.toLong())
    rlpOutput.writeInt(value.round.toInt())
    rlpOutput.writeLong(value.timestamp.toLong())
    validatorSerializer.writeTo(value.proposer, rlpOutput)
    rlpOutput.writeBytes(Bytes.wrap(value.parentRoot))
    rlpOutput.writeBytes(Bytes.wrap(value.stateRoot))
    rlpOutput.writeBytes(Bytes.wrap(value.bodyRoot))

    rlpOutput.endList()
  }

  override fun readFrom(rlpInput: RLPInput): BeaconBlockHeader {
    rlpInput.enterList()

    val number = rlpInput.readLong().toULong()
    val round = rlpInput.readInt().toUInt()
    val timestamp = rlpInput.readLong().toULong()
    val proposer = validatorSerializer.readFrom(rlpInput)
    val parentRoot = rlpInput.readBytes().toArray()
    val stateRoot = rlpInput.readBytes().toArray()
    val bodyRoot = rlpInput.readBytes().toArray()

    rlpInput.leaveList()

    return BeaconBlockHeader(
      number = number,
      round = round,
      timestamp = timestamp,
      proposer = proposer,
      parentRoot = parentRoot,
      stateRoot = stateRoot,
      bodyRoot = bodyRoot,
      headerHashFunction(this, hasher),
    )
  }
}
