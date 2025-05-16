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

import io.netty.buffer.ByteBuf
import io.netty.buffer.ByteBufUtil
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.networking.p2p.peer.NodeId
import tech.pegasys.teku.networking.p2p.rpc.RpcRequestHandler
import tech.pegasys.teku.networking.p2p.rpc.RpcStream

class MaruIncomingRpcRequestHandler : RpcRequestHandler {
  override fun active(
    nodeId: NodeId,
    rpcStream: RpcStream,
  ) {
  }

  override fun processData(
    nodeId: NodeId,
    rpcStream: RpcStream,
    byteBuffer: ByteBuf,
  ) {
    val bytes = ByteBufUtil.getBytes(byteBuffer)
    rpcStream.writeBytes(Bytes.wrap(bytes).reverse())
  }

  override fun readComplete(
    nodeId: NodeId,
    rpcStream: RpcStream,
  ) {
  }

  override fun closed(
    nodeId: NodeId,
    rpcStream: RpcStream,
  ) {
  }
}
