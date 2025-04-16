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
package maru.consensus.delegated

import fromHexToByteArray
import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.ExecutionPayload
import maru.core.HashUtil
import maru.core.Validator
import maru.serialization.rlp.KeccakHasher
import maru.serialization.rlp.RLPSerializers
import org.web3j.protocol.core.methods.response.EthBlock

object Mapper {
  private val hasher = HashUtil.headerHash(RLPSerializers.BeaconBlockHeaderSerializer, KeccakHasher)

  fun mapWeb3jBlockToBeaconBlock(block: EthBlock.Block): BeaconBlock {
    val executionPayload =
      ExecutionPayload(
        parentHash = block.parentHash.fromHexToByteArray(),
        feeRecipient = block.miner.fromHexToByteArray(),
        stateRoot = block.stateRoot.fromHexToByteArray(),
        receiptsRoot = block.receiptsRoot.fromHexToByteArray(),
        logsBloom = block.logsBloom.fromHexToByteArray(),
        prevRandao = block.mixHash.fromHexToByteArray(),
        blockNumber = block.number.toLong().toULong(),
        gasLimit = block.gasLimit.toLong().toULong(),
        gasUsed = block.gasUsed.toLong().toULong(),
        timestamp = block.timestamp.toLong().toULong(),
        extraData = block.extraData.fromHexToByteArray(),
        baseFeePerGas = block.baseFeePerGas,
        blockHash = block.hash.fromHexToByteArray(),
        transactions = emptyList(), // Transactions are omitted
      )
    val beaconBlockBody = BeaconBlockBody(prevCommitSeals = emptyList(), executionPayload = executionPayload)

    val beaconBlockHeader =
      BeaconBlockHeader(
        number = 0u,
        round = 0u,
        timestamp = block.timestamp.toLong().toULong(),
        proposer = Validator(block.miner.fromHexToByteArray()),
        parentRoot = BeaconBlockHeader.EMPTY_HASH,
        stateRoot = BeaconBlockHeader.EMPTY_HASH,
        bodyRoot = BeaconBlockHeader.EMPTY_HASH,
        headerHashFunction = hasher,
      )
    return BeaconBlock(beaconBlockHeader, beaconBlockBody)
  }
}
