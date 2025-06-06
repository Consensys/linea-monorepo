/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p

import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.networking.p2p.rpc.RpcMethod
import tech.pegasys.teku.networking.p2p.rpc.RpcRequestHandler

class MaruRpcMethod : RpcMethod<MaruOutgoingRpcRequestHandler, Bytes, MaruRpcResponseHandler> {
  override fun getIds(): MutableList<String> = mutableListOf(LINEA_DOMAIN)

  override fun createIncomingRequestHandler(protocolId: String): RpcRequestHandler {
    val maruRpcRequestHandler = MaruIncomingRpcRequestHandler()
    return maruRpcRequestHandler
  }

  override fun createOutgoingRequestHandler(
    protocolId: String,
    request: Bytes,
    responseHandler: MaruRpcResponseHandler,
  ): MaruOutgoingRpcRequestHandler {
    val maruRpcRequestHandler = MaruOutgoingRpcRequestHandler(responseHandler)
    return maruRpcRequestHandler
  }

  override fun encodeRequest(bytes: Bytes): Bytes = bytes

  override fun equals(other: Any?): Boolean {
    if (this === other) {
      return true
    }
    if (other == null || javaClass != other.javaClass) {
      return false
    }
    val rpcMethod: MaruRpcMethod =
      other as MaruRpcMethod
    return LINEA_DOMAIN == rpcMethod.ids.first()
  }

  override fun hashCode(): Int = 42
}
