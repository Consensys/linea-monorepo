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
import java.net.URI
import kotlin.io.path.Path
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import maru.extensions.fromHexToByteArray
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class HopliteFriendlinessTest {
  private val emptyFollowersConfig =
    """
    [persistence]
    data-path="/some/path"

    [sot-eth-endpoint]
    endpoint = "http://localhost:8545"

    [qbft-options]
    communication-margin=100m

    [p2p-config]
    port = 3322

    [validator]
    private-key = "0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"
    jwt-secret-path = "/secret/path"
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
          persistence = Persistence(Path("/some/path")),
          sotEthEndpoint =
            ApiEndpointDtoToml(
              endpoint = URI.create("http://localhost:8545").toURL(),
            ),
          qbftOptions = QbftOptions(100.milliseconds),
          p2pConfig = P2P(port = 3322u),
          validator =
            ValidatorDtoToml(
              elClientEngineApiEndpoint = URI.create("http://localhost:8555").toURL(),
              privateKey = Secret("0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"),
              jwtSecretPath = "/secret/path",
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
          persistence = Persistence(Path("/some/path")),
          sotEthEndpoint =
            ApiEndpointDtoToml(
              endpoint = URI.create("http://localhost:8545").toURL(),
            ),
          qbftOptions = QbftOptions(100.milliseconds),
          p2pConfig = P2P(port = 3322u),
          validator =
            ValidatorDtoToml(
              elClientEngineApiEndpoint = URI.create("http://localhost:8555").toURL(),
              privateKey = Secret("0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"),
              jwtSecretPath = "/secret/path",
            ),
          followerEngineApis = null,
        ),
      )
  }

  @Test
  fun appConfigFileIsConvertableToDomain() {
    val config = Utils.parseTomlConfig<MaruConfigDtoToml>(rawConfig)
    assertThat(config.domainFriendly())
      .isEqualTo(
        MaruConfig(
          persistence = Persistence(Path("/some/path")),
          sotNode =
            ApiEndpointConfig(
              endpoint = URI.create("http://localhost:8545").toURL(),
            ),
          p2pConfig = P2P(port = 3322u),
          validator =
            Validator(
              engineApiClient = ApiEndpointConfig(URI.create("http://localhost:8555").toURL()),
              privateKey = "0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae".fromHexToByteArray(),
            ),
          qbftOptions = QbftOptions(100.milliseconds),
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
          persistence = Persistence(Path("/some/path")),
          sotNode =
            ApiEndpointConfig(
              endpoint = URI.create("http://localhost:8545").toURL(),
            ),
          qbftOptions = QbftOptions(100.milliseconds),
          p2pConfig = P2P(port = 3322u),
          validator =
            Validator(
              engineApiClient = ApiEndpointConfig(URI.create("http://localhost:8555").toURL()),
              privateKey = "0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae".fromHexToByteArray(),
            ),
          followers =
            FollowersConfig(
              emptyMap(),
            ),
        ),
      )
  }

  private val qbftOptions =
    """
    communication-margin=100m
    data-path="/some/path"
    message-queue-limit = 1000
    round-expiry = 1000
    duplicateMessageLimit = 100
    future-message-max-distance = 10
    future-messages-limit = 1000
    """.trimIndent()

  @Test
  fun qbftOptionsAreParseable() {
    val config =
      Utils.parseTomlConfig<QbftOptions>(qbftOptions)
    assertThat(config)
      .isEqualTo(
        QbftOptions(
          communicationMargin = 100.milliseconds,
          messageQueueLimit = 1000,
          roundExpiry = 1.seconds,
          duplicateMessageLimit = 100,
          futureMessageMaxDistance = 10L,
          futureMessagesLimit = 1000L,
        ),
      )
  }
}
