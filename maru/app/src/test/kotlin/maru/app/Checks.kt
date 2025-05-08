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
package maru.app

import java.math.BigInteger
import maru.testutils.besu.BesuFactory
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.methods.response.EthBlock

object Checks {
  fun BesuNode.getMinedBlocks(blocksMined: Int): List<EthBlock.Block> =
    (1..blocksMined)
      .map {
        this
          .nodeRequests()
          .eth()
          .ethGetBlockByNumber(
            DefaultBlockParameter.valueOf(BigInteger.valueOf(it.toLong())),
            false,
          ).sendAsync()
      }.map { it.get().block }

  // Checks that all block times in the list are exactly BesuFactory.MIN_BLOCK_TIME (1 second)
  // First block is skipped, because after startup block time is sometimes floating above 1 second
  fun List<EthBlock.Block>.verifyBlockTime() {
    val timestampsSeconds = this.subList(1, this.size - 1).map { it.timestamp.toLong() }
    timestampsSeconds.reduceIndexed { index, prevTimestamp, timestamp ->
      assertThat(prevTimestamp).isLessThan(timestamp)
      val actualBlockTime = timestamp - prevTimestamp
      assertThat(actualBlockTime)
        .withFailMessage("Timestamps: $timestampsSeconds")
        .isEqualTo(BesuFactory.MIN_BLOCK_TIME)
      timestamp
    }
  }

  // Checks that all block times in the list are exactly BesuFactory.MIN_BLOCK_TIME (1 second)
  fun List<EthBlock.Block>.verifyBlockTimeWithAGapOn(gappedBlockNumber: Int) {
    val blocksPreGap = this.subList(1, gappedBlockNumber)
    // Skipping the first block after gap as well, because first block after startup can be missed for unclear reasons
    val blocksPostGap = this.subList(gappedBlockNumber, this.size - 1)
    // Verify block time is consistent before and after the switch
    blocksPreGap.verifyBlockTime()
    blocksPostGap.verifyBlockTime()
  }
}
