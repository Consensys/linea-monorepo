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

import maru.consensus.qbft.adapters.toSealedBeaconBlock
import maru.consensus.state.StateTransition
import maru.database.BeaconChain
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockImporter

/**
 * Responsible for: transactional  and El node
 * 1. state transition of node's BeaconChain
 * 2. new block import into an EL node
 * The import is transactional, I.e. all or nothing approach
 */
class QbftBlockImportCoordinator(
  private val beaconChain: BeaconChain,
  private val stateTransition: StateTransition,
  private val beaconBlockImporter: BeaconBlockImporter,
) : QbftBlockImporter {
  override fun importBlock(qbftBlock: QbftBlock): Boolean {
    val sealedBeaconBlock = qbftBlock.toSealedBeaconBlock()
    val beaconBlock = sealedBeaconBlock.beaconBlock

    beaconChain.newUpdater().use { updater ->
      try {
        val resultingState = stateTransition.processBlock(beaconBlock).get()
        updater
          .putBeaconState(resultingState)
          .putSealedBeaconBlock(sealedBeaconBlock)
        beaconBlockImporter.importBlock(beaconBlock).get()
      } catch (e: Exception) {
        updater.rollback()
        return false
      }
      updater.commit()
    }

    return true
  }
}
