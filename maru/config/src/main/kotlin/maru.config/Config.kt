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
import kotlin.time.Duration

data class ApiEndpointConfig(
  val endpoint: URL,
  val jwtSecretPath: String? = null,
)

data class ValidatorClientConfig(
  val engineApiClientConfig: ApiEndpointConfig,
  val minTimeBetweenGetPayloadAttempts: Duration,
)

data class FollowersConfig(
  val followers: Map<String, ApiEndpointConfig>,
)

data class P2P(
  val port: UInt = 9000u,
)

data class Validator(
  val key: ByteArray,
  val client: ValidatorClientConfig,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as Validator

    return key.contentEquals(other.key)
  }

  override fun hashCode(): Int = key.contentHashCode()
}

data class DummyConsensusOptions(
  // Since we cannot finish block production instantly at expected time, we need to set some safety margin
  val communicationMargin: Duration,
)

data class MaruConfig(
  val sotNode: ApiEndpointConfig,
  val dummyConsensusOptions: DummyConsensusOptions?,
  val p2pConfig: P2P?,
  val validator: Validator?,
  val followers: FollowersConfig,
)
