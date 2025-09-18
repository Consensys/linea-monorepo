/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft.adapters

import maru.consensus.state.StateTransition
import maru.core.BeaconBlock
import maru.core.EMPTY_HASH
import maru.core.HashUtil
import maru.serialization.rlp.stateRoot
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockInterface

/**
 * Adapter class for QBFT block interface, this provides a way to replace the round number in a block
 */
class QbftBlockInterfaceAdapter(
  private val stateTransition: StateTransition,
) : QbftBlockInterface {
  override fun replaceRoundInBlock(
    proposalBlock: QbftBlock,
    roundNumber: Int,
  ): QbftBlock {
    val beaconBlockHeader = proposalBlock.header.toBeaconBlockHeader()
    val beaconBlockBody = proposalBlock.toBeaconBlock().beaconBlockBody
    val replacedBeaconBlockHeader =
      beaconBlockHeader.copy(
        round = roundNumber.toUInt(),
      )

    // update header with new state root
    val postState = stateTransition.processBlock(BeaconBlock(replacedBeaconBlockHeader, beaconBlockBody)).get()
    val stateRootHeader =
      postState.beaconBlockHeader.copy(
        stateRoot = EMPTY_HASH,
      )
    val stateRoot = HashUtil.stateRoot(postState.copy(beaconBlockHeader = stateRootHeader))
    val finalBlockHeader = stateRootHeader.copy(stateRoot = stateRoot)

    return QbftBlockAdapter(BeaconBlock(finalBlockHeader, beaconBlockBody))
  }
}
