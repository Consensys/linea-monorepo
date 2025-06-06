/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.app

import java.util.Optional
import org.hyperledger.besu.plugin.services.metrics.MetricCategory

enum class MaruMetricsCategory : MetricCategory {
  STORAGE {
    override fun getName(): String = "storage"

    override fun getApplicationPrefix(): Optional<String> = Optional.empty()
  },
}
