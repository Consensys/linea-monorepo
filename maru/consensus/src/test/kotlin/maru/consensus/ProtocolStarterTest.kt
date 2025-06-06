/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import java.time.Clock
import java.time.Instant
import java.time.ZoneOffset
import kotlin.random.Random
import kotlin.random.nextULong
import maru.core.Protocol
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class ProtocolStarterTest {
  private class StubProtocol : Protocol {
    var started = false

    override fun start() {
      started = true
    }

    override fun stop() {
      started = false
    }
  }

  private val chainId = 1337u

  private val protocol1 = StubProtocol()
  private val protocol2 = StubProtocol()

  private val protocolConfig1 = object : ConsensusConfig {}
  private val forkSpec1 = ForkSpec(0, 1, protocolConfig1)
  private val protocolConfig2 = object : ConsensusConfig {}
  private val forkSpec2 =
    ForkSpec(
      15,
      2,
      protocolConfig2,
    )

  @BeforeEach
  fun stopStubProtocols() {
    protocol1.stop()
    protocol2.stop()
  }

  @Test
  fun `ProtocolStarter kickstarts the protocol based on the latest known block metadata`() {
    val forksSchedule =
      ForksSchedule(
        chainId,
        listOf(
          forkSpec1,
          forkSpec2,
        ),
      )
    val metadataProvider = { randomBlockMetadata(15) }
    val protocolStarter =
      createProtocolStarter(
        forksSchedule = forksSchedule,
        metadataProvider = metadataProvider,
        clockMilliseconds = 14000,
      )
    protocolStarter.start()
    val currentProtocolWithConfig = protocolStarter.currentProtocolWithForkReference.get()
    assertThat(currentProtocolWithConfig.fork).isEqualTo(forkSpec2)
    assertThat(protocol2.started).isTrue()
    assertThat(protocol1.started).isFalse()
  }

  @Test
  fun `ProtocolStarter doesn't re-create the existing protocol`() {
    val forksSchedule =
      ForksSchedule(
        chainId,
        listOf(
          forkSpec1,
          forkSpec2,
        ),
      )
    val metadataProvider = { randomBlockMetadata(15) }
    val protocolStarter =
      createProtocolStarter(
        forksSchedule = forksSchedule,
        metadataProvider = metadataProvider,
        clockMilliseconds = 16000,
      )
    protocolStarter.start()
    val initiallyCreatedProtocol = protocolStarter.currentProtocolWithForkReference.get().protocol
    protocolStarter.handleNewBlock(randomBlockMetadata(16))
    val currentProtocol = protocolStarter.currentProtocolWithForkReference.get().protocol
    assertThat(initiallyCreatedProtocol).isSameAs(currentProtocol)
    assertThat(protocol2.started).isTrue()
    assertThat(protocol1.started).isFalse()
  }

  @Test
  fun `ProtocolStarter re-creates the protocol when the switch is needed`() {
    val forksSchedule =
      ForksSchedule(
        chainId,
        listOf(
          forkSpec1,
          forkSpec2,
        ),
      )
    val metadataProvider = { randomBlockMetadata(13) }
    val protocolStarter =
      createProtocolStarter(
        forksSchedule = forksSchedule,
        metadataProvider = metadataProvider,
        clockMilliseconds = 14000,
      )
    protocolStarter.start()

    val initiallyCreatedProtocolWithConfig = protocolStarter.currentProtocolWithForkReference.get()
    assertThat(initiallyCreatedProtocolWithConfig.fork).isEqualTo(forkSpec1)
    assertThat(protocol1.started).isTrue()
    assertThat(protocol2.started).isFalse()

    protocolStarter.handleNewBlock(randomBlockMetadata(14))
    val currentProtocolWithConfig = protocolStarter.currentProtocolWithForkReference.get()
    assertThat(currentProtocolWithConfig.fork).isEqualTo(forkSpec2)
    assertThat(initiallyCreatedProtocolWithConfig.protocol).isNotSameAs(currentProtocolWithConfig.protocol)
    assertThat(protocol1.started).isFalse()
    assertThat(protocol2.started).isTrue()
  }

  @Test
  fun `if latest block is far in the past, current time takes precedence`() {
    val forksSchedule =
      ForksSchedule(
        chainId,
        listOf(
          forkSpec1,
          forkSpec2,
        ),
      )
    val metadataProvider = { randomBlockMetadata(2) }
    val protocolStarter =
      createProtocolStarter(
        forksSchedule = forksSchedule,
        metadataProvider = metadataProvider,
        clockMilliseconds = 15000,
      )
    protocolStarter.start()

    val initiallyCreatedProtocolWithConfig = protocolStarter.currentProtocolWithForkReference.get()
    assertThat(initiallyCreatedProtocolWithConfig.fork).isEqualTo(forkSpec2)
    assertThat(protocol1.started).isFalse()
    assertThat(protocol2.started).isTrue()
  }

  private val protocolFactory =
    object : ProtocolFactory {
      override fun create(forkSpec: ForkSpec): Protocol =
        when (forkSpec.configuration) {
          protocolConfig1 -> protocol1
          protocolConfig2 -> protocol2
          else -> error("invalid protocol config")
        }
    }

  private fun createProtocolStarter(
    forksSchedule: ForksSchedule,
    metadataProvider: MetadataProvider,
    clockMilliseconds: Long,
  ): ProtocolStarter =
    ProtocolStarter(
      forksSchedule = forksSchedule,
      protocolFactory = protocolFactory,
      metadataProvider = metadataProvider,
      nextBlockTimestampProvider =
        NextBlockTimestampProviderImpl(
          clock = Clock.fixed(Instant.ofEpochMilli(clockMilliseconds), ZoneOffset.UTC),
          forksSchedule = forksSchedule,
        ),
    )

  fun randomBlockMetadata(timestamp: Long): BlockMetadata =
    BlockMetadata(
      Random.nextULong(),
      blockHash = Random.nextBytes(32),
      unixTimestampSeconds = timestamp,
    )
}
