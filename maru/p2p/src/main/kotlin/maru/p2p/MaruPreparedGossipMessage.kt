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
import maru.crypto.Hashing
import org.apache.tuweni.bytes.Bytes
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.networking.p2p.gossip.PreparedGossipMessage
import tech.pegasys.teku.spec.logic.common.helpers.MathHelpers

class MaruPreparedGossipMessage(
  private val origMessage: Bytes,
  private val arrTimestamp: Optional<UInt64>,
  private val domain: String,
  private val topicId: String,
) : PreparedGossipMessage {
  companion object {
  }

  // It's for deduplication of messages
  // topic length comes from Altair consensus spec
  // https://github.com/ethereum/consensus-specs/blob/dev/specs/altair/p2p-interface.md#the-gossip-domain-gossipsub
  override fun getMessageId(): Bytes =
    Bytes.wrap(
      Hashing.shortShaHash(
        domain.toByteArray() + encodeTopicLength(topicId.toByteArray()).toArray() +
          topicId.toByteArray() + origMessage.toArray(),
      ),
    )

  override fun getDecodedMessage(): PreparedGossipMessage.DecodedMessageResult =
    PreparedGossipMessage.DecodedMessageResult.successful(origMessage)

  override fun getOriginalMessage(): Bytes = origMessage

  override fun getArrivalTimestamp(): Optional<UInt64> = arrTimestamp

  private fun encodeTopicLength(topicBytes: ByteArray): Bytes = MathHelpers.uint64ToBytes(topicBytes.size.toLong())
}
