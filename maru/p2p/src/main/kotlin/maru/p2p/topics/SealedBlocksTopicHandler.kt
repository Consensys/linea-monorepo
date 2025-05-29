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
package maru.p2p.topics

import java.util.Optional
import maru.core.SealedBeaconBlock
import maru.p2p.MaruPreparedGossipMessage
import maru.p2p.SubscriptionManager
import maru.p2p.ValidationResultCode
import maru.serialization.Serializer
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.networking.p2p.gossip.PreparedGossipMessage
import tech.pegasys.teku.networking.p2p.gossip.TopicHandler
import io.libp2p.core.pubsub.ValidationResult as Libp2pValidationResult

class SealedBlocksTopicHandler(
  val subscriptionManager: SubscriptionManager<SealedBeaconBlock>,
  val sealedBeaconBlockSerializer: Serializer<SealedBeaconBlock>,
) : TopicHandler {
  companion object {
    fun ValidationResultCode.toLibP2P(): Libp2pValidationResult =
      when (this) {
        ValidationResultCode.ACCEPT -> Libp2pValidationResult.Valid
        ValidationResultCode.REJECT -> Libp2pValidationResult.Invalid
        ValidationResultCode.IGNORE -> Libp2pValidationResult.Ignore
        // TODO: We don't have a case for this yet, so maybe it isn't right
        ValidationResultCode.KEEP_FOR_THE_FUTURE -> Libp2pValidationResult.Ignore
      }
  }

  override fun prepareMessage(
    payload: Bytes,
    arrivalTimestamp: Optional<UInt64>,
  ): PreparedGossipMessage = MaruPreparedGossipMessage(payload, arrivalTimestamp)

  override fun handleMessage(message: PreparedGossipMessage): SafeFuture<Libp2pValidationResult> {
    val deserializaedMessage = sealedBeaconBlockSerializer.deserialize(message.originalMessage.toArray())
    return subscriptionManager.handleEvent(deserializaedMessage).thenApply { it.code.toLibP2P() }
  }

  override fun getMaxMessageSize(): Int = 10485760
}
