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

import maru.core.BeaconState
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

class BeaconStateSerializer(
  private val beaconBlockHeaderSerializer: BeaconBlockHeaderSerializer,
  private val validatorSerializer: ValidatorSerializer,
) : RLPSerializer<BeaconState> {
  override fun writeTo(
    value: BeaconState,
    rlpOutput: RLPOutput,
  ) {
    rlpOutput.startList()

    beaconBlockHeaderSerializer.writeTo(value.latestBeaconBlockHeader, rlpOutput)
    rlpOutput.writeList(value.validators) { validator, output ->
      validatorSerializer.writeTo(validator, output)
    }

    rlpOutput.endList()
  }

  override fun readFrom(rlpInput: RLPInput): BeaconState {
    rlpInput.enterList()

    val latestBeaconBlockHeader = beaconBlockHeaderSerializer.readFrom(rlpInput)
    val validators = rlpInput.readList { validatorSerializer.readFrom(it) }.toSet()

    rlpInput.leaveList()

    return BeaconState(
      latestBeaconBlockHeader = latestBeaconBlockHeader,
      validators = validators,
    )
  }
}
