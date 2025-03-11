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
package maru.consensus.qbft.adaptors

import maru.consensus.qbft.adaptors.toBeaconBlock
import maru.consensus.qbft.adaptors.toBeaconBlockHeader
import maru.core.BeaconBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockInterface
import org.hyperledger.besu.consensus.qbft.core.types.QbftHashMode

/**
 * Adaptor class for QBFT block interface, this provides a way to replace the round number in a block
 */
class QbftBlockInterfaceAdaptor : QbftBlockInterface {
  override fun replaceRoundInBlock(
    proposalBlock: QbftBlock,
    roundNumber: Int,
    hashMode: QbftHashMode,
  ): QbftBlock {
    val beaconBlockHeader = proposalBlock.header.toBeaconBlockHeader()
    val replacedBeaconBlockHeader =
      beaconBlockHeader.copy(
        round = roundNumber.toULong(),
      )
    return QbftBlockAdaptor(
      BeaconBlock(replacedBeaconBlockHeader, proposalBlock.toBeaconBlock().beaconBlockBody),
    )
  }
}
