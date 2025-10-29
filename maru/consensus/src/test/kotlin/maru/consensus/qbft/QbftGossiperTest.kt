/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft

import kotlin.test.Test
import maru.consensus.qbft.adapters.P2PValidatorMulticaster
import org.hyperledger.besu.consensus.qbft.core.types.QbftMessage
import org.mockito.Mockito.mock
import org.mockito.Mockito.never
import org.mockito.Mockito.verify
import org.mockito.kotlin.any
import org.mockito.kotlin.whenever
import org.hyperledger.besu.ethereum.p2p.rlpx.wire.MessageData as BesuMessageData

class QbftGossiperTest {
  private val mockP2PValidatorMulticaster = mock<P2PValidatorMulticaster>()
  private val mockMessageData = mock<BesuMessageData>()
  private val mockQbftMessage = mock<QbftMessage>()

  private val gossiper = QbftGossiper(mockP2PValidatorMulticaster)

  @Test
  fun `should send message when replayed is true`() {
    whenever(mockQbftMessage.data).thenReturn(mockMessageData)

    gossiper.send(mockQbftMessage, replayed = true)

    verify(mockP2PValidatorMulticaster).send(mockMessageData)
  }

  @Test
  fun `should not send message when replayed is false`() {
    gossiper.send(mockQbftMessage, replayed = false)

    verify(mockP2PValidatorMulticaster, never()).send(any<BesuMessageData>())
  }

  @Test
  fun `should not send message when replayed is false even if message data is available`() {
    whenever(mockQbftMessage.data).thenReturn(mockMessageData)

    gossiper.send(mockQbftMessage, replayed = false)

    verify(mockP2PValidatorMulticaster, never()).send(any<BesuMessageData>())
  }
}
