package linea.rlp

import org.apache.tuweni.bytes.Bytes
import org.hyperledger.besu.ethereum.core.Block
import org.hyperledger.besu.ethereum.core.BlockBody
import org.hyperledger.besu.ethereum.core.BlockHeader
import org.hyperledger.besu.ethereum.core.BlockHeaderFunctions
import org.hyperledger.besu.ethereum.core.Transaction
import org.hyperledger.besu.ethereum.core.Withdrawal
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions
import org.hyperledger.besu.ethereum.rlp.RLPInput

object BesuRlpBlobDecoder : BesuBlockRlpDecoder {
  val transactionDecoder: NoSignatureTransactionDecoder = NoSignatureTransactionDecoder()
  val hashFunction: BlockHeaderFunctions = MainnetBlockHeaderFunctions()

  override fun decode(block: ByteArray): Block {
    return decode(org.hyperledger.besu.ethereum.rlp.RLP.input(Bytes.wrap(block)), hashFunction)
  }

  fun decode(rlpInput: RLPInput, hashFunction: BlockHeaderFunctions): Block {
    rlpInput.enterList()

    // Read the header
    val header: BlockHeader = BlockHeader.readFrom(rlpInput, hashFunction)

    // Use NoSignatureTransactionDecoder to decode transactions
    val transactions: List<Transaction> = rlpInput.readList(transactionDecoder::decode)

    // Read the ommers
    val ommers: List<BlockHeader> = rlpInput.readList { rlp: RLPInput ->
      BlockHeader.readFrom(rlp, hashFunction)
    }

    // Read the withdrawals
    if (!rlpInput.isEndOfCurrentList) {
      rlpInput.readList<Any>(Withdrawal::readFrom)
    }

    rlpInput.leaveList()
    return Block(header, BlockBody(transactions, ommers))
  }
}
