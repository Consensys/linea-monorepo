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

import maru.consensus.blockimport.SealedBeaconBlockImporter
import maru.core.Protocol
import maru.p2p.P2PNetwork

class QbftConsensusFollower(
  val p2pNetwork: P2PNetwork,
  val blockImporter: SealedBeaconBlockImporter,
) : Protocol {
  private var subscriptionId: Int? = null

  override fun start() {
    subscriptionId = p2pNetwork.subscribeToBlocks(blockImporter::importBlock)
  }

  override fun stop() {
    if (subscriptionId != null) {
      p2pNetwork.unsubscribe(subscriptionId!!)
    }
  }
}
