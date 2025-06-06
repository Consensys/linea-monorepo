/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft.network

import org.hyperledger.besu.consensus.common.bft.network.ValidatorMulticaster
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.ethereum.p2p.rlpx.wire.MessageData

object NoopValidatorMulticaster : ValidatorMulticaster {
  override fun send(message: MessageData) {
  }

  override fun send(
    message: MessageData,
    denyList: Collection<Address?>?,
  ) {
  }
}
