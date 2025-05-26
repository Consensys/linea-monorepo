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
package maru.consensus.validation

import com.github.michaelbull.result.Ok
import maru.core.Seal
import maru.core.Validator
import maru.core.ext.DataGenerators
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory
import org.hyperledger.besu.ethereum.core.Util
import org.junit.jupiter.api.Test

class SCEP256SealVerifierTest {
  private val signatureAlgorithm = SignatureAlgorithmFactory.getInstance()
  private val verifier = SCEP256SealVerifier()
  private val keypair = signatureAlgorithm.generateKeyPair()
  private val validator = Validator(Util.publicKeyToAddress(keypair.publicKey).toArray())

  @Test
  fun `test extract validator`() {
    val beaconBlockHeader = DataGenerators.randomBeaconBlockHeader(101u)
    val signature = signatureAlgorithm.sign(Bytes32.wrap(beaconBlockHeader.hash), keypair)
    val seal = Seal(signature.encodedBytes().toArray())
    val result = verifier.extractValidator(seal, beaconBlockHeader)
    assertThat(result).isInstanceOf(Ok::class.java)
    assertThat((result as Ok).value).isEqualTo(validator)
  }
}
