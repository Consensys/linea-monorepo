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
package maru.consensus.dummy

import maru.consensus.ConsensusConfiguration

data class DummyConsensusConfig(
  private val blockTimeMillis: UInt,
  override val feeRecipient: ByteArray,
) : ConsensusConfiguration {
  /** Leaving blockTimeMillis in the configuration to be future-proof, exposing nextBlockPeriodSeconds to be closer
   * to reality where block timestamp is in seconds
   */
  val nextBlockPeriodSeconds = blockTimeMillis.toInt() / 1000

  init {
    require(feeRecipient.size == 20) {
      "feesRecipient address must be 20 bytes long, " +
        "but it's only ${feeRecipient.size} bytes long!"
    }
    require(blockTimeMillis >= 1000u) { "blockTimeMillis must be greater than 1 second" }
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as DummyConsensusConfig

    if (blockTimeMillis != other.blockTimeMillis) return false
    if (!feeRecipient.contentEquals(other.feeRecipient)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = blockTimeMillis.hashCode()
    result = 31 * result + feeRecipient.contentHashCode()
    return result
  }
}
