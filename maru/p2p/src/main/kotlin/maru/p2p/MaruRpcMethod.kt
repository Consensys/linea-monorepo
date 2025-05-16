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
package maru.p2p

import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.networking.p2p.rpc.RpcMethod
import tech.pegasys.teku.networking.p2p.rpc.RpcRequestHandler

private const val LINEA = "linea"

class MaruRpcMethod : RpcMethod<MaruOutgoingRpcRequestHandler, Bytes, MaruRpcResponseHandler> {
  override fun getIds(): MutableList<String> = mutableListOf(LINEA)

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
    return LINEA == rpcMethod.ids.first()
  }

  override fun hashCode(): Int = 42
}
