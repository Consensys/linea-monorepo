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

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import maru.core.BeaconBlockHeader
import maru.core.Seal
import maru.core.Validator
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.crypto.AbstractSECP256
import org.hyperledger.besu.ethereum.core.Util

interface SealVerifier {
  data class SealValidationError(
    val message: String,
  )

  fun extractValidator(
    seal: Seal,
    beaconBlockHeader: BeaconBlockHeader,
  ): Result<Validator, SealValidationError>
}

class SCEP256SealVerifier(
  private val signatureAlgorithm: AbstractSECP256,
) : SealVerifier {
  override fun extractValidator(
    seal: Seal,
    beaconBlockHeader: BeaconBlockHeader,
  ): Result<Validator, SealVerifier.SealValidationError> {
    val signature = signatureAlgorithm.decodeSignature(Bytes.wrap(seal.signature))
    val blockHash = beaconBlockHeader.hash
    val publicKey = signatureAlgorithm.recoverPublicKeyFromSignature(Bytes32.wrap(blockHash), signature)
    return if (publicKey.isEmpty) {
      Err(
        SealVerifier.SealValidationError(
          "Could not extract validator from seal for block. seal=$seal block=$beaconBlockHeader",
        ),
      )
    } else {
      Ok(Validator(Util.publicKeyToAddress(publicKey.get()).toArray()))
    }
  }
}
