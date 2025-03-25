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

import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.ethereum.rlp.RLP
import org.junit.jupiter.api.Test

class QbftBlockCodecAdapterTest {
  @Test
  fun `can encode and decode same value`() {
    val beaconBlock = DataGenerators.randomBeaconBlock(10U)
    val testValue = QbftBlockAdapter(beaconBlock)
    val qbftBlockCodecAdapter = QbftBlockCodecAdapter()

    val encodedData = RLP.encode { rlpOutput -> qbftBlockCodecAdapter.writeTo(testValue, rlpOutput) }
    val decodedValue = qbftBlockCodecAdapter.readFrom(RLP.input(encodedData))
    assertThat(decodedValue).isEqualTo(testValue)
  }
}
