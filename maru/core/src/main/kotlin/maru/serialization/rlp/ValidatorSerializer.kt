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

import maru.core.Validator
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

class ValidatorSerializer : RLPSerializer<Validator> {
  override fun writeTo(
    value: Validator,
    rlpOutput: RLPOutput,
  ) {
    rlpOutput.startList()
    rlpOutput.writeBytes(Bytes.wrap(value.address))
    rlpOutput.endList()
  }

  override fun readFrom(rlpInput: RLPInput): Validator {
    rlpInput.enterList()
    val address: ByteArray = rlpInput.readBytes().toArray()
    rlpInput.leaveList()
    return Validator(address)
  }
}
