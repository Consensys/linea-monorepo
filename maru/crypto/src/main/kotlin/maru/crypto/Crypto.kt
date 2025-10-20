/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.crypto

import maru.core.Validator
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory
import org.hyperledger.besu.ethereum.core.Util

object Crypto {
  fun privateKeyToValidator(rawPrivateKey: ByteArray): Validator = Validator(privateKeyToAddress(rawPrivateKey))

  fun privateKeyToAddress(rawPrivateKey: ByteArray): ByteArray {
    val signatureAlgorithm = SignatureAlgorithmFactory.getInstance()
    val privateKey = signatureAlgorithm.createPrivateKey(Bytes32.wrap(rawPrivateKey))
    val keyPair = signatureAlgorithm.createKeyPair(privateKey)

    return Util.publicKeyToAddress(keyPair.publicKey).toArray()
  }

  fun privateKeyBytesWithoutPrefix(privateKey: ByteArray) = privateKey.takeLast(32).toByteArray()
}
