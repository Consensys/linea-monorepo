/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

import java.util.Timer
import java.util.TimerTask

// Test implementation that allows controlling the timer execution. Supports only scheduleAtFixedRate
internal class TestablePeriodicTimer : Timer("test-timer", true) {
  var scheduledTask: TimerTask? = null
  var delay: Long? = null
  var period: Long? = null

  override fun scheduleAtFixedRate(
    task: TimerTask,
    delay: Long,
    period: Long,
  ) {
    scheduledTask = task
    this.delay = delay
    this.period = period
  }

  fun runNextTask() {
    scheduledTask?.run()
  }

  override fun cancel() {
    super.cancel()
    scheduledTask = null
    delay = null
    period = null
  }
}
