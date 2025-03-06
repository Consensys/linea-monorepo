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

import maru.core.Protocol
import maru.core.ext.DataGenerators
import maru.executionlayer.client.ExecutionLayerClient
import maru.executionlayer.manager.BlockMetadata
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceUpdatedResult
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV1
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadStatusV1
import tech.pegasys.teku.ethereum.executionclient.schema.Response
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes8

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
    val unexpectedProtocolConfig = protocolConfig1
    val expectedProtocolConfig = protocolConfig2
    val forksSchedule =
      ForksSchedule(
        listOf(
          ForkSpec(0u, unexpectedProtocolConfig),
          ForkSpec(
            15u,
            expectedProtocolConfig,
          ),
        ),
      )
    val executionLayerClient =
      createFakeExecutionLayerClient(latestBlockMetadataToReturn = DataGenerators.randomBlockMetadata(15u))
    val protocolStarter =
      createProtocolStarter(
        forksSchedule = forksSchedule,
        executionLayerClient = executionLayerClient,
      )
    protocolStarter.start()
    val currentProtocolWithConfig = protocolStarter.currentProtocolWithConfig.get()
    assertThat(currentProtocolWithConfig.config).isEqualTo(expectedProtocolConfig)
    assertThat(protocol2.started).isTrue()
    assertThat(protocol1.started).isFalse()
  }

  @Test
  fun `ProtocolStarter doesn't re-create the existing protocol`() {
    val unexpectedProtocolConfig = protocolConfig1
    val expectedProtocolConfig = protocolConfig2
    val forksSchedule =
      ForksSchedule(
        listOf(
          ForkSpec(0u, unexpectedProtocolConfig),
          ForkSpec(
            15u,
            expectedProtocolConfig,
          ),
        ),
      )
    val executionLayerClient =
      createFakeExecutionLayerClient(latestBlockMetadataToReturn = DataGenerators.randomBlockMetadata(15u))
    val protocolStarter =
      createProtocolStarter(
        forksSchedule = forksSchedule,
        executionLayerClient = executionLayerClient,
      )
    protocolStarter.start()
    val initiallyCreatedProtocol = protocolStarter.currentProtocolWithConfig.get().protocol
    protocolStarter.handleNewBlock(DataGenerators.randomBlockMetadata(16u))
    val currentProtocol = protocolStarter.currentProtocolWithConfig.get().protocol
    assertThat(initiallyCreatedProtocol).isSameAs(currentProtocol)
    assertThat(protocol2.started).isTrue()
    assertThat(protocol1.started).isFalse()
  }

  @Test
  fun `ProtocolStarter re-creates the protocol when the switch is needed`() {
    val firstProtocolConfig = protocolConfig1
    val secondProtocolConfig = protocolConfig2
    val forksSchedule =
      ForksSchedule(
        listOf(
          ForkSpec(0u, firstProtocolConfig),
          ForkSpec(
            15u,
            secondProtocolConfig,
          ),
        ),
      )
    val executionLayerClient =
      createFakeExecutionLayerClient(latestBlockMetadataToReturn = DataGenerators.randomBlockMetadata(13u))
    val protocolStarter =
      createProtocolStarter(
        forksSchedule = forksSchedule,
        executionLayerClient = executionLayerClient,
      )
    protocolStarter.start()

    val initiallyCreatedProtocolWithConfig = protocolStarter.currentProtocolWithConfig.get()
    assertThat(initiallyCreatedProtocolWithConfig.config).isEqualTo(firstProtocolConfig)
    assertThat(protocol1.started).isTrue()
    assertThat(protocol2.started).isFalse()

    protocolStarter.handleNewBlock(DataGenerators.randomBlockMetadata(14u))
    val currentProtocolWithConfig = protocolStarter.currentProtocolWithConfig.get()
    assertThat(currentProtocolWithConfig.config).isEqualTo(secondProtocolConfig)
    assertThat(initiallyCreatedProtocolWithConfig.protocol).isNotSameAs(currentProtocolWithConfig.protocol)
    assertThat(protocol1.started).isFalse()
    assertThat(protocol2.started).isTrue()
  }

  private fun createFakeExecutionLayerClient(latestBlockMetadataToReturn: BlockMetadata): ExecutionLayerClient =
    object : ExecutionLayerClient {
      override fun getLatestBlockMetadata(): SafeFuture<BlockMetadata> =
        SafeFuture.completedFuture(
          latestBlockMetadataToReturn,
        )

      override fun getPayload(payloadId: Bytes8): SafeFuture<Response<ExecutionPayloadV1>> {
        TODO("Not yet implemented")
      }

      override fun newPayload(executionPayload: ExecutionPayloadV1): SafeFuture<Response<PayloadStatusV1>> {
        TODO("Not yet implemented")
      }

      override fun forkChoiceUpdate(
        forkChoiceState: ForkChoiceStateV1,
        payloadAttributes: PayloadAttributesV1?,
      ): SafeFuture<Response<ForkChoiceUpdatedResult>> {
        TODO("Not yet implemented")
      }
    }

  private val protocolConfig1 = object : ConsensusConfig {}
  private val protocolConfig2 = object : ConsensusConfig {}

  private val protocolFactory =
    object : ProtocolFactory {
      override fun create(protocolConfig: ConsensusConfig): Protocol =
        when (protocolConfig) {
          protocolConfig1 -> protocol1
          protocolConfig2 -> protocol2
          else -> error("invalid protocol config")
        }
    }

  private fun createProtocolStarter(
    forksSchedule: ForksSchedule,
    executionLayerClient: ExecutionLayerClient,
  ): ProtocolStarter =
    ProtocolStarter(
      forksSchedule = forksSchedule,
      protocolFactory = protocolFactory,
      executionLayerClient = executionLayerClient,
    )
}
