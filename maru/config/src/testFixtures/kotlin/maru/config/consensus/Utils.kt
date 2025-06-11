/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.config.consensus

import com.sksamuel.hoplite.ConfigLoaderBuilder
import com.sksamuel.hoplite.ExperimentalHoplite
import com.sksamuel.hoplite.json.JsonPropertySource
import maru.config.consensus.ForkConfigDecoder
import maru.config.consensus.JsonFriendlyForksSchedule

@OptIn(ExperimentalHoplite::class)
object Utils {
  fun parseBeaconChainConfig(json: String): JsonFriendlyForksSchedule =
    ConfigLoaderBuilder
      .default()
      .addDecoder(ForkConfigDecoder)
      .withExplicitSealedTypes()
      .addSource(JsonPropertySource(json))
      .build()
      .loadConfigOrThrow<JsonFriendlyForksSchedule>()
}
