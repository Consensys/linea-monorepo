/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.consensus

import java.time.Clock
import java.time.Instant
import java.time.ZoneOffset
import kotlin.time.Duration.Companion.milliseconds
import maru.consensus.dummy.NextBlockTimestampProviderImpl
import maru.core.Protocol
import maru.core.ext.DataGenerators
import maru.executionlayer.client.MetadataProvider
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture

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

  private val protocol1 = StubProtocol()
  private val protocol2 = StubProtocol()

  @BeforeEach
  fun stopStubProtocols() {
    protocol1.stop()
    protocol2.stop()
  }

  @Test
  fun `ProtocolStarter kickstarts the protocol based on the latest known block metadata`() {
    val forksSchedule =
      ForksSchedule(
        listOf(
          forkSpec1,
          forkSpec2,
        ),
      )
    val metadataProvider = { SafeFuture.completedFuture(DataGenerators.randomBlockMetadata(15)) }
    val protocolStarter =
      createProtocolStarter(
        forksSchedule = forksSchedule,
        metadataProvider = metadataProvider,
        clockMilliseconds = 16000,
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
        listOf(
          forkSpec1,
          forkSpec2,
        ),
      )
    val metadataProvider = { SafeFuture.completedFuture(DataGenerators.randomBlockMetadata(15)) }
    val protocolStarter =
      createProtocolStarter(
        forksSchedule = forksSchedule,
        metadataProvider = metadataProvider,
        clockMilliseconds = 16000,
      )
    protocolStarter.start()
    val initiallyCreatedProtocol = protocolStarter.currentProtocolWithForkReference.get().protocol
    protocolStarter.handleNewBlock(DataGenerators.randomBlockMetadata(16))
    val currentProtocol = protocolStarter.currentProtocolWithForkReference.get().protocol
    assertThat(initiallyCreatedProtocol).isSameAs(currentProtocol)
    assertThat(protocol2.started).isTrue()
    assertThat(protocol1.started).isFalse()
  }

  @Test
  fun `ProtocolStarter re-creates the protocol when the switch is needed`() {
    val forksSchedule =
      ForksSchedule(
        listOf(
          forkSpec1,
          forkSpec2,
        ),
      )
    val metadataProvider = { SafeFuture.completedFuture(DataGenerators.randomBlockMetadata(13)) }
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

    protocolStarter.handleNewBlock(DataGenerators.randomBlockMetadata(14))
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
        listOf(
          forkSpec1,
          forkSpec2,
        ),
      )
    val metadataProvider = { SafeFuture.completedFuture(DataGenerators.randomBlockMetadata(2)) }
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

  private val protocolConfig1 = object : ConsensusConfig {}
  private val forkSpec1 = ForkSpec(0, 1, protocolConfig1)
  private val protocolConfig2 = object : ConsensusConfig {}
  private val forkSpec2 =
    ForkSpec(
      15,
      2,
      protocolConfig2,
    )

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
          0.milliseconds,
        ),
    )
}
