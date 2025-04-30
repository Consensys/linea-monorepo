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
package maru.consensus.qbft

import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.consensus.common.bft.events.BftEvent
import org.hyperledger.besu.consensus.common.bft.events.BftEvents
import org.hyperledger.besu.consensus.common.bft.events.BftReceivedMessageEvent
import org.hyperledger.besu.consensus.common.bft.events.BlockTimerExpiry
import org.hyperledger.besu.consensus.common.bft.events.RoundExpiry
import org.hyperledger.besu.consensus.qbft.core.types.QbftEventHandler
import org.hyperledger.besu.consensus.qbft.core.types.QbftNewChainHead

class QbftEventMultiplexer(
  private val eventHandler: QbftEventHandler,
) {
  private val log: Logger = LogManager.getLogger(this::class.java)

  fun handleEvent(event: BftEvent) {
    try {
      when (event.type) {
        BftEvents.Type.ROUND_EXPIRY -> eventHandler.handleRoundExpiry(event as RoundExpiry)
        BftEvents.Type.NEW_CHAIN_HEAD -> eventHandler.handleNewBlockEvent(event as QbftNewChainHead)
        BftEvents.Type.BLOCK_TIMER_EXPIRY -> eventHandler.handleBlockTimerExpiry(event as BlockTimerExpiry)
        BftEvents.Type.MESSAGE -> eventHandler.handleMessageEvent(event as BftReceivedMessageEvent)
        else -> {
          throw IllegalStateException("Unhandled event type: ${event.type}")
        }
      }
    } catch (e: Exception) {
      log.error("State machine threw exception while processing event \\{$event\\}", e)
    }
  }
}
