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

import maru.core.BeaconBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockHeader

/**
 * Adaptor class to convert a BeaconBlock to a QBFT block
 */
class QbftBlockAdaptor(
  val beaconBlock: BeaconBlock,
) : QbftBlock {
  val qbftHeader: QbftBlockHeader = QbftBlockHeaderAdaptor(beaconBlock.beaconBlockHeader)

  override fun getHeader(): QbftBlockHeader = qbftHeader

  override fun isEmpty(): Boolean =
    beaconBlock.beaconBlockBody.executionPayload.transactions
      .isEmpty()

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (other !is QbftBlockAdaptor) return false

    if (beaconBlock != other.beaconBlock) return false
    if (qbftHeader != other.qbftHeader) return false

    return true
  }

  override fun hashCode(): Int {
    var result = beaconBlock.hashCode()
    result = 31 * result + qbftHeader.hashCode()
    return result
  }
}
