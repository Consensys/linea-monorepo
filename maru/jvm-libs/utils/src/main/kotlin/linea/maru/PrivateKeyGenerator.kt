/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.maru

import io.libp2p.core.PeerId
import io.libp2p.core.crypto.KeyType
import io.libp2p.core.crypto.generateKeyPair
import io.libp2p.core.crypto.marshalPrivateKey
import io.libp2p.core.crypto.unmarshalPrivateKey
import io.libp2p.crypto.keys.unmarshalSecp256k1PrivateKey
import linea.kotlin.encodeHex
import org.apache.tuweni.bytes.Bytes
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
  ) {
    override fun equals(other: Any?): Boolean {
      if (this === other) return true
      if (javaClass != other?.javaClass) return false

      other as KeyData

      if (!privateKey.contentEquals(other.privateKey)) return false
      if (!prefixedPrivateKey.contentEquals(other.prefixedPrivateKey)) return false
      if (!address.contentEquals(other.address)) return false
      if (peerId != other.peerId) return false

      return true
    }

    override fun hashCode(): Int {
      var result = privateKey.contentHashCode()
      result = 31 * result + prefixedPrivateKey.contentHashCode()
      result = 31 * result + address.contentHashCode()
      result = 31 * result + peerId.hashCode()
      return result
    }

    override fun toString(): String =
      "KeyData(privateKey=${privateKey.contentToString()}, prefixedPrivateKey=${prefixedPrivateKey.contentToString()}, address=${address.contentToString()}, peerId=$peerId)"
  }

  fun logKeyData(keyData: KeyData) {
    println(
      "Generated key: " +
        "prefixedPrivateKey=${keyData.prefixedPrivateKey.encodeHex()} " +
        "privateKey=${keyData.privateKey.encodeHex()} " +
        "ethAddress=${keyData.address.encodeHex()} " +
        "libP2pNodeId=${keyData.peerId}",
    )
  }

  fun getKeyData(privateKey: ByteArray): KeyData {
    val privateKeyTyped = unmarshalSecp256k1PrivateKey(privateKey)
    val publicKey = privateKeyTyped.publicKey()
    val address = Util.publicKeyToAddress(Bytes.wrap(publicKey.bytes()))
    val peerId = PeerId.Companion.fromPubKey(privateKeyTyped.publicKey())
    val libP2PNodeId = LibP2PNodeId(peerId)
    return KeyData(
      privateKey = privateKeyTyped.raw(),
      prefixedPrivateKey = marshalPrivateKey(privateKeyTyped),
      address = address.toArray(),
      peerId = libP2PNodeId,
    )
  }

  fun getKeyDataByPrefixedKey(prefixedPrivateKey: ByteArray): KeyData {
    val privateKeyTyped = unmarshalPrivateKey(prefixedPrivateKey)
    return getKeyData(privateKeyTyped.raw())
  }

  fun generatePrivateKey(): KeyData {
    val (privKey, pubKey) = generateKeyPair(KeyType.SECP256K1)
    // Sometimes keyPair has 1 byte more so we just take the last 32 bytes ¯\_(ツ)_/¯
    val privKeyRaw = privKey.raw().takeLast(32).toByteArray()
    return getKeyData(privKeyRaw)
  }
}
