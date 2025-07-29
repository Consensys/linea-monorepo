/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.config

import com.sksamuel.hoplite.ExperimentalHoplite
import java.net.URI
import kotlin.io.path.Path
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import linea.kotlin.decodeHex
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

@OptIn(ExperimentalHoplite::class)
class HopliteFriendlinessTest {
  private val emptyFollowersConfigToml =
    """
    [persistence]
    data-path="/some/path"
    private-key-path = "/private-key/path"

    [qbft]
    fee-recipient = "0xdead000000000000000000000000000000000000"

    [p2p]
    port = 3322
    ip-address = "127.0.0.1"
    static-peers = ["/dns4/bootnode.linea.build/tcp/3322/p2p/16Uiu2HAmFjVuJoKD6sobrxwyJyysM1rgCsfWKzFLwvdB2HKuHwTg"]
    reconnect-delay = "500 ms"

    [p2p.discovery]
    port = 3324
    bootnodes = ["enr:-Iu4QHk0YN5IRRnufqsWkbO6Tn0iGTx4H_hnyiIEdXDuhIe0KKrxmaECisyvO40mEmmqKLhz_tdIhx2yFBK8XFKhvxABgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQOgBvD-dv0cX5szOeEsiAMtwxnP1q5CA5toYDrgUyOhV4N0Y3CCJBKDdWRwgiQT"]

    [payload-validator]
    engine-api-endpoint = { endpoint = "http://localhost:8555", jwt-secret-path = "/secret/path" }
    eth-api-endpoint = { endpoint = "http://localhost:8545" }

    [observability]
    port = 9090

    [api]
    port = 8080

    [syncing]
    peer-chain-height-polling-interval = "5 seconds"
    peer-chain-height-granularity = 10
    """.trimIndent()
  private val rawConfigToml =
    """
    $emptyFollowersConfigToml

    [follower-engine-apis]
    follower1 = { endpoint = "http://localhost:1234", jwt-secret-path = "/secret/path" }
    follower2 = { endpoint = "http://localhost:4321" }
    """.trimIndent()
  private val dataPath = Path("/some/path")
  private val privateKeyPath = Path("/private-key/path")
  private val persistence = Persistence(dataPath, privateKeyPath)
  private val p2pConfig =
    P2P(
      ipAddress = "127.0.0.1",
      port = 3322u,
      staticPeers =
        listOf(
          "/dns4/bootnode.linea.build/tcp/3322/p2p/16Uiu2HAmFjVuJoKD6sobrxwyJyysM1rgCsfWKzFLwvdB2HKuHwTg",
        ),
      reconnectDelay = 500.milliseconds,
      discovery =
        P2P.Discovery(
          port = 3324u,
          bootnodes =
            listOf(
              "enr:-Iu4QHk0YN5IRRnufqsWkbO6Tn0iGTx4H_hnyiIEdXDuhIe0KKrxmaECisyvO40mEmmqKLhz_tdIhx2yFBK8XFKhvxABgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQOgBvD-dv0cX5szOeEsiAMtwxnP1q5CA5toYDrgUyOhV4N0Y3CCJBKDdWRwgiQT",
            ),
        ),
    )
  private val ethApiEndpoint =
    ApiEndpointConfig(
      endpoint = URI.create("http://localhost:8545").toURL(),
    )
  private val engineApiEndpoint =
    ApiEndpointConfig(
      endpoint = URI.create("http://localhost:8555").toURL(),
      jwtSecretPath = "/secret/path",
    )
  private val payloadValidator =
    PayloadValidatorDto(
      ethApiEndpoint =
        ApiEndpointDto(
          endpoint = URI.create("http://localhost:8545").toURL(),
        ),
      engineApiEndpoint =
        ApiEndpointDto(
          endpoint = URI.create("http://localhost:8555").toURL(),
          jwtSecretPath = "/secret/path",
        ),
    )
  private val follower1 =
    ApiEndpointDto(
      URI.create("http://localhost:1234").toURL(),
      jwtSecretPath = "/secret/path",
    )
  private val follower2 =
    ApiEndpointDto(
      URI.create("http://localhost:4321").toURL(),
    )
  private val followersConfig =
    FollowersConfig(
      mapOf(
        "follower1" to ApiEndpointConfig(URI.create("http://localhost:1234").toURL(), "/secret/path"),
        "follower2" to ApiEndpointConfig(URI.create("http://localhost:4321").toURL()),
      ),
    )
  private val emptyFollowersConfig = FollowersConfig(emptyMap())
  private val qbftOptions =
    QbftOptionsDtoToml(
      minBlockBuildTime = 500.milliseconds,
      messageQueueLimit = 1000,
      roundExpiry = 1.seconds,
      duplicateMessageLimit = 100,
      futureMessageMaxDistance = 10L,
      futureMessagesLimit = 1000L,
      feeRecipient = "0xdead000000000000000000000000000000000000".decodeHex(),
    )
  private val syncingConfig =
    SyncingConfig(
      peerChainHeightPollingInterval = 5.seconds,
      peerChainHeightGranularity = 10u,
    )

  @Test
  fun appConfigFileIsParseable() {
    val config = parseConfig<MaruConfigDtoToml>(rawConfigToml)
    assertThat(config).isEqualTo(
      MaruConfigDtoToml(
        allowEmptyBlocks = false,
        persistence = persistence,
        qbft = qbftOptions,
        p2p = p2pConfig,
        payloadValidator = payloadValidator,
        followerEngineApis = mapOf("follower1" to follower1, "follower2" to follower2),
        observability = ObservabilityOptions(port = 9090u),
        api = ApiConfig(port = 8080u),
        syncing = syncingConfig,
      ),
    )
  }

  @Test
  fun supportsEmptyFollowers() {
    val config = parseConfig<MaruConfigDtoToml>(emptyFollowersConfigToml)
    assertThat(config).isEqualTo(
      MaruConfigDtoToml(
        allowEmptyBlocks = false,
        persistence = persistence,
        qbft = qbftOptions,
        p2p = p2pConfig,
        payloadValidator = payloadValidator,
        followerEngineApis = null,
        observability = ObservabilityOptions(port = 9090u),
        api = ApiConfig(port = 8080u),
        syncing = syncingConfig,
      ),
    )
  }

  @Test
  fun appConfigFileIsConvertableToDomain() {
    val config = parseConfig<MaruConfigDtoToml>(rawConfigToml)
    assertThat(config.domainFriendly()).isEqualTo(
      MaruConfig(
        allowEmptyBlocks = false,
        persistence = persistence,
        p2pConfig = p2pConfig,
        validatorElNode =
          ValidatorElNode(
            engineApiEndpoint = engineApiEndpoint,
            ethApiEndpoint = ethApiEndpoint,
          ),
        qbftOptions = qbftOptions.toDomain(),
        followers = followersConfig,
        observabilityOptions = ObservabilityOptions(port = 9090u),
        apiConfig = ApiConfig(port = 8080u),
        syncing = syncingConfig,
      ),
    )
  }

  @Test
  fun emptyFollowersAreConvertableToDomain() {
    val config = parseConfig<MaruConfigDtoToml>(emptyFollowersConfigToml)
    assertThat(config.domainFriendly()).isEqualTo(
      MaruConfig(
        allowEmptyBlocks = false,
        persistence = persistence,
        qbftOptions = qbftOptions.toDomain(),
        p2pConfig = p2pConfig,
        validatorElNode =
          ValidatorElNode(
            engineApiEndpoint = engineApiEndpoint,
            ethApiEndpoint = ethApiEndpoint,
          ),
        followers = emptyFollowersConfig,
        observabilityOptions = ObservabilityOptions(port = 9090u),
        apiConfig = ApiConfig(port = 8080u),
        syncing = syncingConfig,
      ),
    )
  }

  private val qbftOptionsToml =
    """
    min-block-build-time=200m
    message-queue-limit = 1001
    round-expiry = 900m
    duplicateMessageLimit = 99
    future-message-max-distance = 11
    future-messages-limit = 100
    fee-recipient = "0x0000000000000000000000000000000000000001"
    """.trimIndent()

  @Test
  fun validatorDutiesAreParseable() {
    val config = parseConfig<QbftOptions>(qbftOptionsToml)
    assertThat(config).isEqualTo(
      QbftOptions(
        minBlockBuildTime = 200.milliseconds,
        messageQueueLimit = 1001,
        roundExpiry = 900.milliseconds,
        duplicateMessageLimit = 99,
        futureMessageMaxDistance = 11,
        futureMessagesLimit = 100,
        feeRecipient = "0x0000000000000000000000000000000000000001".decodeHex(),
      ),
    )
  }

  @Test
  fun `should parse allowEmptyBlocks = true`() {
    val configToml =
      """
      allow-empty-blocks = true
      $rawConfigToml
      """.trimIndent()
    val config = parseConfig<MaruConfigDtoToml>(configToml)

    assertThat(config)
      .isEqualTo(
        MaruConfigDtoToml(
          allowEmptyBlocks = true,
          persistence = persistence,
          qbft = qbftOptions,
          p2p = p2pConfig,
          payloadValidator = payloadValidator,
          followerEngineApis = mapOf("follower1" to follower1, "follower2" to follower2),
          observability = ObservabilityOptions(port = 9090u),
          api = ApiConfig(port = 8080u),
          syncing = syncingConfig,
        ),
      )

    assertThat(config.domainFriendly())
      .isEqualTo(
        MaruConfig(
          allowEmptyBlocks = true,
          persistence = persistence,
          p2pConfig = p2pConfig,
          validatorElNode =
            ValidatorElNode(
              engineApiEndpoint = engineApiEndpoint,
              ethApiEndpoint = ethApiEndpoint,
            ),
          qbftOptions = qbftOptions.toDomain(),
          followers = followersConfig,
          observabilityOptions = ObservabilityOptions(port = 9090u),
          apiConfig = ApiConfig(port = 8080u),
          syncing = syncingConfig,
        ),
      )
  }
}
