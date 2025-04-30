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
package maru.consensus.qbft.adapters

import java.math.BigInteger
import java.util.Optional
import maru.config.QbftOptions
import maru.consensus.ForkSpec
import maru.consensus.qbft.QbftConsensusConfig
import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.config.BftConfigOptions
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.consensus.common.ForkSpec as BesuForkSpec
import org.hyperledger.besu.consensus.common.ForksSchedule as BesuForksSchedule

class ForksScheduleAdapter(
  currentSpec: ForkSpec,
  config: QbftOptions,
) : BesuForksSchedule<BftConfigOptions>(maruForkSpecsToBesu(currentSpec, config)) {
  companion object {
    fun maruForkSpecsToBesu(
      currentSpec: ForkSpec,
      config: QbftOptions,
    ): MutableCollection<BesuForkSpec<BftConfigOptions>> =
      mutableListOf(BesuForkSpec(0, createBftConfig(currentSpec, config)))

    private fun createBftConfig(
      spec: ForkSpec,
      config: QbftOptions,
    ): BftConfigOptions =
      object : BftConfigOptions {
        override fun getEpochLength(): Long = 0

        override fun getBlockPeriodSeconds(): Int = spec.blockTimeSeconds

        override fun getEmptyBlockPeriodSeconds(): Int = 0

        override fun getBlockPeriodMilliseconds(): Long = 0

        override fun getRequestTimeoutSeconds(): Int = 0

        override fun getGossipedHistoryLimit(): Int = 0

        override fun getMessageQueueLimit(): Int = config.messageQueueLimit

        override fun getDuplicateMessageLimit(): Int = config.duplicateMessageLimit

        override fun getFutureMessagesLimit(): Int = config.futureMessagesLimit.toInt()

        override fun getFutureMessagesMaxDistance(): Int = config.futureMessageMaxDistance.toInt()

        override fun getMiningBeneficiary(): Optional<Address> =
          Optional.of(Address.wrap(Bytes.wrap((spec.configuration as QbftConsensusConfig).feeRecipient)))

        override fun getBlockRewardWei(): BigInteger = BigInteger.ZERO

        override fun asMap(): Map<String, Any> = emptyMap()
      }
  }
}
