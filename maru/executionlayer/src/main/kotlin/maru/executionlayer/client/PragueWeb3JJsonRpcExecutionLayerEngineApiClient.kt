/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.executionlayer.client

import java.util.Optional
import maru.core.ExecutionPayload
import maru.extensions.captureTimeSafeFuture
import maru.extensions.getEndpoint
import maru.mappers.Mappers.toDomainExecutionPayload
import maru.mappers.Mappers.toExecutionPayloadV3
import maru.metrics.MaruMetricsCategory
import net.consensys.linea.metrics.DynamicTagTimer
import net.consensys.linea.metrics.MetricsFacade
import net.consensys.linea.metrics.Tag
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceUpdatedResult
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV1
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV3
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadStatusV1
import tech.pegasys.teku.ethereum.executionclient.schema.Response
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes8

class PragueWeb3JJsonRpcExecutionLayerEngineApiClient(
  private val web3jClient: Web3JClient,
  private val metricsFacade: MetricsFacade,
) : ExecutionLayerEngineApiClient {
  private val web3jEngineClient: Web3JExecutionEngineClient = Web3JExecutionEngineClient(web3jClient)

  private fun <T> createRequestTimer(method: String): DynamicTagTimer<Response<T>> =
    metricsFacade.createDynamicTagTimer<Response<T>>(
      category = MaruMetricsCategory.ENGINE_API,
      name = "request.latency",
      description = "Execution Engine API request latency",
      commonTags =
        listOf(
          Tag("fork", "prague"),
          Tag("endpoint", web3jClient.getEndpoint()),
          Tag("method", method),
        ),
      tagValueExtractor = {
        when {
          it.payload != null -> listOf(Tag("status", "success"))
          else -> listOf(Tag("status", "failure"))
        }
      },
      tagValueExtractorOnError = { listOf(Tag("status", "failure")) },
    )

  override fun getPayload(payloadId: Bytes8): SafeFuture<Response<ExecutionPayload>> =
    createRequestTimer<ExecutionPayload>(method = "getPayload").captureTimeSafeFuture(
      web3jEngineClient.getPayloadV4(payloadId).thenApply {
        when {
          it.payload != null ->
            Response.fromPayloadReceivedAsJson(it.payload.executionPayload.toDomainExecutionPayload())

          it.errorMessage != null ->
            Response.fromErrorMessage(it.errorMessage)

          else ->
            throw IllegalStateException("Failed to get payload!")
        }
      },
    )

  override fun newPayload(executionPayload: ExecutionPayload): SafeFuture<Response<PayloadStatusV1>> =
    createRequestTimer<PayloadStatusV1>(method = "newPayload").captureTimeSafeFuture(
      web3jEngineClient
        .newPayloadV4(executionPayload.toExecutionPayloadV3(), emptyList(), Bytes32.ZERO, emptyList())
        .thenApply {
          if (it.payload != null) {
            Response.fromPayloadReceivedAsJson(it.payload)
          } else {
            Response.fromErrorMessage(it.errorMessage)
          }
        },
    )

  override fun forkChoiceUpdate(
    forkChoiceState: ForkChoiceStateV1,
    payloadAttributes: PayloadAttributesV1?,
  ): SafeFuture<Response<ForkChoiceUpdatedResult>> =
    createRequestTimer<ForkChoiceUpdatedResult>(
      method = "forkChoiceUpdate",
    ).captureTimeSafeFuture(
      web3jEngineClient.forkChoiceUpdatedV3(forkChoiceState, Optional.ofNullable(payloadAttributes?.toV3())),
    )

  override fun getLatestBlockHash(): SafeFuture<ByteArray> =
    web3jEngineClient.powChainHead.thenApply { powBlock -> powBlock.blockHash.toArray() }

  override fun isOnline(): SafeFuture<Boolean> =
    this
      .getLatestBlockHash()
      .thenApply { true }
      .exceptionally { false }

  private fun PayloadAttributesV1.toV3(): PayloadAttributesV3 =
    PayloadAttributesV3(
      this.timestamp,
      this.prevRandao,
      this.suggestedFeeRecipient,
      emptyList(),
      Bytes32.ZERO,
    )
}
