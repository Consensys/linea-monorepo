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

import com.sksamuel.hoplite.Secret
import fromHexToByteArray
import java.net.URI
import kotlin.time.Duration.Companion.milliseconds
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class HopliteFriendlinessTest {
  private val emptyFollowersConfig =
    """
    [sot-eth-endpoint]
    endpoint = "http://localhost:8545"

    [dummy-consensus-options]
    communication-time-margin=100m

    [p2p-config]
    port = 3322

    [validator]
    private-key = "0xdead"
    jwt-secret-path = "/secret/path"
    min-time-between-get-payload-attempts=800m
    el-client-engine-api-endpoint = "http://localhost:8555"
    """.trimIndent()
  private val rawConfig =
    """
    $emptyFollowersConfig

    [follower-engine-apis]
    follower1 = { endpoint = "http://localhost:1234", jwt-secret-path = "/secret/path" }
    follower2 = { endpoint = "http://localhost:4321" }
    """.trimIndent()

  @Test
  fun appConfigFileIsParseable() {
    val config =
      Utils.parseTomlConfig<MaruConfigDtoToml>(rawConfig)
    assertThat(config)
      .isEqualTo(
        MaruConfigDtoToml(
          sotEthEndpoint =
            ApiEndpointDtoToml(
              endpoint = URI.create("http://localhost:8545").toURL(),
            ),
          dummyConsensusOptions = DummyConsensusOptionsDtoToml(100.milliseconds),
          p2pConfig = P2P(port = 3322u),
          validator =
            ValidatorDtoToml(
              elClientEngineApiEndpoint = URI.create("http://localhost:8555").toURL(),
              privateKey = Secret("0xdead"),
              jwtSecretPath = "/secret/path",
              minTimeBetweenGetPayloadAttempts = 800.milliseconds,
            ),
          followerEngineApis =
            mapOf(
              "follower1" to
                ApiEndpointDtoToml(
                  URI.create("http://localhost:1234").toURL(),
                  jwtSecretPath =
                    "/secret/path",
                ),
              "follower2" to ApiEndpointDtoToml(URI.create("http://localhost:4321").toURL()),
            ),
        ),
      )
  }

  @Test
  fun supportsEmptyFollowers() {
    val config =
      Utils.parseTomlConfig<MaruConfigDtoToml>(emptyFollowersConfig)
    assertThat(config)
      .isEqualTo(
        MaruConfigDtoToml(
          sotEthEndpoint =
            ApiEndpointDtoToml(
              endpoint = URI.create("http://localhost:8545").toURL(),
            ),
          dummyConsensusOptions = DummyConsensusOptionsDtoToml(100.milliseconds),
          p2pConfig = P2P(port = 3322u),
          validator =
            ValidatorDtoToml(
              elClientEngineApiEndpoint = URI.create("http://localhost:8555").toURL(),
              privateKey = Secret("0xdead"),
              jwtSecretPath = "/secret/path",
              minTimeBetweenGetPayloadAttempts = 800.milliseconds,
            ),
          followerEngineApis = null,
        ),
      )
  }

  @Test
  fun appConfigFileIsConvertableToDomain() {
    val config =
      Utils.parseTomlConfig<MaruConfigDtoToml>(rawConfig)
    assertThat(config.domainFriendly())
      .isEqualTo(
        MaruConfig(
          sotNode =
            ApiEndpointConfig(
              endpoint = URI.create("http://localhost:8545").toURL(),
            ),
          dummyConsensusOptions = DummyConsensusOptions(100.milliseconds),
          p2pConfig = P2P(port = 3322u),
          validator =
            Validator(
              client =
                ValidatorClientConfig(
                  engineApiClientConfig = ApiEndpointConfig(URI.create("http://localhost:8555").toURL()),
                  minTimeBetweenGetPayloadAttempts = 800.milliseconds,
                ),
              key = "0xdead".fromHexToByteArray(),
            ),
          followers =
            FollowersConfig(
              mapOf(
                "follower1" to ApiEndpointConfig(URI.create("http://localhost:1234").toURL(), "/secret/path"),
                "follower2" to ApiEndpointConfig(URI.create("http://localhost:4321").toURL()),
              ),
            ),
        ),
      )
  }

  @Test
  fun emptyFollowersAreConvertableToDomain() {
    val config =
      Utils.parseTomlConfig<MaruConfigDtoToml>(emptyFollowersConfig)
    assertThat(config.domainFriendly())
      .isEqualTo(
        MaruConfig(
          sotNode =
            ApiEndpointConfig(
              endpoint = URI.create("http://localhost:8545").toURL(),
            ),
          dummyConsensusOptions = DummyConsensusOptions(100.milliseconds),
          p2pConfig = P2P(port = 3322u),
          validator =
            Validator(
              client =
                ValidatorClientConfig(
                  engineApiClientConfig =
                    ApiEndpointConfig(
                      URI.create("http://localhost:8555").toURL(),
                      "/secret/path",
                    ),
                  minTimeBetweenGetPayloadAttempts = 800.milliseconds,
                ),
              key = "0xdead".fromHexToByteArray(),
            ),
          followers =
            FollowersConfig(
              emptyMap(),
            ),
        ),
      )
  }
}
