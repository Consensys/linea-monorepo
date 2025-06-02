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
package maru.testutils

import maru.core.SealedBeaconBlock
import maru.p2p.NoOpP2PNetwork
import maru.p2p.P2PNetwork
import maru.p2p.SealedBeaconBlockHandler
import maru.p2p.ValidationResult
import tech.pegasys.teku.infrastructure.async.SafeFuture

class InjectableSealedBlocksFakeNetwork : P2PNetwork by NoOpP2PNetwork {
  var handler: SealedBeaconBlockHandler<ValidationResult>? = null

  override fun subscribeToBlocks(subscriber: SealedBeaconBlockHandler<ValidationResult>): Int {
    handler = subscriber
    return 0
  }

  fun injectSealedBlock(sealedBlock: SealedBeaconBlock): SafeFuture<ValidationResult> =
    handler!!.handleSealedBlock(sealedBlock)
}
