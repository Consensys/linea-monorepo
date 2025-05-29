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
package maru.p2p

import maru.core.SealedBeaconBlock
import tech.pegasys.teku.infrastructure.async.SafeFuture

class SealedBeaconBlockBroadcaster(
  val p2PNetwork: P2PNetwork,
) : SealedBeaconBlockHandler<Unit> {
  override fun handleSealedBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<Unit> {
    // TODO: New block message might need an intermediary wrapper in the future
    val message = Message(MessageType.BEACON_BLOCK, payload = sealedBeaconBlock)
    p2PNetwork.broadcastMessage(message)
    return SafeFuture.completedFuture(Unit)
  }
}
