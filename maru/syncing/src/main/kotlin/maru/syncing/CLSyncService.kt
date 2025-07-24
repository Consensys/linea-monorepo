/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

import maru.services.LongRunningService

interface CLSyncService {
  fun setSyncTarget(syncTarget: ULong)

  /**
   * Notifies the handler when the <b>latest<b/> target is reached.
   * If target is updated, onSyncComplete won't be called for previous targets
   */
  fun onSyncComplete(handler: (syncTarget: ULong) -> Unit)
}

class CLSyncPipelineImpl :
  CLSyncService,
  LongRunningService {
  override fun setSyncTarget(syncTarget: ULong) {
    TODO("Not yet implemented")
  }

  override fun onSyncComplete(handler: (ULong) -> Unit) {
    TODO("Not yet implemented")
  }

  override fun start() {
    TODO("Not yet implemented")
  }

  override fun stop() {
    TODO("Not yet implemented")
  }
}
