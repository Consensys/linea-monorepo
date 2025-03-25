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

import maru.core.BeaconBlock
import maru.core.ext.DataGenerators
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class QbftBlockInterfaceAdapterTest {
  @Test
  fun `can replace round number in header`() {
    val beaconBlock =
      BeaconBlock(
        beaconBlockHeader = DataGenerators.randomBeaconBlockHeader(1UL).copy(round = 10UL),
        beaconBlockBody = DataGenerators.randomBeaconBlockBody(),
      )
    val qbftBlock = QbftBlockAdapter(beaconBlock)
    val updatedBlock =
      QbftBlockInterfaceAdapter().replaceRoundInBlock(qbftBlock, 20)
    val updatedBeaconBlockHeader = updatedBlock.header.toBeaconBlockHeader()
    assertEquals(updatedBeaconBlockHeader.round, 20UL)
  }
}
