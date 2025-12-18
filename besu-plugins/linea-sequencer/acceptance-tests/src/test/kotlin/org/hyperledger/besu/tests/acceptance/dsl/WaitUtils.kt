/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package org.hyperledger.besu.tests.acceptance.dsl

import org.awaitility.Awaitility
import org.awaitility.core.ThrowingRunnable
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

/** Contains functionality for timeouts. */
object WaitUtils {
  fun waitFor(
    timeout: Duration,
    pollingInterval: Duration = 1.seconds,
    condition: ThrowingRunnable,
  ) {
    Awaitility.await()
      .pollInterval(pollingInterval.toJavaDuration())
      .ignoreExceptions()
      .atMost(timeout.toJavaDuration())
      .untilAsserted(condition)
  }
}
