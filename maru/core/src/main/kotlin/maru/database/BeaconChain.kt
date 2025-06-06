/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
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
