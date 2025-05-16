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

import io.libp2p.core.pubsub.ValidationResult
import java.util.Optional
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.networking.p2p.gossip.PreparedGossipMessage
import tech.pegasys.teku.networking.p2p.gossip.TopicHandler

class TestTopicHandler : TopicHandler {
  companion object {
    private const val ORIGINAL_MESSAGE = "0xdeadbeef"
  }

  val dataFuture = SafeFuture<Bytes>()

  override fun prepareMessage(
    // This should probably add a sequence number and a from to the message, or something like that. We can check what Teku is doing there :-)
    // e.g. we could compress the message and add a sequence number to it
    payload: Bytes,
    arrivalTimestamp: Optional<UInt64>,
  ): PreparedGossipMessage = MaruPreparedGossipMessage(payload, Optional.empty())

  override fun handleMessage(message: PreparedGossipMessage): SafeFuture<ValidationResult> {
    var data: Bytes
    message.let {
      data = message.originalMessage
    }
    // at this point we have to validate the message (message will only be further distributed if valid!)
    // at this point we should also (asynchronously) do what needs to be done with the data we received
    dataFuture.complete(data)
    return if (data == Bytes.fromHexString(ORIGINAL_MESSAGE)) {
      SafeFuture.completedFuture(ValidationResult.Valid)
    } else {
      SafeFuture.completedFuture(ValidationResult.Invalid)
    }
  }

  override fun getMaxMessageSize(): Int = 43434343 // TODO: what is a good max size here? 10MB?
}
