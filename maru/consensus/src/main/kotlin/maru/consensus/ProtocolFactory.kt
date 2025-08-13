/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import maru.config.consensus.delegated.ElDelegatedConfig
import maru.config.consensus.qbft.QbftConsensusConfig
import maru.core.NoOpProtocol
import maru.core.Protocol

interface ProtocolFactory {
  fun create(forkSpec: ForkSpec): Protocol
}

class OmniProtocolFactory(
  private val qbftConsensusFactory: ProtocolFactory,
) : ProtocolFactory {
  override fun create(forkSpec: ForkSpec): Protocol =
    when (forkSpec.configuration) {
      is QbftConsensusConfig -> {
        qbftConsensusFactory.create(forkSpec)
      }

      is ElDelegatedConfig -> {
        NoOpProtocol()
      }

      else -> {
        throw IllegalArgumentException("Fork $forkSpec is unknown!")
      }
    }
}
