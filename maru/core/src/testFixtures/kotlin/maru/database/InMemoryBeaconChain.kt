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

class InMemoryBeaconChain(
  initialBeaconState: BeaconState,
) : BeaconChain {
  private val beaconStateByBlockRoot = mutableMapOf<ByteArray, BeaconState>()
  private val beaconStateByBlockNumber = mutableMapOf<ULong, BeaconState>()
  private val sealedBeaconBlockByBlockRoot = mutableMapOf<ByteArray, SealedBeaconBlock>()
  private val sealedBeaconBlockByBlockNumber = mutableMapOf<ULong, SealedBeaconBlock>()

  private var latestBeaconState: BeaconState = initialBeaconState

  init {
    newUpdater().putBeaconState(initialBeaconState).commit()
  }

  override fun isInitialized(): Boolean = true

  override fun getLatestBeaconState(): BeaconState = latestBeaconState

  override fun getBeaconState(beaconBlockRoot: ByteArray): BeaconState? = beaconStateByBlockRoot[beaconBlockRoot]

  override fun getBeaconState(beaconBlockNumber: ULong): BeaconState? = beaconStateByBlockNumber[beaconBlockNumber]

  override fun getSealedBeaconBlock(beaconBlockRoot: ByteArray): SealedBeaconBlock? =
    sealedBeaconBlockByBlockRoot[beaconBlockRoot]

  override fun getSealedBeaconBlock(beaconBlockNumber: ULong): SealedBeaconBlock? =
    sealedBeaconBlockByBlockNumber[beaconBlockNumber]

  override fun newUpdater(): BeaconChain.Updater = InMemoryUpdater(this)

  override fun close() {
    // No-op for in-memory beacon chain
  }

  private class InMemoryUpdater(
    private val beaconChain: InMemoryBeaconChain,
  ) : BeaconChain.Updater {
    private val beaconStateByBlockRoot = mutableMapOf<ByteArray, BeaconState>()
    private val beaconStateByBlockNumber = mutableMapOf<ULong, BeaconState>()
    private val sealedBeaconBlockByBlockRoot = mutableMapOf<ByteArray, SealedBeaconBlock>()
    private val sealedBeaconBlockByBlockNumber = mutableMapOf<ULong, SealedBeaconBlock>()

    private var newBeaconState: BeaconState? = null

    override fun putBeaconState(beaconState: BeaconState): BeaconChain.Updater {
      beaconStateByBlockRoot[beaconState.latestBeaconBlockHeader.hash] = beaconState
      beaconStateByBlockNumber[beaconState.latestBeaconBlockHeader.number] = beaconState
      newBeaconState = beaconState
      return this
    }

    override fun putSealedBeaconBlock(sealedBeaconBlock: SealedBeaconBlock): BeaconChain.Updater {
      sealedBeaconBlockByBlockRoot[sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash] = sealedBeaconBlock
      sealedBeaconBlockByBlockNumber[sealedBeaconBlock.beaconBlock.beaconBlockHeader.number] = sealedBeaconBlock
      return this
    }

    override fun commit() {
      beaconChain.beaconStateByBlockRoot.putAll(beaconStateByBlockRoot)
      beaconChain.beaconStateByBlockNumber.putAll(beaconStateByBlockNumber)
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
