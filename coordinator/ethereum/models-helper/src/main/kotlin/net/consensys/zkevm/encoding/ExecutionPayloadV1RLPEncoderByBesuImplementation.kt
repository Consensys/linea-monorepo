package net.consensys.zkevm.encoding

import org.hyperledger.besu.datatypes.Address
import org.hyperledger.besu.datatypes.Hash
import org.hyperledger.besu.datatypes.Wei
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.BlockBody
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder
import org.hyperledger.besu.ethereum.core.Difficulty
import org.hyperledger.besu.ethereum.core.encoding.TransactionDecoder
import org.hyperledger.besu.ethereum.mainnet.BodyValidation
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.evm.log.LogsBloomFilter
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1

object FakeRLPEncoder : ExecutionPayloadV1Encoder {
  override fun encode(payload: ExecutionPayloadV1): ByteArray {
    return ByteArray(0)
  }
}

object ExecutionPayloadV1RLPEncoderByBesuImplementation : ExecutionPayloadV1Encoder {
  override fun encode(payload: ExecutionPayloadV1): ByteArray {
    val parsedTransactions = payload.transactions.map(TransactionDecoder::decodeOpaqueBytes)
    val parsedBody = BlockBody(parsedTransactions, emptyList())
    val blockHeader =
      BlockHeaderBuilder.create()
        .parentHash(Hash.wrap(payload.parentHash))
        .ommersHash(Hash.EMPTY_LIST_HASH)
        .coinbase(Address.wrap(payload.feeRecipient.wrappedBytes))
        .stateRoot(Hash.wrap(payload.stateRoot))
        .transactionsRoot(BodyValidation.transactionsRoot(parsedBody.transactions))
        .receiptsRoot(Hash.wrap(payload.receiptsRoot))
        .logsBloom(LogsBloomFilter(payload.logsBloom))
        .difficulty(Difficulty.ZERO)
        .number(payload.blockNumber.longValue())
        .gasLimit(payload.gasLimit.longValue())
        .gasUsed(payload.gasLimit.longValue())
        .timestamp(payload.timestamp.longValue())
        .extraData(payload.extraData)
        .baseFee(Wei.wrap(payload.baseFeePerGas.toBytes()))
        .mixHash(Hash.wrap(payload.prevRandao))
        .nonce(0) // this works because Linea is not using PoW
        .blockHeaderFunctions(MainnetBlockHeaderFunctions())
        .buildBlockHeader()
    return Block(blockHeader, parsedBody).toRlp().toArray()
  }
}
