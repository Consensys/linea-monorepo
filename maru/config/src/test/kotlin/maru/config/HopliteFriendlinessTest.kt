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
import linea.domain.BlockParameter
import linea.domain.RetryConfig
import linea.kotlin.decodeHex
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

@OptIn(ExperimentalHoplite::class)
class HopliteFriendlinessTest {
  private val protocolTransitionPollingInterval = 2.seconds
  private val emptyFollowersConfigToml =
    """
    protocol-transition-polling-interval = "2s"

    [persistence]
    data-path="/some/path"
    private-key-path = "/private-key/path"

    [qbft]
    fee-recipient = "0xdead000000000000000000000000000000000000"

    [p2p]
    port = 3322
    ip-address = "10.11.12.13"
    static-peers = ["/dns4/bootnode.linea.build/tcp/3322/p2p/16Uiu2HAmFjVuJoKD6sobrxwyJyysM1rgCsfWKzFLwvdB2HKuHwTg"]
    reconnect-delay = "500 ms"

    [p2p.discovery]
    port = 3324
    bootnodes = ["enr:-Iu4QHk0YN5IRRnufqsWkbO6Tn0iGTx4H_hnyiIEdXDuhIe0KKrxmaECisyvO40mEmmqKLhz_tdIhx2yFBK8XFKhvxABgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQOgBvD-dv0cX5szOeEsiAMtwxnP1q5CA5toYDrgUyOhV4N0Y3CCJBKDdWRwgiQT"]
    refresh-interval = "30 seconds"

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
    el-sync-status-refresh-interval = "6 seconds"

    [syncing.download]
    block-range-request-timeout = "10 seconds"
    blocks-batch-size = 64
    blocks-parallelism = 10
    max-retries = 5
    """.trimIndent()
  private val rawConfigToml =
    """
    $emptyFollowersConfigToml

    [follower-engine-apis]
    follower1 = { endpoint = "http://localhost:1234", jwt-secret-path = "/secret/path" }
    follower2 = { endpoint = "http://localhost:4321", timeout = "25 seconds" }
    """.trimIndent()
  private val dataPath = Path("/some/path")
  private val privateKeyPath = Path("/private-key/path")
  private val persistence = Persistence(dataPath, privateKeyPath)
  private val p2pConfig =
    P2P(
      ipAddress = "10.11.12.13",
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
          refreshInterval = 30.seconds,
        ),
    )
  private val ethApiEndpoint =
    ApiEndpointConfig(
      endpoint = URI.create("http://localhost:8545").toURL(),
      requestRetries =
        RetryConfig.endlessRetry(
          backoffDelay = 1.seconds,
          failuresWarningThreshold = 3u,
        ),
    )
  private val engineApiEndpoint =
    ApiEndpointConfig(
      endpoint = URI.create("http://localhost:8555").toURL(),
      jwtSecretPath = "/secret/path",
      requestRetries =
        RetryConfig.endlessRetry(
          backoffDelay = 1.seconds,
          failuresWarningThreshold = 3u,
        ),
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
      endpoint = URI.create("http://localhost:1234").toURL(),
      jwtSecretPath = "/secret/path",
    )
  private val follower2 =
    ApiEndpointDto(
      endpoint = URI.create("http://localhost:4321").toURL(),
      timeout = 25.seconds,
    )
  private val followersConfig =
    FollowersConfig(
      mapOf(
        "follower1" to
          ApiEndpointConfig(
            endpoint = URI.create("http://localhost:1234").toURL(),
            jwtSecretPath = "/secret/path",
          ),
        "follower2" to
          ApiEndpointConfig(
            endpoint = URI.create("http://localhost:4321").toURL(),
            timeout = 25.seconds,
          ),
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
      elSyncStatusRefreshInterval = 6.seconds,
      SyncingConfig.Download(
        blockRangeRequestTimeout = 10.seconds,
        blocksBatchSize = 64u,
        blocksParallelism = 10u,
        maxRetries = 5u,
      ),
    )

  @Test
  fun appConfigFileIsParseable() {
    val config = parseConfig<MaruConfigDtoToml>(rawConfigToml)
    assertThat(config).isEqualTo(
      MaruConfigDtoToml(
        protocolTransitionPollingInterval = protocolTransitionPollingInterval,
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
        protocolTransitionPollingInterval = protocolTransitionPollingInterval,
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
  fun `throws when el validator is also specified as follower`() {
    val exception =
      assertThrows<IllegalArgumentException> {
        MaruConfig(
          protocolTransitionPollingInterval = protocolTransitionPollingInterval,
          allowEmptyBlocks = false,
          persistence = persistence,
          qbftOptions = qbftOptions.toDomain(),
          p2pConfig = p2pConfig,
          validatorElNode = payloadValidator.domainFriendly(),
          followers = FollowersConfig(mapOf("el-validator" to payloadValidator.engineApiEndpoint.domainFriendly())),
          observabilityOptions = ObservabilityOptions(port = 9090u),
          linea = null,
          apiConfig = ApiConfig(port = 8080u),
          syncing = syncingConfig,
        )
      }
    assertThat(exception.message).isEqualTo("Validator EL node cannot be defined as a follower")
  }

  @Test
  fun appConfigFileIsConvertableToDomain() {
    val config = parseConfig<MaruConfigDtoToml>(rawConfigToml)
    assertThat(config.domainFriendly()).isEqualTo(
      MaruConfig(
        protocolTransitionPollingInterval = protocolTransitionPollingInterval,
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
        protocolTransitionPollingInterval = protocolTransitionPollingInterval,
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
      $emptyFollowersConfigToml
      """.trimIndent()
    val config = parseConfig<MaruConfigDtoToml>(configToml)

    assertThat(config)
      .isEqualTo(
        MaruConfigDtoToml(
          protocolTransitionPollingInterval = protocolTransitionPollingInterval,
          allowEmptyBlocks = true,
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

    assertThat(config.domainFriendly())
      .isEqualTo(
        MaruConfig(
          protocolTransitionPollingInterval = protocolTransitionPollingInterval,
          allowEmptyBlocks = true,
          persistence = persistence,
          p2pConfig = p2pConfig,
          validatorElNode =
            ValidatorElNode(
              engineApiEndpoint = engineApiEndpoint,
              ethApiEndpoint = ethApiEndpoint,
            ),
          qbftOptions = qbftOptions.toDomain(),
          followers = emptyFollowersConfig,
          observabilityOptions = ObservabilityOptions(port = 9090u),
          apiConfig = ApiConfig(port = 8080u),
          syncing = syncingConfig,
        ),
      )
  }

  @Test
  fun `should parse config with linea settings`() {
    val contractAddress = "0xB218f8A4Bc926cF1cA7b3423c154a0D627Bdb7E5"
    val l1EthApi = "http://ethereum-mainnet"
    val l1PollingInterval = 6.seconds
    val l1HighestBlockTag = "latest"
    val configToml =
      """
      $emptyFollowersConfigToml

      [linea]
      contract-address = "$contractAddress"
      l1-eth-api = { endpoint = "$l1EthApi" }
      l1-polling-interval = "6 seconds"
      l1-highest-block-tag = "$l1HighestBlockTag"
      """.trimIndent()
    val l1EthApiEndpoint =
      ApiEndpointDto(
        endpoint = URI.create("http://ethereum-mainnet").toURL(),
      )
    val contractAddressBytes = contractAddress.decodeHex()
    val expectedTomlConfig =
      LineaConfigDtoToml(
        contractAddress = contractAddressBytes,
        l1EthApi = l1EthApiEndpoint,
        l1PollingInterval,
        l1HighestBlockTag = l1HighestBlockTag,
      )
    val expectedLineaConfig =
      LineaConfig(
        contractAddress = contractAddressBytes,
        l1EthApi = l1EthApiEndpoint.domainFriendly(),
        l1PollingInterval = l1PollingInterval,
        l1HighestBlockTag = BlockParameter.Tag.LATEST,
      )
    val config = parseConfig<MaruConfigDtoToml>(configToml)

    assertThat(config)
      .isEqualTo(
        MaruConfigDtoToml(
          linea = expectedTomlConfig,
          protocolTransitionPollingInterval = protocolTransitionPollingInterval,
          persistence = persistence,
          qbft = qbftOptions,
          p2p = p2pConfig,
          payloadValidator = payloadValidator,
          observability = ObservabilityOptions(port = 9090u),
          api = ApiConfig(port = 8080u),
          syncing = syncingConfig,
          followerEngineApis = null,
        ),
      )

    assertThat(config.domainFriendly())
      .isEqualTo(
        MaruConfig(
          linea = expectedLineaConfig,
          protocolTransitionPollingInterval = protocolTransitionPollingInterval,
          persistence = persistence,
          p2pConfig = p2pConfig,
          validatorElNode =
            ValidatorElNode(
              engineApiEndpoint = engineApiEndpoint,
              ethApiEndpoint = ethApiEndpoint,
            ),
          qbftOptions = qbftOptions.toDomain(),
          followers = emptyFollowersConfig,
          observabilityOptions = ObservabilityOptions(port = 9090u),
          apiConfig = ApiConfig(port = 8080u),
          syncing = syncingConfig,
        ),
      )
  }
}
