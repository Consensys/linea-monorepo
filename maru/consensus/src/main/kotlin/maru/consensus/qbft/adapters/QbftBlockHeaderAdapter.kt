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

import maru.consensus.qbft.toAddress
import maru.core.BeaconBlockHeader
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockHeader
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.Hash

/**
 * Adapter class to convert a BeaconBlockHeader to a QBFT block header
 */
class QbftBlockHeaderAdapter(
  val beaconBlockHeader: BeaconBlockHeader,
) : QbftBlockHeader {
  override fun getNumber(): Long = beaconBlockHeader.number.toLong()

  override fun getTimestamp(): Long = beaconBlockHeader.timestamp.toLong()

  override fun getCoinbase(): Address = beaconBlockHeader.proposer.toAddress()

  override fun getHash(): Hash = Hash.wrap(Bytes32.wrap(beaconBlockHeader.hash()))

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (other !is QbftBlockHeaderAdapter) return false

    if (beaconBlockHeader != other.beaconBlockHeader) return false

    return true
  }

  override fun hashCode(): Int = beaconBlockHeader.hashCode()

  override fun toString(): String = beaconBlockHeader.toString()
}
