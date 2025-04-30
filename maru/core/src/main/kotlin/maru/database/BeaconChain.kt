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

interface BeaconChain : AutoCloseable {
  fun isInitialized(): Boolean

  fun getLatestBeaconState(): BeaconState

  fun getBeaconState(beaconBlockRoot: ByteArray): BeaconState?

  fun getBeaconState(beaconBlockNumber: ULong): BeaconState?

  fun getSealedBeaconBlock(beaconBlockRoot: ByteArray): SealedBeaconBlock?

  fun getSealedBeaconBlock(beaconBlockNumber: ULong): SealedBeaconBlock?

  fun newUpdater(): Updater

  interface Updater : AutoCloseable {
    fun putBeaconState(beaconState: BeaconState): Updater

    fun putSealedBeaconBlock(sealedBeaconBlock: SealedBeaconBlock): Updater

    fun commit(): Unit

    fun rollback(): Unit
  }
}
