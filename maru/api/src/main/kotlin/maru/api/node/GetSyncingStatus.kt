/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.api.node

import com.fasterxml.jackson.annotation.JsonProperty
import io.javalin.http.Context
import io.javalin.http.Handler

data class GetSyncingStatusResponse(
  val data: SyncingStatusData,
)

data class SyncingStatusData(
  @JsonProperty("head_slot") val headSlot: String,
  @JsonProperty("sync_distance") val syncDistance: String,
  @JsonProperty("is_syncing") val isSyncing: Boolean,
  @JsonProperty("is_optimistic") val isOptimistic: Boolean,
  @JsonProperty("el_offline") val elOffline: Boolean,
)

class GetSyncingStatus : Handler {
  override fun handle(ctx: Context) {
    // This is a placeholder implementation. Replace it with actual logic to retrieve syncing status.
    // TODO(Implement actual syncing status retrieval logic when we have syncing in place)
    ctx.status(200).json(
      GetSyncingStatusResponse(
        data =
          SyncingStatusData(
            headSlot = "12345678",
            syncDistance = "0",
            isSyncing = false,
            isOptimistic = false,
            elOffline = false,
          ),
      ),
    )
  }

  companion object {
    const val ROUTE = "/eth/v1/node/syncing"
  }
}
