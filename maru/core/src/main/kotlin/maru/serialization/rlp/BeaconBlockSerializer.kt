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

import maru.core.BeaconBlock
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

class BeaconBlockSerializer(
  private val beaconBlockHeaderSerializer: BeaconBlockHeaderSerializer,
  private val beaconBlockBodySerializer: BeaconBlockBodySerializer,
) : RLPSerializer<BeaconBlock> {
  override fun writeTo(
    value: BeaconBlock,
    rlpOutput: RLPOutput,
  ) {
    rlpOutput.startList()

    beaconBlockHeaderSerializer.writeTo(value.beaconBlockHeader, rlpOutput)
    beaconBlockBodySerializer.writeTo(value.beaconBlockBody, rlpOutput)

    rlpOutput.endList()
  }

  override fun readFrom(rlpInput: RLPInput): BeaconBlock {
    rlpInput.enterList()

    val beaconBLockHeader = beaconBlockHeaderSerializer.readFrom(rlpInput)
    val beaconBlockBody = beaconBlockBodySerializer.readFrom(rlpInput)

    rlpInput.leaveList()

    return BeaconBlock(beaconBLockHeader, beaconBlockBody)
  }
}
