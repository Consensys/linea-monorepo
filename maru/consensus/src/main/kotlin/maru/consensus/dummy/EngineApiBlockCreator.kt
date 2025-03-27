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

import java.util.Optional
import maru.core.ExecutionPayload
import maru.executionlayer.manager.ExecutionLayerManager
import org.apache.logging.log4j.LogManager
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.BlobGas
import org.hyperledger.besu.datatypes.Hash
import org.hyperledger.besu.datatypes.Wei
import org.hyperledger.besu.ethereum.blockcreation.BlockCreationTiming
import org.hyperledger.besu.ethereum.blockcreation.BlockCreator
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.BlockBody
import org.hyperledger.besu.ethereum.core.BlockHeader
import org.hyperledger.besu.ethereum.core.BlockHeaderFunctions
import org.hyperledger.besu.ethereum.core.Difficulty
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.ethereum.mainnet.BodyValidation
import org.hyperledger.besu.ethereum.rlp.RLP
import org.hyperledger.besu.ethereum.rlp.RLPInput
import org.hyperledger.besu.evm.log.LogsBloomFilter

/**
 * Responsible for block creation with Engine API. createEmptyWithdrawalsBlock is the only available signature, the
 * rest of the methods delegate to it
 * Even though it implements BlockCreator, this interface doesn't really fit Engine API flow due to its asynchronicity
 */
class EngineApiBlockCreator(
  private val manager: ExecutionLayerManager,
  private val state: DummyConsensusState,
  private val blockHeaderFunctions: BlockHeaderFunctions,
  nextBlockTimestamp: Long, // Block creation starts right after, so it's required
  private val feeRecipientProvider: FeeRecipientProvider,
) : BlockCreator {
  init {
    manager
      .setHeadAndStartBlockBuilding(
        headHash = state.latestBlockHash,
        safeHash = state.finalizationState.safeBlockHash,
        finalizedHash = state.finalizationState.finalizedBlockHash,
        nextBlockTimestamp = nextBlockTimestamp,
        feeRecipient = feeRecipientProvider.getFeeRecipient(nextBlockTimestamp),
      ).get()
  }

  private val log = LogManager.getLogger(EngineApiBlockCreator::class.java)

  override fun createBlock(
    timestamp: Long,
    parentHeader: BlockHeader?,
  ): BlockCreator.BlockCreationResult = createEmptyWithdrawalsBlock(timestamp, parentHeader)

  override fun createBlock(
    transactions: MutableList<Transaction>,
    ommers: MutableList<BlockHeader>,
    timestamp: Long,
    parentHeader: BlockHeader?,
  ): BlockCreator.BlockCreationResult = createEmptyWithdrawalsBlock(timestamp, parentHeader)

  override fun createBlock(
    maybeTransactions: Optional<MutableList<Transaction>>,
    maybeOmmers: Optional<MutableList<BlockHeader>>,
    timestamp: Long,
    parentHeader: BlockHeader?,
  ): BlockCreator.BlockCreationResult = createEmptyWithdrawalsBlock(timestamp, parentHeader)

  // Starts creation of the next block with timestamp, returns current block being built
  override fun createEmptyWithdrawalsBlock(
    timestamp: Long,
    parentHeader: BlockHeader?,
  ): BlockCreator.BlockCreationResult {
    val blockCreationTimings = BlockCreationTiming()
    blockCreationTimings.register("Block creation")
    val blockBuildingResult =
      try {
        manager
          .finishBlockBuilding()
          .thenApply {
            state.updateLatestStatus(it.blockHash)
            log.info("Updating latest state")
            it
          }.get()
      } catch (e: Exception) {
        log.warn("Error during block building finish! Starting new attempt!", e)
        null
      }

    val newHeadHash = state.latestBlockHash
    // Mind the atomicity of finalization updates
    val finalizationState = state.finalizationState
    log.debug(" {}", newHeadHash)
    manager
      .setHeadAndStartBlockBuilding(
        headHash = newHeadHash,
        safeHash = finalizationState.safeBlockHash,
        finalizedHash = finalizationState.finalizedBlockHash,
        nextBlockTimestamp = timestamp,
        feeRecipient = feeRecipientProvider.getFeeRecipient(timestamp),
      ).whenException {
        log.error("Error while initiating block building!", it)
      }.get()
    val block =
      blockBuildingResult?.let {
        mapExecutionPayloadToBlock(blockBuildingResult)
      }
    blockCreationTimings.end("Block creation")
    // This return type doesn't fit this case well so stubbing it with dummy values for now
    return BlockCreator.BlockCreationResult(block, null, blockCreationTimings)
  }

  private fun mapExecutionPayloadToBlock(payload: ExecutionPayload): Block {
    val transactions =
      payload.transactions.map { tx ->
        val rlpInput: RLPInput = RLP.input(Bytes.wrap(tx))
        Transaction.readFrom(rlpInput)
      }
    val blockHeader =
      BlockHeader(
        /* parentHash = */ Hash.wrap(Bytes32.wrap(payload.parentHash)),
        /* ommersHash = */ Hash.EMPTY_LIST_HASH,
        /* coinbase = */ Address.wrap(Bytes.wrap(payload.feeRecipient)),
        /* stateRoot = */ Hash.wrap(Bytes32.wrap(payload.stateRoot)),
        /* transactionsRoot = */ BodyValidation.transactionsRoot(transactions),
        /* receiptsRoot = */ Hash.wrap(Bytes32.wrap(payload.receiptsRoot)),
        /* logsBloom = */ LogsBloomFilter(Bytes.wrap(payload.logsBloom)),
        /* difficulty = */ Difficulty.ZERO,
        /* number = */ payload.blockNumber.toLong(),
        /* gasLimit = */ payload.gasLimit.toLong(),
        /* gasUsed = */ payload.gasUsed.toLong(),
        /* timestamp = */ payload.timestamp.toLong(),
        /* extraData = */ Bytes.wrap(payload.extraData),
        /* baseFee = */ Wei.of(payload.baseFeePerGas.toLong()),
        /* mixHashOrPrevRandao = */ Bytes32.wrap(payload.prevRandao),
        /* nonce = */ 0,
        /* withdrawalsRoot = */ Hash.EMPTY_TRIE_HASH,
        /* blobGasUsed = */ 0,
        /* excessBlobGas = */ BlobGas.ZERO,
        /* parentBeaconBlockRoot = TODO: use an actual beacon block root */ Bytes32.ZERO,
        /* requestsHash = */ BodyValidation.requestsHash(listOf()),
        /* blockHeaderFunctions = */ blockHeaderFunctions,
      )
    val blockBody = BlockBody(transactions, listOf())
    return Block(blockHeader, blockBody)
  }
}
