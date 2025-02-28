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

import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.BlobGas
import org.hyperledger.besu.datatypes.Hash
import org.hyperledger.besu.datatypes.Wei
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.BlockBody
import org.hyperledger.besu.ethereum.core.BlockHeader
import org.hyperledger.besu.ethereum.core.Difficulty
import org.hyperledger.besu.ethereum.mainnet.BodyValidation
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.evm.log.LogsBloomFilter
import org.web3j.protocol.core.methods.response.EthBlock

object Mapper {
  private val blockHeaderFunctions = MainnetBlockHeaderFunctions()

  fun mapWeb3jBlockToBesuBlock(block: EthBlock.Block): Block {
    val blockHeader =
      BlockHeader(
        /* parentHash = */ Hash.wrap(Bytes32.fromHexString(block.parentHash)),
        /* ommersHash = */ Hash.EMPTY_LIST_HASH,
        /* coinbase = */ Address.fromHexString(block.miner),
        /* stateRoot = */ Hash.wrap(Bytes32.fromHexString(block.stateRoot)),
        /* transactionsRoot = */ Hash.wrap(Bytes32.fromHexString(block.transactionsRoot)),
        /* receiptsRoot = */ Hash.wrap(Bytes32.fromHexString(block.receiptsRoot)),
        /* logsBloom = */ LogsBloomFilter(Bytes.fromHexString(block.logsBloom)),
        /* difficulty = */ Difficulty.ZERO,
        /* number = */ block.number.toLong(),
        /* gasLimit = */ block.gasLimit.toLong(),
        /* gasUsed = */ block.gasUsed.toLong(),
        /* timestamp = */ block.timestamp.toLong(),
        /* extraData = */ Bytes.fromHexString(block.extraData),
        /* baseFee = */ Wei.of(block.baseFeePerGas.toLong()),
        /* mixHashOrPrevRandao = */ Bytes32.fromHexString(block.mixHash),
        /* nonce = */ 0,
        /* withdrawalsRoot = */ Hash.EMPTY_TRIE_HASH,
        /* blobGasUsed = */ 0,
        /* excessBlobGas = */ BlobGas.ZERO,
        /* parentBeaconBlockRoot = TODO: use an actual beacon block root */ Bytes32.ZERO,
        /* requestsHash = */ BodyValidation.requestsHash(listOf()),
        /* blockHeaderFunctions = */ blockHeaderFunctions,
      )

    // Transactions are omitted
    val blockBody = BlockBody(listOf(), listOf())
    return Block(blockHeader, blockBody)
  }
}
