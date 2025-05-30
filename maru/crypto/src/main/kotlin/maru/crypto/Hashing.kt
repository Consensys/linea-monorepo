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
package maru.crypto

import java.security.MessageDigest
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.datatypes.Hash

object Hashing {
  fun shortShaHash(inputData: ByteArray): ByteArray = sha256(inputData).slice(0 until 20).toByteArray()

  private fun sha256(input: ByteArray): ByteArray {
    val digest: MessageDigest = MessageDigest.getInstance("SHA-256")
    return digest.digest(input)
  }

  fun keccak(serializedBytes: ByteArray): ByteArray = Hash.hash(Bytes.wrap(serializedBytes)).toArray()
}
