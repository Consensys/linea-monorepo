/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.config

import java.net.URL
import java.nio.file.Path
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import linea.domain.BlockParameter
import linea.domain.RetryConfig
import linea.kotlin.assertIs20Bytes
import maru.extensions.encodeHex

data class Persistence(
  val dataPath: Path,
  val privateKeyPath: Path = dataPath.resolve("private-key"),
)

data class ApiEndpointConfig(
  val endpoint: URL,
  val jwtSecretPath: String? = null,
  val requestRetries: RetryConfig =
    RetryConfig.endlessRetry(
      backoffDelay = 1.seconds,
      failuresWarningThreshold = 3u,
    ),
)

data class FollowersConfig(
  val followers: Map<String, ApiEndpointConfig>,
)

data class P2P(
  val ipAddress: String,
  val port: UInt,
  val staticPeers: List<String> = emptyList(),
  val reconnectDelay: Duration = 5.seconds,
  val maxPeers: Int = 25,
  val discovery: Discovery? = null,
  val statusUpdate: StatusUpdateConfig = StatusUpdateConfig(),
) {
  data class Discovery(
    val port: UInt,
    val bootnodes: List<String> = emptyList(),
    val refreshInterval: Duration,
  )

  data class StatusUpdateConfig(
    val renewal: Duration = 5.minutes,
    val renewalLeeway: Duration = 10.seconds,
    val timeout: Duration = 10.seconds,
  )
}

data class ValidatorElNode(
  val ethApiEndpoint: ApiEndpointConfig,
  val engineApiEndpoint: ApiEndpointConfig,
)

data class QbftOptions(
  val minBlockBuildTime: Duration = 500.milliseconds,
  val messageQueueLimit: Int = 1000,
  val roundExpiry: Duration = 1.seconds,
  val duplicateMessageLimit: Int = 100,
  val futureMessageMaxDistance: Long = 10L,
  val futureMessagesLimit: Long = 1000L,
  val feeRecipient: ByteArray,
) {
  init {
    feeRecipient.assertIs20Bytes("feeRecipient")
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as QbftOptions

    if (messageQueueLimit != other.messageQueueLimit) return false
    if (duplicateMessageLimit != other.duplicateMessageLimit) return false
    if (futureMessageMaxDistance != other.futureMessageMaxDistance) return false
    if (futureMessagesLimit != other.futureMessagesLimit) return false
    if (minBlockBuildTime != other.minBlockBuildTime) return false
    if (roundExpiry != other.roundExpiry) return false
    if (!feeRecipient.contentEquals(other.feeRecipient)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = messageQueueLimit
    result = 31 * result + duplicateMessageLimit
    result = 31 * result + futureMessageMaxDistance.hashCode()
    result = 31 * result + futureMessagesLimit.hashCode()
    result = 31 * result + minBlockBuildTime.hashCode()
    result = 31 * result + roundExpiry.hashCode()
    result = 31 * result + feeRecipient.contentHashCode()
    return result
  }

  override fun toString(): String =
    "QbftOptions(" +
      "minBlockBuildTime=$minBlockBuildTime, " +
      "messageQueueLimit=$messageQueueLimit, " +
      "roundExpiry=$roundExpiry, " +
      "duplicateMessageLimit=$duplicateMessageLimit, " +
      "futureMessageMaxDistance=$futureMessageMaxDistance, " +
      "futureMessagesLimit=$futureMessagesLimit, " +
      "feeRecipient=${feeRecipient.encodeHex()}" +
      ")"
}

data class ObservabilityOptions(
  val port: UInt,
  val prometheusMetricsEnabled: Boolean = true,
  val jvmMetricsEnabled: Boolean = true,
)

data class LineaConfig(
  val contractAddress: ByteArray,
  val l1EthApi: ApiEndpointConfig,
  val l1PollingInterval: Duration = 6.seconds,
  val l1HighestBlockTag: BlockParameter = BlockParameter.Tag.FINALIZED,
) {
  init {
    contractAddress.assertIs20Bytes("contractAddress")
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as LineaConfig

    if (!contractAddress.contentEquals(other.contractAddress)) return false
    if (l1EthApi != other.l1EthApi) return false
    if (l1PollingInterval != other.l1PollingInterval) return false
    if (l1HighestBlockTag != other.l1HighestBlockTag) return false

    return true
  }

  override fun hashCode(): Int {
    var result = contractAddress.contentHashCode()
    result = 31 * result + l1EthApi.hashCode()
    result = 31 * result + l1PollingInterval.hashCode()
    result = 31 * result + l1HighestBlockTag.hashCode()
    return result
  }
}

data class ApiConfig(
  val port: UInt,
)

data class SyncingConfig(
  val peerChainHeightPollingInterval: Duration,
  val peerChainHeightGranularity: UInt,
  val elSyncStatusRefreshInterval: Duration,
  val download: Download? = Download(),
) {
  data class Download(
    val blockRangeRequestTimeout: Duration = 5.seconds,
    val blocksBatchSize: UInt = 10u,
    val blocksParallelism: UInt = 1u,
    val maxRetries: UInt = 5u,
  )
}

data class MaruConfig(
  val allowEmptyBlocks: Boolean = false,
  val persistence: Persistence,
  val qbftOptions: QbftOptions?,
  val p2pConfig: P2P?,
  val validatorElNode: ValidatorElNode,
  val followers: FollowersConfig,
  val observabilityOptions: ObservabilityOptions,
  val linea: LineaConfig? = null,
  val apiConfig: ApiConfig,
  val syncing: SyncingConfig,
)
