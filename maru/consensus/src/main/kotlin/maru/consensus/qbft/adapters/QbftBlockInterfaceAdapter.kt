/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft.adapters

import maru.core.BeaconBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockInterface

/**
 * Adapter class for QBFT block interface, this provides a way to replace the round number in a block
 */
class QbftBlockInterfaceAdapter : QbftBlockInterface {
  override fun replaceRoundInBlock(
    proposalBlock: QbftBlock,
    roundNumber: Int,
  ): QbftBlock {
    val beaconBlockHeader = proposalBlock.header.toBeaconBlockHeader()
    val replacedBeaconBlockHeader =
      beaconBlockHeader.copy(
        round = roundNumber.toUInt(),
      )
    return QbftBlockAdapter(
      BeaconBlock(replacedBeaconBlockHeader, proposalBlock.toBeaconBlock().beaconBlockBody),
    )
  }
}
