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
package maru.consensus.adaptors

import maru.consensus.qbft.adaptors.QbftBlockAdaptor
import maru.consensus.qbft.adaptors.QbftBlockCodecAdaptor
import maru.core.ext.DataGenerators
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.consensus.qbft.core.types.QbftHashMode
import org.hyperledger.besu.ethereum.rlp.RLP
import org.junit.jupiter.api.Test

class QbftBlockCodecAdaptorTest {
  @Test
  fun `can encode and decode same value for committed seal`() {
    val beaconBlock = DataGenerators.randomBeaconBlock(10U)
    val testValue = QbftBlockAdaptor(beaconBlock)
    val qbftBlockCodecAdaptor = QbftBlockCodecAdaptor()

    val encodedData = RLP.encode { rlpOutput -> qbftBlockCodecAdaptor.writeTo(testValue, rlpOutput) }
    val decodedValue = qbftBlockCodecAdaptor.readFrom(RLP.input(encodedData), QbftHashMode.COMMITTED_SEAL)
    assertThat(decodedValue).isEqualTo(testValue)
  }

  @Test
  fun `can encode and decode same value for onchain`() {
    val beaconBlock = DataGenerators.randomBeaconBlock(10U)
    val testValue = QbftBlockAdaptor(beaconBlock)
    val qbftBlockCodecAdaptor = QbftBlockCodecAdaptor()

    val encodedData = RLP.encode { rlpOutput -> qbftBlockCodecAdaptor.writeTo(testValue, rlpOutput) }
    val decodedValue = qbftBlockCodecAdaptor.readFrom(RLP.input(encodedData), QbftHashMode.ONCHAIN)
    assertThat(decodedValue).isEqualTo(testValue)
  }
}
