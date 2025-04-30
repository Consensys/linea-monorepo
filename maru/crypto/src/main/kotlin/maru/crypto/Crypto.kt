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

import maru.core.Validator
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory
import org.hyperledger.besu.ethereum.core.Util

object Crypto {
  fun privateKeyToValidator(rawPrivateKey: ByteArray): Validator {
    val signatureAlgorithm = SignatureAlgorithmFactory.getInstance()
    val privateKey = signatureAlgorithm.createPrivateKey(Bytes32.wrap(rawPrivateKey))
    val keyPair = signatureAlgorithm.createKeyPair(privateKey)

    return Validator(Util.publicKeyToAddress(keyPair.publicKey).toArray())
  }
}
