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
package maru.config

import java.net.URL
import java.nio.file.Path
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

data class Persistence(
  val dataPath: Path,
  val privateKeyPath: Path = dataPath.resolve("private-key"),
)

data class ApiEndpointConfig(
  val endpoint: URL,
  val jwtSecretPath: String? = null,
)

data class FollowersConfig(
  val followers: Map<String, ApiEndpointConfig>,
)

data class P2P(
  val ipAddress: String,
  val port: UInt,
  val staticPeers: List<String> = emptyList(),
  val reconnectDelay: Duration = 5.seconds,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as P2P

    if (ipAddress != other.ipAddress) return false
    if (port != other.port) return false
    if (staticPeers != other.staticPeers) return false

    return true
  }

  override fun hashCode(): Int {
    var result = ipAddress.hashCode()
    result = 31 * result + port.hashCode()
    result = 31 * result + staticPeers.hashCode()
    return result
  }
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
    require(feeRecipient.size == 20) {
      "feeRecipient address must be 20 bytes long, " +
        "but it's ${feeRecipient.size} bytes long!"
    }
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
}

data class MaruConfig(
  val persistence: Persistence,
  val qbftOptions: QbftOptions?,
  val p2pConfig: P2P?,
  val validatorElNode: ValidatorElNode,
  val followers: FollowersConfig,
)
