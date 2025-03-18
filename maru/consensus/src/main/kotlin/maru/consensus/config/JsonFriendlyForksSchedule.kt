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

import fromHexToByteArray
import maru.consensus.ConsensusConfig
import maru.consensus.ElFork
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.consensus.delegated.ElDelegatedConsensus
import maru.consensus.dummy.DummyConsensusConfig

data class JsonFriendlyForksSchedule(
  val config: Map<String, Map<String, String>>,
) {
  fun domainFriendly(): ForksSchedule {
    val forkSpecs: List<ForkSpec> =
      config.map { (k, v) ->
        val type = v["type"].toString()
        val blockTimeSeconds = v["blockTimeSeconds"]!!.toInt()
        ForkSpec(
          k.toLong(),
          blockTimeSeconds,
          mapObjectToConfiguration(type, v),
        )
      }
    return ForksSchedule(forkSpecs)
  }

  private fun mapObjectToConfiguration(
    type: String,
    obj: Map<String, String>,
  ): ConsensusConfig =
    when (type) {
      "dummy" -> {
        DummyConsensusConfig(
          feeRecipient = obj["feeRecipient"]!!.fromHexToByteArray(),
          elFork = ElFork.valueOf(obj["elFork"]!!),
        )
      }

      "delegated" -> {
        ElDelegatedConsensus.ElDelegatedConfig
      }

      else -> throw IllegalArgumentException("Unsupported fork type $type!")
    }
}
