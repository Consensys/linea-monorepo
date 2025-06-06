/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft

import maru.consensus.blockimport.SealedBeaconBlockImporter
import maru.core.Protocol
import maru.p2p.P2PNetwork
import maru.p2p.ValidationResult

class QbftConsensusFollower(
  val p2pNetwork: P2PNetwork,
  val blockImporter: SealedBeaconBlockImporter<ValidationResult>,
) : Protocol {
  private var subscriptionId: Int? = null

  override fun start() {
    subscriptionId = p2pNetwork.subscribeToBlocks(blockImporter::importBlock)
  }

  override fun stop() {
    if (subscriptionId != null) {
      p2pNetwork.unsubscribeFromBlocks(subscriptionId!!)
    }
  }
}
