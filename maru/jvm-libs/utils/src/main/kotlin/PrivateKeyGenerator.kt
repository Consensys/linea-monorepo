/*
 * Copyright ConsenSys AG.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import io.libp2p.core.PeerId
import io.libp2p.core.crypto.KeyType
import io.libp2p.core.crypto.generateKeyPair
import io.libp2p.core.crypto.marshalPrivateKey
import linea.kotlin.encodeHex
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.ethereum.core.Util
import tech.pegasys.teku.networking.p2p.libp2p.LibP2PNodeId

/**
 * Utility tool to generate a prefixed private key and corresponding node ID
 */
object PrivateKeyGenerator {
  data class KeyData(
    val privateKey: ByteArray,
    val prefixedPrivateKey: ByteArray,
    val address: ByteArray,
    val peerId: LibP2PNodeId,
  )

  fun logKeyData(keyData: KeyData) {
    println(
      "Generated key: " +
        "prefixedPrivateKey=${keyData.prefixedPrivateKey.encodeHex()} " +
        "privateKey=${keyData.privateKey.encodeHex()} " +
        "ethAddress=${keyData.address.encodeHex()} " +
        "libP2pNodeId=${keyData.peerId}",
    )
  }

  fun generateAndLogPrivateKey(): KeyData {
    val keyPair = generateKeyPair(KeyType.SECP256K1)
    val privateKey = keyPair.component1()
    val privateKeyWithPrefix = marshalPrivateKey(privateKey)
    // Sometimes keyPair has 1 byte more so we just take the last 32 bytes ¯\_(ツ)_/¯
    val address = privateKeyToAddress(privateKey.raw().takeLast(32).toByteArray())
    val peerId = PeerId.fromPubKey(privateKey.publicKey())
    val libP2PNodeId = LibP2PNodeId(peerId)
    return KeyData(
      privateKey = privateKey.raw(),
      prefixedPrivateKey = privateKeyWithPrefix,
      address = address.toArray(),
      peerId = libP2PNodeId,
    )
  }

  fun privateKeyToAddress(privateKey: ByteArray): Address {
    val signatureAlgorithm = SignatureAlgorithmFactory.getInstance()
    val privateKey = signatureAlgorithm.createPrivateKey(Bytes32.wrap(privateKey))
    val keyPair = signatureAlgorithm.createKeyPair(privateKey)

    return Util.publicKeyToAddress(keyPair.publicKey)
  }

  @JvmStatic
  fun main(args: Array<String>) {
    val numberOfKeys = args.firstOrNull()?.toIntOrNull() ?: 5
    repeat(numberOfKeys) {
      generateAndLogPrivateKey()
        .also { logKeyData(it) }
    }
  }
}
