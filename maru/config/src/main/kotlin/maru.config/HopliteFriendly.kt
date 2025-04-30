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

import com.sksamuel.hoplite.Masked
import java.net.URL
import maru.extensions.fromHexToByteArray

data class ValidatorDtoToml(
  val privateKey: Masked,
  val elClientEngineApiEndpoint: URL,
  val jwtSecretPath: String? = null,
) {
  fun domainFriendly(): Validator =
    Validator(
      privateKey = privateKey.value.fromHexToByteArray(),
      engineApiClient = ApiEndpointDtoToml(elClientEngineApiEndpoint, jwtSecretPath).toDomain(),
    )
}

data class ApiEndpointDtoToml(
  val endpoint: URL,
  val jwtSecretPath: String? = null,
) {
  fun toDomain(): ApiEndpointConfig = ApiEndpointConfig(endpoint = endpoint, jwtSecretPath = jwtSecretPath)
}

data class MaruConfigDtoToml(
  private val persistence: Persistence,
  private val qbftOptions: QbftOptions,
  private val sotEthEndpoint: ApiEndpointDtoToml,
  private val p2pConfig: P2P?,
  private val validator: ValidatorDtoToml?,
  private val followerEngineApis: Map<String, ApiEndpointDtoToml>?,
) {
  fun domainFriendly(): MaruConfig =
    MaruConfig(
      persistence = persistence,
      qbftOptions = qbftOptions,
      sotNode = sotEthEndpoint.toDomain(),
      p2pConfig = p2pConfig,
      validator = validator?.domainFriendly(),
      followers = FollowersConfig(followers = followerEngineApis?.mapValues { it.value.toDomain() } ?: emptyMap()),
    )
}
