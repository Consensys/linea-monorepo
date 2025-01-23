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

import maru.core.Seal
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

class SealSerializer : RLPSerializer<Seal> {
  override fun writeTo(
    value: Seal,
    rlpOutput: RLPOutput,
  ) {
    rlpOutput.startList()
    rlpOutput.writeBytes(Bytes.wrap(value.signature))
    rlpOutput.endList()
  }

  override fun readFrom(rlpInput: RLPInput): Seal {
    rlpInput.enterList()
    val signature = rlpInput.readBytes().toArray()
    rlpInput.leaveList()
    return Seal(signature)
  }
}
