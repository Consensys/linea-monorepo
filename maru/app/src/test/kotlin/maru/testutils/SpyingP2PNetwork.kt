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
package maru.testutils

import java.util.concurrent.CopyOnWriteArrayList
import maru.consensus.qbft.adapters.QbftBlockCodecAdapter
import maru.core.SealedBeaconBlock
import maru.p2p.Message
import maru.p2p.MessageType
import maru.p2p.P2PNetwork
import maru.p2p.SealedBeaconBlockHandler
import org.apache.logging.log4j.LogManager
import org.hyperledger.besu.consensus.common.bft.messagewrappers.BftMessage
import org.hyperledger.besu.consensus.qbft.core.messagedata.CommitMessageData
import org.hyperledger.besu.consensus.qbft.core.messagedata.PrepareMessageData
import org.hyperledger.besu.consensus.qbft.core.messagedata.ProposalMessageData
import org.hyperledger.besu.consensus.qbft.core.messagedata.RoundChangeMessageData
import tech.pegasys.teku.infrastructure.async.SafeFuture
import org.hyperledger.besu.ethereum.p2p.rlpx.wire.MessageData as BesuMessageData

class SpyingP2PNetwork(
  val p2pNetwork: P2PNetwork,
) : P2PNetwork {
  companion object {
    private fun Message<*>.toBesuMessageData(): BesuMessageData {
      require(this.type == MessageType.QBFT) {
        "Unsupported message type: ${this.type}"
      }
      require(this.payload is BesuMessageData) {
        "Message is QBFT, but its payload is of type: ${this.payload.javaClass}"
      }
      return this.payload as BesuMessageData
    }
  }

  private val log = LogManager.getLogger(this.javaClass)
  val emittedQbftMessages = CopyOnWriteArrayList<BftMessage<*>>()
  val emittedBlockMessages = CopyOnWriteArrayList<SealedBeaconBlock>()

  private fun decodedMessage(message: BesuMessageData): BftMessage<*> =
    when (message) {
      is CommitMessageData -> message.decode()
      is PrepareMessageData -> message.decode()
      is ProposalMessageData -> message.decode(QbftBlockCodecAdapter)
      is RoundChangeMessageData -> message.decode(QbftBlockCodecAdapter)
      else -> throw IllegalArgumentException("Unknown message $message, don't know how to decode!")
    }

  override fun start(): SafeFuture<Unit> = SafeFuture.completedFuture(Unit)

  override fun stop(): SafeFuture<Unit> = SafeFuture.completedFuture(Unit)

  override fun broadcastMessage(message: Message<*>) {
    when (message.type) {
      MessageType.QBFT -> {
        val decodedMessage = decodedMessage(message.toBesuMessageData())
        log.debug("Got new message {}", decodedMessage)
        emittedQbftMessages.add(decodedMessage)
        p2pNetwork.broadcastMessage(message)
      }
      MessageType.BLOCK -> emittedBlockMessages.add(message.payload as SealedBeaconBlock)
    }
  }

  override fun subscribeToBlocks(subscriber: SealedBeaconBlockHandler): Int = p2pNetwork.subscribeToBlocks(subscriber)

  override fun unsubscribe(subscriptionId: Int) {
    p2pNetwork.unsubscribe(subscriptionId)
  }
}
