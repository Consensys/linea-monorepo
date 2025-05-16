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

import java.util.Optional
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.networking.p2p.rpc.RpcResponseHandler

class MaruRpcResponseHandler : RpcResponseHandler<Bytes> {
  val future = SafeFuture<Bytes>()

  override fun onResponse(response: Bytes): SafeFuture<*> {
    future.complete(response)
    return SafeFuture.completedFuture(response)
  }

  override fun onCompleted(error: Optional<out Throwable>) {
    if (error.isEmpty) {
      return // if needed do something when the response is completed successfully
    } else {
      throw error.get()
    }
  }

  fun response(): SafeFuture<Bytes> = future
}
