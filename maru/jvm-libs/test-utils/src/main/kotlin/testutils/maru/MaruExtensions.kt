/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package testutils.maru

import maru.app.MaruApp
import org.awaitility.Awaitility.await
import org.awaitility.core.ConditionFactory

fun MaruApp.awaitTillMaruHasPeers(
  numberOfPeers: UInt,
  await: ConditionFactory = await(),
) {
  await.until { this.peersConnected() >= numberOfPeers }
}
