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
import maru.core.ext.DataGenerators
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class QbftBlockInterfaceAdapterTest {
  @Test
  fun `can replace round number in header`() {
    val beaconBlock =
      BeaconBlock(
        beaconBlockHeader = DataGenerators.randomBeaconBlockHeader(1UL).copy(round = 10u),
        beaconBlockBody = DataGenerators.randomBeaconBlockBody(),
      )
    val qbftBlock = QbftBlockAdapter(beaconBlock)
    val updatedBlock =
      QbftBlockInterfaceAdapter().replaceRoundInBlock(qbftBlock, 20)
    val updatedBeaconBlockHeader = updatedBlock.header.toBeaconBlockHeader()
    assertEquals(updatedBeaconBlockHeader.round, 20u)
  }
}
