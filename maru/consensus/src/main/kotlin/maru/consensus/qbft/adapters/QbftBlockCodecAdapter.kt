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

import maru.serialization.rlp.RLPSerializers.BeaconBlockSerializer
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockCodec
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.ethereum.rlp.RLPOutput

/**
 * Adapter for QBFT block codec, this provides a way to serialize QBFT blocks
 */
class QbftBlockCodecAdapter : QbftBlockCodec {
  override fun readFrom(rlpInput: RLPInput): QbftBlock = QbftBlockAdapter(BeaconBlockSerializer.readFrom(rlpInput))

  override fun writeTo(
    qbftBlock: QbftBlock,
    rlpOutput: RLPOutput,
  ) {
    qbftBlock.toBeaconBlock().let {
      BeaconBlockSerializer.writeTo(it, rlpOutput)
    }
  }
}
