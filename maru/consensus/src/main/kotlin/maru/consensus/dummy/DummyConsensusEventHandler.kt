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
package maru.consensus.dummy

import maru.consensus.NewBlockHandler
import maru.consensus.NextBlockTimestampProvider
import maru.executionlayer.manager.BlockMetadata
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.hyperledger.besu.consensus.common.bft.events.BftReceivedMessageEvent
import org.hyperledger.besu.consensus.common.bft.events.BlockTimerExpiry
import org.hyperledger.besu.consensus.common.bft.events.NewChainHead
import org.hyperledger.besu.consensus.common.bft.events.RoundExpiry
import org.hyperledger.besu.consensus.common.bft.statemachine.BftEventHandler

class DummyConsensusEventHandler(
  private val blockCreator: DummyEngineApiBlockCreator,
  private val nextBlockTimestampProvider: NextBlockTimestampProvider,
  private val onNewBlock: NewBlockHandler,
  private val blockMetadataCache: LatestBlockMetadataCache,
) : BftEventHandler {
  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun start() {
  }

  override fun stop() {
  }

  override fun handleMessageEvent(p0: BftReceivedMessageEvent) {
    TODO("Unexpected because there should be no peers yet")
  }

  override fun handleNewBlockEvent(p0: NewChainHead) {
    TODO("Unexpected because there should be no peers yet")
  }

  override fun handleBlockTimerExpiry(blockTimerExpiry: BlockTimerExpiry) {
    val roundIdentifier: ConsensusRoundIdentifier = blockTimerExpiry.roundIdentifier

    if (isMsgForCurrentHeight(roundIdentifier)) {
      log.debug("Creating new block by timer {}", blockTimerExpiry.roundIdentifier)

      val latestBlockMetadata = blockMetadataCache.getLatestBlockMetadata()
      val nextBlockTimestamp =
        nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(latestBlockMetadata.unixTimestampSeconds)
      val blockCreationResult =
        blockCreator.createBlock(
          nextBlockTimestamp,
        )
      blockCreationResult?.also(onNewBlock::handleNewBlock)?.also {
        blockMetadataCache.updateLatestBlockMetadata(
          BlockMetadata(
            blockCreationResult.beaconBlockBody.executionPayload.blockNumber,
            blockCreationResult.beaconBlockBody.executionPayload.blockHash,
            blockCreationResult.beaconBlockBody.executionPayload.timestamp
              .toLong(),
          ),
        )
      }
    } else {
      log.trace(
        "Block timer event discarded as it is not for current block height " +
          "latestBlockMetadata={} eventHeight={}",
        blockMetadataCache.getLatestBlockMetadata(),
        roundIdentifier.sequenceNumber,
      )
    }
  }

  override fun handleRoundExpiry(p0: RoundExpiry) {
    TODO("No other validators are supported so nothing to do")
  }

  private fun isMsgForCurrentHeight(roundIdentifier: ConsensusRoundIdentifier): Boolean =
    roundIdentifier.sequenceNumber.toULong() == blockMetadataCache.getLatestBlockMetadata().blockNumber + 1.toULong()
}
