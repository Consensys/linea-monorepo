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
package maru.database

import maru.core.BeaconState
import maru.core.SealedBeaconBlock

class InMemoryBeaconChain(
  initialBeaconState: BeaconState,
) : BeaconChain {
  private val beaconStateByBlockRoot = mutableMapOf<ByteArray, BeaconState>()
  private val sealedBeaconBlockByBlockRoot = mutableMapOf<ByteArray, SealedBeaconBlock>()
  private val sealedBeaconBlockByBlockNumber = mutableMapOf<ULong, SealedBeaconBlock>()

  private var latestBeaconState: BeaconState = initialBeaconState

  override fun getLatestBeaconState(): BeaconState = latestBeaconState

  override fun getBeaconState(beaconBlockRoot: ByteArray): BeaconState? = beaconStateByBlockRoot[beaconBlockRoot]

  override fun getSealedBeaconBlock(beaconBlockRoot: ByteArray): SealedBeaconBlock? =
    sealedBeaconBlockByBlockRoot[beaconBlockRoot]

  override fun getSealedBeaconBlock(beaconBlockNumber: ULong): SealedBeaconBlock? =
    sealedBeaconBlockByBlockNumber[beaconBlockNumber]

  override fun newUpdater(): Updater = InMemoryUpdater(this)

  override fun close() {
    // No-op for in-memory beacon chain
  }

  private class InMemoryUpdater(
    private val beaconChain: InMemoryBeaconChain,
  ) : Updater {
    private val beaconStateByBlockRoot = mutableMapOf<ByteArray, BeaconState>()
    private val sealedBeaconBlockByBlockRoot = mutableMapOf<ByteArray, SealedBeaconBlock>()
    private val sealedBeaconBlockByBlockNumber = mutableMapOf<ULong, SealedBeaconBlock>()

    private var newBeaconState: BeaconState? = null

    override fun putBeaconState(beaconState: BeaconState): Updater {
      beaconStateByBlockRoot[beaconState.latestBeaconBlockHeader.hash] = beaconState
      newBeaconState = beaconState
      return this
    }

    override fun putSealedBeaconBlock(sealedBeaconBlock: SealedBeaconBlock): Updater {
      sealedBeaconBlockByBlockRoot[sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash] = sealedBeaconBlock
      sealedBeaconBlockByBlockNumber[sealedBeaconBlock.beaconBlock.beaconBlockHeader.number] = sealedBeaconBlock
      return this
    }

    override fun commit() {
      beaconChain.beaconStateByBlockRoot.putAll(beaconStateByBlockRoot)
      beaconChain.sealedBeaconBlockByBlockRoot.putAll(sealedBeaconBlockByBlockRoot)
      beaconChain.sealedBeaconBlockByBlockNumber.putAll(sealedBeaconBlockByBlockNumber)
      if (newBeaconState != null) {
        beaconChain.latestBeaconState = newBeaconState!!
      }
    }

    override fun rollback() {
      beaconStateByBlockRoot.clear()
      sealedBeaconBlockByBlockRoot.clear()
      sealedBeaconBlockByBlockNumber.clear()
      newBeaconState = null
    }

    override fun close() {
      // No-op for in-memory updater
    }
  }
}
