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
package maru.consensus.qbft.adapters

import kotlin.test.Test
import maru.consensus.qbft.ProposerSelectorImpl
import maru.consensus.qbft.toAddress
import maru.core.ext.DataGenerators
import maru.database.BeaconChain
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.junit.jupiter.api.assertThrows
import org.mockito.Mockito.mock
import org.mockito.kotlin.whenever

class ProposerSelectorAdapterTest {
  @Test
  fun `test previous block not found`() {
    val beaconChain = mock<BeaconChain>()
    val missingBlockNumber = 3

    val proposerSelectorAdapter = ProposerSelectorAdapter(beaconChain, ProposerSelectorImpl)
    val consensusRoundIdentifier = ConsensusRoundIdentifier(missingBlockNumber.toLong(), 0)

    val exception =
      assertThrows<Exception> {
        proposerSelectorAdapter.selectProposerForRound(consensusRoundIdentifier)
      }
    assertThat(exception).isInstanceOf(IllegalStateException::class.java)
    assertThat(exception.message).isEqualTo("Parent block not found. parentBlockNumber=${missingBlockNumber - 1}")
  }

  @Test
  fun `passes the state for a given block number to proposer selector`() {
    val beaconChain = mock<BeaconChain>()
    val currentBlockNumber = 3uL
    val previousBlockNumber = currentBlockNumber.dec()
    val expectedState = DataGenerators.randomBeaconState(previousBlockNumber)
    whenever(beaconChain.getBeaconState(previousBlockNumber)).thenReturn(expectedState)
    val proposerSelector = ProposerSelectorAdapter(beaconChain, ProposerSelectorImpl)
    val consensusRoundIdentifier = ConsensusRoundIdentifier(currentBlockNumber.toLong(), 0)
    val result = proposerSelector.selectProposerForRound(consensusRoundIdentifier)

    assertThat(result).isIn(expectedState.validators.map { it.toAddress() })
  }
}
