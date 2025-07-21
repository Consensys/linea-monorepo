/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import java.util.Base64
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.crypto.SECP256K1
import org.ethereum.beacon.discovery.schema.IdentitySchemaInterpreter
import org.ethereum.beacon.discovery.schema.NodeRecordBuilder
import org.ethereum.beacon.discovery.schema.NodeRecordFactory

fun getBootnodeEnrString(
  privateKeyBytes: ByteArray,
  ipv4: String,
  discPort: Int,
  tcpPort: Int,
): String {
  val secretKey = SECP256K1.SecretKey.fromBytes(Bytes32.wrap(privateKeyBytes))
  val bootnodeNR =
    NodeRecordBuilder()
      .nodeRecordFactory(NodeRecordFactory(IdentitySchemaInterpreter.V4))
      .seq(1)
      .secretKey(secretKey)
      .address(ipv4, discPort, tcpPort)
      .build()
  val enr = bootnodeNR.serialize()
  val encoded = Base64.getUrlEncoder().encode(enr.toArray())
  val enrString = "enr:${String(encoded)}"
  return enrString
}
