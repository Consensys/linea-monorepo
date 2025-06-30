/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.config

import com.sksamuel.hoplite.ConfigFailure
import com.sksamuel.hoplite.ConfigResult
import com.sksamuel.hoplite.DecoderContext
import com.sksamuel.hoplite.Node
import com.sksamuel.hoplite.decoder.DataClassDecoder
import com.sksamuel.hoplite.decoder.Decoder
import com.sksamuel.hoplite.fp.invalid
import com.sksamuel.hoplite.valueOrNull
import java.net.URL
import kotlin.reflect.KType
import kotlin.reflect.full.createType
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds
import maru.extensions.fromHexToByteArray

data class PayloadValidatorDto(
  val engineApiEndpoint: ApiEndpointDto,
  val ethApiEndpoint: ApiEndpointDto,
) {
  fun domainFriendly(): ValidatorElNode =
    ValidatorElNode(
      ethApiEndpoint = ethApiEndpoint.domainFriendly(),
      engineApiEndpoint = engineApiEndpoint.domainFriendly(),
    )
}

data class ApiEndpointDto(
  val endpoint: URL,
  val jwtSecretPath: String? = null,
) {
  fun domainFriendly(): ApiEndpointConfig = ApiEndpointConfig(endpoint = endpoint, jwtSecretPath = jwtSecretPath)
}

object QbftOptionsDecoder : Decoder<QbftOptions> {
  // This should be private, but Hoplite won't accept a private data class
  data class QbftOptionsDtoToml(
    val minBlockBuildTime: Duration = 500.milliseconds,
    val messageQueueLimit: Int = 1000,
    val roundExpiry: Duration = 1.seconds,
    val duplicateMessageLimit: Int = 100,
    val futureMessageMaxDistance: Long = 10L,
    val futureMessagesLimit: Long = 1000L,
  ) {
    fun toDomain(feeRecipient: ByteArray): QbftOptions =
      QbftOptions(
        minBlockBuildTime = minBlockBuildTime,
        messageQueueLimit = messageQueueLimit,
        roundExpiry = roundExpiry,
        duplicateMessageLimit = duplicateMessageLimit,
        futureMessageMaxDistance = futureMessageMaxDistance,
        futureMessagesLimit = futureMessagesLimit,
        feeRecipient = feeRecipient,
      )
  }

  override fun decode(
    node: Node,
    type: KType,
    context: DecoderContext,
  ): ConfigResult<QbftOptions> {
    val tomlFriendlyPart = DataClassDecoder().safeDecode(node, QbftOptionsDtoToml::class.createType(), context)
    val tomlFriendlyPartTyped = tomlFriendlyPart as ConfigResult<QbftOptionsDtoToml>
    val feeRecipient =
      try {
        node.getString("feerecipient").fromHexToByteArray()
      } catch (throwable: Throwable) {
        return ConfigFailure.ResolverException("Unable to convert feeRecipient to byteArray!", throwable).invalid()
      }

    return tomlFriendlyPartTyped.map { it.toDomain(feeRecipient) }
  }

  private fun Node.getString(key: String): String = this[key].valueOrNull()!!

  override fun supports(type: KType): Boolean = type.classifier == QbftOptions::class
}

data class MaruConfigDtoToml(
  private val persistence: Persistence,
  private val qbftOptions: QbftOptions?,
  private val p2pConfig: P2P?,
  private val payloadValidator: PayloadValidatorDto,
  private val followerEngineApis: Map<String, ApiEndpointDto>?,
  private val observabilityOptions: ObservabilityOptions,
  private val apiConfig: ApiConfig,
) {
  fun domainFriendly(): MaruConfig =
    MaruConfig(
      persistence = persistence,
      qbftOptions = qbftOptions,
      p2pConfig = p2pConfig,
      validatorElNode = payloadValidator.domainFriendly(),
      followers =
        FollowersConfig(
          followers = followerEngineApis?.mapValues { it.value.domainFriendly() } ?: emptyMap(),
        ),
      observabilityOptions = observabilityOptions,
      apiConfig = apiConfig,
    )
}
