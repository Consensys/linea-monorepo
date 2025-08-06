/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing.beaconchain.pipeline

import kotlin.random.Random
import maru.p2p.messages.Status

object DataGenerators {
  fun randomStatus(latestBlockNumber: ULong): Status =
    Status(
      forkIdHash = Random.nextBytes(32),
      latestStateRoot = Random.nextBytes(32),
      latestBlockNumber = latestBlockNumber,
    )
}
