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
import tech.pegasys.teku.networking.p2p.libp2p.LibP2PNodeId

/**
 * Utility tool to generate a prefixed private key and corresponding node ID
 */
object PrivateKeyGenerator {
  fun generateAndLogPrivateKey() {
    val keyPair = generateKeyPair(KeyType.SECP256K1)
    val privateKey = keyPair.component1()
    val privateKeyWithPrefixString = marshalPrivateKey(privateKey).encodeHex()
    val peerId = PeerId.fromPubKey(privateKey.publicKey())
    val libP2PNodeId = LibP2PNodeId(peerId)

    println("Generated private key (prefixed): $privateKeyWithPrefixString")
    println("Corresponding node ID: $libP2PNodeId")
  }

  @JvmStatic
  fun main(args: Array<String>) {
    generateAndLogPrivateKey()
  }
}
