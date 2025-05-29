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
package maru.consensus.config

import com.sksamuel.hoplite.ArrayNode
import com.sksamuel.hoplite.ConfigFailure
import com.sksamuel.hoplite.ConfigResult
import com.sksamuel.hoplite.DecoderContext
import com.sksamuel.hoplite.MapNode
import com.sksamuel.hoplite.Node
import com.sksamuel.hoplite.decoder.Decoder
import com.sksamuel.hoplite.fp.NonEmptyList
import com.sksamuel.hoplite.fp.invalid
import com.sksamuel.hoplite.fp.valid
import com.sksamuel.hoplite.valueOrNull
import kotlin.collections.component1
import kotlin.collections.component2
import kotlin.collections.map
import kotlin.collections.toSet
import kotlin.reflect.KType
import maru.consensus.ConsensusConfig
import maru.consensus.ElFork
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.delegated.ElDelegatedConsensus
import maru.consensus.qbft.QbftConsensusConfig
import maru.core.Validator
import maru.extensions.fromHexToByteArray

class ForkConfigDecoder : Decoder<JsonFriendlyForksSchedule> {
  override fun decode(
    node: Node,
    type: KType,
    context: DecoderContext,
  ): ConfigResult<JsonFriendlyForksSchedule> {
    val chainId = node.getString("chainid").toUInt()
    val config = node["config"]
    val forkSpecs =
      (config as MapNode).map.map { (k, v) ->
        val type = v.getString("type")
        val blockTimeSeconds = v.getString("blocktimeseconds").toInt()
        mapObjectToConfiguration(type, v).map {
          ForkSpec(
            k.toLong(),
            blockTimeSeconds,
            it,
          )
        }
      }
    return if (forkSpecs.all { it.isValid() }) {
      JsonFriendlyForksSchedule(chainId, forkSpecs.map { it.getUnsafe() }.toSet()).valid()
    } else {
      val failures = forkSpecs.filter { it.isInvalid() }.map { it.getInvalidUnsafe() }
      ConfigFailure.MultipleFailures(NonEmptyList(failures)).invalid()
    }
  }

  private fun Node.getString(key: String): String = this[key].valueOrNull()!!

  private fun mapObjectToConfiguration(
    type: String,
    obj: Node,
  ): ConfigResult<ConsensusConfig> =
    when (type) {
      "delegated" -> ElDelegatedConsensus.ElDelegatedConfig.valid()
      "qbft" ->
        QbftConsensusConfig(
          feeRecipient = obj.getString("feerecipient").fromHexToByteArray(),
          validatorSet =
            (obj["validatorset"] as ArrayNode)
              .elements
              .map {
                Validator(
                  it.valueOrNull()!!.fromHexToByteArray(),
                )
              }.toSet(),
          elFork = ElFork.valueOf(obj.getString("elfork")),
        ).valid()

      else -> (ConfigFailure.UnsupportedCollectionType(obj, "Unsupported fork type $type!") as ConfigFailure).invalid()
    }

  override fun supports(type: KType): Boolean = type.classifier == JsonFriendlyForksSchedule::class
}

data class JsonFriendlyForksSchedule(
  val chainId: UInt,
  val config: Set<ForkSpec>,
) {
  override fun equals(other: Any?): Boolean {
    if (other !is JsonFriendlyForksSchedule) {
      false
    }
    val otherTyped = other as JsonFriendlyForksSchedule
    return config.containsAll(otherTyped.config) && config.size == otherTyped.config.size
  }

  fun domainFriendly(): ForksSchedule = ForksSchedule(chainId, config)

  override fun hashCode(): Int = config.hashCode()
}
