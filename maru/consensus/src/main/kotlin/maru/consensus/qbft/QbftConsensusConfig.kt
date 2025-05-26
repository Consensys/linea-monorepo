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
package maru.consensus.qbft

import maru.consensus.ConsensusConfig
import maru.consensus.ElFork
import maru.core.Validator

data class QbftConsensusConfig(
  val feeRecipient: ByteArray,
  val validatorSet: Set<Validator>,
  val elFork: ElFork,
) : ConsensusConfig {
  init {
    require(feeRecipient.size == 20) {
      "feesRecipient address must be 20 bytes long, " +
        "but it's only ${feeRecipient.size} bytes long!"
    }
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as QbftConsensusConfig

    if (!feeRecipient.contentEquals(other.feeRecipient)) return false
    if (elFork != other.elFork) return false

    return true
  }

  override fun hashCode(): Int {
    var result = feeRecipient.contentHashCode()
    result = 31 * result + elFork.hashCode()
    return result
  }
}
