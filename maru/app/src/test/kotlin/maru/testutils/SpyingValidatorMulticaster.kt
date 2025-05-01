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
import org.apache.logging.log4j.LogManager
import org.hyperledger.besu.consensus.common.bft.messagewrappers.BftMessage
import org.hyperledger.besu.consensus.common.bft.network.ValidatorMulticaster
import org.hyperledger.besu.consensus.qbft.core.messagedata.CommitMessageData
import org.hyperledger.besu.consensus.qbft.core.messagedata.PrepareMessageData
import org.hyperledger.besu.consensus.qbft.core.messagedata.ProposalMessageData
import org.hyperledger.besu.consensus.qbft.core.messagedata.RoundChangeMessageData
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.ethereum.p2p.rlpx.wire.MessageData

class SpyingValidatorMulticaster(
  val validatorMulticaster: ValidatorMulticaster,
) : ValidatorMulticaster {
  private val log = LogManager.getLogger(this.javaClass)
  val emittedMessages = CopyOnWriteArrayList<BftMessage<*>>()

  override fun send(message: MessageData) {
    val decodedMessage = decodedMessage(message)
    log.debug("Got new message {}", decodedMessage)
    emittedMessages.add(decodedMessage)
    validatorMulticaster.send(message)
  }

  override fun send(
    message: MessageData,
    denylist: Collection<Address>,
  ) {
    val decodedMessage = decodedMessage(message)
    log.debug("Got new message {}", decodedMessage)
    emittedMessages.add(decodedMessage)
    validatorMulticaster.send(message, denylist)
  }

  private fun decodedMessage(message: MessageData): BftMessage<*> =
    when (message) {
      is CommitMessageData -> message.decode()
      is PrepareMessageData -> message.decode()
      is ProposalMessageData -> message.decode(QbftBlockCodecAdapter)
      is RoundChangeMessageData -> message.decode(QbftBlockCodecAdapter)
      else -> throw IllegalArgumentException("Unknown message $message, don't know how to decode!")
    }
}
